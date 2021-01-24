package maildb

/*
 * Copyright (C) 2020, Jim Lieb <lieb@sea-troll.net>
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public
 * License as published by the Free Software Foundation; either
 * version 3 of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public
 * License along with this library; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 *
 * -------------
 */

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3" // do I really need this here?
)

// RFC822 combined address and domain with their extra gunk
type RFC822 struct {
	id       int64
	lpart    string
	d_id     sql.NullInt64
	a_trans  sql.NullInt64
	a_rclass sql.NullInt64
	name     sql.NullString // should probably be NOT NULL
	class    sql.NullInt64
	d_trans  sql.NullInt64
	d_rclass sql.NullInt64
}

// String
func (a *RFC822) String() string {
	var line strings.Builder

	fmt.Fprintf(&line, "%s", a.lpart)
	if a.d_id.Valid {
		fmt.Fprintf(&line, "@%s", a.name.String)
	}
	return line.String()
}

// Target one each for each alias matched
type Target struct {
	id  int64 // id of alias row
	ext sql.NullString
	t   *RFC822
}

// String
func (tg *Target) String() string {
	var (
		ext  string
		line strings.Builder
	)

	if tg.ext.Valid {
		ext = tg.ext.String
	} else {
		ext = "<NULL>"
	}
	if tg.t == nil {
		fmt.Fprintf(&line, "%s", ext)
	} else {
		if tg.ext.Valid {
			fmt.Fprintf(&line, "%s+%s", tg.t.lpart, ext)
			if tg.t.d_id.Valid {
				fmt.Fprintf(&line, "@%s", tg.t.name.String)
			}
		} else {
			fmt.Fprintf(&line, "%s", tg.t.String())
		}
	}
	return line.String()
}

// Alias
type Alias struct {
	a      RFC822
	recips []*Target
}

// Id
func (al *Alias) Id() int64 {
	return al.a.id
}

// GetVirtual get a virtual mailbox alias
// empty lpart dumps all in domain, both empty dumps all
func (mdb *MailDB) GetVirtual(lpart string, domain string) ([]*Alias, error) {
	var (
		rows *sql.Rows
		err  error
	)

	q := `
SELECT a.id AS id,
       aa.id AS a_id,
       aa.localpart as a_local,
       aa.domain as ad_id,
       aa.transport as a_trans,
       aa.rclass as a_rclass,
       a.target as  t_id,
       a.extension AS ext
FROM alias AS a
JOIN address AS aa ON a.address = aa.id
`
	if len(lpart) == 0 && len(domain) == 0 { // slurp them all up
		q += `WHERE ad_id NOT NULL ORDER BY a_id`
		rows, err = mdb.db.Query(q)
	} else if len(lpart) == 0 && len(domain) >= 0 { // all in domain
		q += `
JOIN domain AS ad ON (aa.domain = ad.id)
WHERE ad.name = ? ORDER BY a_id
`
		rows, err = mdb.db.Query(q, domain)
	} else { // lpart@domain virtual alias
		q += `
JOIN domain AS ad ON (aa.domain = ad.id)
WHERE a_local = ? AND ad.name = ? ORDER BY a_id`
		rows, err = mdb.db.Query(q, lpart, domain)
	}
	if err != nil {
		return nil, fmt.Errorf("GetVirtual: query=%s, %s", q, err)
	}
	defer rows.Close()
	//fmt.Printf("q =%s\n", q)
	return mdb.fillAlias(rows)
}

// GetAlias alias from /etc/aliases
// empty alias implies all
func (mdb *MailDB) GetAlias(alias string) ([]*Alias, error) {
	var (
		rows *sql.Rows
		err  error
	)

	q := `
SELECT a.id AS id,
       aa.id AS a_id,
       aa.localpart as a_local,
       aa.domain as ad_id,
       aa.transport as a_trans,
       aa.rclass as a_rclass,
       a.target as  t_id,
       a.extension AS ext
FROM alias AS a
JOIN address AS aa ON a.address = aa.id
`
	if len(alias) > 0 { // one specific
		q += `WHERE a_local = ? AND ad_id IS NULL ORDER BY a_id`
		rows, err = mdb.db.Query(q, alias)
	} else { // all of them
		q += `WHERE ad_id IS NULL ORDER BY a_id`
		rows, err = mdb.db.Query(q)
	}
	if err != nil {
		return nil, fmt.Errorf("GetAlias: query=%s, %s", q, err)
	}
	defer rows.Close()
	return mdb.fillAlias(rows)
}

// fillAlias from the query
func (mdb *MailDB) fillAlias(rows *sql.Rows) ([]*Alias, error) {
	var (
		err       error
		curr_id   int64
		al        *Alias
		res       []*Alias
		id        int64
		a_id      int64
		a_local   string
		a_trans   sql.NullInt64
		a_rclass  sql.NullInt64
		ad_id     sql.NullInt64
		a_name    sql.NullString
		a_class   sql.NullInt64
		ad_trans  sql.NullInt64
		ad_rclass sql.NullInt64
		t_id      sql.NullInt64
		t_local   string
		t_trans   sql.NullInt64
		t_rclass  sql.NullInt64
		td_id     sql.NullInt64
		t_name    sql.NullString
		t_class   sql.NullInt64
		td_trans  sql.NullInt64
		td_rclass sql.NullInt64
		ext       sql.NullString
	)

	for rows.Next() {
		err = rows.Scan(&id, &a_id, &a_local, &ad_id, &a_trans, &a_rclass,
			&t_id, &ext)
		if err != nil {
			return nil, fmt.Errorf("fillAlias: Alias Scan, %s", err)
		}
		if ad_id.Valid && ad_id.Int64 != 0 {
			qd := `SELECT name, class, transport, rclass FROM domain WHERE id = ?`
			row := mdb.db.QueryRow(qd, ad_id.Int64)
			switch err := row.Scan(&a_name, &a_class, &a_trans, &a_rclass); err {
			case sql.ErrNoRows:
				return nil, fmt.Errorf("fillAlias: No alias domain!")
			case nil:
				break
			default:
				panic(err)
			}
		}
		if curr_id != a_id {
			al = &Alias{
				a: RFC822{
					id:       a_id,
					lpart:    a_local,
					d_id:     ad_id,
					a_trans:  a_trans,
					a_rclass: a_rclass,
					name:     a_name,
					class:    a_class,
					d_trans:  ad_trans,
					d_rclass: ad_rclass,
				},
			}
			res = append(res, al)
			curr_id = a_id
		}
		recp := &Target{
			id:  id,
			ext: ext,
		}
		if t_id.Valid { // is it local+ext or ext?
			qt := `SELECT localpart, domain, transport, rclass FROM address WHERE id = ?`
			row := mdb.db.QueryRow(qt, t_id.Int64)
			switch err := row.Scan(&t_local, &td_id, &t_trans, &t_rclass); err {
			case sql.ErrNoRows:
				return nil, fmt.Errorf("fillAlias: Target not found!")
			case nil:
				break
			default:
				panic(fmt.Errorf("GetAliases: %s", err))
			}
			if td_id.Valid && td_id.Int64 != 0 { // do we have a domain for this target
				qtd := `SELECT name, class, transport, rclass FROM domain WHERE id = ?`
				row := mdb.db.QueryRow(qtd, td_id.Int64)
				switch err := row.Scan(&t_name, &t_class, &td_trans, &td_rclass); err {
				case sql.ErrNoRows:
					return nil, fmt.Errorf("GetAliases: Target domain not found!")
				case nil:
					break
				default:
					panic(fmt.Errorf("GetAliases: %s", err))
				}
			}
			recp.t = &RFC822{
				id:       t_id.Int64,
				lpart:    t_local,
				a_trans:  t_trans,
				a_rclass: t_rclass,
				d_id:     td_id,
				name:     t_name,
				class:    t_class,
				d_trans:  td_trans,
				d_rclass: td_rclass,
			}
		}
		al.recips = append(al.recips, recp)
	}
	return res, nil
}

// String
func (al *Alias) String() string {
	var (
		line   strings.Builder
		commas int
	)

	fmt.Fprintf(&line, "%s:\t", al.a.String())
	for _, r := range al.recips {
		if commas > 0 {
			fmt.Fprintf(&line, ", ")
		}
		fmt.Fprintf(&line, "%s", r.String())
		commas++
	}
	return line.String()
}

// MakeAlias
// This is not 'NewAlias' because we can add recipients to an already made alias
func (mdb *MailDB) MakeAlias(alias string, recipient string) error {
	var (
		err        error
		aliasParts *AddressParts
		recipParts *AddressParts
		aliasAddr  *Address
		recipAddr  *Address
	)
	if aliasParts, err = DecodeRFC822(alias); err != nil {
		return fmt.Errorf("MakeAlias: alias, %s", err)
	}
	if recipParts, err = DecodeTarget(recipient); err != nil {
		return fmt.Errorf("MakeAlias: recipient, %s", err)
	}
	// Enter a transaction for everything else
	if mdb.tx, err = mdb.db.Begin(); err != nil {
		return fmt.Errorf("MakeAlias: begin, %s", err)
	}
	defer func() {
		if err == nil {
			if err = mdb.tx.Commit(); err != nil {
				panic(fmt.Errorf("MakeAlias: commit, %s", err)) // we are screwed
			}
		} else {
			mdb.tx.Rollback()
		}
	}()
	if aliasAddr, err = mdb.lookupAddress(aliasParts); err != nil {
		return fmt.Errorf("MakeAlias: alias, %s", err)
	}
	if aliasAddr == nil { // no such, make one
		if aliasAddr, err = mdb.insertAddress(aliasParts); err != nil {
			return fmt.Errorf("MakeAlias: alias, %s", err)
		}
	}
	// We now have the alias address part, either brand new or an existing
	// Now find the target.
	if recipAddr, err = mdb.lookupAddress(recipParts); err != nil {
		return fmt.Errorf("MakeAlias: target, %s", err)
	}
	if recipAddr == nil {
		if recipAddr, err = mdb.insertAddress(recipParts); err != nil {
			return fmt.Errorf("MakeAlias: target %s", err)
		}
	} else { // this addr already exists. Is it at the target end of an alias and is that a dup?
		var (
			id int64
			e  sql.NullString
		)
		qal := `SELECT id, extension FROM alias WHERE address = ? AND target = ?`
		rows, err := mdb.db.Query(qal, aliasAddr.id, recipAddr.id)
		for rows.Next() {
			if err = rows.Scan(&id, &e); err != nil {
				return fmt.Errorf("MakeAlias: alias scan, %s", err)
			}
			if (!e.Valid && recipParts.extension == "") ||
				(e.Valid && e.String == recipParts.extension) {
				return fmt.Errorf("MakeAlias: alias to this extension already exists")
			}
		}
	}
	// Now have both, no dups, make the link
	ext := sql.NullString{Valid: false}
	if recipParts.extension != "" {
		ext.Valid = true
		ext.String = recipParts.extension
	}
	_, err = mdb.tx.Exec("INSERT INTO alias (address, target, extension) VALUES (?, ?, ?)",
		aliasAddr.id, recipAddr.id, ext)
	if err != nil {
		return fmt.Errorf("MakeAlias: insert alias, %s", err)
	}
	return nil
}

// IsAlias
func (mdb *MailDB) IsAlias(alias string) (bool, error) {
	var (
		aliasAddr  *Address
		aliasParts *AddressParts
		row        *sql.Row
		err        error
		count      int
	)
	if aliasParts, err = DecodeRFC822(alias); err != nil {
		return false, fmt.Errorf("IsAlias: alias, %s", err)
	}
	if aliasAddr, err = mdb.lookupAddress(aliasParts); err != nil {
		return false, fmt.Errorf("IsAlias: alias, %s", err)
	}
	if aliasAddr == nil {
		return false, fmt.Errorf("IsAlias: no such address")
	}
	row = mdb.db.QueryRow("SELECT COUNT(*) FROM alias WHERE address = ?", aliasAddr.id)
	switch err = row.Scan(&count); err {
	case sql.ErrNoRows:
		return false, fmt.Errorf("IsAlias: scan, %s", err)
	case nil:
		if count > 0 {
			return true, nil
		} else {
			return false, nil
		}
	default:
		return false, fmt.Errorf("IsAlias: default, %s", err)
	}
}

// RemoveAlias and all its targets
func (mdb *MailDB) RemoveAlias(alias string) error {
	var (
		err        error
		aliasParts *AddressParts
		aliasAddr  *Address
		recipAddr  *Address
		aliasCnt   int
		aliasID    int64
		targetID   sql.NullInt64
		ext        sql.NullString
	)
	if aliasParts, err = DecodeRFC822(alias); err != nil {
		return fmt.Errorf("RemoveAlias: alias, %s", err)
	}

	// Enter a transaction for everything else
	if mdb.tx, err = mdb.db.Begin(); err != nil {
		return fmt.Errorf("RemoveAlias: begin, %s", err)
	}
	defer func() {
		if err == nil {
			if err = mdb.tx.Commit(); err != nil {
				panic(fmt.Errorf("RemoveAlias: commit, %s", err)) // we are screwed
			}
		} else {
			mdb.tx.Rollback()
		}
	}()
	if aliasAddr, err = mdb.lookupAddress(aliasParts); err != nil {
		return fmt.Errorf("RemoveAlias: alias, %s", err)
	}
	if aliasAddr == nil {
		return fmt.Errorf("RemoveAlias: no such alias")
	}
	qa := `SELECT id, target, extension FROM alias WHERE address = ?`
	rows, err := mdb.db.Query(qa, aliasAddr.id)
	for rows.Next() {
		if err = rows.Scan(&aliasID, &targetID, &ext); err != nil {
			return fmt.Errorf("RemoveAlias: alias scan, %s", err)
		}
		aliasCnt++
		_, err = mdb.tx.Exec("DELETE FROM alias WHERE id = ?", aliasID)
		if err != nil {
			return fmt.Errorf("RemoveAlias: delete alias, %s", err)
		}
		if targetID.Valid {
			err = mdb.deleteAddressByID(recipAddr)
			if err != nil {
				return fmt.Errorf("RemoveAlias: delete recipient, %s", err)
			}
		}
	}
	if aliasCnt > 0 { // Found aliases so delete address
		err = mdb.deleteAddressByID(aliasAddr)
		if err != nil {
			return fmt.Errorf("RemoveAlias: delete alias address, %s", err)
		}
	} else {
		return fmt.Errorf("RemoveAlias: address is not an alias")
	}
	return nil
}

// RemoveRecipient. Remove the alias as well if this is the last target
func (mdb *MailDB) RemoveRecipient(alias string, recipient string) error {
	var (
		err        error
		aliasParts *AddressParts
		recipParts *AddressParts
		aliasAddr  *Address
		recipAddr  *Address
		foundit    bool = false
		aliasCnt   int
		aliasID    int64
		targetID   sql.NullInt64
		ext        sql.NullString
	)
	if aliasParts, err = DecodeRFC822(alias); err != nil {
		return fmt.Errorf("RemoveRecipient: alias, %s", err)
	}
	if recipParts, err = DecodeTarget(recipient); err != nil {
		return fmt.Errorf("RemoveRecipient: recipient, %s", err)
	}

	// Enter a transaction for everything else
	if mdb.tx, err = mdb.db.Begin(); err != nil {
		return fmt.Errorf("RemoveRecipient: begin, %s", err)
	}
	defer func() {
		if err == nil {
			if err = mdb.tx.Commit(); err != nil {
				panic(fmt.Errorf("RemoveRecipient: commit, %s", err)) // we are screwed
			}
		} else {
			mdb.tx.Rollback()
		}
	}()
	if aliasAddr, err = mdb.lookupAddress(aliasParts); err != nil {
		return fmt.Errorf("RemoveRecipient: alias, %s", err)
	}
	if aliasAddr == nil {
		return fmt.Errorf("RemoveRecipient: no such alias")
	}
	if recipParts.domain != "" { // not a file, filter, or include. no address to see
		if recipAddr, err = mdb.lookupAddress(recipParts); err != nil {
			return fmt.Errorf("RemoveRecipient: recipient, %s", err)
		}
		if recipAddr == nil {
			return fmt.Errorf("RemoveRecipient: no such recipient")
		}
	}
	qa := `SELECT id, target, extension FROM alias WHERE address = ?`
	rows, err := mdb.db.Query(qa, aliasAddr.id)
	for rows.Next() {
		if err = rows.Scan(&aliasID, &targetID, &ext); err != nil {
			return fmt.Errorf("RemoveRecipient: alias scan, %s", err)
		}
		aliasCnt++
		if !foundit {
			if recipParts.domain != "" { // a target with possible '+' ext
				if targetID.Valid && targetID.Int64 == recipAddr.id {
					foundit = true
				}
			} else { // looking for a file/pipe/include only
				if !targetID.Valid && ext.Valid && ext.String == recipParts.extension {
					foundit = true
				}
			}
		}
	}
	if foundit {
		_, err = mdb.tx.Exec("DELETE FROM alias WHERE id = ?", aliasID)
		if err != nil {
			return fmt.Errorf("RemoveRecipient: delete alias, %s", err)
		}
		if targetID.Valid {
			err = mdb.deleteAddressByID(recipAddr)
			if err != nil {
				return fmt.Errorf("RemoveRecipient: delete recipient, %s", err)
			}
		}
		if aliasCnt <= 1 { // last one, remove the alias address too
			err = mdb.deleteAddressByID(aliasAddr)
			if err != nil {
				return fmt.Errorf("RemoveRecipient: delete alias address, %s", err)
			}
		}
	} else {
		return fmt.Errorf("RemoveRecipient: recipient not found in alias")
	}
	return nil
}

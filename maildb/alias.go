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

// Recipient in alias list
type Recipient struct {
	id  int64 // id of alias row
	ext sql.NullString
	t   *Address
}

// String
// beware! virtual aliases cannot have /etc/aliases attributes (pipes and stuff)
func (tg *Recipient) String() string {
	var (
		line strings.Builder
	)
	if tg.t != nil {
		if tg.ext.Valid {
			fmt.Fprintf(&line, "%s+%s", tg.t.localpart, tg.ext.String)
			if !tg.t.IsLocal() {
				fmt.Fprintf(&line, "@%s", tg.t.d.String())
			}
		} else {
			fmt.Fprintf(&line, "%s", tg.t.String())
		}
	} else if tg.ext.Valid {
		fmt.Fprintf(&line, "%s", tg.ext.String)
	} // panic else here? CHECK constraint on alias should apply on insert/update
	return line.String()
}

// Alias
type Alias struct {
	addr   *Address
	recips []*Recipient
}

// Id
func (al *Alias) Id() int64 {
	return al.addr.id
}

// LookupAlias
// get either "local_user" or "mbox@domain" aliases
// name@domain returns that alias recipients for this address
// *           returns all local (/etc/aliases) aliases
// *@domain    returns all aliases in this domain
// name@*      returns all aliases of this name, e.g. abuse@foo.com, abuse@example.org
// *@*         returns all virtual aliases in the database
func (mdb *MailDB) LookupAlias(alias string) ([]*Alias, error) {
	var (
		al_list []*Alias
		a_list  []*Address
		err     error
		rowCnt  int
	)

	if a_list, err = mdb.FindAddress(alias); err != nil {
		return nil, err
	}
	for _, a := range a_list {
		al, err := mdb.lookupAliasByAddr(a)
		if err != nil {
			if err == ErrMdbNotAlias {
				continue
			} else {
				break
			}
		}
		al_list = append(al_list, al)
		rowCnt++
	}
	if err == nil && rowCnt == 0 {
		err = ErrMdbNoAliases
	}
	return al_list, err
}

// lookupAliasByAddr
func (mdb *MailDB) lookupAliasByAddr(a *Address) (*Alias, error) {
	var (
		aID    int64
		ta     *Address
		target sql.NullInt64
		ext    sql.NullString
		rows   *sql.Rows
		rowCnt int64
		err    error
	)

	al := &Alias{
		addr: a,
	}
	qal := `SELECT id, target, extension FROM alias WHERE address IS ? ORDER BY id`
	rows, err = mdb.db.Query(qal, a.id)
	for rows.Next() {
		if err = rows.Scan(&aID, &target, &ext); err != nil {
			return nil, err
		}
		if target.Valid {
			if ta, err = mdb.lookupAddressByID(target.Int64); err != nil {
				return nil, err
			}
		} else {
			ta = nil
			if !a.IsLocal() {
				return nil, ErrMdbAddressTarget
			}
		}
		r := &Recipient{
			id:  aID,
			t:   ta,
			ext: ext,
		}
		rowCnt++
		al.recips = append(al.recips, r)

	}
	if err = rows.Close(); err != nil {
		return nil, err
	}
	if rowCnt == 0 {
		return nil, ErrMdbNotAlias
	}
	return al, nil
}

// String
// return a line for this alias
// Note that /etc/aliases is a different syntax from virtual(5)
func (al *Alias) String() string {
	var (
		line   strings.Builder
		commas int
	)

	if al.addr.IsLocal() {
		fmt.Fprintf(&line, "%s: ", al.addr.String())
	} else {
		fmt.Fprintf(&line, "%s ", al.addr.String())
	}
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
// do all the address decoding first. That way the transaction is working with
// already parsed arguments saving complications in rollback on errors. We will
// only have to rollback on db errors.
func (mdb *MailDB) MakeAlias(alias string, recipients []string) error {
	var (
		err       error
		ap        *AddressParts
		aliasAddr *Address
	)

	if len(recipients) < 1 {
		return ErrMdbNoRecipients
	}
	ap, err = DecodeRFC822(alias)
	if err != nil {
		return err
	}

	// Enter a transaction for everything else
	mdb.Begin()
	defer mdb.End(&err)

	if aliasAddr, err = mdb.GetOrInsAddress(alias); err != nil {
		return err
	}

	// We now have the alias address part, either brand new or an existing
	// Now cycle through the recipient list and stuff them in
	for _, r := range recipients {
		var (
			rp      *AddressParts
			rAddr   *Address
			recipID sql.NullInt64
			ext     sql.NullString
		)

		if rp, err = DecodeTarget(r); err != nil {
			break
		}
		if !ap.IsLocal() && rp.IsPipe() { // a virtual alias cannot have a pipe target
			err = ErrMdbAddressTarget
			break
		}
		if rp.extension != "" {
			ext.Valid = true
			ext.String = rp.extension
		} else {
			ext.Valid = false
		}
		if !rp.IsPipe() { // we have a foo@baz address
			if rAddr, err = mdb.GetOrInsAddress(r); err != nil {
				break
			}
			recipID = sql.NullInt64{Valid: true, Int64: rAddr.id}
		} else {
			recipID.Valid = false
		}
		// Now make the link
		if _, err = mdb.tx.Exec("INSERT INTO alias (address, target, extension) VALUES (?, ?, ?)",
			aliasAddr.id, recipID, ext); err != nil {
			break
		}
	}
	return err
}

// RemoveAlias and all its targets
// All we need to do here is delete the aliases that aliasAddr points to
// As the set of aliases disappear, their delete triggers clean up all the
// orphan targets (and the alias address itself) on the way out
func (mdb *MailDB) RemoveAlias(alias string) error {
	var (
		ap  *AddressParts
		err error
		c   int64
		res sql.Result
	)

	if ap, err = DecodeRFC822(alias); err != nil {
		return err
	}
	if ap.IsLocal() {
		qd := `
DELETE FROM alias WHERE address =
(SELECT a.id FROM address a  WHERE a.domain IS NULL AND a.localpart = ?)
`
		res, err = mdb.db.Exec(qd, ap.lpart)
	} else {
		qd := `
DELETE FROM alias WHERE address =
(SELECT a.id FROM address a, domain d
  WHERE a.domain = d.id AND a.localpart = ? AND d.name = ?)
`
		res, err = mdb.db.Exec(qd, ap.lpart, ap.domain)
	}
	if err == nil {
		c, err = res.RowsAffected()
		if err == nil {
			if c < 1 {
				err = ErrMdbNotAlias
			}
		}
	}
	return err
}

// RemoveRecipient. Remove the alias as well if this is the last target
func (mdb *MailDB) RemoveRecipient(alias string, recipient string) error {
	var (
		ap  *AddressParts
		err error
		c   int64
		rp  *AddressParts
		res sql.Result
	)

	if ap, err = DecodeRFC822(alias); err != nil {
		return err
	}
	if rp, err = DecodeTarget(recipient); err != nil {
		return err
	}
	if ap.IsLocal() {
		if rp.IsPipe() {
			qd := `
DELETE FROM alias WHERE target IS NULL AND extension IS ? AND address =
  (SELECT id FROM address WHERE localpart = ? AND domain IS NULL)
`
			res, err = mdb.db.Exec(qd, rp.extension, ap.lpart)
		} else {
			qd := `
DELETE FROM alias WHERE address =
  (SELECT id FROM address WHERE localpart = ? AND domain IS NULL)
`
			if rp.IsLocal() {
				qd += `
 AND target = (SELECT id from address WHERE localpart = ? AND domain IS NULL)
`
				res, err = mdb.db.Exec(qd, ap.lpart, rp.lpart)
			} else {
				qd += `
 AND target = (SELECT id from address a, domain d
   WHERE a.localpart = ? AND a.domain = d.id AND d.name = ?)
`
				res, err = mdb.db.Exec(qd, ap.lpart, rp.lpart, rp.domain)
			}
		}
	} else { // name@domain
		qd := `
DELETE FROM alias WHERE address =
  (SELECT a.id FROM address a, domain d WHERE a.localpart = ? AND a.domain = d.id AND d.name = ? )
`
		if !rp.IsPipe() {
			if rp.IsLocal() {
				qd += `
 AND target = (SELECT id from address WHERE localpart = ? AND domain IS NULL)
`
				res, err = mdb.db.Exec(qd, ap.lpart, ap.domain, rp.lpart)
			} else {
				qd += `
 AND target = (SELECT a.id from address a, domain d
   WHERE a.localpart = ? AND a.domain = d.id AND d.name = ?)
`
				res, err = mdb.db.Exec(qd, ap.lpart, ap.domain, rp.lpart, rp.domain)
			}
		} else {
			err = ErrMdbNoLocalPipe // for name@domain virtuals
		}
	}
	if err == nil {
		c, err = res.RowsAffected()
		if err == nil {
			if c < 1 {
				err = ErrMdbRecipientNotFound
			}
		}
	}
	return err
}

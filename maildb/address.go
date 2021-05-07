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
)

// DB query/insert helpers

type Address struct {
	mdb       *MailDB
	id        int64
	d         *Domain
	localpart string
	transport sql.NullInt64
	rclass    sql.NullString
	access    sql.NullInt64
}

// IsLocal
// a "local" address has no domain (address.domain IS NULL)
func (a *Address) IsLocal() bool {
	if a.d == nil {
		return true
	} else {
		return false
	}
}

// InVmailDomain
func (a *Address) InVMailDomain() bool {
	return a.d.IsVmailbox()
}

// Id
func (a *Address) Id() int64 {
	return a.id
}

// String
func (a *Address) String() string {
	var (
		line strings.Builder
	)
	fmt.Fprintf(&line, "%s", a.localpart)
	if a.d != nil {
		fmt.Fprintf(&line, "@%s", a.d.String())
	}
	return line.String()
}

// dump
func (a *Address) dump() string {
	var (
		line strings.Builder
	)
	fmt.Fprintf(&line, "id:%d, localpart: %s, ", a.id, a.localpart)
	if a.d != nil {
		fmt.Fprintf(&line, "domain id: %d, dname: %s, ", a.d.Id(), a.d.String())
	} else {
		fmt.Fprintf(&line, "domain id: <NULL>, dname: <empty>, ")
	}
	if a.transport.Valid {
		fmt.Fprintf(&line, "transport: %d, ", a.transport.Int64)
	} else {
		fmt.Fprintf(&line, "transport: <NULL>, ")
	}
	if a.rclass.Valid {
		fmt.Fprintf(&line, "rclass: %s, ", a.rclass.String)
	} else {
		fmt.Fprintf(&line, "rclass: <NULL>, ")
	}
	if a.access.Valid {
		fmt.Fprintf(&line, "access: %d, ", a.access.Int64)
	} else {
		fmt.Fprintf(&line, "access: <NULL>")
	}
	return line.String()
}

// It would be nice to have everything in one query but the best we can do is
// join address and domain. Counting alias and vmailbox references is seriously
// messy and expensive.
//
// query for local (no domain) addresses
var qaLocal string = `
SELECT id, localpart, transport, rclass, access FROM address
 WHERE localpart = ? AND domain IS NULL
`

// query for full localpart@domain addresses
var qaRFC822 string = `
SELECT a.id, a.localpart, a.transport, a.rclass, a.access,
       d.id, d.name, d.class, d.transport, d.access, d.vuid, d.vgid, d.rclass
 FROM address AS a, domain AS d
 WHERE a.localpart = ? AND a.domain IS d.id AND d.name = ?
`

// LookupAddress
// Lookup an address without an active transaction
func (mdb *MailDB) LookupAddress(addr string) (*Address, error) {
	var (
		ap  *AddressParts
		row *sql.Row
		err error
	)

	if ap, err = DecodeRFC822(addr); err != nil {
		return nil, err
	}
	a := &Address{
		mdb: mdb,
	}
	d := &Domain{
		mdb: mdb,
	}
	if ap.domain == "" { // A "local" address
		row = mdb.db.QueryRow(qaLocal, ap.lpart)
		err = row.Scan(
			&a.id, &a.localpart, &a.transport, &a.rclass, &a.access)
	} else { // A full RFC822 address
		row = mdb.db.QueryRow(qaRFC822, ap.lpart, ap.domain)
		err = row.Scan(
			&a.id, &a.localpart, &a.transport, &a.rclass, &a.access,
			&d.id, &d.name, &d.class, &d.transport, &d.access, &d.vuid, &d.vgid, &d.rclass)
	}
	switch err {
	case sql.ErrNoRows:
		return nil, ErrMdbAddressNotFound
	case nil:
		a.mdb = mdb
		if ap.domain != "" {
			a.d = d
		}
		return a, nil
	default:
		return nil, err
	}
}

// GetAddress
// Lookup an address under an active transaction
// really a copy of LookupAddress with transaction queries...
func (mdb *MailDB) GetAddress(addr string) (*Address, error) {
	var (
		ap  *AddressParts
		row *sql.Row
		err error
	)

	if ap, err = DecodeRFC822(addr); err != nil {
		return nil, err
	}
	if mdb.tx == nil {
		return nil, ErrMdbTransaction
	}
	a := &Address{
		mdb: mdb,
	}
	d := &Domain{
		mdb: mdb,
	}
	if ap.domain == "" { // A "local" address
		row = mdb.tx.QueryRow(qaLocal, ap.lpart)
		err = row.Scan(
			&a.id, &a.localpart, &a.transport, &a.rclass, &a.access)
	} else { // A full RFC822 address
		row = mdb.tx.QueryRow(qaRFC822, ap.lpart, ap.domain)
		err = row.Scan(
			&a.id, &a.localpart, &a.transport, &a.rclass, &a.access,
			&d.id, &d.name, &d.class, &d.transport, &d.access, &d.vuid, &d.vgid, &d.rclass)
	}
	switch err {
	case sql.ErrNoRows:
		return nil, ErrMdbAddressNotFound
	case nil:
		a.mdb = mdb
		if ap.domain != "" {
			a.d = d
		}
		return a, nil
	default:
		return nil, err
	}
}

// GetOrInsAddress get the address and if not found, insert it.
// Make this common pattern a function on its own. Transaction required
func (mdb *MailDB) GetOrInsAddress(addr string) (*Address, error) {
	a, err := mdb.GetAddress(addr)
	if err != nil {
		if err == ErrMdbAddressNotFound || err == ErrMdbDomainNotFound {
			a, err = mdb.InsertAddress(addr)
		}
	}
	return a, err
}

// lookupAddressByID
func (mdb *MailDB) lookupAddressByID(addrID int64) (*Address, error) {
	var (
		row    *sql.Row
		err    error
		addr   *Address
		domain sql.NullInt64
	)

	qa := `
SELECT localpart, domain, transport, rclass, access
FROM address WHERE id IS ?
`
	addr = &Address{
		mdb: mdb,
		id:  addrID,
	}
	row = mdb.db.QueryRow(qa, addrID)
	switch err = row.Scan(&addr.localpart, &domain,
		&addr.transport, &addr.rclass, &addr.access); err {
	case sql.ErrNoRows:
		return nil, ErrMdbAddressNotFound
	case nil:
		break
	default:
		return nil, err
	}
	if domain.Valid {
		d, err := mdb.LookupDomainByID(domain.Int64)
		if err == nil {
			addr.d = d
		} else {
			return nil, err
		}
	}
	addr.mdb = mdb
	return addr, nil
}

// FindAddress
func (mdb *MailDB) FindAddress(address string) ([]*Address, error) {
	var (
		err  error
		ap   *AddressParts
		q    string
		rows *sql.Rows
		al   []*Address
		dl   []*Domain
	)

	if ap, err = DecodeRFC822(address); err != nil {
		return nil, err
	}
	q = "SELECT id, localpart, transport, rclass, access FROM address"
	if ap.domain == "" { // "*" is for locals only
		qa := q + " WHERE domain IS NULL"
		if ap.lpart == "*" {
			qa += " ORDER BY localpart"
			rows, err = mdb.db.Query(qa)
		} else {
			lp := strings.ReplaceAll(ap.lpart, "*", "%")
			qa += " AND localpart LIKE ? ORDER BY localpart"
			rows, err = mdb.db.Query(qa, lp)
		}
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			a := &Address{
				mdb: mdb,
			}
			err = rows.Scan(&a.id, &a.localpart, &a.transport, &a.rclass, &a.access)
			if err != nil {
				break
			}
			a.mdb = mdb
			al = append(al, a)
		}
		if e := rows.Close(); e != nil {
			if err == nil {
				err = e
			}
		}
		if err != nil {
			return nil, err
		}
	} else { // must be "*@*" do get all non-locals
		if dl, err = mdb.FindDomain(ap.domain); err != nil {
			return nil, err
		}
		for _, d := range dl {
			if ap.lpart == "*" {
				qd := q + " WHERE domain IS ? ORDER BY localpart"
				rows, err = mdb.db.Query(qd, d.Id())
			} else {
				lp := strings.ReplaceAll(ap.lpart, "*", "%")
				qd := q + " WHERE domain IS ? AND localpart LIKE ? ORDER BY localpart"
				rows, err = mdb.db.Query(qd, d.Id(), lp)
			}
			if err != nil {
				break
			}
			for rows.Next() {
				a := &Address{
					mdb: mdb,
				}
				err = rows.Scan(&a.id, &a.localpart, &a.transport, &a.rclass, &a.access)
				if err != nil {
					break
				}
				a.mdb = mdb
				a.d = d
				al = append(al, a)
			}
			if e := rows.Close(); e != nil {
				if err == nil {
					err = e
				}
			}
			if err != nil {
				break
			}
		}
	}
	if err == nil && len(al) == 0 {
		err = ErrMdbAddressNotFound
	}
	return al, err
}

// InsertAddress
// Insert an address MUST be under a transaction
func (mdb *MailDB) InsertAddress(address string) (*Address, error) {
	var (
		ap  *AddressParts
		a   *Address
		d   *Domain
		row *sql.Row
		res sql.Result
		err error
	)

	if mdb.tx == nil {
		return nil, ErrMdbTransaction
	}
	if ap, err = DecodeRFC822(address); err != nil {
		return nil, err
	}
	if ap.domain == "" { // A "local user" entry
		res, err = mdb.tx.Exec("INSERT INTO address (localpart) VALUES (?)", ap.lpart)
		if err != nil {
			if strings.Contains(err.Error(), "Duplicate insert") {
				err = ErrMdbDupAddress // caught by trigger, not constraint
			}
		}
	} else { // A Virtual alias entry
		if d, err = mdb.GetDomain(ap.domain); err != nil {
			if err == ErrMdbDomainNotFound {
				d, err = mdb.InsertDomain(ap.domain, "")
			}
		}
		if err == nil {
			res, err = mdb.tx.Exec("INSERT INTO address (localpart, domain) VALUES (?, ?)",
				ap.lpart, d.Id())
			if err != nil {
				if IsErrConstraintUnique(err) {
					err = ErrMdbDupAddress
				}
			}
		} else {
			return nil, err // error with domain
		}
	}
	if err == nil {
		if aid, err := res.LastInsertId(); err == nil {
			a = &Address{
				mdb:       mdb,
				id:        aid,
				localpart: ap.lpart,
			}
			row = mdb.tx.QueryRow(
				"SELECT transport, rclass, access FROM address WHERE id IS ?", aid)
			if err = row.Scan(&a.transport, &a.rclass, &a.access); err == nil {
				a.d = d
				return a, nil
			}
		}
	}
	return nil, err
}

// DeleteAddress
// does not need a transaction because the cleanup delete
// to an unreferenced domain is done by a trigger
func (mdb *MailDB) DeleteAddress(addr string) error {
	var (
		err error
		ap  *AddressParts
		res sql.Result
	)

	if ap, err = DecodeRFC822(addr); err != nil {
		return err
	}
	if ap.domain == "" {
		dq := "DELETE FROM address WHERE localpart = ? AND domain iS NULL"
		res, err = mdb.db.Exec(dq, ap.lpart)
	} else {
		dq := `
DELETE FROM address WHERE id = 
  (SELECT a.id FROM address a, domain d
    WHERE a.domain = d.id AND a.localpart = ? AND d.name = ?)
`
		res, err = mdb.db.Exec(dq, ap.lpart, ap.domain)
	}
	if err != nil {
		return err
	} else {
		c, err := res.RowsAffected()
		if err != nil {
			return err
		} else if c == 0 {
			return ErrMdbAddressNotFound
		}
	}
	return nil
}

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

// IsMailbox
func (a *Address) IsMailbox() bool {
	var (
		cnt int
		err error
	)
	row := a.mdb.db.QueryRow("SELECT COUNT(*) FROM address WHERE id = ?", a.id)
	if err := row.Scan(&cnt); err == nil {
		return cnt > 0
	}
	panic(fmt.Errorf("Select count() should not fail, %s", err))
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

// lookupAddress
// helper to find 'lpart@domain' and return an address id.
// return nil, err for bad stuff
func (mdb *MailDB) lookupAddress(ap *AddressParts) (*Address, error) {
	var (
		row *sql.Row
		err error
	)

	addr := &Address{
		mdb: mdb,
	}
	d := &Domain{
		mdb: mdb,
	}
	if ap.domain == "" { // A "local" address
		qa := `
SELECT id, localpart, transport, rclass, access FROM address
 WHERE localpart = ? AND domain IS NULL
`
		row = mdb.db.QueryRow(qa, ap.lpart)
		err = row.Scan(
			&addr.id, &addr.localpart, &addr.transport, &addr.rclass, &addr.access)
	} else { // A full RFC822 address
		qa := `
SELECT a.id, a.localpart, a.transport, a.rclass, a.access,
       d.id, d.name, d.class, d.transport, d.access, d.vuid, d.vgid, d.rclass
 FROM address AS a, domain AS d
 WHERE a.localpart = ? AND a.domain IS d.id AND d.name = ?
`
		row = mdb.db.QueryRow(qa, ap.lpart, ap.domain)
		err = row.Scan(
			&addr.id, &addr.localpart, &addr.transport, &addr.rclass, &addr.access,
			&d.id, &d.name, &d.class, &d.transport, &d.access, &d.vuid, &d.vgid, &d.rclass)
	}
	switch err {
	case sql.ErrNoRows:
		return nil, ErrMdbAddressNotFound
	case nil:
		addr.mdb = mdb
		if ap.domain != "" {
			addr.d = d
		}
		return addr, nil
	default:
		return nil, err
	}
}

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
		qa := `
SELECT id, localpart, transport, rclass, access FROM address
 WHERE localpart = ? AND domain IS NULL
`
		row = mdb.db.QueryRow(qa, ap.lpart)
		err = row.Scan(
			&a.id, &a.localpart, &a.transport, &a.rclass, &a.access)
	} else { // A full RFC822 address
		qa := `
SELECT a.id, a.localpart, a.transport, a.rclass, a.access,
       d.id, d.name, d.class, d.transport, d.access, d.vuid, d.vgid, d.rclass
 FROM address AS a, domain AS d
 WHERE a.localpart = ? AND a.domain IS d.id AND d.name = ?
`
		row = mdb.db.QueryRow(qa, ap.lpart, ap.domain)
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
// Lookup an address under and active transaction
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
		qa := `
SELECT id, localpart, transport, rclass, access FROM address
 WHERE localpart = ? AND domain IS NULL
`
		row = mdb.tx.QueryRow(qa, ap.lpart)
		err = row.Scan(
			&a.id, &a.localpart, &a.transport, &a.rclass, &a.access)
	} else { // A full RFC822 address
		qa := `
SELECT a.id, a.localpart, a.transport, a.rclass, a.access,
       d.id, d.name, d.class, d.transport, d.access, d.vuid, d.vgid, d.rclass
 FROM address AS a, domain AS d
 WHERE a.localpart = ? AND a.domain IS d.id AND d.name = ?
`
		row = mdb.tx.QueryRow(qa, ap.lpart, ap.domain)
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
	return addr, nil
}

// FindAddress
func (mdb *MailDB) FindAddress(address string) ([]*Address, error) {
	var (
		err        error
		ap         *AddressParts
		q          string
		rows       *sql.Rows
		al         []*Address
		dl         []*Domain
		addressCnt int
	)

	if ap, err = DecodeRFC822(address); err != nil {
		return nil, err
	}
	q = "SELECT id, localpart, transport, rclass, access FROM address"
	if ap.domain == "" { // if "*", start with locals
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
			al = append(al, a)
			addressCnt++
		}
		if e := rows.Close(); e != nil {
			if err == nil {
				err = e
			}
		}
		if err != nil {
			return nil, err
		}
	}
	/*	if address == "*" {
			ap.domain = "*" // make it all domains too
		}
	*/if ap.domain != "" {
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
			if err == nil {
				for rows.Next() {
					a := &Address{
						mdb: mdb,
					}
					err = rows.Scan(&a.id, &a.localpart, &a.transport, &a.rclass, &a.access)
					if err != nil {
						break
					}
					a.d = d
					al = append(al, a)
					addressCnt++
				}
				if e := rows.Close(); e != nil {
					if err == nil {
						err = e
					}
				}
			}
			if err != nil {
				break
			}
		}
	}
	if addressCnt == 0 {
		err = ErrMdbAddressNotFound
	}
	return al, err
}

// insertAddress
// Insert an address MUST be under a transaction
func (mdb *MailDB) insertAddress(ap *AddressParts) (*Address, error) {
	var (
		addr *Address
		d    *Domain
		row  *sql.Row
		res  sql.Result
		err  error
	)

	if mdb.tx == nil {
		return nil, ErrMdbTransaction
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
			addr = &Address{
				mdb:       mdb,
				id:        aid,
				localpart: ap.lpart,
			}
			row = mdb.tx.QueryRow(
				"SELECT transport, rclass, access FROM address WHERE id IS ?", aid)
			if err = row.Scan(&addr.transport, &addr.rclass, &addr.access); err == nil {
				addr.d = d
				return addr, nil
			}
		}
	}
	return nil, err
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

// deleteAddress
func (mdb *MailDB) deleteAddress(ap *AddressParts) error {
	addr, err := mdb.lookupAddress(ap)
	if err != nil {
		return err
	}
	return mdb.deleteAddressByAddr(addr)
}

// deleteAddressByAddr
// we consider foreign key on domain is not really an error here. throw other errors
func (mdb *MailDB) deleteAddressByAddr(addr *Address) error {
	res, err := mdb.tx.Exec("DELETE FROM address WHERE id = ?", addr.id)
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

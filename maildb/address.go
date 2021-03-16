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
	id        int64
	localpart string
	dname     string
	domain    sql.NullInt64
	transport sql.NullInt64
	rclass    sql.NullString
	access    sql.NullInt64
}

// IsLocal
// a "local" address has no domain (address.domain IS NULL)
func (a *Address) IsLocal() bool {
	if !a.domain.Valid {
		return true
	} else {
		return false
	}
}

func (a *Address) String() string {
	var (
		line strings.Builder
	)
	fmt.Fprintf(&line, "%s", a.localpart)
	if a.domain.Valid {
		fmt.Fprintf(&line, "@%s", a.dname)
	}
	return line.String()
}

// dump
func (a *Address) dump() string {
	var (
		line strings.Builder
	)
	fmt.Fprintf(&line, "id:%d, localpart: %s, ", a.id, a.localpart)
	if a.domain.Valid {
		fmt.Fprintf(&line, "domain id: %d, dname: %s, ", a.domain.Int64, a.dname)
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

	if ap.domain == "" { // An /etc/aliases entry
		qa := `
SELECT id, localpart, domain, transport, rclass, access, "" FROM address
 WHERE localpart = ? AND domain IS NULL
`
		row = mdb.db.QueryRow(qa, ap.lpart)
	} else { // A virtual alias entry
		qa := `
SELECT a.id, a.localpart, a.domain, a.transport, a.rclass, a.access, d.name
 FROM address AS a, domain AS d
 WHERE a.localpart = ? AND a.domain IS d.id AND d.name = ?
`
		row = mdb.db.QueryRow(qa, ap.lpart, ap.domain)
	}
	addr := &Address{}
	switch err = row.Scan(&addr.id, &addr.localpart, &addr.domain,
		&addr.transport, &addr.rclass, &addr.access, &addr.dname); err {
	case sql.ErrNoRows:
		return nil, ErrMdbAddressNotFound
	case nil:
		return addr, nil
	default:
		return nil, err
	}
}

// lookupAddressByID
func (mdb *MailDB) lookupAddressByID(addrID int64) (*Address, error) {
	var (
		row   *sql.Row
		err   error
		addr  *Address
		dname string
	)

	qa := `
SELECT id, localpart, domain, transport, rclass, access
FROM address WHERE id IS ?
`
	addr = &Address{}
	row = mdb.db.QueryRow(qa, addrID)
	switch err = row.Scan(&addr.id, &addr.localpart, &addr.domain,
		&addr.transport, &addr.rclass, &addr.access); err {
	case sql.ErrNoRows:
		return nil, ErrMdbAddressNotFound
	case nil:
		break
	default:
		return nil, err
	}
	if addr.domain.Valid {
		row = mdb.db.QueryRow("SELECT name FROM domain WHERE id IS ?", addr.domain)
		if err = row.Scan(&dname); err == nil {
			addr.dname = dname
		} else {
			return nil, err
		}
	}
	return addr, nil
}

// insertAddress
// Insert an address MUST be under a transaction
func (mdb *MailDB) insertAddress(ap *AddressParts) (*Address, error) {
	var (
		domID  sql.NullInt64
		addr   *Address
		addrID int64
		row    *sql.Row
		err    error
	)

	if ap.domain == "" { // An /etc/aliases entry
		domID = sql.NullInt64{
			Valid: false,
			Int64: 0,
		}
	} else { // A Virtual alias entry
		row = mdb.tx.QueryRow("SELECT id FROM domain WHERE name = ?", ap.domain)
		switch err = row.Scan(&domID); err {
		case sql.ErrNoRows: // Make a new virtual domain, assume its class is the default...
			res, err := mdb.tx.Exec("INSERT INTO domain (name) VALUES (?)", ap.domain)
			if err != nil {
				return nil, err
			}
			if id, err := res.LastInsertId(); err == nil {
				domID = sql.NullInt64{
					Valid: true,
					Int64: id,
				}
			} else {
				return nil, err
			}
		case nil: // existing domain
			break
		default:
			return nil, err
		}
	}
	// FIXME: just insert and detect dup IsErrConstraintUnique
	row = mdb.tx.QueryRow("SELECT id FROM address WHERE localpart = ? AND domain IS ?",
		ap.lpart, domID)
	switch err = row.Scan(&addrID); err {
	case sql.ErrNoRows: // Make a new alias
		res, err := mdb.tx.Exec("INSERT INTO address (localpart, domain) VALUES (?, ?)",
			ap.lpart, domID)
		if err != nil {
			return nil, err
		}
		if addrID, err = res.LastInsertId(); err != nil {
			return nil, err
		}
	case nil: // already exists.
		return nil, ErrMdbDupAddress
	default:
		return nil, err
	}
	addr = &Address{ // the rest of Address is not init'd. DB may have other defaults
		id:        addrID,
		localpart: ap.lpart,
		domain:    domID,
	}
	if domID.Valid {
		addr.dname = ap.domain
	}
	return addr, nil
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

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

// Class
// domain classes are used for easy categorizing domains, some of
// which we actually own and others (like relay) are somewhere else.
type Class int

const (
	internet Class = iota // obviously... for those "others"
	local                 // local domains, i.e. "localhost", my-domain etc.
	relay                 // domains that will be relayed by us
	virtual               // mainly virtual aliases
	vmailbox              // domains handled by dovecot
)

// Domain
type Domain struct {
	mdb       *MailDB // only valid after successful GetDomain
	errs      int     // count of errors in transaction
	id        int64
	name      string
	class     Class
	transport sql.NullInt64
	access    sql.NullInt64
	vuid      sql.NullInt64
	vgid      sql.NullInt64
	rclass    string
}

var domainClass = []string{
	internet: "internet",
	local:    "local",
	relay:    "relay",
	virtual:  "virtual",
	vmailbox: "vmailbox",
}

// String
func (d *Domain) String() string {
	var (
		line strings.Builder
	)

	fmt.Fprintf(&line, "%s", d.name)
	return line.String()
}

// dump
func (d *Domain) dump() string {
	var (
		line strings.Builder
	)

	fmt.Fprintf(&line, "id=%d, name=%s, class=%s, ", d.id, d.name, domainClass[d.class])
	if d.transport.Valid {
		fmt.Fprintf(&line, "transport=%d, ", d.transport.Int64)
	} else {
		fmt.Fprintf(&line, "transport=<NULL>, ")
	}
	if d.access.Valid {
		fmt.Fprintf(&line, "access=%d, ", d.access.Int64)
	} else {
		fmt.Fprintf(&line, "access=<NULL>, ")
	}
	if d.vuid.Valid {
		fmt.Fprintf(&line, "vuid=%d, ", d.vuid.Int64)
	} else {
		fmt.Fprintf(&line, "vuid=<NULL>, ")
	}
	if d.vgid.Valid {
		fmt.Fprintf(&line, "vgid=%d, ", d.vgid.Int64)
	} else {
		fmt.Fprintf(&line, "vgid=<NULL>, ")
	}
	fmt.Fprintf(&line, "rclass=%s.", d.rclass)
	return line.String()
}

// IsInternet
func (d *Domain) IsInternet() bool {
	if d.class == internet {
		return true
	} else {
		return false
	}
}

// IsLocal
func (d *Domain) IsLocal() bool {
	if d.class == local {
		return true
	} else {
		return false
	}
}

// IsRelay
func (d *Domain) IsRelay() bool {
	if d.class == relay {
		return true
	} else {
		return false
	}
}

// IsVirtual
func (d *Domain) IsVirtual() bool {
	if d.class == virtual {
		return true
	} else {
		return false
	}
}

// IsVmailbox
func (d *Domain) IsVmailbox() bool {
	if d.class == vmailbox {
		return true
	} else {
		return false
	}
}

// LookupDomain
// Does lookup outside a transaction
func (mdb *MailDB) LookupDomain(name string) (*Domain, error) {
	if name == "" {
		return nil, ErrMdbBadName
	}
	d := &Domain{
		name: name,
	}
	row := mdb.db.QueryRow(
		"SELECT id, class, transport, access, vuid, vgid, rclass FROM domain WHERE name = ?",
		name)
	switch err := row.Scan(&d.id, &d.class, &d.transport, &d.access, &d.vuid, &d.vgid, &d.rclass); err {
	case sql.ErrNoRows:
		return nil, ErrMdbDomainNotFound
	case nil:
		return d, nil
	default:
		return nil, err
	}
}

//InsertDomain
// requires transaction. May need a non-tx version, i.e. insertDomainTx
func (mdb *MailDB) InsertDomain(name string, class string) (*Domain, error) {
	var (
		dclass Class
		res    sql.Result
		err    error
	)

	if name == "" || strings.ContainsAny(name, "\n\r\t\f{}()[];\"") {
		return nil, ErrMdbBadName
	}
	switch strings.ToLower(class) {
	case "":
		dclass = internet
	case "internet":
		dclass = internet
	case "local":
		dclass = local
	case "relay":
		dclass = relay
	case "virtual":
		dclass = virtual
	case "vmailbox":
		dclass = vmailbox
	default:
		return nil, ErrMdbBadClass
	}

	if class == "" { // use the schema default
		res, err = mdb.tx.Exec("INSERT INTO domain (name) VALUES (?)", name)
	} else {
		res, err = mdb.tx.Exec("INSERT INTO domain (name, class) VALUES (?, ?)", name, int64(dclass))
	}
	if err != nil {
		if IsErrConstraintUnique(err) {
			return nil, ErrMdbDupDomain
		} else {
			return nil, err
		}
	}
	dID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	// Now query it to pick up the schema defaults
	d := &Domain{
		id:   dID,
		name: name,
	}
	row := mdb.tx.QueryRow("SELECT class, transport, access, vuid, vgid, rclass FROM domain WHERE id = ?",
		dID)
	if err = row.Scan(&d.class, &d.transport, &d.access, &d.vuid, &d.vgid, &d.rclass); err != nil {
		return nil, err
	}
	return d, nil
}

// GetDomain
// fetch the domain under transaction
func (mdb *MailDB) GetDomain(name string) (*Domain, error) {
	var err error

	if name == "" {
		return nil, ErrMdbBadName
	}
	d := &Domain{
		name: name,
	}
	if mdb.tx != nil {
		return nil, ErrMdbInTransaction
	}
	if err = mdb.begin(); err != nil {
		return nil, err
	}
	row := mdb.tx.QueryRow(
		"SELECT id, class, transport, access, vuid, vgid, rclass FROM domain WHERE name = ?",
		name)
	switch err := row.Scan(&d.id, &d.class, &d.transport, &d.access, &d.vuid, &d.vgid, &d.rclass); err {
	case sql.ErrNoRows:
		mdb.end(false)
		return nil, ErrMdbDomainNotFound
	case nil:
		d.mdb = mdb
		return d, nil
	default:
		mdb.end(false)
		return nil, err
	}
}

// SetTransport

// SetAccess

// SetVUid
func (d *Domain) SetVUid(vuid int) error {
	var err error

	res, err := d.mdb.tx.Exec("UPDATE domain SET vuid = ? WHERE id = ?", vuid, d.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil && c == 1 {
			d.vuid = sql.NullInt64{Valid: true, Int64: int64(vuid)}
			return nil
		} else {
			err = ErrMdbDomainNotFound
		}
	}
	d.errs++
	return err
}

// SetVGid
func (d *Domain) SetVGid(vgid int) error {
	var err error

	res, err := d.mdb.tx.Exec("UPDATE domain SET vgid = ? WHERE id = ?", vgid, d.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil && c == 1 {
			d.vgid = sql.NullInt64{Valid: true, Int64: int64(vgid)}
			return nil
		} else {
			err = ErrMdbDomainNotFound
		}
	}
	d.errs++
	return err
}

// SetRclass
func (d *Domain) SetRclass(rclass string) error {
	var err error

	res, err := d.mdb.tx.Exec("UPDATE domain SET rclass = ? WHERE id = ?", rclass, d.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil && c == 1 {
			d.rclass = rclass
			return nil
		} else {
			err = ErrMdbDomainNotFound
		}
	}
	d.errs++
	return nil
}

// Release Domain
// Commit and release the transaction acquired by GetDomain
// Rollback if there were any intervening errors
func (d *Domain) Release() {
	if d.mdb != nil && d.mdb.tx != nil {
		d.mdb.end(d.errs == 0)
		d.mdb = nil
	}
}

// DeleteDomain
func (mdb *MailDB) DeleteDomain(name string) error {
	res, err := mdb.db.Exec("DELETE FROM domain WHERE name = ?", name)
	if err != nil {
		if !IsErrConstraintForeignKey(err) {
			return err
		}
	} else {
		c, err := res.RowsAffected()
		if err != nil {
			return err
		} else if c == 0 {
			return ErrMdbDomainNotFound
		}
	}
	return nil
}

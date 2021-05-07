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

var className = map[string]Class{
	"internet": internet,
	"local":    local,
	"relay":    relay,
	"virtual":  virtual,
	"vmailbox": vmailbox,
}

// Id
func (d *Domain) Id() int64 {
	return d.id
}

// String just the name
func (d *Domain) String() string {
	var (
		line strings.Builder
	)

	fmt.Fprintf(&line, "%s", d.name)
	return line.String()
}

// Export is export/import file format
func (d *Domain) Export() string {
	var line strings.Builder

	fmt.Fprintf(&line, "%s %s", d.name, domainClass[d.class])
	return line.String()
}

// Class
func (d *Domain) Class() string {
	var line strings.Builder

	fmt.Fprintf(&line, "%s", domainClass[d.class])
	return line.String()
}

// Transport
// full transport stuff NYI
func (d *Domain) Transport() string {
	var line strings.Builder

	if d.transport.Valid {
		fmt.Fprintf(&line, "NYI(%d)", d.transport.Int64)
	} else {
		fmt.Fprintf(&line, "--")
	}
	return line.String()
}

// Access
// full access stuff NYI
func (d *Domain) Access() string {
	var line strings.Builder

	if d.access.Valid {
		fmt.Fprintf(&line, "NYI(%d)", d.access.Int64)
	} else {
		fmt.Fprintf(&line, "--")
	}
	return line.String()
}

// Vuid
func (d *Domain) Vuid() string {
	var line strings.Builder

	if d.vuid.Valid {
		fmt.Fprintf(&line, "%d", d.vuid.Int64)
	} else {
		fmt.Fprintf(&line, "--")
	}
	return line.String()
}

// Vgid
func (d *Domain) Vgid() string {
	var line strings.Builder

	if d.vgid.Valid {
		fmt.Fprintf(&line, "%d", d.vgid.Int64)
	} else {
		fmt.Fprintf(&line, "--")
	}
	return line.String()
}

// Rclass
func (d *Domain) Rclass() string {
	var line strings.Builder

	fmt.Fprintf(&line, "%s", d.rclass)
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
		d.mdb = mdb
		return d, nil
	default:
		return nil, err
	}
}

// FindDomain
// LookupDomain but with wildcards
// '*' - find all domains
// '*.somedomain' - find all subdomains of somedomain
// '*.*' - find all domains with subdomains
func (mdb *MailDB) FindDomain(name string) ([]*Domain, error) {
	var (
		err       error
		q         string
		d         *Domain
		dl        []*Domain
		domainCnt int
	)
	if name == "*" {
		q = `
SELECT id, name, class, transport, access, vuid, vgid, rclass FROM domain ORDER BY NAME`
	} else {
		name = strings.ReplaceAll(name, "*", "%")
		q = `
SELECT id, name, class, transport, access, vuid, vgid, rclass FROM domain WHERE name LIKE ? ORDER BY name`
	}
	rows, err := mdb.db.Query(q, name)
	for rows.Next() {
		d = &Domain{}
		if err = rows.Scan(&d.id, &d.name, &d.class, &d.transport,
			&d.access, &d.vuid, &d.vgid, &d.rclass); err != nil {
			break
		}
		d.mdb = mdb
		dl = append(dl, d)
		domainCnt++
	}
	if e := rows.Close(); e != nil {
		if err == nil {
			err = e
		}
	}
	if domainCnt == 0 {
		err = ErrMdbDomainNotFound
	}
	if err == nil {
		return dl, nil
	} else {
		return nil, err
	}
}

// InsertDomain and start a transaction which must be commited with Release()
// returns a *Domain. If error, rollback the transaction.
func (mdb *MailDB) InsertDomain(name string, class string) (*Domain, error) {
	var (
		dclass Class
		res    sql.Result
		err    error
		ok     bool
	)

	if name == "" || strings.ContainsAny(name, "\n\r\t\f{}()[];\"") ||
		strings.Contains(name, "..") { // '..' means an empty sub-domain. not allowed
		return nil, ErrMdbBadName
	}
	if class != "" {
		if dclass, ok = className[strings.ToLower(class)]; !ok {
			return nil, ErrMdbBadClass
		}
	}

	if mdb.tx == nil {
		return nil, ErrMdbTransaction
	}
	if class == "" { // use the schema default
		res, err = mdb.tx.Exec("INSERT INTO domain (name) VALUES (?)", name)
	} else {
		res, err = mdb.tx.Exec("INSERT INTO domain (name, class) VALUES (?, ?)", name, int64(dclass))
	}
	if err != nil {
		if IsErrConstraintUnique(err) {
			err = ErrMdbDupDomain
		}
	} else {
		if dID, err := res.LastInsertId(); err == nil {
			// Now query it to pick up the schema defaults
			d := &Domain{
				id:   dID,
				name: name,
			}
			row := mdb.tx.QueryRow(
				"SELECT class, transport, access, vuid, vgid, rclass FROM domain WHERE id = ?",
				dID)
			if err = row.Scan(&d.class, &d.transport, &d.access, &d.vuid, &d.vgid, &d.rclass); err == nil {
				d.mdb = mdb
				return d, nil
			}
		}
	}
	return nil, err
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
	if mdb.tx == nil {
		return nil, ErrMdbTransaction
	}
	row := mdb.tx.QueryRow(
		"SELECT id, class, transport, access, vuid, vgid, rclass FROM domain WHERE name = ?",
		name)
	switch err = row.Scan(&d.id, &d.class, &d.transport, &d.access, &d.vuid, &d.vgid, &d.rclass); err {
	case sql.ErrNoRows:
		return nil, ErrMdbDomainNotFound
	case nil:
		d.mdb = mdb
		return d, nil
	default:
		return nil, err
	}
}

// SetClass
func (d *Domain) SetClass(class string) error {
	var (
		dclass Class
		err    error
		ok     bool
	)

	if class == "" {
		dclass = Class(d.mdb.DefaultInt("domain.class"))
	} else {
		if dclass, ok = className[strings.ToLower(class)]; !ok {
			return ErrMdbBadClass
		}
	}
	res, err := d.mdb.tx.Exec("UPDATE domain SET class = ? WHERE id = ?", dclass, d.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				d.class = dclass
			} else {
				err = ErrMdbDomainNotFound
			}
		}
	}
	return err
}

// SetTransport

// SetAccess

// SetVUid
func (d *Domain) SetVUid(vuid int) error {
	var err error

	res, err := d.mdb.tx.Exec("UPDATE domain SET vuid = ? WHERE id = ?", vuid, d.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				d.vuid = sql.NullInt64{Valid: true, Int64: int64(vuid)}
			} else {
				err = ErrMdbDomainNotFound
			}
		}
	}
	return err
}

// SetVGid
func (d *Domain) SetVGid(vgid int) error {
	var err error

	res, err := d.mdb.tx.Exec("UPDATE domain SET vgid = ? WHERE id = ?", vgid, d.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				d.vgid = sql.NullInt64{Valid: true, Int64: int64(vgid)}
			} else {
				err = ErrMdbDomainNotFound
			}
		}
	}
	return err
}

// SetRclass
func (d *Domain) SetRclass(rclass string) error {
	var err error

	res, err := d.mdb.tx.Exec("UPDATE domain SET rclass = ? WHERE id = ?", rclass, d.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				d.rclass = rclass
			} else {
				err = ErrMdbDomainNotFound
			}
		}
	}
	return err
}

// DeleteDomain
func (mdb *MailDB) DeleteDomain(name string) error {
	res, err := mdb.db.Exec("DELETE FROM domain WHERE name = ?", name)
	if err != nil {
		if IsErrConstraintForeignKey(err) {
			err = ErrMdbDomainBusy
		}
	} else {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 0 {
				return ErrMdbDomainNotFound
			}
		}
	}
	return err
}

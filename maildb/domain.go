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
	transport *Transport
	access    *Access
	vuid      sql.NullInt64
	vgid      sql.NullInt64
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

// Name just the name
func (d *Domain) Name() string {
	return d.name
}

// Export is export/import file format
func (d *Domain) Export() string {
	var line strings.Builder

	fmt.Fprintf(&line, "%s class=%s", d.name, domainClass[d.class])
	if d.transport != nil {
		fmt.Fprintf(&line, ", transport=%s", d.transport.Name())
	}
	if d.vuid.Valid {
		fmt.Fprintf(&line, ", vuid=%d", d.vuid.Int64)
	}
	if d.vgid.Valid {
		fmt.Fprintf(&line, ", vgid=%d", d.vgid.Int64)
	}
	if d.access != nil {
		fmt.Fprintf(&line, ", rclass=%s", d.access.Name())
	}
	return line.String()
}

// Class
func (d *Domain) Class() string {
	var line strings.Builder

	fmt.Fprintf(&line, "%s", domainClass[d.class])
	return line.String()
}

// Transport
func (d *Domain) Transport() string {
	if d.transport != nil {
		return d.transport.Name()
	} else {
		return "--"
	}
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
	if d.access != nil {
		return d.access.Name()
	} else {
		return "--"
	}
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
	var (
		access sql.NullInt64
		trans  sql.NullInt64
		err    error
	)

	if name == "" {
		return nil, ErrMdbBadName
	}
	d := &Domain{
		mdb:  mdb,
		name: name,
	}
	row := mdb.db.QueryRow(
		"SELECT id, class, transport, access, vuid, vgid FROM domain WHERE name = ?",
		name)
	switch err := row.Scan(&d.id, &d.class, &trans, &access, &d.vuid, &d.vgid); err {
	case sql.ErrNoRows:
		return nil, ErrMdbDomainNotFound
	case nil:
		if access.Valid {
			if ac, err := mdb.getAccessById(access.Int64); err == nil {
				d.access = ac
			}
		}
		if err == nil && trans.Valid {
			if tr, err := mdb.getTransportById(trans.Int64); err == nil {
				d.transport = tr
			}
		}
	default:
		break
	}
	if err == nil {
		return d, nil
	} else {
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
		err    error
		access sql.NullInt64
		trans  sql.NullInt64
		q      string
		d      *Domain
		dl     []*Domain
	)
	if name == "*" {
		q = `
SELECT id, name, class, transport, access, vuid, vgid FROM domain ORDER BY NAME`
	} else {
		name = strings.ReplaceAll(name, "*", "%")
		q = `
SELECT id, name, class, transport, access, vuid, vgid FROM domain WHERE name LIKE ? ORDER BY name`
	}
	rows, err := mdb.db.Query(q, name)
	if err == nil {
		for rows.Next() {
			d = &Domain{mdb: mdb}
			if err = rows.Scan(&d.id, &d.name, &d.class, &trans,
				&access, &d.vuid, &d.vgid); err != nil {
				break
			}
			if access.Valid {
				if ac, err := mdb.getAccessById(access.Int64); err == nil {
					d.access = ac
				}
			}
			if err == nil && trans.Valid {
				if tr, err := mdb.getTransportById(trans.Int64); err == nil {
					d.transport = tr
				}
			}
			if err != nil {
				break
			}
			dl = append(dl, d)
		}
		if e := rows.Close(); e != nil {
			if err == nil {
				err = e
			}
		}
	}
	if err == nil && len(dl) == 0 {
		err = ErrMdbDomainNotFound
	}
	if err == nil {
		return dl, nil
	} else {
		return nil, err
	}
}

// InsertDomain
// returns a *Domain. If error, rollback the transaction.
func (mdb *MailDB) InsertDomain(name string) (*Domain, error) {
	var (
		res sql.Result
		err error
	)

	if name == "" || strings.ContainsAny(name, "\n\r\t\f{}()[];\"") ||
		strings.Contains(name, "..") { // '..' means an empty sub-domain. not allowed
		return nil, ErrMdbBadName
	}

	if mdb.tx == nil {
		return nil, ErrMdbTransaction
	}
	res, err = mdb.tx.Exec("INSERT INTO domain (name) VALUES (?)", name)
	if err != nil {
		if IsErrConstraintUnique(err) {
			err = ErrMdbDupDomain
		}
	} else {
		if dID, err := res.LastInsertId(); err == nil {
			// Now query it to pick up the schema defaults
			d := &Domain{
				mdb:  mdb,
				id:   dID,
				name: name,
			}
			// pick up the default class
			row := mdb.tx.QueryRow("SELECT class FROM domain WHERE id = ?", dID)
			if err = row.Scan(&d.class); err == nil {
				return d, nil
			}
		}
	}
	return nil, err
}

// GetDomain
// fetch the domain under transaction
func (mdb *MailDB) GetDomain(name string) (*Domain, error) {
	var (
		access sql.NullInt64
		trans  sql.NullInt64
		err    error
	)

	if name == "" {
		return nil, ErrMdbBadName
	}
	d := &Domain{
		mdb:  mdb,
		name: name,
	}
	if mdb.tx == nil {
		return nil, ErrMdbTransaction
	}
	row := mdb.tx.QueryRow(
		"SELECT id, class, transport, access, vuid, vgid FROM domain WHERE name = ?",
		name)
	switch err = row.Scan(&d.id, &d.class, &trans, &access, &d.vuid, &d.vgid); err {
	case sql.ErrNoRows:
		err = ErrMdbDomainNotFound
	case nil:
		if access.Valid {
			if ac, err := mdb.getAccessByIdTx(access.Int64); err == nil {
				d.access = ac
			}
		}
		if err == nil && trans.Valid {
			if tr, err := mdb.getTransportByIdTx(trans.Int64); err == nil {
				d.transport = tr
			}
		}
	default:
		break
	}
	if err == nil {
		return d, nil
	} else {
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
			fmt.Printf("Bad class (%s)\n", class)
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

// FIX must have unset too for transport, vuid, vgid, rclass etc.

// SetTransport
func (d *Domain) SetTransport(name string) error {
	var (
		tr  *Transport
		err error
	)

	if tr, err = d.mdb.GetTransport(name); err != nil {
		return err
	}
	res, err := d.mdb.tx.Exec("UPDATE domain SET transport = ? WHERE id = ?", tr.id, d.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				d.transport = tr
			} else {
				err = ErrMdbDomainNotFound
			}
		}
	}
	return err
}

// ClearTransport
func (d *Domain) ClearTransport() error {
	res, err := d.mdb.tx.Exec("UPDATE domain SET transport = NULL WHERE id = ?", d.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				d.transport = nil
			} else {
				err = ErrMdbDomainNotFound
			}
		}
	}
	return err
}

// SetVUid
func (d *Domain) SetVUid(vuid int64) error {
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

// ClearVUid
func (d *Domain) ClearVUid() error {
	var err error

	res, err := d.mdb.tx.Exec("UPDATE domain SET vuid = NULL WHERE id = ?", d.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				d.vuid = sql.NullInt64{Valid: false}
			} else {
				err = ErrMdbDomainNotFound
			}
		}
	}
	return err
}

// SetVGid
func (d *Domain) SetVGid(vgid int64) error {
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

// ClearVGid
func (d *Domain) ClearVGid() error {
	var err error

	res, err := d.mdb.tx.Exec("UPDATE domain SET vgid = NULL WHERE id = ?", d.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				d.vgid = sql.NullInt64{Valid: false}
			} else {
				err = ErrMdbDomainNotFound
			}
		}
	}
	return err
}

// SetRclass
func (d *Domain) SetRclass(rclass string) error {
	var (
		a   *Access
		err error
	)

	if a, err = d.mdb.GetAccess(rclass); err != nil {
		return err
	}
	res, err := d.mdb.tx.Exec("UPDATE domain SET access = ? WHERE id = ?", a.id, d.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				d.access = a
			} else {
				err = ErrMdbDomainNotFound
			}
		}
	}
	return err
}

// ClearRclass
func (d *Domain) ClearRclass() error {
	res, err := d.mdb.tx.Exec("UPDATE domain SET access = NULL WHERE id = ?", d.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				d.access = nil
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

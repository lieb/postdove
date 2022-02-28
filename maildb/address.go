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
	transport *Transport
	access    *Access
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

// Address
func (a *Address) Address() string {
	var (
		line strings.Builder
	)
	fmt.Fprintf(&line, "%s", a.localpart)
	if a.d != nil {
		fmt.Fprintf(&line, "@%s", a.d.Name())
	}
	return line.String()
}

// Transport
func (a *Address) Transport() string {
	return "--"
}

// Access
func (a *Address) Access() string {
	return "--"
}

// Rclass
func (a *Address) Rclass() string {
	var line strings.Builder

	if a.access != nil {
		fmt.Fprintf(&line, "%s", a.access.Name())
	} else {
		fmt.Fprintf(&line, "--")
	}
	return line.String()
}

// Export
func (a *Address) Export() string {
	var (
		line strings.Builder
	)

	fmt.Fprintf(&line, "%s", a.Address())
	if a.access != nil {
		fmt.Fprintf(&line, " rclass=%s", a.access.Name())
	} else {
		fmt.Fprintf(&line, " rclass=\"\"")
	}
	if a.transport != nil {
		fmt.Fprintf(&line, ", transport=%s", a.transport.Name())
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
		fmt.Fprintf(&line, "domain id: %d, dname: %s, ", a.d.Id(), a.d.Name())
	} else {
		fmt.Fprintf(&line, "domain id: <NULL>, dname: <empty>, ")
	}
	if a.transport != nil {
		fmt.Fprintf(&line, "transport: %s, ", a.transport.Name())
	} else {
		fmt.Fprintf(&line, "transport: <NULL>, ")
	}
	if a.access != nil {
		fmt.Fprintf(&line, "rclass: %s, ", a.access.Name())
	} else {
		fmt.Fprintf(&line, "rclass: <NULL>.")
	}
	return line.String()
}

// It would be nice to have everything in one query but the best we can do is
// join address and domain. Counting alias and vmailbox references is seriously
// messy and expensive.
//
// query for local (no domain) addresses
var qaLocal string = `
SELECT id, localpart, transport, access FROM address
 WHERE localpart = ? AND domain IS NULL
`

// query for full localpart@domain addresses
var qaRFC822 string = `
SELECT a.id, a.localpart, a.transport, a.access,
       d.id, d.name, d.class, d.transport, d.access, d.vuid, d.vgid
 FROM address AS a, domain AS d
 WHERE a.localpart = ? AND a.domain IS d.id AND d.name = ?
`

// LookupAddress
// Lookup an address without an active transaction
func (mdb *MailDB) LookupAddress(addr string) (*Address, error) {
	var (
		ap      *AddressParts
		row     *sql.Row
		aAccess sql.NullInt64
		dAccess sql.NullInt64
		aTrans  sql.NullInt64
		dTrans  sql.NullInt64
		err     error
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
			&a.id, &a.localpart, &aTrans, &aAccess)
	} else { // A full RFC822 address
		row = mdb.db.QueryRow(qaRFC822, ap.lpart, ap.domain)
		err = row.Scan(
			&a.id, &a.localpart, &aTrans, &aAccess,
			&d.id, &d.name, &d.class, &dTrans, &dAccess, &d.vuid, &d.vgid)
	}
	switch err {
	case sql.ErrNoRows:
		return nil, ErrMdbAddressNotFound
	case nil:
		if aAccess.Valid {
			if ac, err := mdb.getAccessById(aAccess.Int64); err == nil {
				a.access = ac
			}
		}
		if err == nil && aTrans.Valid {
			if tr, err := mdb.getTransportById(aTrans.Int64); err == nil {
				a.transport = tr
			}
		}
		if err == nil && ap.domain != "" {
			if dAccess.Valid {
				if ac, err := mdb.getAccessById(dAccess.Int64); err == nil {
					d.access = ac
				}
			}
			if err == nil && dTrans.Valid {
				if tr, err := mdb.getTransportById(dTrans.Int64); err == nil {
					d.transport = tr
				}
			}
			a.d = d
		}
	default:
		break
	}
	if err == nil {
		return a, nil
	} else {
		return nil, err
	}
}

// GetAddress
// Lookup an address under an active transaction
// really a copy of LookupAddress with transaction queries...
func (mdb *MailDB) GetAddress(addr string) (*Address, error) {
	var (
		ap      *AddressParts
		row     *sql.Row
		aAccess sql.NullInt64
		dAccess sql.NullInt64
		aTrans  sql.NullInt64
		dTrans  sql.NullInt64
		err     error
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
			&a.id, &a.localpart, &aTrans, &aAccess)
	} else { // A full RFC822 address
		row = mdb.tx.QueryRow(qaRFC822, ap.lpart, ap.domain)
		err = row.Scan(
			&a.id, &a.localpart, &aTrans, &aAccess,
			&d.id, &d.name, &d.class, &dTrans, &dAccess, &d.vuid, &d.vgid)
	}
	switch err {
	case sql.ErrNoRows:
		return nil, ErrMdbAddressNotFound
	case nil:
		if aAccess.Valid {
			if ac, err := mdb.getAccessByIdTx(aAccess.Int64); err == nil {
				a.access = ac
			}
		}
		if err == nil && aTrans.Valid {
			if tr, err := mdb.getTransportByIdTx(aTrans.Int64); err == nil {
				a.transport = tr
			}
		}
		if err == nil && ap.domain != "" {
			if dAccess.Valid {
				if ac, err := mdb.getAccessByIdTx(dAccess.Int64); err == nil {
					d.access = ac
				}
			}
			if err == nil && dTrans.Valid {
				if tr, err := mdb.getTransportByIdTx(dTrans.Int64); err == nil {
					d.transport = tr
				}
			}
			a.d = d
		}
	default:
		break
	}
	if err == nil {
		return a, nil
	} else {
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

// Alias
// Return the alias recipients for this address
func (a *Address) Alias() (*Alias, error) {
	var (
		rows    *sql.Rows
		row     *sql.Row
		domain  sql.NullInt64
		aAccess sql.NullInt64
		dAccess sql.NullInt64
		aTrans  sql.NullInt64
		dTrans  sql.NullInt64
		err     error
	)

	qal := `SELECT id, target, extension FROM alias WHERE address IS ? ORDER BY id`
	qa := `
SELECT localpart, domain, transport, access
FROM address WHERE id IS ?
`
	qd := `
SELECT name, class, transport, access, vuid, vgid FROM domain WHERE id IS ?
`
	al := &Alias{
		addr: a,
	}
	rows, err = a.mdb.db.Query(qal, a.id)
	for rows.Next() {
		var (
			target sql.NullInt64
		)

		r := &Recipient{}
		if err = rows.Scan(&r.id, &target, &r.ext); err != nil {
			return nil, err
		}
		if target.Valid {
			ta := &Address{
				mdb: a.mdb, id: target.Int64,
			}
			row = a.mdb.db.QueryRow(qa, target.Int64)
			switch err = row.Scan(&ta.localpart, &domain, &aTrans, &aAccess); err {
			case sql.ErrNoRows:
				err = ErrMdbAddressNotFound
			case nil:
				if aAccess.Valid {
					if ac, err := a.mdb.getAccessById(aAccess.Int64); err == nil {
						a.access = ac
					}
				}
				if err != nil && aTrans.Valid {
					if tr, err := a.mdb.getTransportById(aTrans.Int64); err == nil {
						a.transport = tr
					}
				}
				if err == nil && domain.Valid {
					d := &Domain{mdb: a.mdb, id: domain.Int64}
					row = a.mdb.db.QueryRow(qd, domain.Int64)
					switch err = row.Scan(&d.name, &d.class, &dTrans, &dAccess, &d.vuid, &d.vgid); err {
					case sql.ErrNoRows:
						err = ErrMdbDomainNotFound
					case nil:
						if dAccess.Valid {
							if ac, err := a.mdb.getAccessById(dAccess.Int64); err == nil {
								d.access = ac
							}
						}
						if err != nil && dTrans.Valid {
							if tr, err := a.mdb.getTransportById(dTrans.Int64); err == nil {
								d.transport = tr
							}
						}
					default:
						break
					}
					if err == nil {
						ta.d = d
					}
				}
			default:
				break
			}
			if err == nil {
				r.t = ta
			} else {
				break
			}
		} else {
			if !a.IsLocal() {
				err = ErrMdbAddressTarget
				break
			}
		}
		al.recips = append(al.recips, r)

	}
	if e := rows.Close(); e != nil {
		if err == nil {
			err = e
		}
	}
	if err == nil && len(al.recips) == 0 {
		return nil, ErrMdbNotAlias
	}
	return al, nil
}

// FindAddress
func (mdb *MailDB) FindAddress(address string) ([]*Address, error) {
	var (
		err     error
		ap      *AddressParts
		q       string
		rows    *sql.Rows
		aAccess sql.NullInt64
		aTrans  sql.NullInt64
		al      []*Address
		dl      []*Domain
	)

	if ap, err = DecodeRFC822(address); err != nil {
		return nil, err
	}
	q = "SELECT id, localpart, transport, access FROM address"
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
			err = rows.Scan(&a.id, &a.localpart, &aTrans, &aAccess)
			if err != nil {
				break
			}
			if aAccess.Valid {
				if ac, err := mdb.getAccessById(aAccess.Int64); err == nil {
					a.access = ac
				}
			}
			if err != nil && aTrans.Valid {
				if tr, err := mdb.getTransportById(aTrans.Int64); err == nil {
					a.transport = tr
				}
			}
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
				err = rows.Scan(&a.id, &a.localpart, &aTrans, &aAccess)
				if err != nil {
					break
				}
				if aAccess.Valid {
					if ac, err := mdb.getAccessById(aAccess.Int64); err == nil {
						a.access = ac
					}
				}
				if err != nil && aTrans.Valid {
					if tr, err := mdb.getTransportById(aTrans.Int64); err == nil {
						a.transport = tr
					}
				}
				if err == nil {
					a.d = d
					al = append(al, a)
				} else {
					break
				}
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
				d, err = mdb.InsertDomain(ap.domain)
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
				d:         d,
			}
		}
	}
	if err == nil {
		return a, nil
	} else {
		return nil, err
	}
}

// AttachAlias
// Attach an alias recipient to this address. If it was just a simple address before,
// it is now an alias with one or more recipients
func (a *Address) AttachAlias(target string) error {
	var (
		err     error
		rp      *AddressParts
		rAddr   *Address
		recipID sql.NullInt64
		ext     sql.NullString
	)

	if rp, err = DecodeTarget(target); err != nil {
		return err
	}
	if !a.IsLocal() && rp.IsPipe() { // a virtual alias cannot have a pipe target
		err = ErrMdbAddressTarget
		return err
	}
	if rp.extension != "" {
		ext = sql.NullString{Valid: true, String: rp.extension}
	}
	if !rp.IsPipe() { // we have a foo@baz address
		rAddr, err = a.mdb.GetOrInsAddress(target)
		if err == nil {
			recipID = sql.NullInt64{Valid: true, Int64: rAddr.id}
		}
	}
	if err == nil {
		// Now make the link
		_, err = a.mdb.tx.Exec("INSERT INTO alias (address, target, extension) VALUES (?, ?, ?)",
			a.id, recipID, ext)
	}
	return err
}

// SetTransport
func (a *Address) SetTransport(name string) error {
	var (
		tr  *Transport
		err error
	)

	if tr, err = a.mdb.GetTransport(name); err != nil {
		return err
	}
	res, err := a.mdb.tx.Exec("UPDATE address SET transport = ? WHERE id = ?", tr.id, a.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				a.transport = tr
			} else {
				err = ErrMdbTransNotFound
			}
		}
	}
	return err
}

// ClearTransport
func (a *Address) ClearTransport() error {
	res, err := a.mdb.tx.Exec("UPDATE address SET transport = NULL WHERE id = ?", a.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				a.transport = nil
			} else {
				err = ErrMdbTransNotFound
			}
		}
	}
	return err
}

// SetRclass
func (a *Address) SetRclass(name string) error {
	var (
		ac  *Access
		err error
	)

	if ac, err = a.mdb.GetAccess(name); err != nil {
		return err
	}
	res, err := a.mdb.tx.Exec("UPDATE address SET access = ? WHERE id = ?", ac.id, a.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				a.access = ac
			} else {
				err = ErrMdbAddressNotFound
			}
		}
	}
	return err
}

// ClearRclass
func (a *Address) ClearRclass() error {
	res, err := a.mdb.tx.Exec("UPDATE address SET access = NULL WHERE id = ?", a.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				a.access = nil
			} else {
				err = ErrMdbAddressNotFound
			}
		}
	}
	return err
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

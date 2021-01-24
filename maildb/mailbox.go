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
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3" // do I really need this?
)

// VMailbox
type VMailbox struct {
	id       int64
	active   int64
	user     string
	dname    string
	uid      sql.NullInt64
	gid      sql.NullInt64
	home     sql.NullString
	password sql.NullString
	d_id     sql.NullInt64
}

// GetVmailbox
// return one or more matching mailboxes
func (mdb *MailDB) GetVmailbox(vaddr string) ([]*VMailbox, error) {
	var (
		id       int64
		active   int64
		uid      sql.NullInt64
		gid      sql.NullInt64
		home     sql.NullString
		password sql.NullString
		user     string
		dname    string
		d_id     sql.NullInt64

		rows    *sql.Rows
		err     error
		ap      *AddressParts
		vb_list []*VMailbox
	)
	q := `
SELECT vm.id AS id, vm.active AS active, vm.uid AS uid, vm.gid AS gid,
       vm.home AS home, vm.password AS password, a.localpart AS user,
       d.name AS dname, d.id AS d_id
FROM vmailbox AS vm
JOIN address AS a ON (vm.id = a.id)
JOIN domain AS d ON (a.domain = d.id)
`
	if vaddr == "" { // get all mailboxes
		rows, err = mdb.db.Query(q)
	} else { // one mailbox
		ap, err = DecodeRFC822(vaddr)
		if err != nil {
			return nil, fmt.Errorf("GetVmailbox: decode %s, %s",
				vaddr, err)
		}
		if ap.extension != "" || ap.domain == "" {
			return nil, fmt.Errorf(
				"GetVmailbox: %s not user@domain", vaddr)
		}
		q += `
WHERE user = ? AND dname = ?
`
		rows, err = mdb.db.Query(q, ap.lpart, ap.domain)
	}
	if err != nil {
		return nil, fmt.Errorf("GetVmailbox: query = %s, %s", q, err)
	}
	for rows.Next() {
		err = rows.Scan(&id, &active, &uid, &gid, &home,
			&password, &user, &dname, &d_id)
		if err != nil {
			return nil, fmt.Errorf("GetVmailbox: Scan, %s", err)
		}
		vbox := &VMailbox{
			id:       id,
			active:   active,
			uid:      uid,
			gid:      gid,
			home:     home,
			password: password,
			user:     user,
			dname:    dname,
			d_id:     d_id,
		}
		vb_list = append(vb_list, vbox)
	}
	return vb_list, nil
}

//String
func (vm *VMailbox) String() string {
	var line strings.Builder

	home := "<NULL>"
	if vm.home.Valid {
		home = vm.home.String
	}
	password := "<NULL>"
	if vm.password.Valid {
		password = vm.password.String
	}
	uid := "<NULL>"
	if vm.uid.Valid {
		uid = strconv.FormatInt(vm.uid.Int64, 10)
	}
	gid := "<NULL>"
	if vm.gid.Valid {
		gid = strconv.FormatInt(vm.gid.Int64, 10)
	}
	fmt.Fprintf(&line, "%s@%s:%s:%s:%s:%s",
		vm.user, vm.dname, password, uid, gid, home)
	return line.String()
}

// NewVmailbox
func (mdb *MailDB) NewVmailbox(vaddr string, password sql.NullString,
	uid sql.NullInt64, gid sql.NullInt64, home sql.NullString) (*VMailbox, error) {
	var (
		err      error
		res      sql.Result
		row      *sql.Row
		ap       *AddressParts
		addr     *Address
		m_id     int64
		rowCount int64
	)
	if ap, err = DecodeRFC822(vaddr); err != nil {
		return nil, fmt.Errorf("NewVmailbox: %s", err)
	}
	if ap.domain == "" {
		return nil, fmt.Errorf("NewVmailbox: Mailbox must have a domain")
	}
	// Enter a transaction for everything else
	if mdb.tx, err = mdb.db.Begin(); err != nil {
		return nil, fmt.Errorf("NewVmailbox: Begin, %s", err)
	}
	defer func() {
		if err == nil {
			if err = mdb.tx.Commit(); err != nil {
				panic(fmt.Errorf("NewVmailbox: Commit, %s", err)) // really screwed...
			}
		} else {
			mdb.tx.Rollback()
		}
	}()
	// The domain must exist. All that dovecot wiring must be in place first
	// Think of this as a spellcheck...
	row = mdb.db.QueryRow("SELECT COUNT(*) FROM domain WHERE name = ?", ap.domain)
	switch err = row.Scan(&rowCount); err {
	case sql.ErrNoRows:
		return nil, fmt.Errorf("NewVmailbox: No such virtual domain")
	case nil:
		if rowCount != 1 {
			return nil, fmt.Errorf("NewVmailbox: virtual domain must exist first!")
		}
	default:
		return nil, fmt.Errorf("NewVmailbox: %s", err)
	}
	if addr, err = mdb.lookupAddress(ap); err != nil {
		return nil, fmt.Errorf("NewVmailbox: address lookup, %s", err)
	}
	if addr == nil { // must be new user
		if addr, err = mdb.insertAddress(ap); err != nil {
			return nil, fmt.Errorf("NewVmailbox: new address, %s", err)
		}
	} else { // make sure it's not an alias
		row = mdb.db.QueryRow("SELECT COUNT(*) FROM alias WHERE address = ?", addr.id)
		switch err = row.Scan(&rowCount); err {
		case sql.ErrNoRows:
			return nil, fmt.Errorf("NewVmailbox: count(alias) failed, %s", err)
		case nil:
			if rowCount > 0 {
				return nil, fmt.Errorf("NewVmailbox: Already an alias")
			}
		default:
			return nil, fmt.Errorf("NewVmailbox: select count, %s", err)
		}
	}

	// Now we can insert the mailbox.

	res, err = mdb.tx.Exec("INSERT INTO vmailbox (id, uid, gid, home, password) VALUES (?, ?, ?, ?, ?)",
		addr.id, uid, gid, home, password)
	if err != nil { // we'll be rolling back the new address we just created...
		return nil, fmt.Errorf("NewVmailbox: could not insert new mailbox, %s", err)
	}
	if m_id, err = res.LastInsertId(); err != nil {
		return nil, fmt.Errorf("Newmailbox: Cannot get id of new mailbox, %s", err)
	}
	vm := &VMailbox{
		id:       m_id,
		active:   1, // start off enabled or disabled??
		user:     ap.lpart,
		dname:    ap.domain,
		uid:      uid,
		gid:      gid,
		home:     home,
		password: password,
		d_id:     addr.domain,
	}
	return vm, nil
}

// ChangePassword

// EnableVmailbox
func (mdb *MailDB) EnableVmailbox(vaddr string) error {
	return nil
}

// DisableVmailbox
func (mdb *MailDB) DisableVmailbox(vaddr string) error {
	return nil
}

// ActiveVmailbox
func (mdb *MailDB) ActiveVmailbox(vaddr string) (bool, error) {
	return false, nil
}

//RemoveVmailbox
func (mdb *MailDB) RemoveVmailbox(vaddr string) error {
	return nil
}

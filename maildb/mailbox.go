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

	_ "github.com/mattn/go-sqlite3" // do I really need this?
)

// VMailbox
type VMailbox struct {
	addr     *Address
	pw_type  string
	password sql.NullString
	uid      sql.NullInt64
	gid      sql.NullInt64
	quota    sql.NullInt64
	home     sql.NullString
	enable   int64
}

// String
func (vm *VMailbox) String() string {
	var (
		line strings.Builder
	)

	fmt.Fprintf(&line, "%s:", vm.addr.String())
	if vm.password.Valid {
		fmt.Fprintf(&line, "{%s}%s:", vm.pw_type, vm.password.String)
	} else {
		fmt.Fprintf(&line, "{%s}*:", vm.pw_type) // maybe this should be a required?
	}
	if vm.uid.Valid {
		fmt.Fprintf(&line, "%d:", vm.uid.Int64)
	} else {
		fmt.Fprintf(&line, ":")
	}
	if vm.gid.Valid {
		fmt.Fprintf(&line, "%d:", vm.gid.Int64)
	} else {
		fmt.Fprintf(&line, ":")
	}
	if vm.quota.Valid {
		fmt.Fprintf(&line, "%d:", vm.quota.Int64)
	} else {
		fmt.Fprintf(&line, ":")
	}
	if vm.home.Valid {
		fmt.Fprintf(&line, "%s:", vm.home.String)
	} else {
		fmt.Fprintf(&line, ":")
	}
	if vm.enable != 0 {
		fmt.Fprintf(&line, "true")
	} else {
		fmt.Fprintf(&line, "false")
	}

	return line.String()
}

// LookupVMailbox
// name@domain username for mailbox
// *@domain all users in this domain
// *@* all users in all domains
// * error. No local system users (for now)
func (mdb *MailDB) LookupVMailbox(mbox string) ([]*VMailbox, error) {
	var (
		ap      *AddressParts
		mb_list []*VMailbox
		rows    *sql.Rows
		err     error
	)

	if ap, err = DecodeRFC822(mbox); err != nil {
		return nil, err
	}
	q := `SELECT a.id, a.localpart, a.domain, a.transport, a.rclass, a.access, d.name `
	if ap.lpart == "*" || ap.domain == "*" { // wildcard
		rowCnt := 0
		if ap.lpart == "*" && ap.domain == "*" { // all mailboxes
			q += `FROM address AS a, domain AS d WHERE a.domain = d.id ORDER by a.domain, a.id`
			rows, err = mdb.db.Query(q)
		} else if ap.lpart == "*" && len(ap.domain) > 0 && ap.domain != "*" { // all in this domain
			q += `FROM address AS a, domain AS d WHERE a.domain IS d.id AND d.name = ? ORDER BY d.id, a.id`
			rows, err = mdb.db.Query(q, ap.domain)
		} else {
			return nil, ErrMdbBadMboxWild
		}
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			a := &Address{}
			if err = rows.Scan(&a.id, &a.localpart, &a.domain, &a.transport, &a.rclass, &a.access,
				&a.dname); err != nil {
				return nil, err
			}
			mb, err := mdb.lookupVMailboxByAddr(a)
			if err != nil {
				if err == ErrMdbNotMbox {
					continue // just skip these guys
				} else {
					return nil, err
				}
			}
			mb_list = append(mb_list, mb)
			rowCnt++
		}
		if err = rows.Close(); err != nil {
			return nil, err
		}
		if rowCnt == 0 {
			return nil, ErrMdbNotMbox
		}
	} else { // single mailbox
		a, err := mdb.lookupAddress(ap)
		if err != nil {
			return nil, err
		}
		mb, err := mdb.lookupVMailboxByAddr(a)
		if err != nil {
			return nil, err
		}
		mb_list = append(mb_list, mb)
	}
	return mb_list, nil
}

// lookupVMailboxByAddr
func (mdb *MailDB) lookupVMailboxByAddr(addr *Address) (*VMailbox, error) {
	mb := &VMailbox{
		addr: addr,
	}
	qmb := `SELECT pw_type, password, uid, gid, quota, home, enable FROM vmailbox WHERE id IS ?`
	row := mdb.db.QueryRow(qmb, addr.id)
	switch err := row.Scan(&mb.pw_type, &mb.password, &mb.uid, &mb.gid, &mb.quota, &mb.home, &mb.enable); err {
	case sql.ErrNoRows:
		return nil, ErrMdbNotMbox
	case nil:
		return mb, nil
	default:
		return nil, err
	}
}

// NewVmailbox
func (mdb *MailDB) NewVmailbox(vaddr string, pw_type string, password sql.NullString,
	uid sql.NullInt64, gid sql.NullInt64, quota sql.NullInt64,
	home sql.NullString, enable sql.NullInt64) (*VMailbox, error) {
	var (
		err      error
		row      *sql.Row
		ap       *AddressParts
		addr     *Address
		rowCount int64
	)
	if ap, err = DecodeRFC822(vaddr); err != nil {
		return nil, err
	}
	if ap.domain == "" {
		return nil, ErrMdbMboxNoDomain
	}
	switch strings.ToLower(pw_type) { // This is not an exhaustive list ATM
	case "":
		break // use default
	case "plain":
		pw_type = "PLAIN"
	case "crypt":
		pw_type = "CRYPT"
	case "sha256":
		pw_type = "SHA256"
	default:
		return nil, ErrMdbMboxBadPw
	}
	// Enter a transaction for everything else
	if err = mdb.begin(); err != nil {
		return nil, err
	}
	defer mdb.end(err == nil)

	// The domain must exist. All that dovecot wiring must be in place first
	// Think of this as a spellcheck...
	_, err = mdb.LookupDomain(ap.domain)
	if err != nil {
		return nil, err
	}
	if addr, err = mdb.lookupAddress(ap); err != nil {
		if err == ErrMdbAddressNotFound { // must be a new user
			if addr, err = mdb.insertAddress(ap); err != nil {
				return nil, err
			}
		} else { // Something bad
			return nil, err
		}
	}
	// make sure it's not an alias
	row = mdb.db.QueryRow("SELECT COUNT(*) FROM alias WHERE address = ?", addr.id)
	if err = row.Scan(&rowCount); err != nil {
		return nil, err
	}
	if rowCount > 0 {
		return nil, ErrMdbIsAlias
	}

	// Now we can insert the mailbox.
	_, err = mdb.tx.Exec("INSERT INTO vmailbox (id) VALUES (?)", addr.id)
	if err != nil { // we'll be rolling back the new address we just created...
		return nil, fmt.Errorf("NewVmailbox: could not insert new mailbox, %s", err)
	}
	// This is a little convoluted but this way we can set the defaults in the schema, not the app code
	if pw_type != "" {
		_, err = mdb.tx.Exec("UPDATE vmailbox SET pw_type = ? WHERE id IS ?", pw_type, addr.id)
		if err != nil {
			return nil, err
		}
	}
	if password.Valid {
		_, err = mdb.tx.Exec("UPDATE vmailbox SET password = ? WHERE id IS ?", password.String, addr.id)
		if err != nil {
			return nil, err
		}
	}
	if uid.Valid {
		_, err = mdb.tx.Exec("UPDATE vmailbox SET uid = ? WHERE id IS ?", uid.Int64, addr.id)
		if err != nil {
			return nil, err
		}
	}
	if gid.Valid {
		_, err = mdb.tx.Exec("UPDATE vmailbox SET gid = ? WHERE id IS ?", gid.Int64, addr.id)
		if err != nil {
			return nil, err
		}
	}
	if quota.Valid {
		_, err = mdb.tx.Exec("UPDATE vmailbox SET quota = ? WHERE id IS ?", quota.Int64, addr.id)
		if err != nil {
			return nil, err
		}
	}
	if home.Valid {
		_, err = mdb.tx.Exec("UPDATE vmailbox SET home = ? WHERE id IS ?", home.String, addr.id)
		if err != nil {
			return nil, err
		}
	}
	if enable.Valid {
		_, err = mdb.tx.Exec("UPDATE vmailbox SET enable = ? WHERE id IS ?", enable.Int64, addr.id)
		if err != nil {
			return nil, err
		}
	}
	// now let's see what we actually got...
	vm := &VMailbox{
		addr: addr,
	}
	row = mdb.tx.QueryRow("SELECT pw_type, password, uid, gid, quota, home, enable FROM vmailbox WHERE id IS ?",
		addr.id)
	if err = row.Scan(&vm.pw_type, &vm.password, &vm.uid, &vm.gid, &vm.quota, &vm.home, &vm.enable); err != nil {
		return nil, err
	}
	return vm, nil
}

// ChangePassword

// EnableVmailbox
func (mdb *MailDB) EnableVmailbox(vaddr string) error {
	var (
		ap  *AddressParts
		a   *Address
		err error
	)

	if ap, err = DecodeRFC822(vaddr); err != nil {
		return err
	}

	// Enter a transaction for everything else
	if err = mdb.begin(); err != nil {
		return err
	}
	defer mdb.end(err == nil)

	if a, err = mdb.lookupAddress(ap); err != nil {
		return err
	}
	_, err = mdb.tx.Exec("UPDATE vmailbox SET enable = 1 WHERE id IS ?", a.id)
	if err != nil {
		return err
	}
	return nil
}

// DisableVmailbox
func (mdb *MailDB) DisableVmailbox(vaddr string) error {
	var (
		ap  *AddressParts
		a   *Address
		err error
	)

	if ap, err = DecodeRFC822(vaddr); err != nil {
		return err
	}

	// Enter a transaction for everything else
	if err = mdb.begin(); err != nil {
		return err
	}
	defer mdb.end(err == nil)

	if a, err = mdb.lookupAddress(ap); err != nil {
		return err
	}
	_, err = mdb.tx.Exec("UPDATE vmailbox SET enable = 0 WHERE id IS ?", a.id)
	if err != nil {
		return err
	}
	return nil
}

// ActiveVmailbox
func (mdb *MailDB) ActiveVmailbox(vaddr string) bool {
	var (
		ap     *AddressParts
		a      *Address
		enable int
		err    error
	)

	if ap, err = DecodeRFC822(vaddr); err != nil {
		return false
	}
	if a, err = mdb.lookupAddress(ap); err != nil {
		return false
	}
	row := mdb.db.QueryRow("SELECT enable FROM vmailbox WHERE id IS ?", a.id)
	switch err = row.Scan(&enable); err {
	case sql.ErrNoRows:
		return false // not a mailbox, not active QED
	case nil:
		if enable != 0 {
			return true
		} else {
			return false
		}
	default:
		return false
	}
}

//DeleteVMailbox
func (mdb *MailDB) DeleteVMailbox(vaddr string) error {
	var (
		ap  *AddressParts
		a   *Address
		err error
	)

	if ap, err = DecodeRFC822(vaddr); err != nil {
		return err
	}

	// Enter a transaction for everything else
	if err = mdb.begin(); err != nil {
		return err
	}
	defer mdb.end(err == nil)

	if a, err = mdb.lookupAddress(ap); err != nil {
		return err
	}
	res, err := mdb.tx.Exec("DELETE FROM vmailbox WHERE id IS ?", a.id)
	if err != nil {
		return err
	} else {
		c, err := res.RowsAffected()
		if err != nil {
			return err
		} else if c == 0 {
			return ErrMdbNotMbox
		}
	}
	if err = mdb.deleteAddressByAddr(a); err != nil {
		return err
	}
	return nil
}

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
	id       RowID
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
func (mdb *MailDB) GetVmailbox(vaddr string) ([]*VMailbox, error) {
	var (
		id       RowID
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

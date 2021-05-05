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
	a        *Address
	pw_type  string
	password sql.NullString
	uid      sql.NullInt64
	gid      sql.NullInt64
	home     sql.NullString
	quota    sql.NullString
	enable   int64
}

// String
// FIX: should be Export()...
func (vm *VMailbox) String() string {
	var (
		line strings.Builder
	)

	fmt.Fprintf(&line, "%s:", vm.a.String())
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
		fmt.Fprintf(&line, "%s:", vm.quota.String) // flip this with home for export
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

// Export
func (vm *VMailbox) Export() string {
	var (
		line strings.Builder
	)

	fmt.Fprintf(&line, "%s:", vm.a.String())
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
		fmt.Fprintf(&line, "%d::", vm.gid.Int64)
	} else {
		fmt.Fprintf(&line, "::")
	}
	// skip GECOS field
	if vm.home.Valid {
		fmt.Fprintf(&line, "%s::", vm.home.String)
	} else {
		fmt.Fprintf(&line, "::")
	}
	// skip shell field
	if vm.quota.Valid {
		fmt.Fprintf(&line, "userdb_quota_rule=%s ", vm.quota.String)
	} else {
		fmt.Fprintf(&line, "userdb_quota_rule=none ")
	}
	if vm.enable != 0 {
		fmt.Fprintf(&line, "mbox_enabled=true")
	} else {
		fmt.Fprintf(&line, "mbox_enabled=false")
	}

	return line.String()
}

// User
func (vm *VMailbox) User() string {
	return vm.a.String()
}

// PwType
func (vm *VMailbox) PwType() string {
	var line strings.Builder

	fmt.Fprintf(&line, "%s", vm.pw_type)
	return line.String()
}

// Password
func (vm *VMailbox) Password() string {
	var line strings.Builder

	if vm.password.Valid {
		fmt.Fprintf(&line, "%s", vm.password.String)
	} else {
		fmt.Fprintf(&line, "--")
	}
	return line.String()
}

// Uid
func (vm *VMailbox) Uid() string {
	var line strings.Builder

	if vm.uid.Valid {
		fmt.Fprintf(&line, "%d", vm.uid.Int64)
	} else {
		fmt.Fprintf(&line, "--")
	}
	return line.String()
}

// Gid
func (vm *VMailbox) Gid() string {
	var line strings.Builder

	if vm.gid.Valid {
		fmt.Fprintf(&line, "%d", vm.gid.Int64)
	} else {
		fmt.Fprintf(&line, "--")
	}
	return line.String()
}

//Home
func (vm *VMailbox) Home() string {
	var line strings.Builder

	if vm.home.Valid {
		fmt.Fprintf(&line, "%s", vm.home.String)
	} else {
		fmt.Fprintf(&line, "--")
	}
	return line.String()
}

//Quota
// only the value here. Caller has to wrap appropriately if going to Dovecot
func (vm *VMailbox) Quota() string {
	var line strings.Builder

	if vm.quota.Valid {
		fmt.Fprintf(&line, "%s", vm.quota.String)
	} else {
		fmt.Fprintf(&line, "none")
	}
	return line.String()
}

// IsEnabled
func (mb *VMailbox) IsEnabled() bool {
	if mb.enable != 0 {
		return true
	} else {
		return false
	}
}

// FindVMailbox
// name@domain username for mailbox
// *@domain all users in this domain
// *@* all users in all domains
// * error. No local system users (for now)
// return a list of matched mailboxes
func (mdb *MailDB) FindVMailbox(user string) ([]*VMailbox, error) {
	var (
		mb_list []*VMailbox
		a_list  []*Address
		err     error
		rowCnt  int
	)

	if a_list, err = mdb.FindAddress(user); err != nil {
		return nil, err
	}
	for _, a := range a_list {
		mb := &VMailbox{
			a: a,
		}
		qmb := `SELECT pw_type, password, uid, gid, quota, home, enable FROM vmailbox WHERE id IS ?`
		row := mdb.db.QueryRow(qmb, a.id)
		switch err := row.Scan(&mb.pw_type, &mb.password, &mb.uid, &mb.gid, &mb.quota, &mb.home, &mb.enable); err {
		case sql.ErrNoRows:
			continue // not a mailbox
		case nil:
			mb_list = append(mb_list, mb)
			rowCnt++
		default:
			return nil, err
		}
	}
	if err == nil && rowCnt == 0 {
		err = ErrMdbNoMailboxes
	}
	return mb_list, err
}

// LookupVmailbox
// lookup a mailbox without transactions
func (mdb *MailDB) LookupVMailbox(user string) (*VMailbox, error) {
	var (
		a   *Address
		err error
	)

	if a, err = mdb.LookupAddress(user); err != nil {
		return nil, err
	}
	mb := &VMailbox{
		a: a,
	}
	qmb := `SELECT pw_type, password, uid, gid, quota, home, enable FROM vmailbox WHERE id IS ?`
	row := mdb.db.QueryRow(qmb, a.id)
	switch err := row.Scan(&mb.pw_type, &mb.password, &mb.uid, &mb.gid, &mb.quota, &mb.home, &mb.enable); err {
	case sql.ErrNoRows:
		return nil, ErrMdbNotMbox
	case nil:
		return mb, nil
	default:
		return nil, err
	}
}

// GetVmailbox
// lookup a mailbox under a transaction
func (mdb *MailDB) GetVMailbox(user string) (*VMailbox, error) {
	var (
		a   *Address
		err error
	)
	if mdb.tx == nil {
		return nil, ErrMdbTransaction
	}
	if a, err = mdb.GetAddress(user); err != nil {
		return nil, err
	}
	mb := &VMailbox{
		a: a,
	}
	qmb := `SELECT pw_type, password, uid, gid, quota, home, enable FROM vmailbox WHERE id IS ?`
	row := mdb.tx.QueryRow(qmb, a.id)
	switch err := row.Scan(&mb.pw_type, &mb.password, &mb.uid, &mb.gid, &mb.quota, &mb.home, &mb.enable); err {
	case sql.ErrNoRows:
		return nil, ErrMdbNotMbox
	case nil:
		return mb, nil
	default:
		return nil, err
	}
}

// InsertVMailbox
// must be under a transaction
func (mdb *MailDB) InsertVMailbox(user string) (*VMailbox, error) {
	var (
		a   *Address
		err error
	)

	if mdb.tx == nil {
		return nil, ErrMdbTransaction
	}
	// the domain must exist and be a vmailbox class
	// if we fail with a dup entry that could be either an already existing mbox
	// or this address is an alias or something (which must be deleted before we can proceed)
	if a, err = mdb.InsertAddress(user); err != nil {
		return nil, err
	}
	// if we just created a new domain, it will be the default (not vmailbox) and fail
	if !a.InVMailDomain() {
		return nil, ErrMdbMboxNotMboxDomain
	}
	// Now we can insert the mailbox.
	_, err = mdb.tx.Exec("INSERT INTO vmailbox (id) VALUES (?)", a.Id())
	if err != nil {
		return nil, err
	}
	// now let's see what we actually got...
	vm := &VMailbox{
		a: a,
	}
	row := mdb.tx.QueryRow("SELECT pw_type, password, uid, gid, quota, home, enable FROM vmailbox WHERE id IS ?",
		a.Id())
	if err = row.Scan(&vm.pw_type, &vm.password, &vm.uid, &vm.gid, &vm.quota, &vm.home, &vm.enable); err != nil {
		return nil, err
	}
	return vm, nil
}

// SetPwType
func (m *VMailbox) SetPwType(pwType string) error {
	var err error

	// Check for legit type
	switch strings.ToLower(pwType) { // This is not an exhaustive list ATM
	case "":
		pwType = m.a.mdb.DefaultString("vmailbox.pw_type")
		break // use default
	case "plain":
		pwType = "PLAIN"
	case "crypt":
		pwType = "CRYPT"
	case "sha256":
		pwType = "SHA256"
	default:
		return ErrMdbMboxBadPw
	}
	res, err := m.a.mdb.tx.Exec("UPDATE vmailbox SET pw_type = ? WHERE id = ?", pwType, m.a.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				m.pw_type = pwType
			} else {
				err = ErrMdbBadUpdate
			}
		}
	}
	return err
}

// SetPassword
func (m *VMailbox) SetPassword(ps string) error {
	var (
		err error
		pw  sql.NullString
	)

	if ps == "" {
		pw = NullStr
	} else {
		pw = sql.NullString{Valid: true, String: ps}
	}
	res, err := m.a.mdb.tx.Exec("UPDATE vmailbox SET password = ? WHERE id = ?", pw, m.a.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				m.password = pw
			} else {
				err = ErrMdbBadUpdate
			}
		}
	}
	return err
}

// ClearPassword
func (m *VMailbox) ClearPassword() error {
	var (
		err error
	)

	res, err := m.a.mdb.tx.Exec("UPDATE vmailbox SET password = NULL WHERE id = ?", m.a.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				m.password = NullStr
			} else {
				err = ErrMdbBadUpdate
			}
		}
	}
	return err
}

// SetUid
func (m *VMailbox) SetUid(uid int64) error {
	var err error

	res, err := m.a.mdb.tx.Exec("UPDATE vmailbox SET uid = ? WHERE id = ?", uid, m.a.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				m.uid = sql.NullInt64{Valid: true, Int64: uid}
			} else {
				err = ErrMdbBadUpdate
			}
		}
	}
	return err
}

// ClearUid
func (m *VMailbox) ClearUid() error {
	var err error

	res, err := m.a.mdb.tx.Exec("UPDATE vmailbox SET uid = NULL WHERE id = ?", m.a.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				m.uid = NullInt
			} else {
				err = ErrMdbBadUpdate
			}
		}
	}
	return err
}

// SetGid
func (m *VMailbox) SetGid(gid int64) error {
	var err error

	res, err := m.a.mdb.tx.Exec("UPDATE vmailbox SET gid = ? WHERE id = ?", gid, m.a.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				m.gid = sql.NullInt64{Valid: true, Int64: gid}
			} else {
				err = ErrMdbBadUpdate
			}
		}
	}
	return err
}

// ClearGid
func (m *VMailbox) ClearGid() error {
	var err error

	res, err := m.a.mdb.tx.Exec("UPDATE vmailbox SET gid = NULL WHERE id = ?", m.a.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				m.gid = NullInt
			} else {
				err = ErrMdbBadUpdate
			}
		}
	}
	return err
}

// SetHome
func (m *VMailbox) SetHome(home string) error {
	var (
		err error
		hm  sql.NullString
	)

	if home == "" {
		hm = NullStr
	} else {
		hm = sql.NullString{Valid: true, String: home}
	}
	res, err := m.a.mdb.tx.Exec("UPDATE vmailbox SET home = ? WHERE id = ?", hm, m.a.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				m.home = hm
			} else {
				err = ErrMdbBadUpdate
			}
		}
	}
	return err
}

// ClearHome
func (m *VMailbox) ClearHome() error {
	var (
		err error
	)

	res, err := m.a.mdb.tx.Exec("UPDATE vmailbox SET home = NULL WHERE id = ?", m.a.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				m.home = NullStr
			} else {
				err = ErrMdbBadUpdate
			}
		}
	}
	return err
}

// SetQuota
func (m *VMailbox) SetQuota(quota string) error {
	var err error

	res, err := m.a.mdb.tx.Exec("UPDATE vmailbox SET quota = ? WHERE id = ?", quota, m.a.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				m.quota = sql.NullString{Valid: true, String: quota}
			} else {
				err = ErrMdbBadUpdate
			}
		}
	}
	return err
}

// ClearQuota
func (m *VMailbox) ClearQuota() error {
	var err error

	res, err := m.a.mdb.tx.Exec("UPDATE vmailbox SET quota = NULL WHERE id = ?", m.a.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				m.quota = NullStr
			} else {
				err = ErrMdbBadUpdate
			}
		}
	}
	return err
}

// ResetQuota
func (m *VMailbox) ResetQuota() error {
	var err error

	res, err := m.a.mdb.tx.Exec("UPDATE vmailbox SET quota = ? WHERE id = ?",
		m.a.mdb.DefaultString("vmailbox.quota"), m.a.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				row := m.a.mdb.tx.QueryRow("SELECT quota FROM vmailbox WHERE id IS ?",
					m.a.Id())
				err = row.Scan(&m.quota)
			} else {
				err = ErrMdbBadUpdate
			}
		}
	}
	return err
}

// Enable
func (m *VMailbox) Enable() error {
	var err error

	res, err := m.a.mdb.tx.Exec("UPDATE vmailbox SET enable = 1 WHERE id = ?", m.a.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				m.enable = 1
			} else {
				err = ErrMdbBadUpdate
			}
		}
	}
	return err
}

// Disable
func (m *VMailbox) Disable() error {
	var err error

	res, err := m.a.mdb.tx.Exec("UPDATE vmailbox SET enable = 0 WHERE id = ?", m.a.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				m.enable = 0
			} else {
				err = ErrMdbBadUpdate
			}
		}
	}
	return err
}

// DeleteVMailbox
// We also delete the address part but leave the domain alone. Domains are special
// here because dovecot admin has directory structure that needs the domain intact.
func (mdb *MailDB) DeleteVMailbox(vaddr string) error {
	var (
		err error
	)

	// Enter a transaction for everything else
	mdb.Begin()
	defer mdb.End(&err)

	vm, err := mdb.GetVMailbox(vaddr)
	if err != nil {
		return err
	}
	res, err := mdb.tx.Exec("DELETE FROM vmailbox WHERE id IS ?", vm.a.id)
	if err != nil {
		if err.Error() == "ErrMdbMboxIsRecip" {
			return ErrMdbMboxIsRecip
		} else {
			return err
		}
	} else {
		c, err := res.RowsAffected()
		if err != nil {
			return err
		} else if c == 0 {
			return ErrMdbNotMbox
		}
	}
	return nil
}

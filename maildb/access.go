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

// Access
type Access struct {
	mdb    *MailDB
	id     int64
	name   string
	action string
}

// getAccessById
// make an Access without transaction
func (mdb *MailDB) getAccessById(id int64) (*Access, error) {
	ac := &Access{mdb: mdb, id: id}
	row := mdb.db.QueryRow("SELECT name, action FROM access WHERE id = ?", id)
	switch err := row.Scan(&ac.name, &ac.action); err {
	case sql.ErrNoRows:
		return nil, ErrMdbAccessNotFound
	case nil:
		return ac, nil
	default:
		return nil, err
	}
}

// getAccessByIdTx
// make an Access within a transaction
func (mdb *MailDB) getAccessByIdTx(id int64) (*Access, error) {
	ac := &Access{mdb: mdb, id: id}
	row := mdb.tx.QueryRow("SELECT name, action FROM access WHERE id = ?", id)
	switch err := row.Scan(&ac.name, &ac.action); err {
	case sql.ErrNoRows:
		return nil, ErrMdbAccessNotFound
	case nil:
		return ac, nil
	default:
		return nil, err
	}
}

// LookupAccess
// outside transactions
func (mdb *MailDB) LookupAccess(name string) (*Access, error) {
	if name == "" {
		return nil, ErrMdbBadName
	}
	a := &Access{
		name: name,
		mdb:  mdb,
	}
	row := mdb.db.QueryRow("SELECT id, action FROM access WHERE name = ?", name)
	switch err := row.Scan(&a.id, &a.action); err {
	case sql.ErrNoRows:
		return nil, ErrMdbAccessNotFound
	case nil:
		return a, nil
	default:
		return nil, err
	}
}

// FindAccess
// '*' find all access rules
// 'something*something' find matching names
func (mdb *MailDB) FindAccess(name string) ([]*Access, error) {
	var (
		err  error
		rows *sql.Rows
		al   []*Access
		a    *Access
		q    string
	)
	if name == "*" {
		q = `SELECT id, name, action FROM access ORDER BY name`
		rows, err = mdb.db.Query(q)
	} else {
		name = strings.ReplaceAll(name, "*", "%")
		q = `SELECT id, name, action FROM access WHERE name LIKE ? ORDER BY name`
		rows, err = mdb.db.Query(q, name)
	}
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		a = &Access{
			mdb: mdb,
		}
		if err = rows.Scan(&a.id, &a.name, &a.action); err != nil {
			break
		}
		al = append(al, a)
	}
	if e := rows.Close(); e != nil {
		if err == nil {
			err = e
		}
	}
	if err == nil && len(al) == 0 {
		err = ErrMdbAccessNotFound
	}
	if err == nil {
		return al, nil
	} else {
		return nil, err
	}

}

// GetAccess
// inside transactions
func (mdb *MailDB) GetAccess(name string) (*Access, error) {
	if name == "" {
		return nil, ErrMdbBadName
	}
	a := &Access{
		name: name,
		mdb:  mdb,
	}
	if mdb.tx == nil {
		return nil, ErrMdbTransaction
	}
	row := mdb.tx.QueryRow("SELECT id, action FROM access WHERE name = ?", name)
	switch err := row.Scan(&a.id, &a.action); err {
	case sql.ErrNoRows:
		return nil, ErrMdbAccessNotFound
	case nil:
		return a, nil
	default:
		return nil, err
	}
}

// InsertAccess
func (mdb *MailDB) InsertAccess(name string, action string) (*Access, error) {
	var (
		res sql.Result
		err error
	)
	if name == "" {
		return nil, ErrMdbBadName
	}
	if action == "" {
		return nil, ErrMdbAccessBadAction
	}
	if mdb.tx == nil {
		return nil, ErrMdbTransaction
	}
	res, err = mdb.tx.Exec("INSERT INTO access (name, action) VALUES (?, ?)", name, action)
	if err != nil {
		if IsErrConstraintUnique(err) {
			err = ErrMdbDupAccess
		}
	} else {
		if aID, err := res.LastInsertId(); err == nil {
			a := &Access{
				mdb:    mdb,
				id:     aID,
				name:   name,
				action: action,
			}
			return a, nil
		}
	}
	return nil, err
}

// Name
func (a *Access) Name() string {
	return a.name
}

// Action
func (a *Access) Action() string {
	return a.action
}

// Export
func (a *Access) Export() string {
	var line strings.Builder

	fmt.Fprintf(&line, "%s %s", a.name, a.action)
	return line.String()
}

// SetAction
// under transaction
func (a *Access) SetAction(action string) error {
	var err error

	if action == "" {
		return ErrMdbAccessBadAction
	} else {
		if a.mdb.tx == nil {
			return ErrMdbTransaction
		}
		res, err := a.mdb.tx.Exec("UPDATE access SET action = ? WHERE id = ?",
			action, a.id)
		if err == nil {
			c, err := res.RowsAffected()
			if err == nil {
				if c == 1 {
					a.action = action
				} else {
					err = ErrMdbAccessNotFound
				}
			}
		}
	}
	return err
}

// DeleteAccess
func (mdb *MailDB) DeleteAccess(name string) error {
	res, err := mdb.db.Exec("DELETE FROM access WHERE name = ?", name)
	if err != nil {
		if IsErrConstraintForeignKey(err) {
			err = ErrMdbAccessBusy
		}
	} else {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 0 {
				return ErrMdbAccessNotFound
			}
		}
	}
	return err
}

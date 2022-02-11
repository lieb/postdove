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
	//"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3" // do I really need this?
)

// Transport DB table
type Transport struct {
	mdb       *MailDB
	id        int64
	name      string
	transport sql.NullString
	nexthop   sql.NullString
}

// getTransportById
// make a Transport without transaction
func (mdb *MailDB) getTransportById(id int64) (*Transport, error) {
	tr := &Transport{mdb: mdb, id: id}
	row := mdb.db.QueryRow(
		"SELECT name, transport, nexthop FROM transport WHERE id = ?", id)
	switch err := row.Scan(&tr.name, &tr.transport, &tr.nexthop); err {
	case sql.ErrNoRows:
		return nil, ErrMdbTransNotFound
	case nil:
		return tr, nil
	default:
		return nil, err
	}
}

// getTransportByIdTx
// make a Transport witn a transaction
func (mdb *MailDB) getTransportByIdTx(id int64) (*Transport, error) {
	tr := &Transport{mdb: mdb, id: id}
	row := mdb.tx.QueryRow(
		"SELECT name, transport, nexthop FROM transport WHERE id = ?", id)
	switch err := row.Scan(&tr.name, &tr.transport, &tr.nexthop); err {
	case sql.ErrNoRows:
		return nil, ErrMdbTransNotFound
	case nil:
		return tr, nil
	default:
		return nil, err
	}
}

// LookupTransport
// outside transactions
func (mdb *MailDB) LookupTransport(name string) (*Transport, error) {
	if name == "" {
		return nil, ErrMdbBadName
	}
	tr := &Transport{
		name: name,
		mdb:  mdb,
	}
	row := mdb.db.QueryRow("SELECT id, transport, nexthop FROM transport WHERE name = ?", name)
	switch err := row.Scan(&tr.id, &tr.transport, &tr.nexthop); err {
	case sql.ErrNoRows:
		return nil, ErrMdbTransNotFound
	case nil:
		return tr, nil
	default:
		return nil, err
	}
}

// FindTransport
// '*' find all transport rules
// 'something*something' find matching names
func (mdb *MailDB) FindTransport(name string) ([]*Transport, error) {
	var (
		err  error
		rows *sql.Rows
		tl   []*Transport
		tr   *Transport
		q    string
	)
	if name == "*" {
		q = `SELECT id, name, transport, nexthop FROM transport ORDER BY name`
		rows, err = mdb.db.Query(q)
	} else {
		name = strings.ReplaceAll(name, "*", "%")
		q = `SELECT id, name, transport, nexthop FROM transport WHERE name LIKE ? ORDER BY name`
		rows, err = mdb.db.Query(q, name)
	}
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		tr = &Transport{
			mdb: mdb,
		}
		if err = rows.Scan(&tr.id, &tr.name, &tr.transport, &tr.nexthop); err != nil {
			break
		}
		tl = append(tl, tr)
	}
	if e := rows.Close(); e != nil {
		if err == nil {
			err = e
		}
	}
	if len(tl) == 0 && err == nil {
		err = ErrMdbTransNotFound
	}
	if err == nil {
		return tl, nil
	} else {
		return nil, err
	}

}

// GetTransport
// inside transactions
func (mdb *MailDB) GetTransport(name string) (*Transport, error) {
	if name == "" {
		return nil, ErrMdbBadName
	}
	tr := &Transport{
		name: name,
		mdb:  mdb,
	}
	if mdb.tx == nil {
		return nil, ErrMdbTransaction
	}
	row := mdb.tx.QueryRow("SELECT id, transport, nexthop FROM transport WHERE name = ?", name)
	switch err := row.Scan(&tr.id, &tr.transport, &tr.nexthop); err {
	case sql.ErrNoRows:
		return nil, ErrMdbTransNotFound
	case nil:
		return tr, nil
	default:
		return nil, err
	}
}

// InsertTransport
// for now, empty string == NULL
func (mdb *MailDB) InsertTransport(name string, transport string, nexthop string) (*Transport, error) {
	var (
		res   sql.Result
		trans sql.NullString
		next  sql.NullString
		err   error
	)
	if name == "" {
		return nil, ErrMdbBadName
	}
	if mdb.tx == nil {
		return nil, ErrMdbTransaction
	}
	if transport == "" {
		trans = sql.NullString{Valid: false}
	} else {
		trans = sql.NullString{Valid: true, String: transport}
	}
	if nexthop == "" {
		next = sql.NullString{Valid: false}
	} else {
		next = sql.NullString{Valid: true, String: nexthop}
	}
	res, err = mdb.tx.Exec("INSERT INTO transport (name, transport, nexthop) VALUES (?, ?, ?)",
		name, trans, next)
	if err != nil {
		if IsErrConstraintUnique(err) {
			err = ErrMdbDupTrans
		}
	} else {
		if trID, err := res.LastInsertId(); err == nil {
			tr := &Transport{
				mdb:       mdb,
				id:        trID,
				name:      name,
				transport: trans,
				nexthop:   next,
			}
			return tr, nil
		}
	}
	return nil, err
}

// Name
func (tr *Transport) Name() string {
	return tr.name
}

// Transport
func (tr *Transport) Transport() string {
	if tr.transport.Valid {
		return tr.transport.String
	} else {
		return ""
	}
}

// Nexthop
func (tr *Transport) Nexthop() string {
	if tr.nexthop.Valid {
		return tr.nexthop.String
	} else {
		return ""
	}
}

// SetTransport
func (tr *Transport) SetTransport(trans string) error {
	var (
		transport sql.NullString
		err       error
	)

	if tr.mdb.tx == nil {
		return ErrMdbTransaction
	}
	if trans == "" {
		transport = sql.NullString{Valid: false}
	} else {
		transport = sql.NullString{Valid: true, String: trans}
	}
	res, err := tr.mdb.tx.Exec("UPDATE transport SET transport = ? WHERE id = ?", transport, tr.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				tr.transport = transport
			} else {
				err = ErrMdbTransNotFound
			}
		}
	}
	return err
}

// SetNexthop
func (tr *Transport) SetNexthop(hop string) error {
	var (
		nexthop sql.NullString
		err     error
	)

	if tr.mdb.tx == nil {
		return ErrMdbTransaction
	}
	if hop == "" {
		nexthop = sql.NullString{Valid: false}
	} else {
		nexthop = sql.NullString{Valid: true, String: hop}
	}
	res, err := tr.mdb.tx.Exec("UPDATE transport SET nexthop = ? WHERE id = ?", nexthop, tr.id)
	if err == nil {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 1 {
				tr.nexthop = nexthop
			} else {
				err = ErrMdbTransNotFound
			}
		}
	}
	return err
}

// DeleteTransport
func (mdb *MailDB) DeleteTransport(name string) error {
	res, err := mdb.db.Exec("DELETE FROM transport WHERE name = ?", name)
	if err != nil {
		if IsErrConstraintForeignKey(err) {
			err = ErrMdbTransBusy
		}
	} else {
		c, err := res.RowsAffected()
		if err == nil {
			if c == 0 {
				return ErrMdbTransNotFound
			}
		}
	}
	return err
}

/*
// NewTransport
func (mdb *MailDB) NewTransport(t string) error {
	var (
		err     error
		tr      *TransportParts
		trans   sql.NullString
		nexthop sql.NullString
	)

	if tr, err = DecodeTransport(t); err != nil {
		return fmt.Errorf("NewTransport: %s", err)
	}
	if tr.transport == "" {
		trans = sql.NullString{Valid: false}
	} else {
		trans = sql.NullString{Valid: true, String: tr.transport}
	}
	if tr.nexthop == "" {
		nexthop = sql.NullString{Valid: false}
	} else {
		nexthop = sql.NullString{Valid: true, String: tr.nexthop}
	}
	// Enter a transaction for everything else
	if mdb.tx, err = mdb.db.Begin(); err != nil {
		return fmt.Errorf("MakeAlias: begin, %s", err)
	}
	defer func() {
		if err == nil {
			if err = mdb.tx.Commit(); err != nil {
				panic(fmt.Errorf("MakeAlias: commit, %s", err)) // we are screwed
			}
		} else {
			mdb.tx.Rollback()
		}
	}()
	_, err = mdb.tx.Exec("INSERT INTO transport (transport, nexthop) VALUES( ?, ?)",
		trans, nexthop)
	if err != nil {
		return fmt.Errorf("NewTransport: insert, %s", err)
	}
	return nil
}
*/

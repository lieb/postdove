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
	"errors"
	"fmt"
	"strings"

	"github.com/mattn/go-sqlite3" // do I really need this here?
)

var (
	ErrMdbAddressEmpty      = errors.New("address is empty")
	ErrMdbTargetEmpty       = errors.New("target is empty")
	ErrMdbAddrIllegalChars  = errors.New("illegal chars in address")
	ErrMdbAddrNoAddr        = errors.New("address extension without user part")
	ErrMdbNoLocalPipe       = errors.New("no local pipe or redirect")
	ErrMdbBadInclude        = errors.New("badly formed or empty include")
	ErrMdbTransNoColon      = errors.New("No ':' separator")
	ErrMdbAddressNotFound   = errors.New("address not found")
	ErrMdbDomainNotFound    = errors.New("domain not found")
	ErrMdbDupAddress        = errors.New("Address already exists")
	ErrMdbDupDomain         = errors.New("Domain already exists")
	ErrMdbTransNotFound     = errors.New("transport not found")
	ErrMdbDupTrans          = errors.New("transport already exists")
	ErrMdbNotAlias          = errors.New("address is not an alias")
	ErrMdbBadAliasWild      = errors.New("Badly formed alias lookup")
	ErrMdbAddressTarget     = errors.New("virtual alias must have an addressable target")
	ErrMdbNoRecipients      = errors.New("No recipients supplied for alias")
	ErrMdbRecipientNotFound = errors.New("alias recipient not found")
)

// Sqlite3 errors we are interested in

// IsErrConstraintForeignKey
// attempting insert with either non-existent ref or
// delete with refs pointing to it.
func IsErrConstraintForeignKey(err error) bool {
	if e, ok := err.(sqlite3.Error); ok {
		if e.Code == sqlite3.ErrConstraint &&
			e.ExtendedCode == sqlite3.ErrConstraintForeignKey {
			return true
		} else {
			return false
		}
	} else {
		panic(err)
	}
}

// IsErrConstraintUnique
func IsErrConstraintUnique(err error) bool {
	if e, ok := err.(sqlite3.Error); ok {
		if e.Code == sqlite3.ErrConstraint &&
			e.ExtendedCode == sqlite3.ErrConstraintUnique {
			return true
		} else {
			return false
		}
	} else {
		panic(err)
	}
}

// IsErrConstraintNotNull
func IsErrConstraintNotNull(err error) bool {
	if e, ok := err.(sqlite3.Error); ok {
		if e.Code == sqlite3.ErrConstraint &&
			e.ExtendedCode == sqlite3.ErrConstraintNotNull {
			return true
		} else {
			return false
		}
	} else {
		panic(err)
	}
}

// MailDB
type MailDB struct {
	db *sql.DB
	tx *sql.Tx
}

// NewMailDB
// Sqlite DB open.  ":memory:" for testing...
func NewMailDB(dbPath string) (*MailDB, error) {

	db, err := sql.Open("sqlite3", "file:"+dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("NewMailDB: open, %s", err)
	}
	mdb := &MailDB{
		db: db,
	}
	return mdb, nil
}

// loadDB
func loadDB(dbPath string, schema string) (*MailDB, error) {
	var (
		mdb *MailDB
		err error
	)

	if mdb, err = NewMailDB(dbPath); err != nil {
		return nil, fmt.Errorf("loadDB: %s", err)
	}
	lines := strings.Split(schema, ";\n")
	for line, req := range lines {
		if _, err = mdb.db.Exec(req); err != nil {
			return nil, fmt.Errorf("loadDB: line %d: %s, %s", line, req, err)
		}
	}
	return mdb, nil
}

// begin
func (mdb *MailDB) begin() error {
	if tx, err := mdb.db.Begin(); err != nil {
		return fmt.Errorf("begin(): failed %s", err)
	} else {
		mdb.tx = tx
		return nil
	}
}

// end
func (mdb *MailDB) end(makeItSo bool) {
	if mdb.tx == nil {
		panic("End(): not in a transaction")
	}
	if makeItSo {
		if err := mdb.tx.Commit(); err != nil {
			panic(fmt.Errorf("end(): commit, %s", err)) // we are really screwed
		}
	} else {
		mdb.tx.Rollback()
	}
	mdb.tx = nil
}

// Close
func (mdb *MailDB) Close() {
	mdb.db.Close()
}

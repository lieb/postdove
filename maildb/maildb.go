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
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/mattn/go-sqlite3" // do I really need this here?
)

// Error return constants
var (
	ErrMdbTransaction       = errors.New("Not in a transaction")
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
	ErrMdbAddressBusy       = errors.New("Address still in use")
	ErrMdbDomainBusy        = errors.New("Domain still in use")
	ErrMdbTransNotFound     = errors.New("transport not found")
	ErrMdbDupTrans          = errors.New("transport already exists")
	ErrMdbNotAlias          = errors.New("address is not an alias")
	ErrMdbNoAliases         = errors.New("No Aliases")
	ErrMdbAddressTarget     = errors.New("virtual alias must have an addressable target")
	ErrMdbNoRecipients      = errors.New("No recipients supplied for alias")
	ErrMdbRecipientNotFound = errors.New("alias recipient not found")
	ErrMdbNoMailboxes       = errors.New("No Mailboxes")
	ErrMdbMboxNoDomain      = errors.New("Mailbox must have a domain")
	ErrMdbMboxNotMboxDomain = errors.New("Mailbox must be in a vmailbox domain")
	ErrMdbNotMbox           = errors.New("address is not a mailbox")
	ErrMdbIsAlias           = errors.New("New mailbox already an alias")
	ErrMdbIsMbox            = errors.New("New alias already a mailbox")
	ErrMdbMboxBadPw         = errors.New("Unrecognized password type")
	ErrMdbBadName           = errors.New("Not a correct name")
	ErrMdbBadClass          = errors.New("Unknown domain class")
	ErrMdbBadUid            = errors.New("User ID must be unsigned decimal integer")
	ErrMdbBadGid            = errors.New("Group ID must be unsigned decimal integer")
	ErrMdbBadUpdate         = errors.New("Update did not happen")
	ErrMdbMboxIsRecip       = errors.New("Mailbox is an alias recipient")
)

// Useful constants
var (
	NullStr = sql.NullString{Valid: false}
	NullInt = sql.NullInt64{Valid: false}
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
	db    *sql.DB
	tx    *sql.Tx
	dflts map[string]TableInfo
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
	mdb.dflts = make(map[string]TableInfo)
	return mdb, nil
}

type TableInfo struct {
	cid     int64
	name    string
	colType string
	notNull int64
	dflt    sql.NullString
	pk      int64
}

// FindDefaults
// This is deep Sqlite3 magic. The problem is that one can
// do and INSERT and get a column default (from the schema) for
// any column un-named in the INSERT. Wouldn't it be also nice to be able to:
//
//     UPDATE table SET foo = DEFAULT
//
// Yes it would but SQL doesn't allow it so everybody does a workaround.
// Here is how we do it for Sqlite3. We build a map of default fields by
// doing SELECTs from magic tables and functions. Porting to another DB
// will require engine specific incantations/supplications to the query god.
func (mdb *MailDB) findDefaults() error {
	var (
		info   TableInfo
		tables []string
		table  string
		rows   *sql.Rows
		err    error
	)

	rows, err = mdb.db.Query("SELECT name FROM sqlite_master WHERE type = 'table'")
	if err != nil {
		return fmt.Errorf("table lookup broke: %s", err)
	}
	for rows.Next() {
		if err = rows.Scan(&table); err != nil {
			return fmt.Errorf("Table lookup scan failed: %s", err)
		}
		tables = append(tables, strings.ToLower(table))
	}
	if err = rows.Close(); err != nil {
		return fmt.Errorf("table lookup scan close broke: %s", err)
	}
	for _, table = range tables {

		rows, err = mdb.db.Query("SELECT * FROM pragma_table_info(?)", table)
		if err != nil {
			return fmt.Errorf("table_info broke: %s", err)
		}
		for rows.Next() {
			if err = rows.Scan(&info.cid, &info.name, &info.colType,
				&info.notNull, &info.dflt, &info.pk); err != nil {
				return fmt.Errorf("table_info scan broke: %s", err)
			}
			if info.dflt.Valid { // we are only interested in cols with defaults
				mdb.dflts[table+"."+info.name] = info
			}
		}
	}
	if err = rows.Close(); err != nil {
		return fmt.Errorf("table_info scan close broke: %s", err)
	}
	if len(mdb.dflts) == 0 {
		return fmt.Errorf("No defaults found")
	}
	return nil
}

// DefaultString
// We panic() here because an error here means someone has
// changed/mis-matched to the schema and at that point
// the best thing to do is crash, not mess up the DB
func (mdb *MailDB) DefaultString(sym string) string {
	var (
		i  TableInfo
		ok bool
	)

	if len(mdb.dflts) == 0 {
		err := mdb.findDefaults()
		if err != nil {
			panic(fmt.Errorf("findDefaults: %s", err))
		}
	}
	if i, ok = mdb.dflts[sym]; !ok {
		panic(fmt.Errorf("%s not found", sym))
	}
	if i.colType != "TEXT" {
		panic(fmt.Errorf("DefaultString: %s is %s, should be 'TEXT'", sym, i.colType))
	}
	return strings.Trim(i.dflt.String, "'\"") // gets stored with the schema quote marks...
}

func (mdb *MailDB) DefaultInt(sym string) int64 {
	var (
		i  TableInfo
		ok bool
	)

	if len(mdb.dflts) == 0 {
		err := mdb.findDefaults()
		if err != nil {
			panic(fmt.Errorf("findDefaults: %s", err))
		}
	}
	if i, ok = mdb.dflts[sym]; !ok {
		panic(fmt.Errorf("%s not found", sym))
	}
	if i.colType != "INTEGER" {
		panic(fmt.Errorf("DefaultInt: %s is %s, should be 'INTEGER'", sym, i.colType))
	}
	if num, err := strconv.ParseInt(i.dflt.String, 10, 64); err != nil {
		panic(fmt.Errorf("DefaultInt: parse of %s(%s) to integer failed, %s",
			sym, i.dflt.String, err))
	} else {
		return num
	}
}

// LoadSchema
func (mdb *MailDB) LoadSchema(schema string) error {
	c, err := ioutil.ReadFile(schema)
	if err != nil {
		return fmt.Errorf("LoadSchema: ReadFile, %s", err)
	}
	lines := strings.Split(string(c), ";\n")
	for line, req := range lines {
		if _, err = mdb.db.Exec(req); err != nil {
			return fmt.Errorf("loadSchema: line %d: %s, %s", line, req, err)
		}
	}
	return nil
}

// Begin
// If a begin() goes bad, we are in serious trouble. Just crash
func (mdb *MailDB) Begin() {
	if tx, err := mdb.db.Begin(); err != nil {
		panic(fmt.Errorf("begin(): failed %s", err))
	} else {
		mdb.tx = tx
	}
}

// End
// This is deferred so pass a reference to the error var
// Commit on no errors, rollback otherwise
func (mdb *MailDB) End(err *error) {
	if mdb.tx == nil {
		panic("End(): not in a transaction")
	}
	if *err == nil {
		if err := mdb.tx.Commit(); err != nil {
			panic(fmt.Errorf("end(): commit, %s", err)) // we are really screwed
		}
	} else {
		mdb.tx.Rollback()
	}
	mdb.tx = nil
}

// Close
// This must match a successful NewMailDB or it will panic
// best practice is to defer a call here in the same function
// that did the open
func (mdb *MailDB) Close() {
	if mdb.db == nil {
		panic("mdb.Close called with database not open")
	}
	mdb.db.Close()
	mdb.db = nil
}

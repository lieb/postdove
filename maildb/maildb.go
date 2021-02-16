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

	"github.com/mattn/go-sqlite3" // do I really need this here?
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

// AddressParts
type AddressParts struct {
	lpart     string
	domain    string
	extension string
}

// DecodeRFC822 Decode an RFC822 address into its constituent parts
// Actually, we decode per RFC5322
func DecodeRFC822(addr string) (*AddressParts, error) {
	var (
		local  string = ""
		domain string = ""
		// extension is transparent here and embedded in local
	)
	a := strings.ToLower(strings.Trim(addr, " "))    // clean up and lower everything
	if strings.ContainsAny(a, "\n\r\t\f{}()[];\"") { // contains illegal cruft
		return nil, fmt.Errorf("DecodeRFC822: %s contains illegal characters", addr)
	}
	if strings.Contains(a, "@") { // local@fqdn
		at := strings.Index(a, "@")
		local = a[0:at]
		domain = a[at+1:]
	} else { // just local
		local = a
	}
	return &AddressParts{
		lpart:     local,
		domain:    domain,
		extension: "",
	}, nil
}

// DecodeTarget Decode an RFC822 address and the various options for extensions
func DecodeTarget(addr string) (*AddressParts, error) {
	ap := &AddressParts{
		lpart:     "",
		domain:    "",
		extension: addr,
	}
	if addr[0] == '/' || addr[0] == '|' { // a local pipe or file redirect
		if len(addr) > 1 {
			return ap, nil
		} else {
			return nil, fmt.Errorf("DecodeTarget: no local pipe or redirect")
		}
	} else if addr[0] == ':' {
		if len(addr) > 10 && addr[:9] == ":include:" { // an include
			return ap, nil
		} else {
			return nil, fmt.Errorf("DecodeTarget: badly formed or empty include")
		}
	} else {
		ap, err := DecodeRFC822(addr)
		if err != nil {
			return nil, fmt.Errorf("DecodeTarget: %s", err)
		}
		if strings.Contains(ap.lpart, "+") { // we have an address extension
			pl := strings.Index(ap.lpart, "+")
			loc := ap.lpart[0:pl]
			ext := ap.lpart[pl+1:]
			ap.lpart = loc
			ap.extension = ext
		}
		return ap, nil
	}
}

func (ap *AddressParts) String() string {
	var (
		line strings.Builder
	)
	if ap.lpart != "" {
		fmt.Fprintf(&line, "%s", ap.lpart)
		if ap.extension != "" {
			fmt.Fprintf(&line, "+%s", ap.extension)
		}
		if ap.domain != "" {
			fmt.Fprintf(&line, "@%s", ap.domain)
		}
	} else if ap.domain != "" {
		fmt.Fprintf(&line, "@%s", ap.domain)
	} else {
		fmt.Fprintf(&line, ap.extension)
	}
	return line.String()
}

type TransportParts struct {
	transport string
	nexthop   string
}

// DecodeTransport
func DecodeTransport(trans string) (*TransportParts, error) {
	i := strings.Index(trans, ":")
	if i >= 0 {
		t := &TransportParts{
			transport: trans[0:i],
			nexthop:   trans[i+1:],
		}
		return t, nil
	} else {
		return nil, fmt.Errorf("DecodeTransport: No ':' separator")
	}
}

// DB query/insert helpers

type Address struct {
	id        int64
	localpart string
	domain    sql.NullInt64
	transport sql.NullInt64
	rclass    sql.NullString
	access    sql.NullInt64
}

// dump
func (a *Address) dump() string {
	var (
		line strings.Builder
	)
	fmt.Fprintf(&line, "id:%d, localpart: %s, ", a.id, a.localpart)
	if a.domain.Valid {
		fmt.Fprintf(&line, "domain id: %d, ", a.domain.Int64)
	} else {
		fmt.Fprintf(&line, "domain id: <NULL>, ")
	}
	if a.transport.Valid {
		fmt.Fprintf(&line, "transport: %d, ", a.transport.Int64)
	} else {
		fmt.Fprintf(&line, "transport: <NULL>, ")
	}
	if a.rclass.Valid {
		fmt.Fprintf(&line, "rclass: %s, ", a.rclass.String)
	} else {
		fmt.Fprintf(&line, "rclass: <NULL>, ")
	}
	if a.access.Valid {
		fmt.Fprintf(&line, "access: %d, ", a.access.Int64)
	} else {
		fmt.Fprintf(&line, "access: <NULL>")
	}
	return line.String()
}

// lookupAddress
// helper to find 'lpart@domain' and return an address id.
// return nil, nil for "not found"
// return nil, err for bad stuff
func (mdb *MailDB) lookupAddress(ap *AddressParts) (*Address, error) {
	var (
		domID sql.NullInt64
		row   *sql.Row
		err   error
	)

	if ap.domain == "" { // An /etc/aliases entry
		domID = sql.NullInt64{
			Valid: false,
			Int64: 0,
		}
	} else { // A Virtual alias entry
		row = mdb.db.QueryRow("SELECT id FROM domain WHERE name = ?", ap.domain)
		switch err = row.Scan(&domID); err {
		case sql.ErrNoRows:
			return nil, nil // no such domain so not found address
		case nil: // existing domain
			break
		default:
			return nil, fmt.Errorf("lookupAddress: select address domain, %s", err)
		}
	}
	addr := &Address{}
	qa := `
SELECT id, localpart, domain, transport, rclass, access
FROM address WHERE localpart = ? AND domain = ?
`
	row = mdb.db.QueryRow(qa, ap.lpart, domID)
	switch err = row.Scan(&addr.id, &addr.localpart, &addr.domain,
		&addr.transport, &addr.rclass, &addr.access); err {
	case sql.ErrNoRows:
		return nil, nil // not found address in this domain
	case nil:
		return addr, nil
	default:
		return nil, fmt.Errorf("lookupAddress: select address localpart, %s", err)
	}
}

// insertAddress
// Insert an address MUST be under a transaction
func (mdb *MailDB) insertAddress(ap *AddressParts) (*Address, error) {
	var (
		domID  sql.NullInt64
		addr   *Address
		addrID int64
		row    *sql.Row
		err    error
	)

	if ap.domain == "" { // An /etc/aliases entry
		domID = sql.NullInt64{
			Valid: false,
			Int64: 0,
		}
	} else { // A Virtual alias entry
		row = mdb.db.QueryRow("SELECT id FROM domain WHERE name = ?", ap.domain)
		switch err = row.Scan(&domID); err {
		case sql.ErrNoRows: // Make a new virtual domain, assume its class is the default...
			res, err := mdb.tx.Exec("INSERT INTO domain (name) VALUES (?)", ap.domain)
			if err != nil {
				return nil, fmt.Errorf("insertAddress: new domain, %s", err)
			}
			if id, err := res.LastInsertId(); err == nil {
				domID = sql.NullInt64{
					Valid: true,
					Int64: id,
				}
			} else {
				return nil, fmt.Errorf(
					"insertAddress: Cannot get id of new domain, %s", err)
			}
		case nil: // existing domain
			break
		default:
			return nil, fmt.Errorf("insertAddress: select alias domain, %s", err)
		}
	}
	// FIXME: just insert and detect dup IsErrConstraintUnique
	row = mdb.db.QueryRow("SELECT id FROM address WHERE localpart = ? AND domain = ?",
		ap.lpart, domID)
	switch err = row.Scan(&addrID); err {
	case sql.ErrNoRows: // Make a new alias
		res, err := mdb.tx.Exec("INSERT INTO address (localpart, domain) VALUES (?, ?)",
			ap.lpart, domID)
		if err != nil {
			return nil, fmt.Errorf("insertAddress: new alias, %s", err)
		}
		if addrID, err = res.LastInsertId(); err != nil {
			return nil, fmt.Errorf("insertAddress: cannot get id of new alias, %s", err)
		}
	case nil: // already exists.
		return nil, nil
	default:
		return nil, fmt.Errorf("insertAddress: select alias localpart, %s", err)
	}
	addr = &Address{ // the rest of Address is not init'd. DB may have other defaults
		id:        addrID,
		localpart: ap.lpart,
		domain:    domID,
	}
	return addr, nil
}

// deleteAddress
func (mdb *MailDB) deleteAddress(ap *AddressParts) error {
	addr, err := mdb.lookupAddress(ap)
	if err != nil {
		fmt.Errorf("deleteAddress: %s", err)
	}
	if addr != nil {
		return mdb.deleteAddressByID(addr)
	} else {
		return fmt.Errorf("deleteAddress: address not found")
	}
}

// deleteAddressByID
// we consider foreign key on domain is not really an error here. throw other errors
func (mdb *MailDB) deleteAddressByID(addr *Address) error {
	if addr == nil {
		return fmt.Errorf("deleteAddressByID: nil addr")
	}
	_, err := mdb.tx.Exec("DELETE FROM address WHERE id = ?", addr.id)
	if err != nil && !IsErrConstraintForeignKey(err) {
		return fmt.Errorf("deleteAddressByID: delete address, %s", err)
	}
	if addr.domain.Valid { // See if we can delete the domain too
		_, err = mdb.tx.Exec("DELETE FROM domain WHERE id = ?", addr.domain.Int64)
		if err != nil && !IsErrConstraintForeignKey(err) {
			return fmt.Errorf("deleteAddressByID: delete domain, %s", err)
		}
	}
	return nil
}

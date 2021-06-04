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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3" // do I really need this here?
)

// makeTestDB
func makeTestDB(dbFile string) (*MailDB, error) {
	var (
		mdb *MailDB
		err error
	)

	if mdb, err = NewMailDB(dbFile); err != nil {
		return nil, fmt.Errorf("makeTestDB: %s", err)
	}
	if err = mdb.LoadSchema(""); err != nil {
		return nil, fmt.Errorf("makeTestDB: %s", err)
	}
	return mdb, nil
}

// countAddresses
func countAddresses(mdb *MailDB) (aCnt int, dCnt int) {
	row := mdb.db.QueryRow("SELECT count(*) FROM address")
	if err := row.Scan(&aCnt); err != nil {
		panic(fmt.Errorf("countAddresses: addresses, %s", err))
	}
	row = mdb.db.QueryRow("SELECT count(*) FROM domain")
	if err := row.Scan(&dCnt); err != nil {
		panic(fmt.Errorf("countAddresses: domains, %s", err))
	}
	return
}

// TestDBdefaults
func TestDBdefaults(t *testing.T) {
	var (
		err error
		mdb *MailDB
		dir string
	)

	fmt.Printf("Database load test\n")

	dir, err = ioutil.TempDir("", "TestDBLoad-*")
	defer os.RemoveAll(dir)
	mdb, err = makeTestDB(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Errorf("Database load failed, %s", err)
		return
	}
	defer mdb.Close()

	// test known table col defaults
	doDefaultString(mdb, t)
	doDefaultInt(mdb, t)
}

// test DefaultString
// in a separate function to recover the panics
func doDefaultString(mdb *MailDB, t *testing.T) {

	defer func() {
		if p := recover(); p != nil {
			if strings.Contains(string(p.(error).Error()), "not found") {
				return // intentional error
			} else { // all others really bad
				t.Errorf("DefaultString: unexpected panic: %v", p)
			}
		}
	}()

	// now play
	s := mdb.DefaultString("domain.rclass")
	if s != "DEFAULT" {
		t.Errorf("DefaultString: expected 'DEFAULT', got '%s'", s)
	}
	s = mdb.DefaultString("domain.name") // no default, should panic
	if len(s) >= 0 {                     // succeeded with something
		t.Errorf("DefaultString: should have panic'd on 'domain.name'")
	}
}

// test DefaultInt
func doDefaultInt(mdb *MailDB, t *testing.T) {

	defer func() {
		if p := recover(); p != nil {
			if strings.Contains(p.(error).Error(), "not found") {
				return // intentional error
			} else { // all others really bad
				t.Errorf("DefaultInt: unexpected panic: %v", p)
			}
		}
	}()

	// now play
	i := mdb.DefaultInt("vmailbox.enable")
	if i != 1 {
		t.Errorf("DefaultInt: expected 1, got %d", i)
	}
	i = mdb.DefaultInt("vmailbox.uid")
	if i >= 0 { // succeeded with something
		t.Errorf("DefaultInt: should have panic'd on 'vmailbox.uid'")
	}
}

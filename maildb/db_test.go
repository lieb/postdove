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
	//"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	//"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3" // do I really need this here?
)

// makeTestDB
func makeTestDB(dbFile string, schema string) (*MailDB, error) {
	var (
		mdb *MailDB
		err error
	)

	if mdb, err = NewMailDB(dbFile); err != nil {
		return nil, fmt.Errorf("makeTestDB: %s", err)
	}
	if err = mdb.LoadSchema(schema); err != nil {
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

// TestDBLoad
func TestDBLoad(t *testing.T) {
	var (
		err            error
		mdb            *MailDB
		dir            string
		aCount, dCount int
	)

	fmt.Printf("Database load test\n")

	dir, err = ioutil.TempDir("", "TestDBLoad-*")
	defer os.RemoveAll(dir)
	mdb, err = makeTestDB(filepath.Join(dir, "test.db"), "../schema.sql")
	if err != nil {
		t.Errorf("Database load failed, %s", err)
		return
	}
	defer mdb.Close()

	// test basic local address insert
	mdb.Begin()
	a, err := mdb.InsertAddress("dmr")
	mdb.End(&err)
	if err != nil {
		t.Errorf("insert of dmr failed %s", err)
	} else {
		aCount, dCount = countAddresses(mdb)
		if aCount != 1 || dCount != 0 {
			t.Errorf("insert of dmr: expected 1 addr, 0 domains, got %d, %d",
				aCount, dCount)
		}
		if a.String() != "dmr" {
			t.Errorf("dmr: bad String(), %s", a.String())
		}
		if a.dump() != "id:1, localpart: dmr, domain id: <NULL>, dname: <empty>, transport: <NULL>, rclass: <NULL>, access: <NULL>" {
			t.Errorf("dmr: bad dump(), %s", a.dump())
		}
	}

	// try to insert it again
	mdb.Begin()
	a, err = mdb.InsertAddress("dmr")
	mdb.End(&err)
	if err != nil && err != ErrMdbDupAddress {
		t.Errorf("duplicate insert of dmr, unexpected error %s", err)
	}

	// test basic insert. Should have one address row and one domain row
	mdb.Begin()
	a, err = mdb.InsertAddress("mary@goof.com")
	mdb.End(&err)
	if err != nil {
		t.Errorf("insert of mary@goof.com failed %s", err)
	} else {
		aCount, dCount = countAddresses(mdb)
		if aCount != 2 || dCount != 1 {
			t.Errorf("insert of mary@goof.com: expected 2 addr, 1 domain, got %d, %d",
				aCount, dCount)
		}
		if a.String() != "mary@goof.com" {
			t.Errorf("mary@goof.com: bad String(), %s", a.String())
		}
		if a.dump() != "id:2, localpart: mary, domain id: 1, dname: goof.com, transport: <NULL>, rclass: <NULL>, access: <NULL>" {
			t.Errorf("mary@goof.com: bad dump(), %s", a.dump())
		}
	}

	// try inserting it again
	mdb.Begin()
	a, err = mdb.InsertAddress("mary@goof.com")
	mdb.End(&err)
	if err != nil && err != ErrMdbDupAddress {
		t.Errorf("duplicate insert of mary@goof.com, unexpected error %s", err)
	}

	// second insert, same domain. should now have 2 address rows and 1 domain
	mdb.Begin()
	a, err = mdb.InsertAddress("bill@goof.com")
	mdb.End(&err)
	if err != nil {
		t.Errorf("insert of bill@goof.com failed %s", err)
	} else {
		aCount, dCount = countAddresses(mdb)
		if aCount != 3 || dCount != 1 {
			t.Errorf("insert of bill@goof.com: expected 3 addr, 1 domain, got %d, %d",
				aCount, dCount)
		}
		if a.dump() != "id:3, localpart: bill, domain id: 1, dname: goof.com, transport: <NULL>, rclass: <NULL>, access: <NULL>" {
			t.Errorf("bill@goof.com: bad dump(), %s", a.dump())
		}
	}

	// third insert is new domain. should have 4 addresses and 2 domains
	mdb.Begin()
	a, err = mdb.InsertAddress("dave@slip.com")
	mdb.End(&err)
	if err != nil {
		t.Errorf("insert of dave@slip.com failed %s", err)
	} else {
		aCount, dCount = countAddresses(mdb)
		if aCount != 4 || dCount != 2 {
			t.Errorf("dave@slip.com: expected 4 addr, 2 domain, got %d, %d",
				aCount, dCount)
		}
	}

	// lookup a bogus address at legit domain.
	a, err = mdb.LookupAddress("foo@goof.com")
	if err != nil && err != ErrMdbAddressNotFound {
		t.Errorf("lookup of foo@goof.com failed unexpectedly: %s", err)
	}

	// now look up a legit...
	a, err = mdb.LookupAddress("dmr")
	if err != nil {
		t.Errorf("lookup of dmr failed: %s", err)
	}
	if a.dump() != "id:1, localpart: dmr, domain id: <NULL>, dname: <empty>, transport: <NULL>, rclass: <NULL>, access: <NULL>" {
		t.Errorf("dmr: bad dump(), %s", a.dump())
	}

	// now delete it and check. We should have 3 addresses and 2 domains
	if err = mdb.DeleteAddress("dmr"); err != nil {
		t.Errorf("delete of dmr failed: %s", err)
	}
	aCount, dCount = countAddresses(mdb)
	if aCount != 3 || dCount != 2 {
		t.Errorf("delete of dmr: expected 3 addresses, 2 domains, got %d, %d", aCount, dCount)
	}

	a, err = mdb.LookupAddress("mary@goof.com")
	if err != nil {
		t.Errorf("lookup of mary@goof.com failed: %s", err)
	}
	if a.dump() != "id:2, localpart: mary, domain id: 1, dname: goof.com, transport: <NULL>, rclass: <NULL>, access: <NULL>" {
		t.Errorf("mary@goof.com: bad dump(), %s", a.dump())
	}

	// now delete it and check. We should have 2 addresses and 2 domains
	if err = mdb.DeleteAddress("mary@goof.com"); err != nil {
		t.Errorf("delete of mary@goof.com failed: %s", err)
	}
	aCount, dCount = countAddresses(mdb)
	if aCount != 2 || dCount != 2 {
		t.Errorf("delete of mary@goof.com: expected 2 addresses, 2 domains, got %d, %d", aCount, dCount)
	}

	// delete dave@slip.com and see if his domain also gets deleted
	if err = mdb.DeleteAddress("dave@slip.com"); err != nil {
		t.Errorf("delete of dave@slip.com failed: %s", err)
	}
	aCount, dCount = countAddresses(mdb)
	if aCount != 1 || dCount != 1 {
		t.Errorf("delete of dave@slip.com: expected 1 address, 1 domain, got %d, %d", aCount, dCount)
	}

	// delete a bogus address in a legit domain. We should see an error
	if err = mdb.DeleteAddress("foo@goof.com"); err != nil {
		if err != ErrMdbAddressNotFound {
			t.Errorf("delete of foo@goof.com failed: %s", err)
		}
	} else {
		t.Errorf("delete of foo@goof.com should have failed")
	}
	aCount, dCount = countAddresses(mdb)
	if aCount != 1 || dCount != 1 {
		t.Errorf("delete of foo@goof.com: expected 1 address, 1 domain, got %d, %d", aCount, dCount)
	}

	// delete a bogus address in a bogus domain
	if err = mdb.DeleteAddress("foo@baz"); err != nil {
		if err != ErrMdbAddressNotFound {
			t.Errorf("delete of foo@baz failed: %s", err)
		}
	} else {
		t.Errorf("delete of foo@baz should have failed")
	}
	aCount, dCount = countAddresses(mdb)
	if aCount != 1 || dCount != 1 {
		t.Errorf("delete of foo@baz: expected 1 address, 1 domain, got %d, %d", aCount, dCount)
	}

	// now delete bill@goof.com. That should be it. no more rows
	if err = mdb.DeleteAddress("bill@goof.com"); err != nil {
		t.Errorf("delete of bill@goof.com failed: %s", err)
	}
	aCount, dCount = countAddresses(mdb)
	if aCount != 0 || dCount != 0 {
		t.Errorf("delete of bill@goof.com: expected 0 addresses, 0 domains, got %d, %d", aCount, dCount)
	}
}

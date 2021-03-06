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
	"testing"

	_ "github.com/mattn/go-sqlite3" // do I really need this here?
)

// TestAddress
func TestAddress(t *testing.T) {
	var (
		err            error
		mdb            *MailDB
		dir            string
		aCount, dCount int
	)

	fmt.Printf("Address Test\n")

	dir, err = ioutil.TempDir("", "TestDBLoad-*")
	defer os.RemoveAll(dir)
	mdb, err = makeTestDB(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Errorf("Database load failed, %s", err)
		return
	}
	defer mdb.Close()

	// First add an access and transport so we can see if we can set them
	mdb.Begin()
	if _, err = mdb.InsertAccess("spam", "gooberfilter"); err != nil {
		t.Errorf("Insert spam entry unexpectedly failed, %s", err)
	}
	if tr, err := mdb.InsertTransport("relay"); err != nil {
		t.Errorf("Insert relay unexpectedly failed, %s", err)
	} else {
		if err = tr.SetTransport("relay"); err != nil {
			t.Errorf("Insert relay: set transport relay failed, %s", err)
		} else if err = tr.SetNexthop("localhost:56"); err != nil {
			t.Errorf("Insert relay: set nexthop localhost:56 failed, %s", err)
		}
	}
	mdb.End(&err)

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
		if a.Address() != "dmr" {
			t.Errorf("dmr: bad Address(), %s", a.Address())
		}
		if a.Transport() != "--" {
			t.Errorf("dmr: expected transport --, got %s", a.Transport())
		}
		if a.Rclass() != "--" {
			t.Errorf("dmr: expected rclass --, got %s", a.Rclass())
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
		if a.Address() != "mary@goof.com" {
			t.Errorf("mary@goof.com: bad Address(), %s", a.Address())
		}
		if a.Transport() != "--" {
			t.Errorf("mary@goof.com: expected transport --, got %s", a.Transport())
		}
		if a.Rclass() != "--" {
			t.Errorf("mary@goof.com: expected rclass --, got %s", a.Rclass())
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
		if a.Address() != "bill@goof.com" {
			t.Errorf("bill@goof.com: bad Address(), %s", a.Address())
		}
		if a.Transport() != "--" {
			t.Errorf("bill@goof.com: expected transport --, got %s", a.Transport())
		}
		if a.Rclass() != "--" {
			t.Errorf("bill@goof.com: expected rclass --, got %s", a.Rclass())
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
	if a.Address() != "dmr" {
		t.Errorf("dmr: bad Address(), %s", a.Address())
	}
	if a.Transport() != "--" {
		t.Errorf("dmr: expected transport --, got %s", a.Transport())
	}
	if a.Rclass() != "--" {
		t.Errorf("dmr: expected rclass --, got %s", a.Rclass())
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
	if a.Address() != "mary@goof.com" {
		t.Errorf("mary@goof.com: bad Address(), %s", a.Address())
	}
	if a.Transport() != "--" {
		t.Errorf("mary@goof.com: expected transport --, got %s", a.Transport())
	}
	if a.Rclass() != "--" {
		t.Errorf("mary@goof.com: expected rclass --, got %s", a.Rclass())
	}

	// Set and clear Rclass and transport for poor mary
	mdb.Begin()
	a, err = mdb.GetAddress("mary@goof.com")
	if err != nil {
		t.Errorf("Get mary@goof.com: unexpected error %s", err)
		mdb.End(&err)
	} else {
		if err = a.SetTransport("relay"); err != nil {
			t.Errorf("mary@goof.com: SetTransport relay, %s", err)
		}
		if err = a.SetRclass("spam"); err != nil {
			t.Errorf("mary@goof.com: SetRclass spam, %s", err)
		}
	}
	mdb.End(&err)

	var exportLine string = "mary@goof.com rclass=spam, transport=relay"
	a, err = mdb.LookupAddress("mary@goof.com")
	if err != nil {
		t.Errorf("lookup of mary@goof.com after sets failed: %s", err)
	}
	if a.Transport() != "relay" {
		t.Errorf("mary@goof.com: expected transport relay, got %s", a.Transport())
	}
	if a.Rclass() != "spam" {
		t.Errorf("mary@goof.com: expected rclass spam, got %s", a.Rclass())
	}
	if a.Export() != exportLine {
		t.Errorf("export mary@goof.com: expected %s, got %s", exportLine, a.Export())
	}

	// Now clear them
	mdb.Begin()
	a, err = mdb.GetAddress("mary@goof.com")
	if err != nil {
		t.Errorf("Get mary@goof.com to clear: unexpected error %s", err)
		mdb.End(&err)
	} else {
		if err = a.ClearTransport(); err != nil {
			t.Errorf("mary@goof.com: ClearTransport, %s", err)
		}
		if err = a.ClearRclass(); err != nil {
			t.Errorf("mary@goof.com: ClearRclass, %s", err)
		}
	}
	mdb.End(&err)

	a, err = mdb.LookupAddress("mary@goof.com")
	if err != nil {
		t.Errorf("lookup of mary@goof.com failed: %s", err)
	}
	if a.Transport() != "--" {
		t.Errorf("mary@goof.com: expected transport --, got %s", a.Transport())
	}
	if a.Rclass() != "--" {
		t.Errorf("mary@goof.com: expected rclass --, got %s", a.Rclass())
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

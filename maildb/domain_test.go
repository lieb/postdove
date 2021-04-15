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

// TestDomain
func TestDomain(t *testing.T) {
	var (
		err error
		mdb *MailDB
		dir string
		d   *Domain
	)

	fmt.Printf("Domain Test\n")

	dir, err = ioutil.TempDir("", "TestDBLoad-*")
	defer os.RemoveAll(dir)
	mdb, err = makeTestDB(filepath.Join(dir, "test.db"), "../schema.sql")
	if err != nil {
		t.Errorf("Database load failed, %s", err)
		return
	}
	defer mdb.Close()

	// Try to insert a domain without a transaction
	d, err = mdb.InsertDomain("foo", "")
	if err == nil {
		t.Errorf("Insert with no transaction did not fail")
		return
	} else if err != ErrMdbTransaction {
		t.Errorf("Expected error: Not in a transaction, got %s", err)
		return
	}

	// Try to insert a domain
	mdb.Begin()
	d, err = mdb.InsertDomain("foo", "")
	mdb.End(err == nil)
	if err != nil {
		t.Errorf("Insert foo: %s", err)
		return // no need to go further this early
	} else {
		if d.String() != "foo" {
			t.Errorf("Insert foo: bad String(), %s", d.String())
		}
		// NOTE: this will fail if you change schema defaults...
		if d.dump() != "id=1, name=foo, class=internet, transport=<NULL>, access=<NULL>, vuid=<NULL>, vgid=<NULL>, rclass=DEFAULT." {
			t.Errorf("Insert foo: bad dump(), %s", d.dump())
		}
		if !d.IsInternet() {
			t.Errorf("IsInternet should be true")
		} else if d.IsLocal() {
			t.Errorf("IsLocal should be false")
		} else if d.IsRelay() {
			t.Errorf("IsRelay should be false")
		} else if d.IsVirtual() {
			t.Errorf("IsVirtual should be false")
		} else if d.IsVmailbox() {
			t.Errorf("IsVmailbox should be false")
		}
	}

	// Try some bad args...
	mdb.Begin()
	d, err = mdb.InsertDomain("", "")
	mdb.End(err == nil)
	if err == nil {
		t.Errorf("Insert \"\" should have failed")
	} else if err != ErrMdbBadName {
		t.Errorf("Insert of \"\": %s", err)
	}
	mdb.Begin()
	d, err = mdb.InsertDomain(";bogus", "")
	mdb.End(err == nil)
	if err == nil {
		t.Errorf("Insert \";bogus\" should have failed")
	} else if err != ErrMdbBadName {
		t.Errorf("Insert of \";bogus\": %s", err)
	}

	mdb.Begin()
	d, err = mdb.InsertDomain("baz", "jazz")
	mdb.End(err == nil)
	if err == nil {
		t.Errorf("Insert \"jazz\" should have failed")
	} else if err != ErrMdbBadClass {
		t.Errorf("Insert of \"jazz\": %s", err)
	}

	// Lookup should agree with Insert...
	d, err = mdb.LookupDomain("foo")
	if err != nil {
		t.Errorf("Lookup foo: %s", err)
	} else {
		if d.String() != "foo" {
			t.Errorf("Lookup foo: bad String(), %s", d.String())
		}
		// NOTE: this will fail if you change schema defaults...
		if d.dump() != "id=1, name=foo, class=internet, transport=<NULL>, access=<NULL>, vuid=<NULL>, vgid=<NULL>, rclass=DEFAULT." {
			t.Errorf("Lookup foo: bad dump(), %s", d.dump())
		}
	}

	// Set some of the fields, first get the domain for transactions
	mdb.Begin()
	if d, err = mdb.GetDomain("foo"); err != nil {
		t.Errorf("Get foo: %s", err)
	} else {
		if err = d.SetVUid(53); err != nil {
			t.Errorf("SetVUid foo, %s", err)
		}
		if err = d.SetVGid(42); err != nil {
			t.Errorf("SetVGid foo, %s", err)
		}
		if err = d.SetRclass("spam"); err != nil {
			t.Errorf("SetRclassid foo, %s", err)
		}
		mdb.End(err == nil)
		// now check it
		if dn, err := mdb.LookupDomain("foo"); err != nil {
			t.Errorf("Lookup foo after sets, %s", err)
		} else {
			if d.dump() != dn.dump() {
				t.Errorf("Lookup to set mismatch: d=%s, dn=%s", d.dump(), dn.dump())
			}
			if dn.dump() != "id=1, name=foo, class=internet, transport=<NULL>, access=<NULL>, vuid=53, vgid=42, rclass=spam." {
				t.Errorf("domain not expected after transactions, %s", d.dump())
			}
		}
		if d.Class() != "internet" {
			t.Errorf("Domain.Class(): expected \"internet\", got %s", d.Class())
		}
		if d.Transport() != "--" {
			t.Errorf("Domain.Transport(): expected --, got %s", d.Transport())
		}
		if d.Access() != "--" {
			t.Errorf("Domain.Access(): expected --, got %s", d.Access())
		}
		if d.Vuid() != "53" {
			t.Errorf("Domain.Vuid(): expected --, got %s", d.Vuid())
		}
		if d.Vgid() != "42" {
			t.Errorf("Domain.Vgid(): expected --, got %s", d.Vgid())
		}
		if d.Rclass() != "spam" {
			t.Errorf("Domain.Rclass(): expected --, got %s", d.Rclass())
		}

	}
	// Lookup something not there
	d, err = mdb.LookupDomain("baz")
	if err == nil {
		t.Errorf("Lookup baz should have failed: got %s", d.dump())
	} else if err != ErrMdbDomainNotFound {
		t.Errorf("Lookup baz: %s", err)
	}

	// Delete stuff
	err = mdb.DeleteDomain("baz")
	if err == nil {
		t.Errorf("Delete baz should have failed")
	}
	err = mdb.DeleteDomain("foo")
	if err != nil {
		t.Errorf("Delete foo: %s", err)
	}
}

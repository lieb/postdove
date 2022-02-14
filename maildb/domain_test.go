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
	mdb, err = makeTestDB(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Errorf("Database load failed, %s", err)
		return
	}
	defer mdb.Close()

	// Try to insert a domain without a transaction
	d, err = mdb.InsertDomain("foo")
	if err == nil {
		t.Errorf("Insert with no transaction did not fail")
		return
	} else if err != ErrMdbTransaction {
		t.Errorf("Expected error: Not in a transaction, got %s", err)
		return
	}

	// Try to insert a domain
	mdb.Begin()
	d, err = mdb.InsertDomain("foo")
	mdb.End(&err)
	if err != nil {
		t.Errorf("Insert foo: %s", err)
		return // no need to go further this early
	} else {
		if d.Id() != 1 {
			t.Errorf("Insert foo: foolishly expected row id 1, got %d", d.Id())
		}
		if d.Name() != "foo" {
			t.Errorf("Insert foo: expected name 'foo', got %s", d.Name())
		}
		if d.Class() != "internet" {
			t.Errorf("Insert foo: expected class 'internet', got %s", d.Class())
		}
		if d.Transport() != "--" {
			t.Errorf("Insert foo: expected transport '--', got %s", d.Transport())
		}
		if d.Rclass() != "--" {
			t.Errorf("Insert foo: expected rclass '--', got %s", d.Rclass())
		}
		if d.Vuid() != "--" {
			t.Errorf("Insert foo expected vuid '--', got %s", d.Vuid())
		}
		if d.Vgid() != "--" {
			t.Errorf("Insert foo expected vgid '--', got %s", d.Vgid())
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
	d, err = mdb.InsertDomain("")
	mdb.End(&err)
	if err == nil {
		t.Errorf("Insert \"\" should have failed")
	} else if err != ErrMdbBadName {
		t.Errorf("Insert of \"\": %s", err)
	}
	mdb.Begin()
	d, err = mdb.InsertDomain(";bogus")
	mdb.End(&err)
	if err == nil {
		t.Errorf("Insert \";bogus\" should have failed")
	} else if err != ErrMdbBadName {
		t.Errorf("Insert of \";bogus\": %s", err)
	}

	mdb.Begin()
	d, err = mdb.InsertDomain("baz")
	if err == nil {
		err = d.SetClass("jazz")
	}
	mdb.End(&err)
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
		if d.Id() != 1 {
			t.Errorf("Insert foo: foolishly expected row id 1, got %d", d.Id())
		}
		if d.Name() != "foo" {
			t.Errorf("Insert foo: expected name 'foo', got %s", d.Name())
		}
		if d.Class() != "internet" {
			t.Errorf("Insert foo: expected class 'internet', got %s", d.Class())
		}
		if d.Transport() != "--" {
			t.Errorf("Insert foo: expected transport '--', got %s", d.Transport())
		}
		if d.Rclass() != "--" {
			t.Errorf("Insert foo: expected rclass '--', got %s", d.Rclass())
		}
		if d.Vuid() != "--" {
			t.Errorf("Insert foo expected vuid '--', got %s", d.Vuid())
		}
		if d.Vgid() != "--" {
			t.Errorf("Insert foo expected vgid '--', got %s", d.Vgid())
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

	// First add an access and transport so we can see if we can set them
	// get the domain for transactions, set some of the fields, and check
	mdb.Begin()
	if _, err = mdb.InsertAccess("spam", "gooberfilter"); err != nil {
		t.Errorf("Insert spam entry unexpectedly failed, %s", err)
	}
	if _, err = mdb.InsertTransport("relay", "relay", "localhost:56"); err != nil {
		t.Errorf("Insert relay unexpectedly failed, %s", err)
	}
	mdb.End(&err)
	// new transaction
	mdb.Begin()
	d, err = mdb.GetDomain("foo")
	if err != nil {
		mdb.End(&err)
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
		if err = d.SetTransport("relay"); err != nil {
			t.Errorf("SetTransport foo, %s", err)
		}
		mdb.End(&err)
		// now check it
		if dn, err := mdb.LookupDomain("foo"); err != nil {
			t.Errorf("Lookup foo after sets, %s", err)
		} else {
			if dn.Class() != "internet" {
				t.Errorf("Domain.Class(): expected \"internet\", got %s", dn.Class())
			}
			if dn.Transport() != "relay" {
				t.Errorf("Domain.Transport(): expected relay, got %s", dn.Transport())
			}
			if dn.Vuid() != "53" {
				t.Errorf("Domain.Vuid(): expected 42, got %s", dn.Vuid())
			}
			if dn.Vgid() != "42" {
				t.Errorf("Domain.Vgid(): expected 42, got %s", dn.Vgid())
			}
			if dn.Rclass() != "spam" {
				t.Errorf("Domain.Rclass(): expected spam, got %s", dn.Rclass())
			}
		}
	}
	// Lookup something not there
	d, err = mdb.LookupDomain("baz")
	if err == nil {
		t.Errorf("Lookup baz should have failed")
	} else if err != ErrMdbDomainNotFound {
		t.Errorf("Lookup baz: %s", err)
	}

	// Add some more FQDN domains
	domainList := []string{
		"buz.com",
		"buz.org",
		"buz.net",
		"bar.com",
		"bar.org",
		"bar.net",
		"zip.bar.net",
		"tie.bar.net",
	}
	dlists := map[string][]string{
		"*": []string{
			"bar.com",
			"bar.net",
			"bar.org",
			"buz.com",
			"buz.net",
			"buz.org",
			"foo",
			"tie.bar.net",
			"zip.bar.net"},
		"*.org": []string{
			"bar.org",
			"buz.org"},
		"*.bar.*": []string{
			"tie.bar.net",
			"zip.bar.net"},
		"*bar*": []string{
			"bar.com",
			"bar.net",
			"bar.org",
			"tie.bar.net",
			"zip.bar.net"},
	}
	mdb.Begin()
	for _, dom := range domainList {
		if d, err = mdb.InsertDomain(dom); err != nil {
			t.Errorf("Insert domains: unexpected error, %s", err)
			break
		}
	}
	mdb.End(&err)
	if err == nil { // only check if inserts passed
		for q, l := range dlists {
			dl, err := mdb.FindDomain(q)
			if err != nil {
				t.Errorf("Find domain \"%s\": Unexpected error, %s", q, err)
			} else if len(dl) != len(l) {
				t.Errorf("Find domain \"%s\": Should have found %d, found %d", q, len(l), len(dl))
			} else {
				for i, d := range dl {
					if d.name != l[i] {
						t.Errorf("Domain list for \"%s\", expected (%s), got (%s)", q, l[i], d.name)
					}
				}
			}
		}
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

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

// TestMailbox
func TestMailbox(t *testing.T) {
	var (
		err error
		mdb *MailDB
		d   *Domain
		dir string
		mb  *VMailbox
	)

	fmt.Printf("Mailbox Test\n")

	dir, err = ioutil.TempDir("", "TestDBLoad-*")
	defer os.RemoveAll(dir)
	mdb, err = makeTestDB(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Errorf("Database load failed, %s", err)
		return
	}
	defer mdb.Close()

	// Test creating a new mailbox with all of the empties and defaults
	// we first need a real domain so create a few because
	// we need a domain before we can add mailboxes
	mdb.Begin()
	d, err = mdb.InsertDomain("skywalker")
	if err == nil {
		err = d.SetClass("vmailbox")
	}
	mdb.End(&err)
	if err != nil {
		t.Errorf("Insert of skywalker failed, %s", err)
		return
	}

	mdb.Begin()
	d, err = mdb.InsertDomain("nowhere")
	if err != nil {
		err = d.SetClass("relay") // fodder for busted mailboxes
	}
	mdb.End(&err)
	if err != nil {
		t.Errorf("Insert of nowhere failed, %s", err)
		return
	}

	// See if we can create a mailbox in nowhere
	mdb.Begin()
	mb, err = mdb.InsertVMailbox("lost@nowhere")
	mdb.End(&err)
	if err == nil {
		t.Errorf("Add of lost@nowhere should have failed")
	} else if err != ErrMdbMboxNotMboxDomain {
		t.Errorf("Add of lost@nowhere, %s", err)
	}
	// see if we can add a user
	mdb.Begin()
	mb, err = mdb.InsertVMailbox("luke@skywalker")
	mdb.End(&err)
	if err != nil {
		t.Errorf("luke@skywalker: %s", err)
		return // no sense continuing if we can do this...
	}
	// NOTE: this will change with schema default changes
	if mb.String() != "luke@skywalker:{PLAIN}*:::*:bytes=300M::true" {
		t.Errorf("Luke@skywalker: expected \"luke@skywalker:{PLAIN}*:::*:bytes=300M::true\", got %s", mb.String())
	}

	// Now try and get it back
	mb, err = mdb.LookupVMailbox("luke@skywalker")
	if err != nil {
		t.Errorf("Lookup luke@skywalker, %s", err)
	} else {
		if mb.String() != "luke@skywalker:{PLAIN}*:::*:bytes=300M::true" {
			t.Errorf("Luke@skywalker: expected \"luke@skywalker:{PLAIN}*:::*:bytes=300M::true\", got %s",
				mb.String())
		}
		if !mb.IsEnabled() {
			t.Errorf("Mailbox should start out as enabled")
		}
	}

	// Play with it
	mdb.Begin()
	if mb, err = mdb.GetVMailbox("luke@skywalker"); err != nil {
		t.Errorf("Get luke@skywalker failed, %s", err)
	} else {
		if err = mb.Disable(); err != nil {
			t.Errorf("Disable luke@skywalker failed, %s", err)
		} else {
			if mb.IsEnabled() {
				t.Errorf("luke@skywalker should be disabled")
			}
		}
	}
	mdb.End(&err)
	mb, err = mdb.LookupVMailbox("luke@skywalker")
	if err != nil {
		t.Errorf("luke@skywalker after disable, %s", err)
	} else {
		if mb.IsEnabled() {
			t.Errorf("luke@skywalker should be disabled")
		}
	}
	mdb.Begin()
	if mb, err = mdb.GetVMailbox("luke@skywalker"); err != nil {
		t.Errorf("Get luke@skywalker failed, %s", err)
	} else {
		if err = mb.Enable(); err != nil {
			t.Errorf("Enable luke@skywalker failed, %s", err)
		} else {
			if !mb.IsEnabled() {
				t.Errorf("luke@skywalker should be enabled")
			}
		}
	}
	mdb.End(&err)
	mb, err = mdb.LookupVMailbox("luke@skywalker")
	if err != nil {
		t.Errorf("lookup luke@skywalker after enable, %s", err)
	} else {
		if !mb.IsEnabled() {
			t.Errorf("luke@skywalker should be enabled")
		}
	}

	// Change password
	mdb.Begin()
	if mb, err = mdb.GetVMailbox("luke@skywalker"); err != nil {
		t.Errorf("Get luke@skywalker failed, %s", err)
	} else {
		if err = mb.SetPassword("Not123456"); err != nil {
			t.Errorf("Set password for luke@skywalker failed, %s", err)
		}
	}
	mdb.End(&err)
	// See if it changes
	mb, err = mdb.LookupVMailbox("luke@skywalker")
	if err != nil {
		t.Errorf("Lookup luke@skywalker, %s", err)
	} else {
		if mb.String() != "luke@skywalker:{PLAIN}Not123456:::*:bytes=300M::true" {
			t.Errorf("Luke@skywalker: expected \"luke@skywalker:{PLAIN}Not123456:::*:bytes=300M::true\", got %s",
				mb.String())
		}
	}
	// Change password and type
	mdb.Begin()
	if mb, err = mdb.GetVMailbox("luke@skywalker"); err != nil {
		t.Errorf("Get luke@skywalker failed, %s", err)
	} else {
		if err = mb.SetPassword("Sn3@kyB1ts"); err != nil {
			t.Errorf("Set password for luke@skywalker failed, %s", err)
		} else {
			if err = mb.SetPwType("sha256"); err != nil {
				t.Errorf("Set password type failed, %s", err)
			}
		}
	}
	mdb.End(&err)
	// See if it changed
	mb, err = mdb.LookupVMailbox("luke@skywalker")
	if err != nil {
		t.Errorf("Lookup luke@skywalker, %s", err)
	} else {
		if mb.String() != "luke@skywalker:{SHA256}Sn3@kyB1ts:::*:bytes=300M::true" {
			t.Errorf("Luke@skywalker: expected \"luke@skywalker:{SHA256}Sn3@kyB1ts:::*:bytes=300M::true\", got %s",
				mb.String())
		}
	}

	// Remove it first with alias pointing to it
	luke := []string{"luke@skywalker"}
	if err = makeAlias(mdb, "rebel@skywalker", luke); err != nil {
		t.Errorf("Make rebel@skywalker, %s", err)
	}
	err = mdb.DeleteVMailbox("luke@skywalker")
	if err == nil {
		t.Errorf("First delete of luke@skywalker should have failed")
	} else {
		if err != ErrMdbMboxIsRecip {
			t.Errorf("Delete luke@skywalker, %s", err)
		}
	}
	if err = mdb.RemoveAlias("rebel@skywalker"); err != nil {
		t.Errorf("remove alias rebel@skywalker, %s", err)
	}
	if err = mdb.DeleteVMailbox("luke@skywalker"); err != nil {
		t.Errorf("Delete mbox luke@skywalker, %s", err)
	}

	// Delete a bogus
	err = mdb.DeleteVMailbox("yoda@skywalker")
	if err == nil {
		t.Errorf("Delete yoda@skywalker should have failed")
	} else if err != ErrMdbNotMbox {
		t.Errorf("Delete yoda@skywalker, %s", err)
	}

	// Clean up by deleting the domains too
	if err = mdb.DeleteDomain("nowhere"); err != nil {
		t.Errorf("Failed to remove nowhere, %s", err)
	}
	if err = mdb.DeleteDomain("skywalker"); err != nil {
		t.Errorf("Failed to remove skywalker, %s", err)
	}
}

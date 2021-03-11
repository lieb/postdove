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
		err     error
		mdb     *MailDB
		dir     string
		mb      *VMailbox
		mb_list []*VMailbox
	)

	fmt.Printf("Mailbox Test\n")

	dir, err = ioutil.TempDir("", "TestDBLoad-*")
	defer os.RemoveAll(dir)
	mdb, err = makeTestDB(filepath.Join(dir, "test.db"), "../schema.sql")
	if err != nil {
		t.Errorf("Database load failed, %s", err)
		return
	}
	defer mdb.Close()

	// Test creating a new mailbox with all of the empties and defaults
	// we first need a real domain so create one
	if err = mdb.begin(); err != nil {
		t.Errorf("Transaction begin failed: %s", err)
		return
	}
	_, err = mdb.InsertDomain("skywalker")
	mdb.end(err == nil)
	if err != nil {
		t.Errorf("Insert of skywalker failed, %s", err)
		return
	}

	// see if we can add a user
	mb, err = mdb.NewVmailbox("luke@skywalker", "", "", "", "", "", "")
	if err != nil {
		t.Errorf("luke@skywalker: %s", err)
		return // no sense continuing if we can do this...
	}
	// NOTE: this will change with schema default changes
	if mb.String() != "luke@skywalker:{PLAIN}*:::1000::true" {
		t.Errorf("Luke@skywalker: expected \"luke@skywalker:{PLAIN}*:::1000::true\", got %s", mb.String())
	}

	// Now try and get it back
	mb_list, err = mdb.LookupVMailbox("luke@skywalker")
	if err != nil {
		t.Errorf("Lookup luke@skywalker, %s", err)
	} else {
		if len(mb_list) != 1 {
			t.Errorf("Lookup luke@skywalker: expected 1 returned, got %d", len(mb_list))
		} else {
			if mb_list[0].String() != "luke@skywalker:{PLAIN}*:::1000::true" {
				t.Errorf("Luke@skywalker: expected \"luke@skywalker:{PLAIN}*:::1000::true\", got %s",
					mb_list[0].String())
			}
			if !mb_list[0].IsEnabled() {
				t.Errorf("Mailbox should start out as enabled")
			}
		}
	}

	// Play with it
	if err = mdb.DisableVMailbox("luke@skywalker"); err != nil {
		t.Errorf("Disable of luke@skywalker, %s", err)
	}
	mb_list, err = mdb.LookupVMailbox("luke@skywalker")
	if err != nil {
		t.Errorf("luke@skywalker after disable, %s", err)
	} else {
		if mb_list[0].IsEnabled() {
			t.Errorf("luke@skywalker should be disabled")
		}
	}
	if err = mdb.EnableVMailbox("luke@skywalker"); err != nil {
		t.Errorf("Enable of luke@skywalker, %s", err)
	}
	mb_list, err = mdb.LookupVMailbox("luke@skywalker")
	if err != nil {
		t.Errorf("luke@skywalker after enable, %s", err)
	} else {
		if !mb_list[0].IsEnabled() {
			t.Errorf("luke@skywalker should be enabled")
		}
	}

	// Change password
	if mdb.ChangePassword("luke@skywalker", "Not123456", ""); err != nil {
		t.Errorf("Change password luke@skywalker, %s", err)
	}
	// See if it changes
	mb_list, err = mdb.LookupVMailbox("luke@skywalker")
	if err != nil {
		t.Errorf("Lookup luke@skywalker, %s", err)
	} else {
		if len(mb_list) != 1 {
			t.Errorf("Lookup luke@skywalker: expected 1 returned, got %d", len(mb_list))
		} else {
			if mb_list[0].String() != "luke@skywalker:{PLAIN}Not123456:::1000::true" {
				t.Errorf("Luke@skywalker: expected \"luke@skywalker:{PLAIN}Not123456:::1000::true\", got %s",
					mb_list[0].String())
			}
			if !mb_list[0].IsEnabled() {
				t.Errorf("Mailbox should start out as enabled")
			}
		}
	}
	// Change password and type
	if mdb.ChangePassword("luke@skywalker", "Sn3@kyB1ts", "sha256"); err != nil {
		t.Errorf("Change password luke@skywalker, %s", err)
	}
	// See if it changed
	mb_list, err = mdb.LookupVMailbox("luke@skywalker")
	if err != nil {
		t.Errorf("Lookup luke@skywalker, %s", err)
	} else {
		if len(mb_list) != 1 {
			t.Errorf("Lookup luke@skywalker: expected 1 returned, got %d", len(mb_list))
		} else {
			if mb_list[0].String() != "luke@skywalker:{SHA256}Sn3@kyB1ts:::1000::true" {
				t.Errorf("Luke@skywalker: expected \"luke@skywalker:{SHA256}Sn3@kyB1ts:::1000::true\", got %s",
					mb_list[0].String())
			}
		}
	}

	// Remove it
	err = mdb.DeleteVMailbox("luke@skywalker")
	if err != nil {
		t.Errorf("Delete luke@skywalker, %s", err)
	}

	// Delete a bogus
	err = mdb.DeleteVMailbox("yoda@skywalker")
	if err == nil {
		t.Errorf("Delete yoda@skywalker should have failed")
	} else if err != ErrMdbAddressNotFound {
		t.Errorf("Delete yoda@skywalker, %s", err)
	}

}

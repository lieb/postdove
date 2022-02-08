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
	//"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3" // do I really need this here?
)

// TestAccess
func TestAccess(t *testing.T) {
	var (
		err error
		mdb *MailDB
		dir string
		a   *Access
		al  []*Access
	)

	fmt.Printf("Access Test\n")

	dir, err = ioutil.TempDir("", "TestDBLoad-*")
	defer os.RemoveAll(dir)
	mdb, err = makeTestDB(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Errorf("Database load failed, %s", err)
		return
	}
	defer mdb.Close()

	// Try to find access rules in an empty database
	al, err = mdb.FindAccess("*")
	if err == nil {
		t.Errorf("Find in empty database did not fail")
		return
	}

	// Try to insert an action without a transaction
	a, err = mdb.InsertAccess("permit", "permissive")
	if err == nil {
		t.Errorf("Insert with no transaction did not fail")
		return
	} else if err != ErrMdbTransaction {
		t.Errorf("Expected error: Not in a transaction, got %s", err)
		return
	}

	// Insert an access rule
	mdb.Begin()
	a, err = mdb.InsertAccess("permit", "permissive")
	mdb.End(&err)
	if err != nil {
		t.Errorf("Insert permit: %s", err)
		return
	}

	// See if it has the right action
	action := a.Action()
	if action != "permissive" {
		t.Errorf("Insert permit: did not set action, got %s", action)
	}

	// look it up
	a, err = mdb.LookupAccess("permit")
	if err != nil {
		t.Errorf("Lookup permit: %s", err)
	} else {
		if a.Name() != "permit" {
			t.Errorf("Lookup permit: bad Name(), got %s", a.Name())
		}
		if a.Action() != "permissive" {
			t.Errorf("Lookup permit: bad Action(), got %s", a.Action())
		}
	}
	// look up a bogus rule
	a, err = mdb.LookupAccess("bogus")
	if err == nil {
		t.Errorf("Lookup of bogus should have failed")
	}

	// get it without a transaction
	a, err = mdb.GetAccess("permit")
	if err == nil {
		t.Errorf("Get of permit with no transaction did not fail")
	} else if err != ErrMdbTransaction {
		t.Errorf("Expected error: Not in transaction, got %s", err)
	}

	// get it
	mdb.Begin()
	a, err = mdb.GetAccess("permit")
	if err != nil {
		t.Errorf("Get permit: %s", err)
		mdb.End(&err)
		return
	}
	// set the action
	err = a.SetAction("freedum")
	if err != nil {
		t.Errorf("Set freedum: %s", err)
	}
	mdb.End(&err)
	// try to set it again after closing transaction
	err = a.SetAction("lost_cause")
	if err == nil {
		t.Errorf("SetAction lost_cause should have failed")
	} else if err != ErrMdbTransaction {
		t.Errorf("SetAction lost_cause: expected not in transaction, got %s", err)
	}
	// check set to correct value
	a, err = mdb.LookupAccess("permit")
	if err != nil {
		t.Errorf("Lookup permit unexpected fail: %s", err)
	} else if a.Action() != "freedum" {
		t.Errorf("Expected action freedum, got %s", a.Action())
	}

	// load some more rules and then find them
	mdb.Begin()
	a, err = mdb.InsertAccess("reject", "OverTheSide")
	if err != nil {
		mdb.End(&err)
		t.Errorf("Insert reject: %s", err)
		return
	}
	a, err = mdb.InsertAccess("defer", "DeadLetter")
	if err != nil {
		mdb.End(&err)
		t.Errorf("Insert defer: %s", err)
		return
	}
	mdb.End(&err)
	if al, err = mdb.FindAccess("reject"); err != nil {
		t.Errorf("FindAccess reject, unexpected error, %s", err)
	} else if len(al) != 1 {
		t.Errorf("FindAccess reject expected 1 result, got %d", len(al))
	} else if al[0].Action() != "OverTheSide" {
		t.Errorf("FindAccess reject expected action OverTheSide, got %s", al[0].Action())
	}
	if al, err = mdb.FindAccess("bogus"); err != nil {
		if err != ErrMdbAccessNotFound {
			t.Errorf("FindAccess bogus, unexpected error, %s", err)
		}
	} else if len(al) != 0 {
		t.Errorf("FindAccess bogus, expected 0 entries, got %d", len(al))
	}
	if al, err = mdb.FindAccess("*"); err != nil {
		t.Errorf("FindAccess *, unexpected error, %s", err)
	} else if len(al) != 3 {
		t.Errorf("FindAccess *, expected 3 entries, got %d", len(al))
	} else {
		for _, a = range al {
			if (a.Name() == "permit" && a.Action() == "freedum") ||
				(a.Name() == "reject" && a.Action() == "OverTheSide") ||
				(a.Name() == "defer" && a.Action() == "DeadLetter") {
				continue
			} else {
				t.Errorf("FindAccess *, got unexpected %s or %s", a.Name(), a.Action())
			}
		}
	}
}

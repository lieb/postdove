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

// Test_Transport
func Test_Transport(t *testing.T) {
	var (
		err error
		mdb *MailDB
		dir string
		tr  *Transport
		tl  []*Transport
	)

	fmt.Printf("Transport Test\n")

	dir, err = ioutil.TempDir("", "TestDBLoad-*")
	defer os.RemoveAll(dir)
	mdb, err = makeTestDB(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Errorf("Database load failed, %s", err)
		return
	}
	defer mdb.Close()

	// Try to find transport in an empty database
	tl, err = mdb.FindTransport("*")
	if err == nil {
		t.Errorf("Find in empty database did not fail")
		return
	}

	// Try to insert an transport without a transaction
	tr, err = mdb.InsertTransport("dovecot", "lmtp", "localhost:24")
	if err == nil {
		t.Errorf("Insert with no transaction did not fail")
		return
	} else if err != ErrMdbTransaction {
		t.Errorf("Expected error: Not in a transaction, got %s", err)
		return
	}

	// Insert a transport
	mdb.Begin()
	tr, err = mdb.InsertTransport("dovecot", "lmtp", "localhost:24")
	mdb.End(&err)
	if err != nil {
		t.Errorf("Insert transport dovecot: %s", err)
		return
	}
	// See if it has the right fields
	trans := tr.Transport()
	if trans != "lmtp" {
		t.Errorf("Insert dovecot: did not set transport, got %s", trans)
	}
	hop := tr.Nexthop()
	if hop != "localhost:24" {
		t.Errorf("Insert dovecot: did not set nexthop, got %s", hop)
	}

	// now look it up
	tr, err = mdb.LookupTransport("dovecot")
	if err != nil {
		t.Errorf("Lookup dovecot: %s", err)
	} else {
		if tr.Name() != "dovecot" {
			t.Errorf("Lookup dovecot: bad Name(), got %s", tr.Name())
		}
		trans := tr.Transport()
		if trans != "lmtp" {
			t.Errorf("Lookup dovecot: did not set transport, got %s", trans)
		}
		hop := tr.Nexthop()
		if hop != "localhost:24" {
			t.Errorf("Lookup dovecot: did not set nexthop, got %s", hop)
		}

	}

	// look up a bogus transport
	tr, err = mdb.LookupTransport("nowhere")
	if err == nil {
		t.Errorf("Lookup of nowhere should have failed")
	}

	// get it without a transaction
	tr, err = mdb.GetTransport("dovecot")
	if err == nil {
		t.Errorf("Get of dovecot with no transaction did not fail")
	} else if err != ErrMdbTransaction {
		t.Errorf("Expected error: Not in transaction, got %s", err)
	}

	// get it
	mdb.Begin()
	tr, err = mdb.GetTransport("dovecot")
	if err != nil {
		t.Errorf("Get dovecot: %s", err)
		mdb.End(&err)
		return
	}
	// set transport and nexthop
	err = tr.SetTransport("virtual")
	if err != nil {
		t.Errorf("Set virtual: %s", err)
	}
	err = tr.SetNexthop("nomad")
	if err != nil {
		t.Errorf("Set nomad: %s", err)
	}
	mdb.End(&err)

	// Now check it
	tr, err = mdb.LookupTransport("dovecot")
	if err != nil {
		t.Errorf("Lookup dovecot: %s", err)
	} else {
		if tr.Name() != "dovecot" {
			t.Errorf("Lookup dovecot: bad Name(), got %s", tr.Name())
		}
		trans := tr.Transport()
		if trans != "virtual" {
			t.Errorf("Lookup dovecot: did not set transport, got %s", trans)
		}
		hop := tr.Nexthop()
		if hop != "nomad" {
			t.Errorf("Lookup dovecot: did not set nexthop, got %s", hop)
		}
	}

	// get it to test for null entry
	mdb.Begin()
	tr, err = mdb.GetTransport("dovecot")
	if err != nil {
		t.Errorf("Get dovecot: %s", err)
		mdb.End(&err)
		return
	}
	// set transport and nexthop
	err = tr.SetTransport("")
	if err != nil {
		t.Errorf("Set transport null: %s", err)
	}
	err = tr.SetNexthop("")
	if err != nil {
		t.Errorf("Set nexthop null: %s", err)
	}
	mdb.End(&err)

	// Now check it
	tr, err = mdb.LookupTransport("dovecot")
	if err != nil {
		t.Errorf("Lookup dovecot: %s", err)
	} else {
		trans := tr.Transport()
		if trans != "" {
			t.Errorf("Lookup dovecot: did not set transport to null, got %s", trans)
		}
		hop := tr.Nexthop()
		if hop != "" {
			t.Errorf("Lookup dovecot: did not set nexthop to null, got %s", hop)
		}
	}

	// load some more transports and then find them
	mdb.Begin()
	tr, err = mdb.InsertTransport("spam", "", "/dev/null")
	if err != nil {
		mdb.End(&err)
		t.Errorf("Insert spam: %s", err)
		return
	}
	tr, err = mdb.InsertTransport("local", "mailbox", "")
	if err != nil {
		mdb.End(&err)
		t.Errorf("Insert local: %s", err)
		return
	}
	mdb.End(&err)
	if tl, err = mdb.FindTransport("dovecot"); err != nil {
		t.Errorf("FindTransport dovecot, unexpected error, %s", err)
	} else if len(tl) != 1 {
		t.Errorf("FindTransport dovecot expected 1 result, got %d", len(tl))
	} else if tl[0].Transport() != "" {
		t.Errorf("FindTransport dovecot expected transport \"\", got %s", tl[0].Transport())
	} else if tl[0].Nexthop() != "" {
		t.Errorf("FindTransport dovecot expected nexthop \"\", got %s", tl[0].Nexthop())
	}
	if tl, err = mdb.FindTransport("bogus"); err != nil {
		if err != ErrMdbTransNotFound {
			t.Errorf("FindTransport bogus, unexpected error, %s", err)
		}
	} else if len(tl) != 0 {
		t.Errorf("FindTransport bogus, expected 0 entries, got %d", len(tl))
	}
	if tl, err = mdb.FindTransport("*"); err != nil {
		t.Errorf("FindTransport *, unexpected error, %s", err)
	} else if len(tl) != 3 {
		t.Errorf("FindTransport *, expected 3 entries, got %d", len(tl))
	} else {
		for _, tr = range tl {
			if (tr.Name() == "dovecot" && tr.Transport() == "" && tr.Nexthop() == "") ||
				(tr.Name() == "spam" && tr.Transport() == "" && tr.Nexthop() == "/dev/null") ||
				(tr.Name() == "local" && tr.Transport() == "mailbox" && tr.Nexthop() == "") {
				continue
			} else {
				t.Errorf("FindTransport *, got unexpected %s or %s or %s",
					tr.Name(), tr.Transport(), tr.Nexthop())
			}
		}
	}
}

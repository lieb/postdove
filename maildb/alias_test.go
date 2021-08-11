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

// makeAlias
func makeAlias(mdb *MailDB, alias string, recipients []string) error {
	var (
		err       error
		aliasAddr *Address
	)

	if len(recipients) < 1 {
		return ErrMdbNoRecipients
	}
	// Enter a transaction for everything else
	mdb.Begin()
	defer mdb.End(&err)

	if aliasAddr, err = mdb.GetOrInsAddress(alias); err != nil {
		return err
	}

	// We now have the alias address part, either brand new or an existing
	// Now cycle through the recipient list and stuff them in
	for _, r := range recipients {
		if err = aliasAddr.AttachAlias(r); err != nil {
			break
		}
	}
	return err
}

// TestAliasOps
func TestAliasOps(t *testing.T) {
	var (
		err            error
		mdb            *MailDB
		dir            string
		al             *Alias
		al_list        []*Alias
		recips         []string
		aCount, dCount int
	)

	fmt.Printf("Alias ops Test\n")

	dir, err = ioutil.TempDir("", "TestDBLoad-*")
	defer os.RemoveAll(dir)
	mdb, err = makeTestDB(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Errorf("Database load failed, %s", err)
		return
	}
	defer mdb.Close()

	// Test simple MakeAlias
	recips = []string{"rednose@clown.com"}
	if err = makeAlias(mdb, "bozo@clown.com", recips); err != nil {
		t.Errorf("MakeAlias: bozo@clown.com: %s", err)
	}
	aCount, dCount = countAddresses(mdb)
	if aCount != 2 || dCount != 1 {
		t.Errorf("bozo@clown.com: expected 2 addr, 2 domain, got %d, %d",
			aCount, dCount)
	}
	// Now look it up
	if al_list, err = mdb.LookupAlias("bozo@clown.com"); err != nil {
		t.Errorf("lookup bozo@clown.com: %s", err)
	}
	if len(al_list) != 1 {
		t.Errorf("LookupAlias: bozo@clown.com should be 1 alias, got %d", len(al_list))
	} else {
		al = al_list[0]
		if al.String() != "bozo@clown.com rednose@clown.com" {
			t.Errorf("bozo@clown.com: expected \"bozo@clown.com rednose@clown.com\", got %s\n",
				al.String())
		}
	}

	// Test /etc/aliases type
	recips = []string{"\"| cat > /dev/null\""}
	if err = makeAlias(mdb, "rebar", recips); err != nil {
		t.Errorf("MakeAlias: rebar: %s", err)
	}
	aCount, dCount = countAddresses(mdb)
	if aCount != 3 || dCount != 1 {
		t.Errorf("rebar: expected 3 addr, 1 domain, got %d, %d",
			aCount, dCount)
	}
	// Now look it up
	if al_list, err = mdb.LookupAlias("rebar"); err != nil {
		t.Errorf("lookup rebar: %s", err)
	}
	if len(al_list) != 1 {
		t.Errorf("LookupAlias: rebar should be 1 alias, got %d", len(al_list))
	} else {
		al = al_list[0]
		if al.String() != "rebar: \"| cat > /dev/null\"" {
			t.Errorf("rebar: expected 'rebar: \"| cat > /dev/null\"' got %s\n",
				al.String())
		}
	}

	// Add another to bozo@clown
	recips = []string{"micky@clown.com"}
	if err = makeAlias(mdb, "bozo@clown.com", recips); err != nil {
		t.Errorf("MakeAlias: add micky to bozo@clown.com: %s", err)
	}
	aCount, dCount = countAddresses(mdb)
	if aCount != 4 || dCount != 1 {
		t.Errorf("bozo@clown.com: after add expected 4 addr, 1 domain, got %d, %d",
			aCount, dCount)
	}
	// Now look it up
	if al_list, err = mdb.LookupAlias("bozo@clown.com"); err != nil {
		t.Errorf("lookup bozo@clown.com: %s", err)
	}
	if len(al_list) != 1 {
		t.Errorf("LookupAlias: bozo@clown.com should be 1 alias, got %d", len(al_list))
	} else {
		al = al_list[0]
		if al.String() != "bozo@clown.com rednose@clown.com, micky@clown.com" {
			t.Errorf("bozo@clown.com: expected \"bozo@clown.com rednose@clown.com, micky@clown.com\", got %s",
				al.String())
		}
	}

	// Add another to rebar
	recips = []string{"/tmp/rubbish"}
	if err = makeAlias(mdb, "rebar", recips); err != nil {
		t.Errorf("MakeAlias: rebar: %s", err)
	}
	aCount, dCount = countAddresses(mdb)
	if aCount != 4 || dCount != 1 {
		t.Errorf("rebar: expected 4 addr, 1 domain, got %d, %d",
			aCount, dCount)
	}
	// Now look it up
	if al_list, err = mdb.LookupAlias("rebar"); err != nil {
		t.Errorf("lookup rebar: %s", err)
	}
	if len(al_list) != 1 {
		t.Errorf("LookupAlias: rebar should be 1 alias, got %d", len(al_list))
	} else {
		al = al_list[0]
		if al.String() != "rebar: \"| cat > /dev/null\", /tmp/rubbish" {
			t.Errorf("rebar: expected \"rebar: \"| cat > /dev/null\", /tmp/rubbish\" got %s",
				al.String())
		}
	}

	// Test a virtual type with pipe for failure
	recips = nil
	recips = []string{"/drain.txt"}
	err = makeAlias(mdb, "pipe@plumbing", recips)
	if err != nil {
		if err != ErrMdbAddressTarget {
			t.Errorf("MakeAlias: pipe@plumbing: %s", err)
		}
	} else {
		t.Errorf("MakeAlias: inserted pipe@plumbing without errors")
	}
	aCount, dCount = countAddresses(mdb)
	if aCount != 4 || dCount != 1 {
		t.Errorf("pipe@plumbing: expected 4 addr, 1 domain, got %d, %d",
			aCount, dCount)
	}
	al_list, err = mdb.LookupAlias("pipe@plumbing")
	if err == nil {
		t.Errorf("Lookup of pipe@plumbing should have failed")
	} else {
		if err != ErrMdbDomainNotFound {
			t.Errorf("Lookup of pipe@plumbing: unexpected error %s", err)
		}
	}

	// multiple recips in same call
	recips = []string{"dave@work", "dave@home", "dave@mobile"}
	if err = makeAlias(mdb, "miller@office", recips); err != nil {
		t.Errorf("MakeAlias: miller@office: %s", err)
	}
	aCount, dCount = countAddresses(mdb)
	if aCount != 8 || dCount != 5 {
		t.Errorf("miller@office: expected 8 addr, 5 domains, got %d, %d",
			aCount, dCount)
	}
	if al_list, err = mdb.LookupAlias("miller@office"); err != nil {
		t.Errorf("lookup miller@office: %s", err)
	}
	if len(al_list) != 1 {
		t.Errorf("LookupAlias: miller@office should be 1 alias, got %d", len(al_list))
	} else {
		al = al_list[0]
		if al.String() != "miller@office dave@work, dave@home, dave@mobile" {
			t.Errorf("miller@office: expected \"miller@office dave@work, dave@home, dave@mobile\", got %s",
				al.String())
		}
	}

	// Now try wildcards. First add in some more aliases
	recips = []string{"bill@plumbers.com", "mike@shovel.org"}
	if err = makeAlias(mdb, "steve@office", recips); err != nil {
		t.Errorf("MakeAlias: steve@office: %s", err)
	}
	recips = []string{"willy@circus", "grocho@marx", "chico@marx"}
	if err = makeAlias(mdb, "steve@clown.com", recips); err != nil {
		t.Errorf("MakeAlias: steve@clown.com: %s", err)
	}
	recips = []string{"root", "daemon@server", "postmaster@usps.gov"}
	if err = makeAlias(mdb, "postfix", recips); err != nil {
		t.Errorf("MakeAlias: postfix: %s", err)
	}

	// *@office
	if al_list, err = mdb.LookupAlias("*@office"); err != nil {
		t.Errorf("lookup *@office: %s", err)
	}
	res := []string{
		"miller@office dave@work, dave@home, dave@mobile",
		"steve@office bill@plumbers.com, mike@shovel.org",
	}
	if len(al_list) != 2 {
		t.Errorf("LookupAlias: *@office should be 2 aliases, got %d", len(al_list))
	} else {
		for i, a := range al_list {
			if a.String() != res[i] {
				t.Errorf("expected %s, got %s", res[i], a.String())
			}
		}
	}

	// *@clown.com
	if al_list, err = mdb.LookupAlias("*@clown.com"); err != nil {
		t.Errorf("lookup *@clown.com: %s", err)
	}
	res = []string{
		"bozo@clown.com rednose@clown.com, micky@clown.com",
		"steve@clown.com willy@circus, grocho@marx, chico@marx",
	}
	if len(al_list) != 2 {
		t.Errorf("LookupAlias: *@clown.com should be 2 aliases, got %d", len(al_list))
	} else {
		for i, a := range al_list {
			if a.String() != res[i] {
				t.Errorf("expected %s, got %s", res[i], a.String())
			}
		}
	}

	// steve@*
	if al_list, err = mdb.LookupAlias("steve@*"); err != nil {
		t.Errorf("lookup steve@*: %s", err)
	}
	res = []string{
		"steve@clown.com willy@circus, grocho@marx, chico@marx",
		"steve@office bill@plumbers.com, mike@shovel.org",
	}
	if len(al_list) != 2 {
		t.Errorf("LookupAlias: steve@* should be 2 aliases, got %d", len(al_list))
	} else {
		for i, a := range al_list {
			if a.String() != res[i] {
				t.Errorf("expected %s, got %s", res[i], a.String())
			}
		}
	}

	// *@*
	if al_list, err = mdb.LookupAlias("*@*"); err != nil {
		t.Errorf("lookup *@*: %s", err)
	}
	res = []string{
		"bozo@clown.com rednose@clown.com, micky@clown.com",
		"steve@clown.com willy@circus, grocho@marx, chico@marx",
		"miller@office dave@work, dave@home, dave@mobile",
		"steve@office bill@plumbers.com, mike@shovel.org",
	}
	if len(al_list) != 4 {
		t.Errorf("LookupAlias: *@* should be 2 aliases, got %d", len(al_list))
	} else {
		for i, a := range al_list {
			if a.String() != res[i] {
				t.Errorf("expected %s, got %s", res[i], a.String())
			}
		}
	}

	// *
	if al_list, err = mdb.LookupAlias("*"); err != nil {
		t.Errorf("lookup *: %s", err)
	}
	res = []string{
		"postfix: root, daemon@server, postmaster@usps.gov",
		"rebar: \"| cat > /dev/null\", /tmp/rubbish",
	}
	if len(al_list) != 2 {
		t.Errorf("LookupAlias: * should be 2 aliases, got %d", len(al_list))
	} else {
		for i, a := range al_list {
			if a.String() != res[i] {
				t.Errorf("expected %s, got %s", res[i], a.String())
			}
		}
	}

	// Now delete bill@plumbers.com of steve@office
	if err = mdb.RemoveRecipient("steve@office", "bill@plumbers.com"); err != nil {
		t.Errorf("Remove bill@plumbers.com: %s", err)
	} else if al_list, err = mdb.LookupAlias("steve@office"); err != nil {
		t.Errorf("Lookup truncated steve@office: %s", err)
	} else if len(al_list) != 1 {
		t.Errorf("Look up of modified steve@office expected 1 alias, got %d",
			len(al_list))
	} else {
		a := al_list[0]
		if a.String() != "steve@office mike@shovel.org" {
			t.Errorf("Truncated steve@office should be \"steve@office mike@shovel.org\", got %s",
				a.String())
		}
	}

	// delete a bogus recipient
	err = mdb.RemoveRecipient("steve@office", "bronco.billy@the.ranch")
	if err == nil {
		t.Errorf("delete of bronco.billy should have failed")
	} else if err != ErrMdbRecipientNotFound {
		t.Errorf("delete of bronco.billy unexpected error, %s", err)
	}

	// delete a bogus pipe recipient
	err = mdb.RemoveRecipient("steve@office", "\"| cat > /dev/null\"")
	if err == nil {
		t.Errorf("delete of '\"|cat > /dev/null\"' should have failed")
	} else if err != ErrMdbNoLocalPipe {
		t.Errorf("delete of '\"|cat > /dev/null\"' from steve@office  unexpected error, %s", err)
	}

	// try to delete a recipient as an alias
	err = mdb.RemoveAlias("mike@shovel.org")
	if err == nil {
		t.Errorf("delete of a mike@shovel.org as an alias did not fail")
	} else if err != ErrMdbNotAlias {
		t.Errorf("delete of alias mike@shovel.org unexpected error, %s", err)
	}

	// then the other (last) from steve@office
	if err = mdb.RemoveRecipient("steve@office", "mike@shovel.org"); err != nil {
		t.Errorf("Remove mike@shovel.org: %s", err)
	}
	al_list, err = mdb.LookupAlias("steve@office")
	if err == nil {
		t.Errorf("Lookup of deleted steve@office should have failed")
	} else if err != ErrMdbAddressNotFound && err != ErrMdbDomainNotFound {
		t.Errorf("Lookup of deleted steve@office: %s", err)
	}

	// remove a pipe recipient
	if err = mdb.RemoveRecipient("rebar", "\"| cat > /dev/null\""); err != nil {
		t.Errorf("Remove | cat > /dev/null: %s", err)
	} else if al_list, err = mdb.LookupAlias("rebar"); err != nil {
		t.Errorf("Lookup truncated rebar: %s", err)
	} else if len(al_list) != 1 {
		t.Errorf("Look up of modified rebar expected 1 alias, got %d",
			len(al_list))
	} else {
		a := al_list[0]
		if a.String() != "rebar: /tmp/rubbish" {
			t.Errorf("Truncated rebar should be \"rebar: /tmp/rubbish\", got %s",
				a.String())
		}
	}

	// then the other (last) from rebar
	if err = mdb.RemoveRecipient("rebar", "/tmp/rubbish"); err != nil {
		t.Errorf("Remove bill@plumbers.com: %s", err)
	}
	al_list, err = mdb.LookupAlias("rebar")
	if err == nil {
		t.Errorf("Lookup of deleted rebar should have failed")
	} else if err != ErrMdbAddressNotFound && err != ErrMdbDomainNotFound {
		t.Errorf("Lookup of deleted rebar: %s", err)
	}

	// try to remove a bogus alias
	err = mdb.RemoveAlias("orange.one@putz")
	if err == nil {
		t.Errorf("delete of a putz did not fail")
	} else if err != ErrMdbNotAlias {
		t.Errorf("delete of a putz unexpected error, %s", err)
	}

	// try to remove a domain out from under an alias
	if err = mdb.DeleteDomain("office"); err == nil {
		t.Errorf("delete of office out from under alias should have failed")
	} else if err != ErrMdbDomainBusy {
		t.Errorf("delete of office: unexpected error, %s", err)
	}
	// now remove the whole alias of all that remain
	if err = mdb.RemoveAlias("miller@office"); err != nil {
		t.Errorf("Remove miller@office: %s", err)
	}
	al_list, err = mdb.LookupAlias("miller@office")
	if err == nil {
		t.Errorf("Lookup of deleted miller@office should have failed")
	} else if err != ErrMdbAddressNotFound && err != ErrMdbDomainNotFound {
		t.Errorf("Lookup of deleted miller@office: %s", err)
	}

	if err = mdb.RemoveAlias("bozo@clown.com"); err != nil {
		t.Errorf("Remove bozo@clown.com: %s", err)
	}
	al_list, err = mdb.LookupAlias("bozo@clown.com")
	if err == nil {
		t.Errorf("Lookup of deleted bozo@clown.com should have failed")
	} else if err != ErrMdbAddressNotFound && err != ErrMdbDomainNotFound {
		t.Errorf("Lookup of deleted bozo@clown.com: %s", err)
	}

	if err = mdb.RemoveAlias("steve@clown.com"); err != nil {
		t.Errorf("Remove steve@clown.com: %s", err)
	}
	al_list, err = mdb.LookupAlias("steve@clown.com")
	if err == nil {
		t.Errorf("Lookup of deleted steve@clown.com should have failed")
	} else if err != ErrMdbAddressNotFound && err != ErrMdbDomainNotFound {
		t.Errorf("Lookup of deleted steve@clown.com: %s", err)
	}

	// now remove the whole alias
	if err = mdb.RemoveAlias("postfix"); err != nil {
		t.Errorf("Remove postfix: %s", err)
	}
	al_list, err = mdb.LookupAlias("postfix")
	if err == nil {
		t.Errorf("Lookup of deleted postfix should have failed")
	} else if err != ErrMdbAddressNotFound && err != ErrMdbDomainNotFound {
		t.Errorf("Lookup of deleted postfix: %s", err)
	}

	// the DB should now be empty
	aCount, dCount = countAddresses(mdb)
	if aCount != 0 || dCount != 0 {
		t.Errorf("count after all deletes: expected 0 addr, 0 domain, got %d, %d",
			aCount, dCount)
		al_list, err = mdb.LookupAlias("*@*")
		if err == nil {
			for _, al = range al_list {
				t.Errorf("*@* should be gone: %s", al.String())
			}
		} else {
			t.Errorf("LookupAlias of *@* after bad counts, %s", err)
		}
		al_list, err = mdb.LookupAlias("*")
		if err == nil {
			for _, al = range al_list {
				t.Errorf("* should be gone: %s", al.String())
			}
		} else {
			t.Errorf("LookupAlias of * after bad counts, %s", err)
		}
	}

	// test + extension stuff
	recips = []string{"bill+spam@soho.org", "dave+spam@soho.org"}
	if err = makeAlias(mdb, "spam@soho.org", recips); err != nil {
		t.Errorf("makeAlias of spam@soho.org: %s", err)
	}
	recips = []string{"bill+junk@soho.org", "sue+junk@soho.org"}
	if err = makeAlias(mdb, "junk@soho.org", recips); err != nil {
		t.Errorf("makeAlias of junk@soho.org, %s", err)
	}
	if err = mdb.RemoveRecipient("spam@soho.org", "bill+spam@soho.org"); err != nil {
		t.Errorf("RemoveRecipient bill+spam@soho.org: %s", err)
	}
	if al_list, err = mdb.LookupAlias("spam@soho.org"); err != nil {
		t.Errorf("Lookup of spam@soho.org: %s", err)
	} else if len(al_list) != 1 {
		t.Errorf("Lookup of modified spam@soho.org expected 1 alias, go %d",
			len(al_list))
	} else {
		al := al_list[0]
		if al.String() != "spam@soho.org dave+spam@soho.org" {
			t.Errorf("Modified spam@soho.org should be \"spam@soho.org dave+spam@soho.org\", got %s",
				al.String())
		}
	}

	// make sure we didn't mess with the other, similar one
	if al_list, err = mdb.LookupAlias("junk@soho.org"); err != nil {
		t.Errorf("Lookup of junk@soho.org: %s", err)
	} else if len(al_list) != 1 {
		t.Errorf("Lookup of modified spam@soho.org expected 1 alias, go %d",
			len(al_list))
	} else {
		al := al_list[0]
		if al.String() != "junk@soho.org bill+junk@soho.org, sue+junk@soho.org" {
			t.Errorf("Modified spam@soho.org should be \"junk@soho.org bill+junk@soho.org, sue+junk@soho.org\", got %s",
				al.String())
		}
	}
}

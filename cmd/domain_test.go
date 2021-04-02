/*
Copyright © 2021 Jim Lieb <lieb@sea-troll.net>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	//"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lieb/postdove/maildb"
	// "github.com/spf13/cobra"
)

// Test_Domain
func Test_Domain(t *testing.T) {
	var (
		err         error
		dir         string
		dbfile      string
		args        []string
		out, errout string
	)

	fmt.Println("Test_Domain")

	// Make a database and test it
	dir, err = ioutil.TempDir("", "TestCmds-*")
	defer os.RemoveAll(dir)
	dbfile = filepath.Join(dir, "test.db")

	// Now create a good database
	args = []string{"create", "-d", dbfile, "-s", "../schema.sql"}
	out, errout, err = doTest(rootCmd, args)
	if err != nil {
		t.Errorf("Create DB: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Create DB: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Create DB: did not expect error output, got %s", errout)
	}

	// Add some domains, first with just defaults
	args = []string{"-d", dbfile, "add", "domain", "somewhere.org"} // using default class
	out, errout, err = doTest(rootCmd, args)
	if err != nil {
		t.Errorf("Add somewhere.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Add somewhere.org: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add somewhere.org: did not expect error output, got %s", errout)
	}
	args = []string{"-d", dbfile, "show", "domain", "somewhere.org"} // now look it up
	out, errout, err = doTest(rootCmd, args)
	if err != nil {
		t.Errorf("Show of somewhere.org in good DB: Unexpected error, %s", err)
	}
	if out != "Name:\t\tsomewhere.org\nClass:\t\tinternet\nTransport:\t--\nAccess:\t\t--\nUserID:\t\t--\nGroup ID:\t--\nRestrictions:\tDEFAULT\n" {
		t.Errorf("Show of somewhere.org in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of somewhere.org in good DB: did not expect error output, got %s", errout)
	}

	// Now a "virtual" domain (for mailboxes)
	args = []string{"-d", dbfile, "add", "domain", "home.net", "virtual"} // using default class
	out, errout, err = doTest(rootCmd, args)
	if err != nil {
		t.Errorf("Add home.net: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Add home.net: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add home.net: did not expect error output, got %s", errout)
	}
	args = []string{"-d", dbfile, "show", "domain", "home.net"} // now look it up
	out, errout, err = doTest(rootCmd, args)
	if err != nil {
		t.Errorf("Show of home.net in good DB: Unexpected error, %s", err)
	}
	if out != "Name:\t\thome.net\nClass:\t\tvirtual\nTransport:\t--\nAccess:\t\t--\nUserID:\t\t--\nGroup ID:\t--\nRestrictions:\tDEFAULT\n" {
		t.Errorf("Show of home.net in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of home.net in good DB: did not expect error output, got %s", errout)
	}

	// Now edit it
	args = []string{"-d", dbfile, "edit", "domain", "home.net", "--uid", "43", "--gid", "88", "--rclass", "STALL"}
	out, errout, err = doTest(rootCmd, args)
	if err != nil {
		t.Errorf("Edit home.net: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Edit home.net: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Edit home.net: did not expect error output, got %s", errout)
	}
	args = []string{"-d", dbfile, "show", "domain", "home.net"} // now look it up
	out, errout, err = doTest(rootCmd, args)
	if err != nil {
		t.Errorf("Show of home.net in good DB: Unexpected error, %s", err)
	}
	if out != "Name:\t\thome.net\nClass:\t\tvirtual\nTransport:\t--\nAccess:\t\t--\nUserID:\t\t43\nGroup ID:\t88\nRestrictions:\tSTALL\n" {
		t.Errorf("Show of home.net in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Show of home.net in good DB: did not expect error output, got %s", errout)
	}

	// delete a domain and check, starting with a non-existent
	args = []string{"-d", dbfile, "delete", "domain", "nowhere.org"}
	out, errout, err = doTest(rootCmd, args)
	if err == nil {
		t.Errorf("Delete nowhere.org: should have failed")
	} else if err != maildb.ErrMdbDomainNotFound {
		t.Errorf("Delete nowhere.org: Unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Delete nowhere.org: Expected formatted help output, got nothing")
	}
	if !strings.Contains(errout, "domain not found") {
		t.Errorf("Delete nowhere.org: Expected error output, got %s", errout)
	}

	args = []string{"-d", dbfile, "delete", "domain", "somewhere.org"}
	out, errout, err = doTest(rootCmd, args)
	if err != nil {
		t.Errorf("Delete nowhere.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Delete nowhere.org: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Delete nowhere.org: Expected no error output, got %s", errout)
	}

	// Now see if it is still there
	args = []string{"-d", dbfile, "show", "domain", "somewhere.org"} // now look it up
	out, errout, err = doTest(rootCmd, args)
	if err == nil {
		t.Errorf("Show somewhere.org: should have failed")
	} else if err != maildb.ErrMdbDomainNotFound {
		t.Errorf("Show of somewhere.org in good DB: Unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Show of somewhere.org: Expected formatted help output, got nothing")
	}
	if !strings.Contains(errout, "domain not found") {
		t.Errorf("Show of somewhere.org: Expected error output, got %s", errout)
	}

}

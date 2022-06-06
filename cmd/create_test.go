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
	//	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	//	"strings"
	"testing"
	//"github.com/lieb/postdove/maildb"
	//	"github.com/spf13/cobra"
)

// Test_Create
// Test create command with options
func Test_Create(t *testing.T) {
	var (
		err         error
		dir         string
		dbfile      string
		args        []string
		out, errout string
	)

	fmt.Println("Test_Create")

	// Make a database and test it
	dir, err = ioutil.TempDir("", "TestCreate-*")
	defer os.RemoveAll(dir)
	dbfile = filepath.Join(dir, "test.db")

	// Now create a good database with default initialization
	args = []string{"create", "-d", dbfile}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Create good DB: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Create good DB: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Create good DB: did not expect error output, got %s", errout)
	}
	// Domain testing. Check to see if the pre-loaded domains are there
	args = []string{"-d", dbfile, "show", "domain", "localhost"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of localhost in good DB: Unexpected error, %s", err)
	}
	if out != "Name:\t\tlocalhost\nClass:\t\tlocal\nTransport:\t--\nUserID:\t\t99\nGroup ID:\t99\nRestrictions:\t--\n" {
		t.Errorf("Show of localhost in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Show of localhost in good DB: did not expect error output, got %s", errout)
	}

	args = []string{"-d", dbfile, "show", "domain", "localhost.localdomain"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of localhost.localdomain in good DB: Unexpected error, %s", err)
	}
	if out != "Name:\t\tlocalhost.localdomain\nClass:\t\tlocal\nTransport:\t--\nUserID:\t\t--\nGroup ID:\t--\nRestrictions:\t--\n" {
		t.Errorf("Show of localhost.localdomain in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Show of localhost.localdomain in good DB: did not expect error output, got %s", errout)
	}

	// Show a bogus domain
	args = []string{"-d", dbfile, "show", "domain", "lost.mars"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Show of lost.mars should have failed")
	} else if err.Error() != "domain not found" {
		t.Errorf("Show of lost.mars: Unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Show of lost.mars: expect help message to output, got nothing")
	}
	if errout == "" {
		t.Errorf("Show of lost.mars: expected formatted error output")
	}

	// See if we have some RFC 2142 aliases. Must have "postmaster"
	args = []string{"-d", dbfile, "show", "alias", "postmaster"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of postmaster in good DB: Unexpected error, %s", err)
	}
	if out != "Alias:\t\tpostmaster\nTargets:\troot\n" {
		t.Errorf("Show of postmaster in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of postmaster in good DB: did not expect error output, got %s", errout)
	}

	// and "abuse". Cuz nobody does that, right?
	args = []string{"-d", dbfile, "show", "alias", "abuse"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of abuse in good DB: Unexpected error, %s", err)
	}
	if out != "Alias:\t\tabuse\nTargets:\troot\n" {
		t.Errorf("Show of abuse in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of abuse in good DB: did not expect error output, got %s", errout)
	}

	// just whack the database file to start clean..
	os.RemoveAll(dbfile)

	// now just load up domains
	args = []string{"create", "-d", dbfile, "--no-aliases", "-l", "../maildb/files/domains"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Create good DB: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Create good DB: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Create good DB: did not expect error output, got %s", errout)
	}
	// should have localhost in domains
	args = []string{"-d", dbfile, "show", "domain", "localhost"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of localhost in good DB: Unexpected error, %s", err)
	}
	if out != "Name:\t\tlocalhost\nClass:\t\tlocal\nTransport:\t--\nUserID:\t\t99\nGroup ID:\t99\nRestrictions:\t--\n" {
		t.Errorf("Show of localhost in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Show of localhost in good DB: did not expect error output, got %s", errout)
	}

	// but not postmaster
	args = []string{"-d", dbfile, "show", "alias", "postmaster"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Show of postmaster should have failed")
	} else if err.Error() != "address not found" {
		t.Errorf("Show of postmaster: Unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Show of postmaster: expect help message to output, got nothing")
	}
	if errout == "" {
		t.Errorf("Show of postmaster: expected formatted error output")
	}
}

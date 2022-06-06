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
func TestCreateNoAliases(t *testing.T) {
	var (
		err         error
		dir         string
		dbfile      string
		args        []string
		out, errout string
	)

	fmt.Println("TestCreateNoAliases")

	// Make a database and test it
	dir, err = ioutil.TempDir("", "TestCreate-*")
	defer os.RemoveAll(dir)
	dbfile = filepath.Join(dir, "test.db")

	// now just load up aliases
	args = []string{"create", "-d", dbfile, "--no-locals", "-a", "../maildb/files/aliases"}
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

	// should have postmaster
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

	// but not localhost
	args = []string{"-d", dbfile, "show", "domain", "localhost"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Show of localhost should have failed")
	} else if err.Error() != "domain not found" {
		t.Errorf("Show of localhost: Unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Show of localhost: expect help message to output, got nothing")
	}
	if errout == "" {
		t.Errorf("Show of localhost: expected formatted error output")
	}
}

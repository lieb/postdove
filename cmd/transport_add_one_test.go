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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	//"strings"
	"testing"
	//"github.com/lieb/postdove/maildb"
)

// TestTransport
func TestTransportAddOne(t *testing.T) {
	var (
		err         error
		dir         string
		dbfile      string
		args        []string
		out, errout string
	)

	fmt.Println("TestTransportAddOne")

	// Make a database and test it
	dir, err = ioutil.TempDir("", "TestTransport-*")
	defer os.RemoveAll(dir)
	dbfile = filepath.Join(dir, "test.db")

	// Now create a good database
	args = []string{"create", "-d", dbfile, "--no-locals", "--no-aliases"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Create DB: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Create DB: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Create DB: did not expect error output, got %s", errout)
	}

	// Add transport "domain :next.com"
	args = []string{"-d", dbfile, "add", "transport", "domain", "--nexthop", "next.com"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Add domain transport: unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Add domain transport: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add domain transport: did not expect error output, got %s", errout)
	}

	// show "domain :next.com"
	args = []string{"-d", dbfile, "show", "transport", "domain"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of domain: Unexpected error, %s", err)
	}
	if out != "Name:\t\tdomain\nTransport:\t--\nNexthop:\tnext.com" {
		t.Errorf("Show of domain: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Show of domain: did not expect error output, got %s", errout)
	}
}

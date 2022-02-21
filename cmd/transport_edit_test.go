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
	"testing"
)

// TestTransportEdit
func TestTransportEdit(t *testing.T) {
	var (
		err         error
		dir         string
		dbfile      string
		args        []string
		out, errout string
	)

	fmt.Println("TestTransportEdit")

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

	// Import transports
	inputStr := `
# some transports
empty :
relay smtp:foo.com:24
local local:
`
	args = []string{"-d", dbfile, "import", "transport"}
	out, errout, err = doTest(rootCmd, inputStr, args)
	if err != nil {
		t.Errorf("Import transport unexpectedly failed")
	}
	if out != "" {
		t.Errorf("Import transport: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Import transport: did not expect error output, got %s", errout)
	}

	// edit relay to remove transport
	args = []string{"-d", dbfile, "edit", "transport", "relay", "--no-transport"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Edit transport relay foo: failed, %s", err)
	}
	if out != "" {
		t.Errorf("Edit transport relay: expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Edit transport relay: expected no output, got %s", errout)
	}
	// check it
	args = []string{"-d", dbfile, "show", "transport", "relay"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of relay: Unexpected error, %s", err)
	}
	if out != "Name:\t\trelay\nTransport:\t--\nNexthop:\tfoo.com:24" {
		t.Errorf("Show of relay: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Show of relay: did not expect error output, got %s", errout)
	}

	args = []string{"-d", dbfile, "edit", "transport", "empty", "--nexthop", "somewhere.org"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Edit transport empty foo: failed, %s", err)
	}
	if out != "" {
		t.Errorf("Edit transport empty: expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Edit transport empty: expected no output, got %s", errout)
	}
	// check it
	args = []string{"-d", dbfile, "show", "transport", "empty"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of empty: Unexpected error, %s", err)
	}
	if out != "Name:\t\tempty\nTransport:\t--\nNexthop:\tsomewhere.org" {
		t.Errorf("Show of empty: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Show of empty: did not expect error output, got %s", errout)
	}
}

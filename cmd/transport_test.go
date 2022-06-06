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
	"strings"
	"testing"
)

// TestTransport_
func TestTransport_(t *testing.T) {
	var (
		err         error
		dir         string
		dbfile      string
		args        []string
		out, errout string
	)

	fmt.Println("TestTransport")

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

	// test bad import just stdin is enough here...
	args = []string{"-d", dbfile, "import", "transport"}
	inputStr = `
# some bogus transports
bogus foo
`
	out, errout, err = doTest(rootCmd, inputStr, args)
	if err != nil {
		if !strings.Contains(err.Error(), "must include ':'") {
			t.Errorf("Import transport unexpected error, %s", err)
		}
	}
	if out == "" {
		t.Errorf("Import transport: expected output, got %s", out)
	}
	if errout == "" {
		t.Errorf("Import transport: expected error output, got %s", errout)
	}

	inputStr = `
# some bogus transports
goofy
`
	out, errout, err = doTest(rootCmd, inputStr, args)
	if err != nil {
		if !strings.Contains(err.Error(), "key but no value") {
			t.Errorf("Import transport unexpected error, %s", err)
		}
	}
	if out == "" {
		t.Errorf("Import transport: expected output, got none")
	}
	if errout == "" {
		t.Errorf("Import transport: expected error output, got none")
	}

	// show "bogus foo" just in case...
	args = []string{"-d", dbfile, "show", "transport", "bogus"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Show of bogus: should have failed")
	} else if !strings.Contains(err.Error(), "Transport not found") {
		t.Errorf("Show of bogus: unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Show of bogus: expected output, got none")
	}
	if errout == "" {
		t.Errorf("Show of bogus: expected error output, got none")
	}

	// show "empty :"
	args = []string{"-d", dbfile, "show", "transport", "empty"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of empty: Unexpected error, %s", err)
	}
	if out != "Name:\t\tempty\nTransport:\t--\nNexthop:\t--\n" {
		t.Errorf("Show of empty: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Show of empty: did not expect error output, got %s", errout)
	}

	// show "local local:"
	args = []string{"-d", dbfile, "show", "transport", "local"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of local: Unexpected error, %s", err)
	}
	if out != "Name:\t\tlocal\nTransport:\tlocal\nNexthop:\t--\n" {
		t.Errorf("Show of local: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Show of local: did not expect error output, got %s", errout)
	}

	// show "relay smtp:foo.com:24
	args = []string{"-d", dbfile, "show", "transport", "relay"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of relay: Unexpected error, %s", err)
	}
	if out != "Name:\t\trelay\nTransport:\tsmtp\nNexthop:\tfoo.com:24\n" {
		t.Errorf("Show of relay: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Show of relay: did not expect error output, got %s", errout)
	}

	// Now export it and compare
	exportList := "empty :\n" +
		"local local:\n" +
		"relay smtp:foo.com:24\n"
	args = []string{"-d", dbfile, "export", "transport"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Export transport: unexpected error, %s", err)
	}
	if out != exportList {
		t.Errorf("Export transport: expected export list[%d](%s), got [%d](%s)",
			len(exportList), exportList, len(out), out)
	}
	if errout != "" {
		t.Errorf("Export transport: unexpected error output, got %s", errout)
	}

	// just export relay
	exportList = "relay smtp:foo.com:24\n"
	args = []string{"-d", dbfile, "export", "transport", "relay"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Export transport relay: unexpected error, %s", err)
	}
	if out != exportList {
		t.Errorf("Export transport relay: expected export list[%d](%s), got [%d](%s)",
			len(exportList), exportList, len(out), out)
	}
	if errout != "" {
		t.Errorf("Export transport relay: unexpected error output, got %s", errout)
	}

	// export bogus transport
	args = []string{"-d", dbfile, "export", "transport", "bogus"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		if !strings.Contains(err.Error(), "Transport not found") {
			t.Errorf("Export transport bogus: unexpected error, %s", err)
		}
	}
	if out == "" {
		t.Errorf("Export transport bogus: expected output, got nothing")
	}
	if errout == "" {
		t.Errorf("Export transport bogus: expected error output, got nothing")
	}

	// delete one of them
	args = []string{"-d", dbfile, "delete", "transport", "local"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Delete transport local: unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Delete transport local: unexpected output, %s", out)
	}
	if errout != "" {
		t.Errorf("Delete transport local: unexpected err output, %s", errout)
	}

	// see if it really got deleted
	exportList = "empty :\n" +
		"relay smtp:foo.com:24\n"
	args = []string{"-d", dbfile, "export", "transport"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Export transport after delete of local: unexpected error, %s", err)
	}
	if out != exportList {
		t.Errorf("Export transport after delete of local: expected export list[%d](%s), got [%d](%s)",
			len(exportList), exportList, len(out), out)
	}
	if errout != "" {
		t.Errorf("Export transport after delete of local: unexpected error output, got %s", errout)
	}

	// now delete a bogus entry
	args = []string{"-d", dbfile, "delete", "transport", "bogus"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Delete transport bogus: should have failed")
	} else if !strings.Contains(err.Error(), "Transport not found") {
		t.Errorf("Delete transport bogus: unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Delete transport bogus: no output, %s", out)
	}
	if errout == "" {
		t.Errorf("Delete transport bogus: unexpected err output, %s", errout)
	}

	// edit bogus transport
	args = []string{"-d", dbfile, "edit", "transport", "bogus"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Edit transport empty bogus did not failed")
	} else if !strings.Contains(err.Error(), "") {
		t.Errorf("Edit transport bogus: unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Edit transport bogusd expected output, got none")
	}
	if errout == "" {
		t.Errorf("Edit transport bogus: expected error output, got none")
	}

	// edit relay to remove nexthop
	args = []string{"-d", dbfile, "edit", "transport", "relay", "--no-nexthop"}
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
	if out != "Name:\t\trelay\nTransport:\tsmtp\nNexthop:\t--\n" {
		t.Errorf("Show of relay: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Show of relay: did not expect error output, got %s", errout)
	}

	// edit empty to add a transport
	args = []string{"-d", dbfile, "edit", "transport", "empty", "--transport", "foo"}
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
	if out != "Name:\t\tempty\nTransport:\tfoo\nNexthop:\t--\n" {
		t.Errorf("Show of empty: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Show of empty: did not expect error output, got %s", errout)
	}
}

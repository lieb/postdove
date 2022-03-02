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
	"github.com/lieb/postdove/maildb"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	// "github.com/spf13/cobra"
)

// Test_Domain
func Test_Address(t *testing.T) {
	var (
		err         error
		dir         string
		dbfile      string
		args        []string
		out, errout string
	)

	fmt.Println("Test_Address")

	// Make a database and test it
	dir, err = ioutil.TempDir("", "TestAddress-*")
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

	// Add in some access and transport entries
	args = []string{"-d", dbfile, "add", "access", "STALL", "x-stall"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Add access STALL: unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Add access STALL: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add access STALL: did not expect error output, got %s", errout)
	}
	args = []string{"-d", dbfile, "add", "access", "DUMP", "x-dump"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Add access DUMP: unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Add access DUMP: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add access DUMP: did not expect error output, got %s", errout)
	}

	args = []string{"-d", dbfile, "add", "transport", "relay",
		"--transport", "smtp", "--nexthop", "foo.com:24"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Add relay transport: unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Add relay transport: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add relay transport: did not expect error output, got %s", errout)
	}

	// Add some addresses, first with just defaults
	args = []string{"-d", dbfile, "add", "address", "bill@somewhere.org"} // using default class
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Add bill@somewhere.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Add bill@somewhere.org: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add bill@somewhere.org: did not expect error output, got %s", errout)
	}
	args = []string{"-d", dbfile, "show", "address", "bill@somewhere.org"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of bill@somewhere.org in good DB: Unexpected error, %s", err)
	}
	if out != "Address:\t\tbill@somewhere.org\nTransport:\t--\nRestrictions:\t--\n" {
		t.Errorf("Show of bill@somewhere.org in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of bill@somewhere.org in good DB: did not expect error output, got %s", errout)
	}
	// and a local address
	args = []string{"-d", dbfile, "add", "address", "postmaster"} // using default class
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Add postmaster: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Add postmaster: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add postmaster: did not expect error output, got %s", errout)
	}
	args = []string{"-d", dbfile, "show", "address", "postmaster"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of postmaster in good DB: Unexpected error, %s", err)
	}
	if out != "Address:\t\tpostmaster\nTransport:\t--\nRestrictions:\t--\n" {
		t.Errorf("Show of postmaster in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of postmaster in good DB: did not expect error output, got %s", errout)
	}

	// now with a valid option
	args = []string{"-d", dbfile, "add", "address", "dave@somewhere.org",
		"--rclass", "DUMP", "--transport", "relay"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Add dave@somewhere.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Add dave@somewhere.org: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add dave@somewhere.org: did not expect error output, got %s", errout)
	}
	args = []string{"-d", dbfile, "show", "address", "dave@somewhere.org"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of dave@somewhere.org in good DB: Unexpected error, %s", err)
	}
	if out != "Address:\t\tdave@somewhere.org\nTransport:\trelay\nRestrictions:\tDUMP\n" {
		t.Errorf("Show of dave@somewhere.org in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of dave@somewhere.org in good DB: did not expect error output, got %s", errout)
	}

	// now with a bogus option
	args = []string{"-d", dbfile, "add", "address", "steve@somewhere.org", "-B", "88"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Bad add of steve@somewhere.org: Should have failed")
	} else if !strings.Contains(err.Error(), "unknown shorthand flag: 'B'") {
		t.Errorf("Bad add of steve@somewhere.org: Unexpected error %s", err)
	}
	if out == "" {
		t.Errorf("Bad add steve@somewhere.org: Expected output, got nothing")
	}
	if errout == "" {
		t.Errorf("Bdd add steve@somewhere.org: Expected error output, got nothing")
	}

	// Now edit bill@somewhere.org
	args = []string{"-d", dbfile, "edit", "address", "bill@somewhere.org", "--rclass", "STALL"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Edit bill@somewhere.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Edit bill@somewhere.org: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Edit bill@somewhere.org: did not expect error output, got %s", errout)
	}
	args = []string{"-d", dbfile, "show", "address", "bill@somewhere.org"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of bill@somewhere.org in good DB: Unexpected error, %s", err)
	}
	if out != "Address:\t\tbill@somewhere.org\nTransport:\t--\nRestrictions:\tSTALL\n" {
		t.Errorf("Show of bill@somewhere.org in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Show of bill@somewhere.org in good DB: did not expect error output, got %s", errout)
	}

	// edit dave@somewhere to remove rclass and transport
	args = []string{"-d", dbfile, "edit", "address", "dave@somewhere.org", "--no-rclass", "--no-transport"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Edit bill@somewhere.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Edit bill@somewhere.org: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Edit bill@somewhere.org: did not expect error output, got %s", errout)
	}
	args = []string{"-d", dbfile, "show", "address", "dave@somewhere.org"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of dave@somewhere.org in good DB: Unexpected error, %s", err)
	}
	if out != "Address:\t\tdave@somewhere.org\nTransport:\t--\nRestrictions:\t--\n" {
		t.Errorf("Show of dave@somewhere.org in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Show of dave@somewhere.org in good DB: did not expect error output, got %s", errout)
	}

	// delete a domain and check, starting with a non-existent
	args = []string{"-d", dbfile, "delete", "address", "steve@somewhere.org"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Delete steve@somewhere.org: should have failed")
	} else if err != maildb.ErrMdbAddressNotFound {
		t.Errorf("Delete steve@somwhere.org: Unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Delete steve@somewhere.org: Expected formatted help output, got nothing")
	}
	if !strings.Contains(errout, "address not found") {
		t.Errorf("Delete steve@somewhere.org: Expected error output, got %s", errout)
	}

	args = []string{"-d", dbfile, "delete", "address", "dave@somewhere.org"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Delete dave@somewhere.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Delete dave@somewhere.org: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Delete dave@somewhere.org: Expected no error output, got %s", errout)
	}

	// Now see if it is still there
	args = []string{"-d", dbfile, "show", "address", "dave@somewhere.org"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Show dave@somewhere.org: should have failed")
	} else if err != maildb.ErrMdbAddressNotFound {
		t.Errorf("Show of dave@somewhere.org in good DB: Unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Show of dave@somewhere.org: Expected formatted help output, got nothing")
	}
	if !strings.Contains(errout, "address not found") {
		t.Errorf("Show of dave@somewhere.org: Expected error output, got %s", errout)
	}

	// import some addresses first from stdin
	args = []string{"-d", dbfile, "import", "address"}
	inputStr := `
# Just one to test stdin
mike@bill.org
`
	out, errout, err = doTest(rootCmd, inputStr, args)
	if err != nil {
		t.Errorf("Import of mike@bill.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Import of mike@bill.org: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Import of mike@bill.org: Expected no error output, got %s", errout)
	}
	// and now a file
	args = []string{"-d", dbfile, "import", "address", "-i", "./test_addresses.txt"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Import of addresses from file: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Import of addresses from file: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Import of addresses from file: Expected no error output, got %s", errout)
	}

	// export list to date...
	exportList := "mike@bill.org\n" +
		"gramma@cottage\n" +
		"wolf@forest\n" +
		"mary@little.lamb\n" +
		"bill@somewhere.org rclass=STALL\n"
	// now check the contents.
	args = []string{"-d", dbfile, "export", "address"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Export addresses: Unexpected error, %s", err)
	}
	if out != exportList {
		t.Errorf("Export addresses: Expected export list[%d](%s), got [%d](%s)",
			len(exportList), exportList, len(out), out)
	}
	if errout != "" {
		t.Errorf("Export addresses: Expected no error output, got %s", errout)
	}
	// now check using wildcard "*" for local addresses
	exportList = "postmaster\n"
	args = []string{"-d", dbfile, "export", "address", "*"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Export * addresses: Unexpected error, %s", err)
	}
	if out != exportList {
		t.Errorf("Export * addresses: Expected export list[%d](%s), got [%d](%s)",
			len(exportList), exportList, len(out), out)
	}
	if errout != "" {
		t.Errorf("Export * addresses: Expected no error output, got %s", errout)
	}
	// now check just *@*.org
	exportList = "mike@bill.org\n" +
		"bill@somewhere.org rclass=STALL\n"
	args = []string{"-d", dbfile, "export", "address", "*@*.org"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Export *@*.org addresses: Unexpected error, %s", err)
	}
	if out != exportList {
		t.Errorf("Export *@*.org addresses: Expected export list[%d](%s), got [%d](%s)",
			len(exportList), exportList, len(out), out)
	}
	if errout != "" {
		t.Errorf("Export *@*.org addresses: Expected no error output, got %s", errout)
	}
}

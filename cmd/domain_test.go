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

	"github.com/lieb/postdove/maildb"
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
	dir, err = ioutil.TempDir("", "TestDomain-*")
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

	// import some domains first from stdin
	// Have to do this here because import is not reentrant due to cobra doing stuff
	// with the cmd structs which, BTW, are global and intended (rightly or wrongly) to be
	// one pass somewhere in initialization...
	fmt.Printf("domains from stdin\n")
	args = []string{"-d", dbfile, "import", "domain"}
	inputStr := `
# Just one to test stdin
bill.org class=local
`
	out, errout, err = doTest(rootCmd, inputStr, args)
	if err != nil {
		t.Errorf("Import of bill.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Import of bill.org: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Import of bill.org: Expected no error output, got %s", errout)
	}
	// Add some domains, first with just defaults
	args = []string{"-d", dbfile, "add", "domain", "somewhere.org"} // using default class
	out, errout, err = doTest(rootCmd, "", args)
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
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of somewhere.org in good DB: Unexpected error, %s", err)
	}
	if out != "Name:\t\tsomewhere.org\nClass:\t\tinternet\nTransport:\t--\nUserID:\t\t--\nGroup ID:\t--\nRestrictions:\t--\n" {
		t.Errorf("Show of somewhere.org in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of somewhere.org in good DB: did not expect error output, got %s", errout)
	}

	// Add in some access and transport entries
	args = []string{"-d", dbfile, "import", "access", "-i", "./test_access.txt"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Import of access rules from file: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Import of access rules from file: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Import of access rules from file: Expected no error output, got %s", errout)
	}
	args = []string{"-d", dbfile, "import", "transport", "-i", "./test_transports.txt"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Import of transports from file: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Import of transports from file: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Import of transports from file: Expected no error output, got %s", errout)
	}

	// try to add home.net with too many args
	args = []string{"-d", dbfile, "add", "domain", "home.net", "-c", "virtual", "more_bits"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Bad add of home.net: Should have failed")
	} else if !strings.Contains(err.Error(), "accepts 1 arg") {
		t.Errorf("Bad add of home.net: Unexpected error %s", err)
	}
	if out == "" {
		t.Errorf("Bad add home.net: Expected output, got nothing")
	}
	if errout == "" {
		t.Errorf("Bdd add home.net: Expected error output, got nothing")
	}

	// Now do it with correct args a "virtual" domain (for mailboxes) and other args set
	args = []string{"-d", dbfile, "add", "domain", "home.net", "-c", "virtual",
		"-u", "88", "-g", "89", "-r", "STALL", "-t", "relay"}
	out, errout, err = doTest(rootCmd, "", args)
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
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of home.net in good DB: Unexpected error, %s", err)
	}
	if out != "Name:\t\thome.net\nClass:\t\tvirtual\nTransport:\trelay\nUserID:\t\t88\nGroup ID:\t89\nRestrictions:\tSTALL\n" {
		t.Errorf("Show of home.net in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of home.net in good DB: did not expect error output, got %s", errout)
	}

	// Now edit it
	args = []string{"-d", dbfile, "edit", "domain", "home.net", "--no-uid", "-G", "--no-rclass", "-T"}
	out, errout, err = doTest(rootCmd, "", args)
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
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of home.net in good DB: Unexpected error, %s", err)
	}
	if out != "Name:\t\thome.net\nClass:\t\tvirtual\nTransport:\t--\nUserID:\t\t--\nGroup ID:\t--\nRestrictions:\t--\n" {
		t.Errorf("Show of home.net in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Show of home.net in good DB: did not expect error output, got %s", errout)
	}

	// delete a domain and check, starting with a non-existent
	args = []string{"-d", dbfile, "delete", "domain", "nowhere.org"}
	out, errout, err = doTest(rootCmd, "", args)
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
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Delete somewhere.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Delete somewhere.org: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Delete somewhere.org: Expected no error output, got %s", errout)
	}

	// Now see if it is still there
	args = []string{"-d", dbfile, "show", "domain", "somewhere.org"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
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

	// now get rid of home.net leaving just the default (no domains)
	args = []string{"-d", dbfile, "delete", "domain", "home.net"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Delete home.net: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Delete home.net: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Delete home.net: Expected no error output, got %s", errout)
	}
	// Now see if it is still there
	args = []string{"-d", dbfile, "show", "domain", "home.net"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Show home.net: should have failed")
	} else if err != maildb.ErrMdbDomainNotFound {
		t.Errorf("Show of home.net in good DB: Unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Show of home.net: Expected formatted help output, got nothing")
	}
	if !strings.Contains(errout, "domain not found") {
		t.Errorf("Show of home.net: Expected error output, got %s", errout)
	}

	// and now a file
	fmt.Printf("domains from file\n")
	args = []string{"-d", dbfile, "import", "domain", "-i", "./test_domains.txt"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Import of domains from file: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Import of domains from file: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Import of domains from file: Expected no error output, got %s", errout)
	}

	// export list to date...
	exportList := "bill.org class=local\n" +
		"cottage class=internet, rclass=permit\n" +
		"dish.net class=relay, rclass=DUMP\n" +
		"foo class=internet\n" +
		"pobox.org class=vmailbox, transport=local\n" +
		"run.com class=virtual, vuid=83, vgid=99\n" +
		"wm.com class=internet, transport=trash\n" +
		"zip.com class=internet\n"
	// now check the contents.
	args = []string{"-d", dbfile, "export", "domain"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Export domains: Unexpected error, %s", err)
	}
	if out != exportList {
		t.Errorf("Export domains: Expected export list[%d](%s), got [%d](%s)",
			len(exportList), exportList, len(out), out)
	}
	if errout != "" {
		t.Errorf("Export domains: Expected no error output, got %s", errout)
	}
	// now check using wildcard "*"
	args = []string{"-d", dbfile, "export", "domain", "*"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Export * domains: Unexpected error, %s", err)
	}
	if out != exportList {
		t.Errorf("Export * domains: Expected export list[%d](%s), got [%d](%s)",
			len(exportList), exportList, len(out), out)
	}
	if errout != "" {
		t.Errorf("Export * domains: Expected no error output, got %s", errout)
	}
	// now check just *.com
	exportList = "run.com class=virtual, vuid=83, vgid=99\n" +
		"wm.com class=internet, transport=trash\n" +
		"zip.com class=internet\n"
	args = []string{"-d", dbfile, "export", "domain", "*.com"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Export *.com domains: Unexpected error, %s", err)
	}
	if out != exportList {
		t.Errorf("Export *.com domains: Expected export list[%d](%s), got [%d](%s)",
			len(exportList), exportList, len(out), out)
	}
	if errout != "" {
		t.Errorf("Export *.com domains: Expected no error output, got %s", errout)
	}
}

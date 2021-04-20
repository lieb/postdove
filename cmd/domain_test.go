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
	dir, err = ioutil.TempDir("", "TestDomain-*")
	defer os.RemoveAll(dir)
	dbfile = filepath.Join(dir, "test.db")

	// Now create a good database
	args = []string{"create", "-d", dbfile, "-s", "../schema.sql"}
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
	if out != "Name:\t\tsomewhere.org\nClass:\t\tinternet\nTransport:\t--\nAccess:\t\t--\nUserID:\t\t--\nGroup ID:\t--\nRestrictions:\tDEFAULT\n" {
		t.Errorf("Show of somewhere.org in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of somewhere.org in good DB: did not expect error output, got %s", errout)
	}

	// try to add home.net with too many args
	args = []string{"-d", dbfile, "add", "domain", "home.net", "virtual", "more_bits"} // using default class
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Bad add of home.net: Should have failed")
	} else if !strings.Contains(err.Error(), "Only one class field") {
		t.Errorf("Bad add of home.net: Unexpected error %s", err)
	}
	if out == "" {
		t.Errorf("Bad add home.net: Expected output, got nothing")
	}
	if errout == "" {
		t.Errorf("Bdd add home.net: Expected error output, got nothing")
	}

	// Now do it with correct args a "virtual" domain (for mailboxes)
	args = []string{"-d", dbfile, "add", "domain", "home.net", "virtual"} // using default class
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
	if out != "Name:\t\thome.net\nClass:\t\tvirtual\nTransport:\t--\nAccess:\t\t--\nUserID:\t\t--\nGroup ID:\t--\nRestrictions:\tDEFAULT\n" {
		t.Errorf("Show of home.net in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of home.net in good DB: did not expect error output, got %s", errout)
	}

	// Now edit it
	args = []string{"-d", dbfile, "edit", "domain", "home.net", "--uid", "43", "--gid", "88", "--rclass", "STALL"}
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
	if out != "Name:\t\thome.net\nClass:\t\tvirtual\nTransport:\t--\nAccess:\t\t--\nUserID:\t\t43\nGroup ID:\t88\nRestrictions:\tSTALL\n" {
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

	// now get rid of home.net leaving just the default base (localhost, localhost.localdomain)
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

	// import some domains first from stdin
	args = []string{"-d", dbfile, "import", "domain"}
	inputStr := `
# Just one to test stdin
bill.org local
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
	// and now a file
	args = []string{"-d", dbfile, "import", "domain", "-i", "./test_domains.txt"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Import of bill.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Import of bill.org: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Import of bill.org: Expected no error output, got %s", errout)
	}

	// export list to date...
	exportList := "bill.org local\ndish.net relay\nfoo internet\nlocalhost local\nlocalhost.localdomain local\nrun.com virtual\nzip.com internet\n"
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
	exportList = "run.com virtual\nzip.com internet\n"
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

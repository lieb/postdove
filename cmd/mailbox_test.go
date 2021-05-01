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

// TestVMailboxCmd
func TestVMailboxCmd(t *testing.T) {
	var (
		err         error
		dir         string
		dbfile      string
		args        []string
		out, errout string
	)

	fmt.Println("Test_VMailboxCmd")

	// Make a database and test it
	dir, err = ioutil.TempDir("", "TestMailbox-*")
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
	// Add a vmail domain
	args = []string{"-d", dbfile, "add", "domain", "pobox.org", "vmailbox"} // using default class
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Add pobox.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Add pobox.org: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add pobox.org: did not expect error output, got %s", errout)
	}

	// Add an ordinary domain
	args = []string{"-d", dbfile, "add", "domain", "pobox.net"} // using default class
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Add pobox.net: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Add pobox.net: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add pobox.net: did not expect error output, got %s", errout)
	}

	// Try to create a mailbox in a non- domain
	args = []string{"-d", dbfile, "add", "mailbox", "jeff@pobox.net"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Add jeff@pobox.net should have failed")
	} else {
		if !strings.Contains(err.Error(), "") {
			t.Errorf("Add jeff@pobox.net: Unexpected error, %s", err)
		}
	}
	if out == "" {
		t.Errorf("Add jeff@pobox.net: expected output, got none")
	}
	if errout == "" {
		t.Errorf("Add jeff@pobox.net: expected output, got none")
	}

	// Try to create a mailbox in the vmailbox domain
	args = []string{"-d", dbfile, "add", "mailbox", "jeff@pobox.org"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Add jeff@pobox.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Add jeff@pobox.org: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add jeff@pobox.org: did not expect error output, got %s", errout)
	}

	// And check it out.

	expectedOut := "Name:\t\tjeff@pobox.org\nPassword Type:\tPLAIN\nPassword:\t--\nUserID:\t\t--\nGroupID:\t--\nHome:\t\t--\nQuota:\t\t*:bytes=300M\nEnabled:\ttrue\n"
	args = []string{"-d", dbfile, "show", "mailbox", "jeff@pobox.org"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Add jeff@pobox.org: Unexpected error, %s", err)
	}
	if out != expectedOut {
		fmt.Printf("jeff@pobox.org: len out=%d, len expected=%d\n", len(out), len(expectedOut))
		t.Errorf("Add jeff@pobox.org: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add jeff@pobox.org: did not expect error output, got %s", errout)
	}

	// and modify it
	args = []string{"-d", dbfile, "edit", "mailbox", "jeff@pobox.org", "-t", "crypt", "--password", "funny", "-u", "42", "--gid", "75", "-m", "black_hole",
		"-q", "none", "-E"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Edit jeff@pobox.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Edit jeff@pobox.org: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Edit jeff@pobox.org: did not expect error output, got %s", errout)
	}

	// check change
	expectedOut = "Name:\t\tjeff@pobox.org\nPassword Type:\tCRYPT\nPassword:\tfunny\nUserID:\t\t42\nGroupID:\t75\nHome:\t\tblack_hole\nQuota:\t\tnone\nEnabled:\tfalse\n"
	args = []string{"-d", dbfile, "show", "mailbox", "jeff@pobox.org"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Edit jeff@pobox.org: Unexpected error, %s", err)
	}
	if out != expectedOut {
		fmt.Printf("jeff@pobox.org: len out=%d, len expected=%d\n", len(out), len(expectedOut))
		t.Errorf("Edit jeff@pobox.org: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Edit jeff@pobox.org: did not expect error output, got %s", errout)
	}

	// and modify it with different args
	args = []string{"-d", dbfile, "edit", "mailbox", "jeff@pobox.org", "-t", "", "--no-password", "funny", "-U", "42", "--no-gid", "75", "-M", "black_hole",
		"-q", "reset", "--enable"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Edit jeff@pobox.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Edit jeff@pobox.org: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Edit jeff@pobox.org: did not expect error output, got %s", errout)
	}

	// check change
	expectedOut = "Name:\t\tjeff@pobox.org\nPassword Type:\tPLAIN\nPassword:\t--\nUserID:\t\t--\nGroupID:\t--\nHome:\t\t--\nQuota:\t\t*:bytes=300M\nEnabled:\ttrue\n"
	args = []string{"-d", dbfile, "show", "mailbox", "jeff@pobox.org"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Edit2 jeff@pobox.org: Unexpected error, %s", err)
	}
	if out != expectedOut {
		fmt.Printf("jeff@pobox.org: len out=%d, len expected=%d\n", len(out), len(expectedOut))
		t.Errorf("Edit2 jeff@pobox.org: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Edit2 jeff@pobox.org: did not expect error output, got %s", errout)
	}

	// now export it
	exportList := "jeff@pobox.org:{PLAIN}*::::::userdb_quota_rule=*:bytes=300M mbox_enabled=true\n"
	args = []string{"-d", dbfile, "export", "mailbox"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Export mailbox: Unexpected error, %s", err)
	}
	if out != exportList {
		t.Errorf("Export mailbox: Expected export list[%d](%s), got [%d](%s)",
			len(exportList), exportList, len(out), out)
	}
	if errout != "" {
		t.Errorf("Export mailbox: Expected no error output, got %s", errout)
	}

	// try import from stdin. Domain has already tested -i
	args = []string{"-d", dbfile, "import", "mailbox"}
	inputStr := `
# only one new user
dave@pobox.org:{sha256}HJJJYGB:56:83::dave::userdb_quota_rule=*:bytes=40G mbox_enabled=false
`
	out, errout, err = doTest(rootCmd, inputStr, args)
	if err != nil {
		t.Errorf("Import of dave@pobox.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Import of dave@pobox.org: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Import of dave@pobox.org: Expected no error output, got %s", errout)
	}
	// check import
	expectedOut = "Name:\t\tdave@pobox.org\nPassword Type:\tSHA256\nPassword:\tHJJJYGB\nUserID:\t\t56\nGroupID:\t83\nHome:\t\tdave\nQuota:\t\t*:bytes=40G\nEnabled:\tfalse\n"
	args = []string{"-d", dbfile, "show", "mailbox", "dave@pobox.org"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Import dave@pobox.org: Unexpected error, %s", err)
	}
	if out != expectedOut {
		fmt.Printf("dave@pobox.org: len out=%d, len expected=%d\n", len(out), len(expectedOut))
		t.Errorf("import dave@pobox.org: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("import dave@pobox.org: did not expect error output, got %s", errout)
	}

	// try export of both
	exportList = `dave@pobox.org:{SHA256}HJJJYGB:56:83::dave::userdb_quota_rule=*:bytes=40G mbox_enabled=false
jeff@pobox.org:{PLAIN}*::::::userdb_quota_rule=*:bytes=300M mbox_enabled=true
`
	args = []string{"-d", dbfile, "export", "mailbox"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Export mailboxes: Unexpected error, %s", err)
	}
	if out != exportList {
		t.Errorf("Export mailboxes: Expected export list[%d](%s), got [%d](%s)",
			len(exportList), exportList, len(out), out)
	}
	if errout != "" {
		t.Errorf("Export mailboxes: Expected no error output, got %s", errout)
	}

	// try delete
	args = []string{"-d", dbfile, "delete", "mailbox", "jeff@pobox.org"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Delete jeff@pobox.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Delete jeff@pobox.org: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Delete jeff@pobox.org: Expected no error output, got %s", errout)
	}
	// Now see if it is still there
	args = []string{"-d", dbfile, "show", "mailbox", "jeff@pobox.org"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Show jeff@pobox.org: should have failed")
	} else if err != maildb.ErrMdbAddressNotFound {
		t.Errorf("Show of jeff@pobox.org in good DB: Unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Show of jeff@pobox.org: Expected formatted help output, got nothing")
	}
	if !strings.Contains(errout, "address not found") {
		t.Errorf("Show of jeff@pobox.org: Expected error output, got %s", errout)
	}

}

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

// TestAccess
func TestAccess(t *testing.T) {
	var (
		err         error
		dir         string
		dbfile      string
		args        []string
		out, errout string
	)

	fmt.Println("TestAccess")

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

	// Test "add access". Add some access rules
	args = []string{"-d", dbfile, "add", "access", "default", "x-permit"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Add default: unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Add default: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add default: did not expect error output, got %s", errout)
	}
	args = []string{"-d", dbfile, "show", "access", "default"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of default: Unexpected error, %s", err)
	}
	if out != "Name:\tdefault\nAction:\tx-permit\n" {
		t.Errorf("Show of default: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Show of default: did not expect error output, got %s", errout)
	}

	// edit default but forget the action option
	args = []string{"-d", dbfile, "edit", "access", "default"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Edit default without action option should have failed")
	} else if err.Error() != "action option for access edit not set" {
		t.Errorf("Edit default without action got unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Edit default without action should have generated output")
	}
	if errout == "" {
		t.Errorf("Edit default without action should have generated error output")
	}

	// now change the action
	args = []string{"-d", dbfile, "edit", "access", "default", "--action", "foo"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Edit default: unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Edit default: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Edit default: did not expect error output, got %s", errout)
	}
	// now check the change
	args = []string{"-d", dbfile, "show", "access", "default"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of new default: Unexpected error, %s", err)
	}
	if out != "Name:\tdefault\nAction:\tfoo\n" {
		t.Errorf("Show of new default: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Show of new default: did not expect error output, got %s", errout)
	}

	// add another
	args = []string{"-d", dbfile, "add", "access", "spam", "x-spammer"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Add spam: unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Add spam: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add spam: did not expect error output, got %s", errout)
	}

	// edit spam but forget the action option
	// doctor accessAction to reset state. Use empty string to force an empty string error
	accessAction = ""
	args = []string{"-d", dbfile, "edit", "access", "spam"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Edit spam without action option should have failed")
	} else if err != maildb.ErrMdbAccessBadAction {
		t.Errorf("Edit spam without action got unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Edit spam without action should have generated output")
	}
	if errout == "" {
		t.Errorf("Edit spam without action should have generated error output")
	}

	// test import just stdin is enough here...
	args = []string{"-d", dbfile, "import", "access"}
	inputStr := `
# some access rules
polite x-polite
nothing x-nothing
`
	out, errout, err = doTest(rootCmd, inputStr, args)
	if err != nil {
		t.Errorf("Import access unexpectedly failed")
	}
	if out != "" {
		t.Errorf("Import access: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Import access: did not expect error output, got %s", errout)
	}

	// test bad import just stdin is enough here...
	args = []string{"-d", dbfile, "import", "access"}
	inputStr = `
# some bogus access rules
bogus
empty
`
	out, errout, err = doTest(rootCmd, inputStr, args)
	if err != nil {
		if !strings.Contains(err.Error(), "only one token") {
			t.Errorf("Import access unexpected error, %s", err)
		}
	}
	if out == "" {
		t.Errorf("Import: expected output, got %s", out)
	}
	if errout == "" {
		t.Errorf("Import: expected error output, got %s", errout)
	}

	// test export
	exportList := "default foo\n" +
		"nothing x-nothing\n" +
		"polite x-polite\n" +
		"spam x-spammer\n"
	args = []string{"-d", dbfile, "export", "access"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Export access: unexpected error, %s", err)
	}
	if out != exportList {
		t.Errorf("Export access: expected export list[%d](%s), got [%d](%s)",
			len(exportList), exportList, len(out), out)
	}
	if errout != "" {
		t.Errorf("Export access: unexpected error output, got %s", errout)
	}

	// just export polite
	exportList = "polite x-polite\n"
	args = []string{"-d", dbfile, "export", "access", "polite"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Export access polite: unexpected error, %s", err)
	}
	if out != exportList {
		t.Errorf("Export access polite: expected export list[%d](%s), got [%d](%s)",
			len(exportList), exportList, len(out), out)
	}
	if errout != "" {
		t.Errorf("Export access polite: unexpected error output, got %s", errout)
	}

	// delete one of them
	args = []string{"-d", dbfile, "delete", "access", "nothing"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Delete nothing: unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Delete nothing: unexpected output, %s", out)
	}
	if errout != "" {
		t.Errorf("Delete nothing: unexpected err output, %s", errout)
	}

	// see if it really got deleted
	exportList = "default foo\n" +
		"polite x-polite\n" +
		"spam x-spammer\n"
	args = []string{"-d", dbfile, "export", "access"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Export access after delete of nothing: unexpected error, %s", err)
	}
	if out != exportList {
		t.Errorf("Export access after delete of nothing: expected export list[%d](%s), got [%d](%s)",
			len(exportList), exportList, len(out), out)
	}
	if errout != "" {
		t.Errorf("Export access after delete of nothing: unexpected error output, got %s", errout)
	}

	// now delete a bogus entry
	args = []string{"-d", dbfile, "delete", "access", "bogus"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Delete bogus: should have failed")
	} else if !strings.Contains(err.Error(), "Access not found") {
		t.Errorf("Delete access bogus: unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Delete access bogus: no output, %s", out)
	}
	if errout == "" {
		t.Errorf("Delete access bogus: unexpected err output, %s", errout)
	}

}

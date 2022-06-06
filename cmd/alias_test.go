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

// TestAliasCmds
func TestAliasCmds(t *testing.T) {
	var (
		err         error
		dir         string
		dbfile      string
		args        []string
		out, errout string
	)

	fmt.Println("TestAliasCmds")

	// Make a database and test it
	dir, err = ioutil.TempDir("", "TestAlias-*")
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

	// Add some aliases
	args = []string{"-d", dbfile, "add", "alias", "postmaster", "bill@sysops"}
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
	args = []string{"-d", dbfile, "show", "alias", "postmaster"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of postmaster in good DB: Unexpected error, %s", err)
	}
	if out != "Alias:\t\tpostmaster\nTargets:\tbill@sysops\n" {
		t.Errorf("Show of postmaster in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of postmaster in good DB: did not expect error output, got %s", errout)
	}
	// Add another target/recipient
	args = []string{"-d", dbfile, "add", "alias", "postmaster", "dave@noc"}
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
	args = []string{"-d", dbfile, "show", "alias", "postmaster"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of postmaster in good DB: Unexpected error, %s", err)
	}
	if out != "Alias:\t\tpostmaster\nTargets:\tbill@sysops\n\tdave@noc\n" {
		t.Errorf("Show of postmaster in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of postmaster in good DB: did not expect error output, got %s", errout)
	}

	// A second alias
	args = []string{"-d", dbfile, "add", "alias", "root", "mary@noc"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Add root: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Add root: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add root: did not expect error output, got %s", errout)
	}
	args = []string{"-d", dbfile, "show", "alias", "root"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of root in good DB: Unexpected error, %s", err)
	}
	if out != "Alias:\t\troot\nTargets:\tmary@noc\n" {
		t.Errorf("Show of root in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of root in good DB: did not expect error output, got %s", errout)
	}

	// Add some virtual aliases
	args = []string{"-d", dbfile, "add", "virtual", "bruce@e-street", "paul@beatles"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Add bruce@e-street: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Add bruce@e-street: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add bruce@e-street: did not expect error output, got %s", errout)
	}
	args = []string{"-d", dbfile, "show", "virtual", "bruce@e-street"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of bruce@e-street in good DB: Unexpected error, %s", err)
	}
	if out != "Virtual Alias:\tbruce@e-street\nTargets:\tpaul@beatles\n" {
		t.Errorf("Show of bruce@e-street in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of bruce@e-street in good DB: did not expect error output, got %s", errout)
	}

	// Another recipient
	args = []string{"-d", dbfile, "add", "virtual", "bruce@e-street", "dorothy@oz"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Add bruce@e-street: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Add bruce@e-street: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add bruce@e-street: did not expect error output, got %s", errout)
	}
	args = []string{"-d", dbfile, "show", "virtual", "bruce@e-street"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of bruce@e-street in good DB: Unexpected error, %s", err)
	}
	if out != "Virtual Alias:\tbruce@e-street\nTargets:\tpaul@beatles\n\t\tdorothy@oz\n" {
		t.Errorf("Show of bruce@e-street in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of bruce@e-street in good DB: did not expect error output, got %s", errout)
	}

	// another virtual alias
	args = []string{"-d", dbfile, "add", "virtual", "mark@fuse.org", "tina@fea"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Add mark@fuse.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Add mark@fuse.org: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Add mark@fuse.org: did not expect error output, got %s", errout)
	}
	args = []string{"-d", dbfile, "show", "virtual", "mark@fuse.org"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of mark@fuse.org in good DB: Unexpected error, %s", err)
	}
	if out != "Virtual Alias:\tmark@fuse.org\nTargets:\ttina@fea\n" {
		t.Errorf("Show of mark@fuse.org in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of mark@fuse.org in good DB: did not expect error output, got %s", errout)
	}

	// edit a virtual to remove one and add two
	args = []string{"-d", dbfile, "edit", "virtual", "mark@fuse.org",
		"-r", "tina@fea", "-a", "john@wayne", "-a", "jimmy@stewart"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Edit mark@fuse.org: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Edit mark@fuse.org: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Edit mark@fuse.org: did not expect error output, got %s", errout)
	}
	args = []string{"-d", dbfile, "show", "virtual", "mark@fuse.org"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of mark@fuse.org in good DB: Unexpected error, %s", err)
	}
	if out != "Virtual Alias:\tmark@fuse.org\nTargets:\tjohn@wayne\n\t\tjimmy@stewart\n" {
		t.Errorf("Show of mark@fuse.org in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of mark@fuse.org in good DB: did not expect error output, got %s", errout)
	}

	// edit an alias adding two and removing one
	args = []string{"-d", dbfile, "edit", "alias", "root",
		"-r", "mary@noc", "-a", "joe@tech", "-a", "ray@ops"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Edit root: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Edit root: did not expect output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Edit root: did not expect error output, got %s", errout)
	}
	args = []string{"-d", dbfile, "show", "alias", "root"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Show of root in good DB: Unexpected error, %s", err)
	}
	if out != "Alias:\t\troot\nTargets:\tjoe@tech\n\tray@ops\n" {
		t.Errorf("Show of root in good DB: did not get expected output, got %s", out)
	}
	if errout != "" {
		t.Errorf("show of root in good DB: did not expect error output, got %s", errout)
	}

	// import some aliases
	args = []string{"-d", dbfile, "import", "alias"}
	inputStr := `
abuse: mike@home, bill@work,
 tech
spam: tech, jones@site
`
	out, errout, err = doTest(rootCmd, inputStr, args)
	if err != nil {
		t.Errorf("Import of aliases: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Import of aliases: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Import of aliases: Expected no error output, got %s", errout)
	}

	// import virtual aliases from a file
	args = []string{"-d", dbfile, "import", "virtual", "-i", "./test_virtuals.txt"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Import of virtual aliases from file: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Import of virtual aliases from file: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Import of virtual aliases from file: Expected no error output, got %s", errout)
	}

	// export aliases
	exportList := "abuse: mike@home, bill@work, tech\n" +
		"postmaster: bill@sysops, dave@noc\n" +
		"root: joe@tech, ray@ops\n" +
		"spam: tech, jones@site\n"
	args = []string{"-d", dbfile, "export", "alias"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Export aliases: Unexpected error, %s", err)
	}
	if out != exportList {
		t.Errorf("Export aliases: Expected export list[%d](%s), got [%d](%s)",
			len(exportList), exportList, len(out), out)
	}
	if errout != "" {
		t.Errorf("Export aliases: Expected no error output, got %s", errout)
	}

	// export virtuals
	exportList = "abuse@disney walt+abuse@disney\n" +
		"walt@disney spamalot\n" +
		"bruce@e-street paul@beatles, dorothy@oz\n" +
		"mark@fuse.org john@wayne, jimmy@stewart\n" +
		"mickey@mouse minnie@mouse, goofy\n" +
		"roadrunner@wb coyote@wb\n"
	args = []string{"-d", dbfile, "export", "virtual"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Export virtual aliases: Unexpected error, %s", err)
	}
	if out != exportList {
		t.Errorf("Export virtual aliases: Expected export list[%d](%s), got [%d](%s)",
			len(exportList), exportList, len(out), out)
	}
	if errout != "" {
		t.Errorf("Export virtual aliases: Expected no error output, got %s", errout)
	}

	// delete an alias
	args = []string{"-d", dbfile, "delete", "alias", "mcgoo"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Delete mcgoo: should have failed")
	} else if err != maildb.ErrMdbNotAlias {
		t.Errorf("Delete mcgoo: Unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Delete mcgoo: Expected formatted help output, got nothing")
	}
	if !strings.Contains(errout, "address is not an alias") {
		t.Errorf("Delete mcgoo: Expected error output, got %s", errout)
	}

	args = []string{"-d", dbfile, "delete", "alias", "spam"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Delete spam: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Delete spam: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Delete spam: Expected no error output, got %s", errout)
	}

	// Now see if it is still there
	args = []string{"-d", dbfile, "show", "alias", "spam"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Show spam: should have failed")
	} else if err != maildb.ErrMdbAddressNotFound {
		t.Errorf("Show of spam in good DB: Unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Show of spam: Expected formatted help output, got nothing")
	}
	if !strings.Contains(errout, "address not found") {
		t.Errorf("Show of spam: Expected error output, got %s", errout)
	}

	// delete a virtual
	args = []string{"-d", dbfile, "delete", "virtual", "lost@art"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Delete lost@art: should have failed")
	} else if err != maildb.ErrMdbNotAlias {
		t.Errorf("Delete lost@art: Unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Delete lost@art: Expected formatted help output, got nothing")
	}
	if !strings.Contains(errout, "address is not an alias") {
		t.Errorf("Delete lost@art: Expected error output, got %s", errout)
	}

	args = []string{"-d", dbfile, "delete", "virtual", "roadrunner@wb"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Delete roadrunner@wb: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Delete roadrunner@wb: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Delete roadrunner@wb: Expected no error output, got %s", errout)
	}

	// Now see if it is still there
	args = []string{"-d", dbfile, "show", "virtual", "roadrunner@wb"} // now look it up
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Show roadrunner@wb: should have failed")
	} else if err != maildb.ErrMdbAddressNotFound {
		t.Errorf("Show of roadrunner@wb in good DB: Unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Show of roadrunner@wb: Expected formatted help output, got nothing")
	}
	if !strings.Contains(errout, "address not found") {
		t.Errorf("Show of roadrunner@wb: Expected error output, got %s", errout)
	}

}

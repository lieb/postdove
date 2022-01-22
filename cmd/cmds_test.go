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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	//"github.com/lieb/postdove/maildb"
	"github.com/spf13/cobra"
)

// doTest
func doTest(cmd *cobra.Command, stdIn string, args []string) (string, string, error) {
	var (
		err error
	)

	inbuf := bytes.NewBufferString(stdIn)
	outbuf := bytes.NewBufferString("")
	errbuf := bytes.NewBufferString("")
	cmd.SetIn(inbuf)
	cmd.SetOut(outbuf)
	cmd.SetErr(errbuf)
	cmd.SetArgs(args)
	err = cmd.Execute()
	out, e := ioutil.ReadAll(outbuf)
	if e != nil {
		return "", "", fmt.Errorf("doTest: ReadAll out failed, %s", e)
	}
	errout, e := ioutil.ReadAll(errbuf)
	if e != nil {
		return "", "", fmt.Errorf("doTest: ReadAll err failed, %s", e)
	}
	return string(out), string(errout), err
}

// Test_Cmds
// Test basic commmands infrastructure and database creation
func Test_Cmds(t *testing.T) {
	var (
		err         error
		dir         string
		dbfile      string
		args        []string
		out, errout string
	)

	fmt.Println("Test_Cmds")

	// first the TUI (no args at all)
	args = []string{"foo", "-x"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Bogus command foo should have failed")
	} else if err.Error() != "unknown command \"foo\" for \"postdove\"" {
		t.Errorf("Bogus command foo unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Bogus foo: did not expect output, got %s", out)
	}
	if errout == "" {
		t.Errorf("Bogus foo: should have gotten the formatted error and a help message")
	}

	// Make a database and test it
	dir, err = ioutil.TempDir("", "TestCmds-*")
	defer os.RemoveAll(dir)
	dbfile = filepath.Join(dir, "test.db")

	// Test create with no flags/args. Assumes test host does not have a dovecot installation
	// so system DB file doesn't exist or unwriteable since we are not root
	args = []string{"create"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Create with no flags should have failed")
	} else if !strings.Contains(err.Error(), "database file: no such file or directory") {
		t.Errorf("Create no flags: Unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Create no flags: expect help message to output, got nothing")
	}
	if errout == "" {
		t.Errorf("Create no flags: expected formatted error output")
	}

	// Create with unwriteable DB but good schema
	args = []string{"create", "-s", "../maildb/files/schema.sql"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Create with bogus schema should have failed")
	} else if !strings.Contains(err.Error(), "unable to open database file: no such file or directory") {
		t.Errorf("Create bad db: Unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Create bad db: expect help message to output, got nothing")
	}
	if errout == "" {
		t.Errorf("Create bad db: expected formatted error output")
	}

	// Create with usable DB but bogus schema
	args = []string{"create", "-d", dbfile, "-s", "/foo/bogus.schema"}
	out, errout, err = doTest(rootCmd, "", args)
	if err == nil {
		t.Errorf("Create with bogus schema should have failed")
	} else if !strings.Contains(err.Error(), "open /foo/bogus.schema: no such file or directory") {
		t.Errorf("Create bogus schema: Unexpected error, %s", err)
	}
	if out == "" {
		t.Errorf("Create bogus schema: expect help message to output, got nothing")
	}
	if errout == "" {
		t.Errorf("Create bogus schema: expected formatted error output")
	}
	schemaFile = "" // cobra warning! can't run init more than once so hack back to something safe...

	// Now create a good database with no initialization (yet)
	args = []string{"create", "-d", dbfile, "--no-locals", "--no-aliases"}
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

}

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
	//"strings"
	"reflect"
	"testing"

	"github.com/lieb/postdove/maildb"
)

// compareCol
// Most type found in Sqlite...
func compareCol(exp, res interface{}) bool {
	if reflect.TypeOf(exp) != reflect.TypeOf(res) {
		// fmt.Printf("compareCol: exp=%T %v != res=%T %v\n", exp, exp, res, res)
		return false
	}
	// fmt.Printf("compareCol: exp=%T %v, res=%T %v\n", exp, exp, res, res)
	switch x := exp.(type) {
	case int64:
		return x == res.(int64)
	case int:
		return x == res.(int)
	case float64:
		return x == res.(float64)
	case bool:
		return x == res.(bool)
	case string:
		return x == res.(string)
	default:
		panic(fmt.Sprintf("CompareCol: unsupported type %T: %v", x, x))
	}
}

// queryView
// and compare to expected result
func queryView(mdb *maildb.MailDB, q string, expected []maildb.QueryRes) error {
	var (
		err     error
		match   int
		viewRes []maildb.QueryRes
	)

	if viewRes, err = mdb.Query(q); err != nil {
		return fmt.Errorf("query failed, %s", err)
	}
	if len(viewRes) != len(expected) {
		return fmt.Errorf("Expected %d results(%v), got %d(%v)",
			len(expected), expected, len(viewRes), viewRes)
	}
	for _, row := range viewRes {
		// fmt.Printf("r = %v\n", row)
		cols := row.NumColumns()
		for _, exp := range expected {
			step := exp.NumColumns()
			if step != cols { // sanity check. SQL _should_always_ return MxN table...
				return fmt.Errorf("Expected %d result columns, got %d",
					step, cols) // probably screwed up expectedRes
			}
			for k, v := range exp { // step through all the expected columns
				if compareCol(v, row[k]) {
					// fmt.Printf("Match r[%s] v %v\n", k, v)
					step--
				} else {
					break // not this expected row
					// fmt.Printf("No match v %v != r[%s] %v\n", v, k, r[k])
				}
			}
			if step == 0 {
				match++
				break // found a match so this row is done.
			}
		}
	}
	if match != len(viewRes) {
		err = fmt.Errorf("Expected %d matched rows from view, got %d(%v)",
			len(viewRes), match, viewRes)
	}
	return err
}

// TestViews
func TestViews(t *testing.T) {
	var (
		err         error
		dir         string
		dbfile      string
		args        []string
		out, errout string
		q           string
		expectedRes []maildb.QueryRes
	)

	fmt.Println("TestViews")

	// Make a database and test it
	dir, err = ioutil.TempDir("", "TestViews-*")
	defer os.RemoveAll(dir)
	dbfile = filepath.Join(dir, "test.db")

	// Now create a good database with init data
	// using CLI commands
	args = []string{"create", "-d", dbfile}
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

	// load some access rules and transports...
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

	// Load some domains
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

	// Load some addresses (for access and transport)
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

	// Load some (/etc/aliases) aliases
	args = []string{"-d", dbfile, "import", "alias", "-i", "./test_aliases.txt"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Import of aliases from file: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Import of aliases from file: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Import of aliases from file: Expected no error output, got %s", errout)
	}

	// Load some virtuals
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

	// Populate with some mailboxes
	args = []string{"-d", dbfile, "import", "mailbox", "-i", "./test_mailboxes.txt"}
	out, errout, err = doTest(rootCmd, "", args)
	if err != nil {
		t.Errorf("Import of mailboxes: Unexpected error, %s", err)
	}
	if out != "" {
		t.Errorf("Import of mailboxes: Expected no output, got %s", out)
	}
	if errout != "" {
		t.Errorf("Import of mailboxes: Expected no error output, got %s", errout)
	}

	// now add an alias for root to a real person mailbox
	args = []string{"-d", dbfile, "add", "alias", "root", "jeff@pobox.org"}
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

	// Open the database directly so we can test views
	if mdb, err = maildb.NewMailDB(dbfile); err != nil {
		t.Errorf("Could not reopen database for view testing, %s", err)
		return
	}
	defer mdb.Close()

	// Postfix related view testing

	// Test domain classes
	// first one that is correct followed by one that is not for that class
	fmt.Printf("Domain class types\n")
	q = `
SELECT name FROM internet_domain WHERE name is 'zip.com'
`
	expectedRes = []maildb.QueryRes{
		{
			"name": "zip.com",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup localhost as local class: %s", err)
	}

	q = `
SELECT name FROM internet_domain WHERE name is 'localhost'
`
	expectedRes = []maildb.QueryRes{}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup localhost as internet class: %s", err)
	}

	q = `
SELECT name FROM local_domain WHERE name is 'localhost'
`
	expectedRes = []maildb.QueryRes{
		{
			"name": "localhost",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup localhost as local class: %s", err)
	}

	q = `
SELECT name FROM local_domain WHERE name is 'dish.net'
`
	expectedRes = []maildb.QueryRes{}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup dish.net as local class: %s", err)
	}

	q = `
SELECT name FROM relay_domain WHERE name is 'dish.net'
`
	expectedRes = []maildb.QueryRes{
		{
			"name": "dish.net",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup dish.net as relay class: %s", err)
	}

	q = `
SELECT name FROM relay_domain WHERE name is 'localhost'
`
	expectedRes = []maildb.QueryRes{}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup localhost as relay class: %s", err)
	}

	q = `
SELECT name FROM virtual_domain WHERE name is 'run.com'
`
	expectedRes = []maildb.QueryRes{
		{
			"name": "run.com",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup run.com as virtual class: %s", err)
	}

	q = `
SELECT name FROM virtual_domain WHERE name is 'dish.net'
`
	expectedRes = []maildb.QueryRes{}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup dish.net as virtual class: %s", err)
	}

	q = `
SELECT name FROM vmailbox_domain WHERE name is 'pobox.org'
`
	expectedRes = []maildb.QueryRes{
		{
			"name": "pobox.org",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup pobox.org as vmailbox class: %s", err)
	}

	q = `
SELECT name FROM vmailbox_domain WHERE name is 'run.com'
`
	expectedRes = []maildb.QueryRes{}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup run.com as vmailbox class: %s", err)
	}

	// Test Access restriction lookups
	// used for smtpd_recipient_access and other postfix rules

	// domain access
	fmt.Printf("Domain access\n")
	q = `
SELECT access_key FROM domain_access WHERE domain_name IS 'dish.net'
		`
	expectedRes = []maildb.QueryRes{
		{
			"access_key": "x-dump",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup dish.net access: %s", err)
	}

	q = `
SELECT access_key FROM domain_access WHERE domain_name IS 'zip.com'
`
	expectedRes = []maildb.QueryRes{} // none specified so no row...
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup zip.com access: %s", err)
	}

	// user@domain access
	fmt.Printf("user@domain access\n")

	// first, key set for address
	q = `
SELECT access_key FROM address_access
WHERE username = 'mary' AND domain_name = 'little.lamb'
`
	expectedRes = []maildb.QueryRes{
		{
			"access_key": "x-dump",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup of mary@little.lamb access: %s", err)
	}

	// next, key for domain
	q = `
SELECT access_key FROM address_access
WHERE username = 'gramma' AND domain_name = 'cottage'
`
	expectedRes = []maildb.QueryRes{
		{
			"access_key": "x-permit",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup of gramma@cottage access: %s", err)
	}

	// and then no key
	q = `
SELECT access_key FROM address_access
WHERE username = 'dave' AND domain_name = 'wm.com'
`
	expectedRes = []maildb.QueryRes{}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup of dave@wm.com access: %s", err)
	}

	// Test Transport lookups
	fmt.Printf("Transport lookups\n")

	// Lookup domain transport
	q = `
SELECT transport FROM domain_transport WHERE domain_name IS 'pobox.org'
`
	expectedRes = []maildb.QueryRes{
		{
			"transport": "lmtp:localhost:24",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup of pobox.org transport: %s", err)
	}

	q = `
SELECT transport FROM domain_transport WHERE domain_name IS 'wm.com'
`
	expectedRes = []maildb.QueryRes{
		{
			"transport": ":", // default case coalesce...
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup of pobox.org transport: %s", err)
	}

	// Lookup domain without transport
	q = `
SELECT transport FROM domain_transport WHERE domain_name IS 'run.com'
`
	expectedRes = []maildb.QueryRes{}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup of pobox.org transport: %s", err)
	}

	// Lookup user@domain transport
	q = `
SELECT transport FROM address_transport
WHERE username IS 'gramma' AND domain_name IS 'cottage'
`
	expectedRes = []maildb.QueryRes{
		{
			"transport": "smtp:faraway.net:25",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup of gramma@cottage transport: %s", err)
	}

	// Lookup user@domain for domain transport
	q = `
SELECT transport FROM address_transport
WHERE username IS 'dave' AND domain_name IS 'wm.com'
`
	expectedRes = []maildb.QueryRes{
		{
			"transport": ":",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup of dave@wm.com transport: %s", err)
	}

	// Lookup user@domain without transport
	q = `
SELECT transport FROM address_transport
WHERE username IS 'wolf' AND domain_name IS 'forest'
`
	expectedRes = []maildb.QueryRes{}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup of wolf@forest transport: %s", err)
	}

	// Test domain class lookups
	fmt.Printf("Domain lookups\n")

	// Domain internet

	// Test alias lookups
	fmt.Printf("Local Alias lookups\n")
	q = `
SELECT recipient FROM etc_aliases WHERE local_user IS 'postmaster'
`
	expectedRes = []maildb.QueryRes{
		{
			"recipient": "root",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup alias postmaster: %s", err)
	}

	q = `
SELECT recipient FROM etc_aliases WHERE local_user IS 'root'
`
	expectedRes = []maildb.QueryRes{
		{
			"recipient": "marc",
		},
		{
			"recipient": "bill@noc",
		},
		{
			"recipient": "\"| cat - >/dev/null\"",
		},
		{
			"recipient": "jeff@pobox.org",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup alias root: %s", err)
	}

	q = `
SELECT recipient FROM etc_aliases WHERE local_user IS 'bogus'
`
	expectedRes = []maildb.QueryRes{}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup alias bogus: %s", err)
	}

	// Test virtual lookups
	// first a simple one
	fmt.Printf("Virtual alias lookups\n")
	q = `
SELECT recipient FROM virt_alias WHERE mailbox IS 'roadrunner' AND domain_name IS 'wb'
`
	expectedRes = []maildb.QueryRes{
		{
			"recipient": "coyote@wb",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup virtual alias roadrunner@wb: %s", err)
	}

	// a list
	q = `
SELECT recipient FROM virt_alias WHERE mailbox IS 'mickey' AND domain_name IS 'mouse'
`
	expectedRes = []maildb.QueryRes{
		{
			"recipient": "minnie@mouse",
		},
		{
			"recipient": "goofy",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup virtual alias mickey@mouse: %s", err)
	}

	// and a target with extension
	q = `
SELECT recipient FROM virt_alias WHERE mailbox IS 'abuse' AND domain_name IS 'disney'
`
	expectedRes = []maildb.QueryRes{
		{
			"recipient": "walt+abuse@disney",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup virtual alias abuse@disney: %s", err)
	}

	// and last, a bogus one
	q = `
SELECT recipient FROM virt_alias WHERE mailbox IS 'bogus' AND domain_name IS 'example.com'
`
	expectedRes = []maildb.QueryRes{}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup virtual alias bogus@example.com: %s", err)
	}

	// Dovecot related view testing

	// Lookup all users
	fmt.Printf("Lookup all users\n")
	q = `
SELECT username, domain FROM user_mailbox
`
	expectedRes = []maildb.QueryRes{
		{
			"username": "jeff",
			"domain":   "pobox.org",
		},
		{
			"username": "dave",
			"domain":   "pobox.org",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup all users: %s", err)
	}

	// Lookup a prefetch password
	fmt.Printf("Lookup prefetch of jeff@pobox.org\n")
	q = `
SELECT username, domain, password,
  uid as userdb_uid, gid as userdb_gid, home as userdb_home,
  quota_rule AS userdb_quota_rule
  FROM user_mailbox WHERE username = 'jeff' AND domain = 'pobox.org'
`
	expectedRes = []maildb.QueryRes{
		{
			"username":          "jeff",
			"domain":            "pobox.org",
			"password":          "{PLAIN}*",
			"userdb_uid":        int64(99),
			"userdb_gid":        int64(99),
			"userdb_home":       "",
			"userdb_quota_rule": "*:bytes=300M",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup a prefetch password: %s", err)
	}

	// Lookup user
	fmt.Printf("Lookup user\n")
	q = `
SELECT home, uid, gid, quota_rule
  FROM user_mailbox WHERE username = 'dave' AND domain = 'pobox.org'
`
	expectedRes = []maildb.QueryRes{
		{
			"home":       "dave",
			"uid":        int64(56),
			"gid":        int64(83),
			"quota_rule": "*:bytes=40G",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup user: %s", err)
	}

	// Lookup a bogus user
	fmt.Printf("Lookup bogus user\n")
	q = `
SELECT home, uid, gid, quota_rule
  FROM user_mailbox WHERE username = 'mary' AND domain = 'pobox.org'
`
	expectedRes = []maildb.QueryRes{} // no results
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Lookup bogus: %s", err)
	}

	// deny allow
	fmt.Printf("Deny allow\n")
	q = `
SELECT deny FROM user_deny
WHERE username = 'jeff' AND domain = 'pobox'
`
	expectedRes = []maildb.QueryRes{} // no result means allow
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Deny allow: %s", err)
	}

	// deny deny
	fmt.Printf("Deny deny\n")
	q = `
SELECT deny FROM user_deny
WHERE username = 'dave' AND domain = 'pobox.org'
`
	expectedRes = []maildb.QueryRes{
		{
			"deny": "true",
		},
	}
	if err = queryView(mdb, q, expectedRes); err != nil {
		t.Errorf("Deny deny: %s", err)
	}
}

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
	//"io/ioutil"
	//"os"
	//"path/filepath"
	"strings"
	"testing"
	//"github.com/lieb/postdove/maildb"
	//"github.com/spf13/cobra"
)

var (
	testNum int
	resLine int
)

// expected tokens and errors
type Res struct {
	test       string // name for error messages
	use        ImportType
	errcode    string     // every test terminates with a deliberate error
	importFile string     // input to parse
	tokens     [][]string // tokens per input line
}

var testSuite = []Res{
	// Simple file format. Also test comment stripping
	{
		test:    "Simple",
		use:     SIMPLE,
		errcode: "only one token",
		importFile: `
# A comment
   # an indented comment

key foo
key baz # baz is a sub
zip dent dump

lump
`,
		tokens: [][]string{
			[]string{"key", "foo"},
			[]string{"key", "baz"},
			[]string{"zip", "dent dump"},
			[]string{"lump"},
		},
	},
	// Simple file format. Also test comment stripping
	{
		test:    "Simple errors",
		use:     SIMPLE,
		errcode: "Indented but not a",
		importFile: `
# A comment
 indented # an indented non-continuation line is an error
`,
		tokens: [][]string{
			[]string{"indented"},
		},
	},
	// Postfix file format
	{
		test:    "Postfix",
		use:     POSTFIX,
		errcode: "got (baz bar)",
		importFile: `
foo baz, bar, zip
new late, old,
 really, old

skip line, for,
 extend, line,
 more, than,
 normal
# now forget a ','
bad foo, baz bar
`,
		tokens: [][]string{
			[]string{"foo", "baz", "bar", "zip"},
			[]string{"new", "late", "old", "really", "old"},
			[]string{"skip", "line", "for", "extend", "line", "more", "than", "normal"},
			[]string{"foo", "baz", "bar"},
		},
	},
	// Postfix file format with continuation error
	{
		test:    "Postfix errors",
		use:     POSTFIX,
		errcode: "expected 5 items, got 3",
		importFile: `
new late, old # and forget the trailing ','
 really, old
`,
		tokens: [][]string{
			[]string{"new", "late", "old", "really", "old"},
		},
	},
	// Aliases file format
	{
		test:    "Aliases",
		use:     ALIASES,
		errcode: "key must be followed by a ':'",
		importFile: `
bill: dave, charlie
dave: steve
steve mike
`,
		tokens: [][]string{
			[]string{"bill", "dave", "charlie"},
			[]string{"dave", "steve"},
			[]string{"steve", "mike"},
		},
	},
}

// test_worker test dummy to compare results to expected tokens
func test_worker(t []string) error {
	var (
		test Res
		err  error
	)

	test = testSuite[testNum]
	//fmt.Printf("test_worker: len t=%d, len tokens=%d\n", len(t), len(test.tokens[resLine]))
	//fmt.Printf("test_worker: t = %v, tokens = %v\n", t, test.tokens[resLine])
	if len(t) != len(test.tokens[resLine]) {
		err = fmt.Errorf("test_worker: expected %d items, got %d\n",
			len(test.tokens[resLine]), len(t))
	} else {
		for j, r := range test.tokens[resLine] {
			if t[j] != r {
				err = fmt.Errorf("test_worker: expected \"%s\", got (%s)", r, t[j])
			}
		}
	}
	resLine++
	return err
}

func testImport(t *testing.T, test Res) {
	var (
		err error
	)

	inbuf := bytes.NewBufferString(test.importFile)
	resLine = 0
	rootCmd.SetIn(inbuf)
	err = procImport(rootCmd, test.use, test_worker)
	if err == nil {
		if test.errcode != "" {
			t.Errorf("Test %s: Expected error %s, got success", test.test, test.errcode)
		}
	} else {
		var e = err.Error()

		if test.errcode == "" || !strings.Contains(e, test.errcode) {
			t.Errorf("Test %s: expected error \"%s\", got %s",
				test.test, test.errcode, err)
		}
	}
}

func Test_Import(t *testing.T) {
	var test Res

	fmt.Println("Test_Import")

	for testNum, test = range testSuite {
		fmt.Printf("Test_%s\n", test.test)
		testImport(t, test)
	}

}

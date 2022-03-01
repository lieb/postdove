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
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	inFilePath string
	inFile     *os.File
	inReader   io.Reader
	savedIn    io.Reader
)

// importRedirect
func importRedirect(cmd *cobra.Command, args []string) error {
	var (
		err error
	)

	if cmd.Flags().Changed("input") {
		if inFilePath == "-" {
			inFile = os.Stdin
		} else {
			if inFile, err = os.Open(inFilePath); err != nil {
				return err
			}
		}
		inReader = inFile
		savedIn = cmd.InOrStdin()
		cmd.SetIn(inReader)
	}
	return nil
}

// importClose
func importClose(cmd *cobra.Command, args []string) error {
	var (
		err error
	)

	if savedIn != nil {
		if err = inFile.Close(); err != nil {
			return err
		}
		cmd.SetIn(savedIn)
		savedIn = nil
	}
	return nil
}

// import helpers

type ImportType int

const (
	SIMPLE  ImportType = iota // key WS+ text
	POSTFIX                   // key WS+ [token ',' WS*]+
	ALIASES                   // key ':' WS* [token ',' WS*]+
	PWFILE                    // fields sep by ':'
)

// procImport
// Read lines from the import stream and tokenize them following the
// aliases(5) and postfix postmap rules
// We are a bit more relaxed here. You can have a comment at the end of a line
// We can get two errors, one syntax and the other from the "worker" we return both
func procImport(cmd *cobra.Command, use ImportType, worker func([]string) error) error {
	var (
		lines   = bufio.NewScanner(cmd.InOrStdin())
		line    string
		segment string
		lineno  int
		imports int
		err     error
	)

	for lines.Scan() {
		segment = lines.Text()
		lineno++
		com := strings.IndexByte(segment, '#')
		if com != -1 { // a comment, strip it
			segment = segment[0:com]
		}
		segment = strings.TrimRight(segment, " \t")
		if segment == "" {
			continue // blank line
		}
		if segment[0] == ' ' || segment[0] == '\t' {
			if line == "" {
				err = fmt.Errorf(
					"At line %d: Indented but not a continuation line", lineno)
				break
			}
			line = line + " " + strings.TrimLeft(segment, " \t")
			continue
		} else if line == "" {
			line = strings.TrimLeft(segment, " \t")
			continue
		} else {
			if err = procLine(line, use, worker); err != nil {
				err = fmt.Errorf("Near line %d: %s", lineno, err)
				break
			}
			imports++
			line = segment
		}
	}
	if err == nil && line != "" {
		if err = procLine(line, use, worker); err != nil {
			err = fmt.Errorf("Near line %d: %s", lineno, err)
		}
		imports++
	}
	if e := lines.Err(); e != nil && err == nil {
		err = e
	}
	if imports == 0 && err == nil {
		err = fmt.Errorf("At line %d: nothing found to import", lineno)
	}
	return err
}

// procLine
func procLine(line string, use ImportType, worker func([]string) error) error {
	var (
		tokens []string
		err    error
	)

	// functions for FieldsFunc delimiting
	onComma := func(c rune) bool { return c == ',' }
	onColon := func(c rune) bool { return c == ':' }

	sp := strings.IndexAny(line, " \t") // split off first token
	switch use {                        // requested line syntax
	case SIMPLE: // key SP value string NL | key NL
		if sp == -1 {
			tokens = []string{line}
		} else {
			tokens = []string{line[0:sp],
				strings.Trim(line[sp:], " \t")}
		}
	case POSTFIX: // key SP value [',' value]* NL | key NL
		if sp == -1 {
			tokens = []string{line}
		} else {
			tokens = []string{line[0:sp]} // first the key by WS
			vals := strings.FieldsFunc(line[sp:], onComma)
			for _, v := range vals {
				v = strings.Trim(v, " \t")
				tokens = append(tokens, v) // then the tokens by ','
			}
		}
	case ALIASES: // key ':' value [',' value]*
		if !strings.Contains(line, ":") {
			return fmt.Errorf("key must be followed by a ':'")
		}
		kv := strings.FieldsFunc(line, onColon) // first key by ':'
		if len(kv) < 2 {
			return fmt.Errorf("no values for key")
		} else {
			tokens = []string{kv[0]}
			vals := strings.FieldsFunc(kv[1], onComma) // then tokens by ','
			for _, v := range vals {
				v = strings.Trim(v, " \t")
				tokens = append(tokens, v) // then the tokens by ','
			}
		}
	case PWFILE: // field [':' field]+
		if !strings.Contains(line, ":") {
			return fmt.Errorf("line must be fields separated by a ':'")
		}
		tokens = strings.Split(line, ":")
	}
	if len(tokens) > 0 {
		err = worker(tokens)
	} else {
		err = fmt.Errorf("Unrecognized import line")
	}
	return err
}

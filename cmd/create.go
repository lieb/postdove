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
	"io"

	"github.com/lieb/postdove/maildb"
	"github.com/spf13/cobra"
)

var (
	schemaFile  string
	domainsFile string
	domainLoad  bool
	aliasFile   string
	aliasLoad   bool
)

// cmdCreate
func cmdCreate(cmd *cobra.Command, args []string) error {
	var (
		inD   *maildb.Input
		inA   *maildb.Input
		cmdIn io.Reader
		err   error
	)
	if cmd.Flags().Changed("schema") {
		err = mdb.LoadSchema(schemaFile)
	} else {
		err = mdb.LoadSchema("")
	}
	if err == nil {
		mdb.Begin()
		defer mdb.End(&err)

		if !cmd.Flags().Changed("no-locals") { // Add default localhost stuff
			if cmd.Flags().Changed("local") {
				inD, err = mdb.NewInput(domainsFile, "files/domains")
			} else {
				inD, err = mdb.NewInput("", "files/domains")
			}
			if err == nil {
				defer inD.Close()
				cmdIn = cmd.InOrStdin()
				cmd.SetIn(inD.Reader())
				err = procImport(cmd, POSTFIX, procDomain)
				cmd.SetIn(cmdIn)
			}
		}
		if err == nil && !cmd.Flags().Changed("no-aliases") { // Add baseline aliases
			if cmd.Flags().Changed("alias") {
				inA, err = mdb.NewInput(aliasFile, "files/aliases")
			} else {
				inA, err = mdb.NewInput("", "files/aliases")
			}

			if err == nil {
				defer inA.Close()
				cmdIn = cmd.InOrStdin()
				cmd.SetIn(inA.Reader())
				err = procImport(cmd, ALIASES, procAlias)
				cmd.SetIn(cmdIn)
			}
		}
	}
	if err != nil {
		err = fmt.Errorf("Create command: %s", err)
	}
	return err
}

func init() {
	createCmd.Flags().StringVarP(&schemaFile, "schema", "s",
		"",
		"Schema file to define tables of database. Default is built in.")
	createCmd.Flags().StringVarP(&aliasFile, "alias", "a",
		"/etc/aliases",
		"RFC 2142 required aliases")
	createCmd.Flags().BoolVarP(&aliasLoad, "no-aliases", "A", false,
		"Do not load RFC 2142 aliases")
	createCmd.Flags().StringVarP(&domainsFile, "local", "l",
		"",
		"default local domains (localhost, localhost.localdomain")
	createCmd.Flags().BoolVarP(&domainLoad, "no-locals", "L", false,
		"Do not load local domain hosts")
}

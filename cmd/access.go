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
	//"strconv"
	//"strings"

	"github.com/lieb/postdove/maildb"
	"github.com/spf13/cobra"
)

var (
	accessAction string
)

// importAccess do import of an access file
var importAccess = &cobra.Command{
	Use:   "access",
	Short: "Import a file containing an access rule name and a restriction class tag",
	Long:  "Import access rules file from the file named by the -i flag (default stdin '-').",
	Args:  cobra.NoArgs,
	RunE:  accessImport,
}

// exportAccess do export of an access file
var exportAccess = &cobra.Command{
	Use:   "access",
	Short: "Export access rules into file one per line",
	Long:  "Export access rules into the file named by the -o flag (default stdout '-'",
	Args:  cobra.MaximumNArgs(1),
	RunE:  accessExport,
}

// addAccess do add of an access rule
var addAccess = &cobra.Command{
	Use:   "access name restriction",
	Short: "Add the named access rule to the database",
	Long: `Add the named access rule to the database. The the value is the key postfix uses to
select a set of recipient restrictions.`,
	Args: cobra.ExactArgs(2),
	RunE: accessAdd,
}

// deleteAccess delete an access rule
var deleteAccess = &cobra.Command{
	Use:   "access name",
	Short: "Delete the named recipient access rule from the database.",
	Long:  "Delete the named rule from the database so long as no address or domain references it.",
	Args:  cobra.ExactArgs(1),
	RunE:  accessDelete,
}

// editAccess edit an access rule
var editAccess = &cobra.Command{
	Use:   "access name",
	Short: "Edit the named access rule.",
	Long:  "Edit the named access rule to change the postfix restriction class key.",
	Args:  cobra.ExactArgs(1),
	RunE:  accessEdit,
}

// showAccess diaplay rule contents
var showAccess = &cobra.Command{
	Use:   "access name",
	Short: "Display the named access rule",
	Long:  "Display the contents of the named access rule to standard output.",
	Args:  cobra.ExactArgs(1),
	RunE:  accessShow,
}

// linkage to top level
func init() {
	importCmd.AddCommand(importAccess)
	exportCmd.AddCommand(exportAccess)
	addCmd.AddCommand(addAccess)
	deleteCmd.AddCommand(deleteAccess)
	editCmd.AddCommand(editAccess)
	editAccess.Flags().StringVarP(&accessAction, "action", "r", "",
		"Access rule action value used by Postfix to process client access restrictions")
	showCmd.AddCommand(showAccess)
}

// accessImport
func accessImport(cmd *cobra.Command, args []string) error {
	var err error

	mdb.Begin()
	defer mdb.End(&err)

	err = procImport(cmd, SIMPLE, procAccess)
	return err
}

// procAccess
func procAccess(tokens []string) error {
	if len(tokens) <= 1 {
		return fmt.Errorf("Access rule has no recipient restriction key")
	}
	_, err := mdb.InsertAccess(tokens[0], tokens[1])
	return err
}

// accessExport
func accessExport(cmd *cobra.Command, args []string) error {
	var name string

	switch len(args) {
	case 0: // All access rules
		name = "*"
	case 1:
		name = args[0]
	default:
		return fmt.Errorf("Only one access rule name can be specified")
	}
	al, err := mdb.FindAccess(name)
	if err == nil {
		for _, ac := range al {
			cmd.Printf("%s\n", ac.Export())
		}
	}
	return err
}

// accessAdd
func accessAdd(cmd *cobra.Command, args []string) error {
	var err error

	mdb.Begin()
	defer mdb.End(&err)

	return procAccess(args)
}

// accessDelete
func accessDelete(cmd *cobra.Command, args []string) error {
	return mdb.DeleteAccess(args[0])
}

// accessEdit
func accessEdit(cmd *cobra.Command, args []string) error {
	var (
		err error
		ac  *maildb.Access
	)

	mdb.Begin()
	defer mdb.End(&err)

	ac, err = mdb.GetAccess(args[0])
	if err == nil {
		if cmd.Flags().Changed("action") {
			err = ac.SetAction(accessAction)
		} else {
			err = fmt.Errorf("action option for access edit not set")
		}
	}
	return err
}

// accessShow
func accessShow(cmd *cobra.Command, args []string) error {
	var (
		err error
		ac  *maildb.Access
	)
	if ac, err = mdb.LookupAccess(args[0]); err != nil {
		return err
	}
	cmd.Printf("Name:\t%s\nAction:\t%s\n", ac.Name(), ac.Action())
	return nil
}

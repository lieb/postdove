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
	aAddRecipient []string
	aDelRecipient []string
)

// importAlias do import of an aliases file
var importAlias = &cobra.Command{
	Use:   "alias",
	Short: "Import an alias file in the aliases(5) format",
	Long: `Import a local aliases file in the aliases(5) format
from the file named by the -i flag (default stdin '-').
This is typically the /etc/aliases file that maps various system
users and email aliases to a specific user or site sysadmin mailbox`,
	Args: cobra.NoArgs,
	RunE: aliasImport,
}

// exportAlias do export of an aliases file
var exportAlias = &cobra.Command{
	Use:   "alias",
	Short: "Export aliases alias file in the aliases(5) format",
	Long: `Export a local aliases file in the aliases(5) format to
the file named by the -o flag (default stdout '-').
This is typically the /etc/aliases file that maps various system
users and email aliases to a specific user or site sysadmin mailbox`,
	Args: cobra.MaximumNArgs(1),
	RunE: aliasExport,
}

// addAlias do add of an aliases file
var addAlias = &cobra.Command{
	Use:   "alias address recipient ...",
	Short: "Add an alias into the database",
	Long: `Add an alias into the database. The address is a local
user or alias target without a "@domain" part, i.e. "postmaster" or "daemon".
One or more recipients can be added. A recipient can either be a single local mailbox,
i.e. "root" or "admin", an RFC2822 format email address, or a file or a pipe to a command.
 See aliases(5) man page for details.`,
	Args: cobra.MinimumNArgs(2), // alias recipient ...
	RunE: aliasAdd,
}

// deleteAlias do delete of an aliases file
var deleteAlias = &cobra.Command{
	Use:   "alias address",
	Short: "Delete an alias from the database.",
	Long: `Delete an address alias from the database.
All of the recipients pointed to by this name will be also deleted`,
	Args: cobra.ExactArgs(1), // alias name
	RunE: aliasDelete,
}

// editAlias do edit of an aliases file
var editAlias = &cobra.Command{
	Use:   "alias address",
	Short: "Edit the alias address recipients in the database",
	Long:  `Edit a local alias address and its list of recipients.`,
	Args:  cobra.ExactArgs(1), // an alias or all aliases
	RunE:  aliasEdit,
}

// showAlias display an alias
var showAlias = &cobra.Command{
	Use:   "alias address",
	Short: "Display the contents of an alias",
	Long: `Display the contents of an alias and all its recipients
to the standard output`,
	Args: cobra.ExactArgs(1),
	RunE: aliasShow,
}

// linkage to top level commands
func init() {
	importCmd.AddCommand(importAlias)
	exportCmd.AddCommand(exportAlias)
	addCmd.AddCommand(addAlias)
	deleteCmd.AddCommand(deleteAlias)
	editCmd.AddCommand(editAlias)
	editAlias.Flags().StringSliceVarP(&aAddRecipient, "add", "a", []string{""},
		"Recipient to add to this alias")
	editAlias.Flags().StringSliceVarP(&aDelRecipient, "remove", "r", []string{""},
		"Recipient to remove from this alias")
	showCmd.AddCommand(showAlias)
}

// aliasImport the aliases in /etc/aliases format from inFile
func aliasImport(cmd *cobra.Command, args []string) error {
	var err error

	mdb.Begin()
	defer mdb.End(&err)

	err = procImport(cmd, ALIASES, procAlias)
	return err
}

// procAlias for both add and import
func procAlias(tokens []string) error {
	var (
		err error
		a   *maildb.Address
	)

	if ap, err := maildb.DecodeRFC822(tokens[0]); err != nil {
		return err
	} else if !ap.IsLocal() {
		return fmt.Errorf("An alias cannot have a domain component")
	}
	if a, err = mdb.GetOrInsAddress(tokens[0]); err == nil {
		for _, r := range tokens[1:] {
			if err = a.AttachAlias(r); err != nil {
				break
			}
		}
	}
	return err
}

// aliasExport the aliases in /etc/aliases format to outFile
func aliasExport(cmd *cobra.Command, args []string) error {
	var (
		err   error
		alias string
		alist []*maildb.Alias
	)

	if len(args) == 0 {
		alias = "*"
	} else {
		if ap, err := maildb.DecodeRFC822(args[0]); err != nil {
			return err
		} else if !ap.IsLocal() {
			return fmt.Errorf("An alias cannot have a domain component")
		}
		alias = args[0]
	}
	if alist, err = mdb.LookupAlias(alias); err == nil {
		for _, al := range alist {
			cmd.Printf("%s\n", al.String())
		}
	}
	return err
}

// aliasAdd the alias and its recipients
func aliasAdd(cmd *cobra.Command, args []string) error {
	var err error

	mdb.Begin()
	defer mdb.End(&err)

	return procAlias(args)
}

// aliasDelete the address in the first arg
func aliasDelete(cmd *cobra.Command, args []string) error {
	if ap, err := maildb.DecodeRFC822(args[0]); err != nil {
		return err
	} else if !ap.IsLocal() {
		return fmt.Errorf("An alias cannot have a domain component")
	}
	return mdb.RemoveAlias(args[0])
}

// aliasEdit the address in the first arg
func aliasEdit(cmd *cobra.Command, args []string) error {
	var (
		err error
		a   *maildb.Address
	)

	if ap, err := maildb.DecodeRFC822(args[0]); err != nil {
		return err
	} else if !ap.IsLocal() {
		return fmt.Errorf("An alias cannot have a domain component")
	}
	// add new ones first so remove doesn't remove an unattached alias...
	if err == nil && cmd.Flags().Changed("add") {
		mdb.Begin()

		if a, err = mdb.GetAddress(args[0]); err == nil {
			for _, r := range aAddRecipient {
				if err = a.AttachAlias(r); err != nil {
					break
				}
			}
		}
		mdb.End(&err)
	}
	if err == nil && cmd.Flags().Changed("remove") {
		for _, r := range aDelRecipient {
			if err = mdb.RemoveRecipient(args[0], r); err != nil {
				break
			}
		}
	}
	return err
}

// aliasShow the address in the first arg
func aliasShow(cmd *cobra.Command, args []string) error {
	var (
		err error
		a   *maildb.Address
		al  *maildb.Alias
	)

	if ap, err := maildb.DecodeRFC822(args[0]); err != nil {
		return err
	} else if !ap.IsLocal() {
		return fmt.Errorf("An alias cannot have a domain component")
	}
	if a, err = mdb.LookupAddress(args[0]); err == nil {
		if al, err = a.Alias(); err == nil {
			cmd.Printf("Alias:\t%s\nTargets:", args[0])
			for _, t := range al.Targets() {
				cmd.Printf("\t%s\n", t.String())
			}
		}
	}
	return err
}

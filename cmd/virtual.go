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
	vAddRecipient []string
	vDelRecipient []string
)

// importVirtual do import of an virtuals file
var importVirtual = &cobra.Command{
	Use:   "virtual",
	Short: "Import an virtual alias file in the postfix virtual(5) format",
	Long: `Import a local virtual alias file in the postfix virtual(5) format
from the file named by the -i flag (default stdin '-').
This is postfix file associated with the $virtual_aliases hash`,
	Args: cobra.NoArgs,
	RunE: virtualImport,
}

// exportVirtual do export of an virtuals file
var exportVirtual = &cobra.Command{
	Use:   "virtual",
	Short: "Export virtual aliases to the named file in postfix virtual(5) format",
	Long: `Export virtual aliases in postfix virtual(5) format to
the file named by the -o flag (default stdout '-').
This is typically the file that maps various email virtual addresses to relay or IMAP/POP3 mailboxes`,
	Args: cobra.MaximumNArgs(1),
	RunE: virtualExport,
}

// addVirtual do add of an virtuals file
var addVirtual = &cobra.Command{
	Use:   "virtual address recipient ...",
	Short: "Add an virtual alias addressinto the database",
	Long: `Add an virtual alias address into the database. The address is an RFC2822
email address. One or more recipients can be added. A recipient can either be a single local mailbox or
an RFC2822 format email address. See postfix virtual(5) man page for details.`,
	Args: cobra.MinimumNArgs(2), // virtual recipient ...
	RunE: virtualAdd,
}

// deleteVirtual do delete of an virtuals file
var deleteVirtual = &cobra.Command{
	Use:   "virtual address",
	Short: "Delete an virtual address aliasfrom the database.",
	Long: `Delete an virtual address alias from the database.
All of the recipients pointed to by this name will be also deleted`,
	Args: cobra.ExactArgs(1), // virtual name
	RunE: virtualDelete,
}

// editVirtual do edit of an virtuals file
var editVirtual = &cobra.Command{
	Use:   "virtual address",
	Short: "Edit the virtual address alias and its recipients in the database",
	Long: `Edit a virtual alias address and its list of recipients. You can edit, add,
or delete recipients`,
	Args: cobra.ExactArgs(1), // an virtual or all virtuales
	RunE: virtualEdit,
}

// showVirtual display an virtual alias
var showVirtual = &cobra.Command{
	Use:   "virtual address",
	Short: "Display the contents of an virtual alias",
	Long: `Display the contents of an virtual alias and all its recipients
to the standard output`,
	Args: cobra.ExactArgs(1),
	RunE: virtualShow,
}

// linkage to top level commands
func init() {
	importCmd.AddCommand(importVirtual)
	exportCmd.AddCommand(exportVirtual)
	addCmd.AddCommand(addVirtual)
	deleteCmd.AddCommand(deleteVirtual)
	editCmd.AddCommand(editVirtual)
	editVirtual.Flags().StringSliceVarP(&vAddRecipient, "add", "a", []string{""},
		"Recipient to add to this virtual alias")
	editVirtual.Flags().StringSliceVarP(&vDelRecipient, "remove", "r", []string{""},
		"Recipient to remove from this virtual alias")
	showCmd.AddCommand(showVirtual)
}

// virtualImport the virtuales in /etc/virtuales format from inFile
func virtualImport(cmd *cobra.Command, args []string) error {
	var err error

	mdb.Begin()
	defer mdb.End(&err)

	err = procImport(cmd, POSTFIX, procVirtual) // once past syntax, they are same
	return err
}

// procVirtual
func procVirtual(tokens []string) error {
	var (
		err error
		a   *maildb.Address
	)

	if ap, err := maildb.DecodeRFC822(tokens[0]); err != nil {
		return err
	} else if ap.IsLocal() {
		return fmt.Errorf("A virtual alias must be 'mailbox@domain'")
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

// virtualExport the virtuales in /etc/virtuales format to outFile
func virtualExport(cmd *cobra.Command, args []string) error {
	var (
		err     error
		virtual string
		alist   []*maildb.Alias
	)

	if len(args) == 0 {
		virtual = "*@*"
	} else {
		if ap, err := maildb.DecodeRFC822(args[0]); err != nil {
			return err
		} else if ap.IsLocal() {
			return fmt.Errorf("A virtual alias must have a domain")
		}
		virtual = args[0]
	}
	if alist, err = mdb.LookupAlias(virtual); err == nil {
		for _, al := range alist {
			cmd.Printf("%s\n", al.Export())
		}
	}
	return err
}

// virtualAdd the virtual and its recipients
func virtualAdd(cmd *cobra.Command, args []string) error {
	var err error

	mdb.Begin()
	defer mdb.End(&err)

	return procVirtual(args)
}

// virtualDelete the address in the first arg
func virtualDelete(cmd *cobra.Command, args []string) error {
	if ap, err := maildb.DecodeRFC822(args[0]); err != nil {
	} else if ap.IsLocal() {
		return fmt.Errorf("A virtual alias must be 'mailbox@domain'")
	}
	return mdb.RemoveAlias(args[0])
}

// virtualEdit the address in the first arg
func virtualEdit(cmd *cobra.Command, args []string) error {
	var (
		err error
		a   *maildb.Address
	)

	if ap, err := maildb.DecodeRFC822(args[0]); err != nil {
	} else if ap.IsLocal() {
		return fmt.Errorf("A virtual alias must be 'mailbox@domain'")
	}
	// add new ones first so remove doesn't remove an unattached alias...
	if err == nil && cmd.Flags().Changed("add") {
		mdb.Begin()

		if a, err = mdb.GetAddress(args[0]); err == nil {
			for _, r := range vAddRecipient {
				if err = a.AttachAlias(r); err != nil {
					break
				}
			}
		}
		mdb.End(&err)
	}
	if err == nil && cmd.Flags().Changed("remove") {
		for _, r := range vDelRecipient {
			if err = mdb.RemoveRecipient(args[0], r); err != nil {
				break
			}
		}
	}
	return err
}

// virtualShow the address in the first arg
func virtualShow(cmd *cobra.Command, args []string) error {
	var (
		err error
		a   *maildb.Address
		al  *maildb.Alias
	)

	if ap, err := maildb.DecodeRFC822(args[0]); err != nil {
	} else if ap.IsLocal() {
		return fmt.Errorf("A virtual alias must be 'mailbox@domain'")
	}
	if a, err = mdb.LookupAddress(args[0]); err == nil {
		if al, err = a.Alias(); err == nil {
			cmd.Printf("Virtual Alias:\t%s\nTargets:", args[0])
			for i, t := range al.Targets() {
				if i == 0 {
					cmd.Printf("\t%s\n", t.Recipient())
				} else {
					cmd.Printf("\t\t%s\n", t.Recipient())
				}
			}
		}
	}
	return err
}

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

	"github.com/spf13/cobra"
)

// importMailbox do import of an mailboxes file
var importMailbox = &cobra.Command{
	Use:   "mailbox",
	Short: "Import a set of mailboxes from a file in a /etc/passwd format",
	Long: `Import a set of mailboxes into the database
from the file named by the -i flag (default stdin '-').`,
	Args: cobra.NoArgs,
	Run:  mailboxImport,
}

// exportMailbox do export of an mailboxes file
var exportMailbox = &cobra.Command{
	Use:   "mailbox",
	Short: "Export mailboxes in a /etc/passwd similar format",
	Long: `Export mailboxes in a /etc/passwd similar format to
the file named by the -o flag (default stdout '-').`,
	Args: cobra.NoArgs,
	Run:  mailboxExport,
}

// addMailbox do add of an mailboxes file
var addMailbox = &cobra.Command{
	Use:   "mailbox address [ flags ]",
	Short: "Add an mailbox and its address into the database",
	Long: `Add an mailbox into the database. The address must be in an already
existing vmailbox domain. The flags set the various login parameters such as password and
quota.`,
	Args: cobra.MinimumNArgs(2), // mailbox recipient ...
	Run:  mailboxAdd,
}

// deleteMailbox do delete of an mailboxes file
var deleteMailbox = &cobra.Command{
	Use:   "mailbox address",
	Short: "Delete an mailbox and its address from the database.",
	Long: `Delete an address mailbox and its address from the database.
All of the aliases that point to it must be changed or deleted first`,
	Args: cobra.ExactArgs(1), // mailbox name
	Run:  mailboxDelete,
}

// editMailbox do edit of an mailboxes file
var editMailbox = &cobra.Command{
	Use:   "mailbox address",
	Short: "Edit the mailbox  for the address in the database",
	Long:  `Edit a mailbox to change attributes such as uid/gid, password, quota.`,
	Args:  cobra.MaximumNArgs(1), // an mailbox or all mailboxes
	Run:   mailboxEdit,
}

// linkage to top level commands
func init() {
	importCmd.AddCommand(importMailbox)
	exportCmd.AddCommand(exportMailbox)
	addCmd.AddCommand(addMailbox)
	deleteCmd.AddCommand(deleteMailbox)
	editCmd.AddCommand(editMailbox)
}

// mailboxImport the mailboxes from inFile
func mailboxImport(cmd *cobra.Command, args []string) {
	fmt.Println("import mailbox called infile", dbFile, inFile)
}

// mailboxExport the mailboxes to outFile
func mailboxExport(cmd *cobra.Command, args []string) {
	fmt.Println("export mailbox called outfile", dbFile, outFile)
}

// mailboxAdd the mailbox and its address
func mailboxAdd(cmd *cobra.Command, args []string) {
	fmt.Println("add mailbox")
}

// mailboxDelete the mailbox and address in the first arg
func mailboxDelete(cmd *cobra.Command, args []string) {
	fmt.Println("delete mailbox called")
}

// mailboxEdit the mailbox of the address in the first arg
func mailboxEdit(cmd *cobra.Command, args []string) {
	fmt.Println("edit mailbox called")
}

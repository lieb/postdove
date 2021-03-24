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

// importVirtual do import of an virtuales file
var importVirtual = &cobra.Command{
	Use:   "virtual",
	Short: "Import an virtual alias file in the postfix virtual(5) format",
	Long: `Import a local virtuales file in the postfixvirtual(5) format
from the file named by the -i flag (default stdin '-').
This is postfix file associated with the $virtual_aliases hash`,
	Args: cobra.NoArgs,
	Run:  virtualImport,
}

// exportVirtual do export of an virtuales file
var exportVirtual = &cobra.Command{
	Use:   "virtual",
	Short: "Export virtual aliases to the named file in postfix virtual(5) format",
	Long: `Export virtual aliases in postfix virtual(5) format to
the file named by the -o flag (default stdout '-').
This is typically the file that maps various email virtual addresses to relay or IMAP/POP3 mailboxes`,
	Args: cobra.NoArgs,
	Run:  virtualExport,
}

// addVirtual do add of an virtuales file
var addVirtual = &cobra.Command{
	Use:   "virtual address recipient ...",
	Short: "Add an virtual alias addressinto the database",
	Long: `Add an virtual alias address into the database. The address is an RFC2822
email address. One or more recipients can be added. A recipient can either be a single local mailbox or
an RFC2822 format email address. See postfix virtual(5) man page for details.`,
	Args: cobra.MinimumNArgs(2), // virtual recipient ...
	Run:  virtualAdd,
}

// deleteVirtual do delete of an virtuales file
var deleteVirtual = &cobra.Command{
	Use:   "virtual address",
	Short: "Delete an virtual address aliasfrom the database.",
	Long: `Delete an virtual address alias from the database.
All of the recipients pointed to by this name will be also deleted`,
	Args: cobra.ExactArgs(1), // virtual name
	Run:  virtualDelete,
}

// editVirtual do edit of an virtuales file
var editVirtual = &cobra.Command{
	Use:   "virtual address",
	Short: "Edit the virtual address alias and its recipients in the database",
	Long: `Edit a virtual alias address and its list of recipients. You can edit, add,
or delete recipients`,
	Args: cobra.MaximumNArgs(1), // an virtual or all virtuales
	Run:  virtualEdit,
}

// linkage to top level commands
func init() {
	importCmd.AddCommand(importVirtual)
	exportCmd.AddCommand(exportVirtual)
	addCmd.AddCommand(addVirtual)
	deleteCmd.AddCommand(deleteVirtual)
	editCmd.AddCommand(editVirtual)
}

// virtualImport the virtuales in /etc/virtuales format from inFile
func virtualImport(cmd *cobra.Command, args []string) {
	fmt.Println("import virtual called infile", dbFile, inFile)
}

// virtualExport the virtuales in /etc/virtuales format to outFile
func virtualExport(cmd *cobra.Command, args []string) {
	fmt.Println("export virtual called outfile", dbFile, outFile)
}

// virtualAdd the virtual and its recipients
func virtualAdd(cmd *cobra.Command, args []string) {
	fmt.Println("add virtual")
}

// virtualDelete the address in the first arg
func virtualDelete(cmd *cobra.Command, args []string) {
	fmt.Println("delete virtual called")
}

// virtualEdit the address in the first arg
func virtualEdit(cmd *cobra.Command, args []string) {
	fmt.Println("edit virtual called")
}

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

// importAlias do import of an aliases file
var importAlias = &cobra.Command{
	Use:   "alias",
	Short: "Import an alias file in the aliases(5) format",
	Long: `Import a local aliases file in the aliases(5) format
from the file named by the -i flag (default stdin '-').
This is typically the /etc/aliases file that maps various system
users and email aliases to a specific user or site sysadmin mailbox`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("import alias called infile", dbFile, inFile)
	},
}

// exportAlias do export of an aliases file
var exportAlias = &cobra.Command{
	Use:   "alias",
	Short: "Export aliases alias file in the aliases(5) format",
	Long: `Export a local aliases file in the aliases(5) format to
the file named by the -o flag (default stdout '-').
This is typically the /etc/aliases file that maps various system
users and email aliases to a specific user or site sysadmin mailbox`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("export alias called outfile", dbFile, outFile)
	},
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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("add alias")
	},
}

// deleteAlias do delete of an aliases file
var deleteAlias = &cobra.Command{
	Use:   "alias address",
	Short: "Delete an alias from the database.",
	Long: `Delete an address alias from the database.
All of the recipients pointed to by this name will be also deleted`,
	Args: cobra.ExactArgs(1), // alias name
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("delete alias called")
	},
}

// editAlias do edit of an aliases file
var editAlias = &cobra.Command{
	Use:   "alias address",
	Short: "Edit the alias address and its recipients in the database",
	Long:  `Edit a local alias address and its list of recipients.`,
	Args:  cobra.MaximumNArgs(1), // an alias or all aliases
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("edit alias called")
	},
}

func init() {
	importCmd.AddCommand(importAlias)
	exportCmd.AddCommand(exportAlias)
	addCmd.AddCommand(addAlias)
	deleteCmd.AddCommand(deleteAlias)
	editCmd.AddCommand(editAlias)
}

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
	//"fmt"
	"github.com/spf13/cobra"
	//"os"
	//homedir "github.com/mitchellh/go-homedir"
	//"github.com/spf13/viper"
)

var (
	dbFile string
)

// rootCmd represents the base command when called without any subcommands
// if we call without any commands, we fall into the TUI app
var rootCmd = &cobra.Command{
	Use:   "postdove",
	Short: "A management tool for aliases and mail users of postfix and dovecot",
	Long: `Postdove is a management tool to manage the sqlite database file that
is used by postfix to manage aliases, domains, and delivery and by dovecot to 
manage email user IMAP/POP3 email accounts`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: cmdTUI,
}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create the Sqlite database and initialize its tables",
	Long: `Create the Sqlite database file and initilize its tables.
You will also have to do some imports and adds to this otherwise empty database.`,
	Args: cobra.NoArgs,
	Run:  cmdCreate,
}

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import [table] ",
	Short: "Import a file to the database",
	Long: `Import a file to the postfix/dovecot database. Most of these files
use the same format required for postfix key/value pair databases`,
}

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export [table]",
	Short: "Export the specified table to a file or stdout",
	Long: `Export the specified table to a file using the expected format
used by postfix and/or dovecot.`,
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [table]",
	Short: "Add an entry into the specified table",
	Long:  `Add an entry into the specified table in the database`,
}

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete [table]",
	Short: "Delete and entry in the specified table",
	Long:  `Delete an entry in the specified table.`,
}

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit [table]",
	Short: "Edit an database entry in this table",
	Long:  `Edit an entry in the specified table.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// dbfile points to a database file that is other than the system default
	rootCmd.PersistentFlags().StringVarP(&dbFile, "dbfile", "d",
		"/etc/dovecot/private/dovecot.sqlite",
		"Sqlite3 database file")

	// Create command and schema arg
	rootCmd.AddCommand(createCmd)
	createCmd.PersistentFlags().StringVarP(&schemaFile, "schema", "s",
		"/etc/dovecot/private/dovecot.schema",
		"Schema file to define tables of database")

	// Import command and input file arg
	rootCmd.AddCommand(importCmd)
	importCmd.PersistentFlags().StringVarP(&inFile, "input", "i", "-",
		"Input file in postfix/dovecot format")

	// Export command and output file arg
	rootCmd.AddCommand(exportCmd)
	exportCmd.PersistentFlags().StringVarP(&outFile, "output", "o", "-",
		"Output file in postfix/dovecot format")

	// Add command
	rootCmd.AddCommand(addCmd)

	// Delete command
	rootCmd.AddCommand(deleteCmd)

	// Edit command
	rootCmd.AddCommand(editCmd)
}

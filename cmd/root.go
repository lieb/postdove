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
	_ "embed"
	"os"

	"github.com/lieb/postdove/maildb"
	"github.com/spf13/cobra"
)

//go:embed version.txt
var Version string

const defaultDB = "/etc/postfix/private/postdove.sqlite"

var (
	dbFile        string
	reportVersion bool
	mdb           *maildb.MailDB
)

// rootCmd represents the base command when called without any subcommands
// if we call without any commands, we fall into the TUI app
var rootCmd = &cobra.Command{
	Use:   "postdove",
	Short: "A management tool for aliases and mail users of postfix and dovecot",
	Long: `Postdove is a management tool to manage the sqlite database file that
is used by postfix to manage aliases, domains, and delivery and by dovecot to 
manage email user IMAP/POP3 email accounts.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("version") {
			cmd.Printf("Version: %s", Version)
			os.Exit(0)
		}
		return openDB(cmd, args)
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run:               cmdTUI,
	PersistentPostRun: closeDB,
}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create the Sqlite database and initialize its tables",
	Long: `Create the Sqlite database file and initilize its tables.
You will also have to do some imports and adds to this otherwise empty database.`,
	Args: cobra.NoArgs,
	RunE: cmdCreate,
}

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import [table] ",
	Short: "Import a file to the database",
	Long: `Import a file to the postfix/dovecot database. Most of these files
use the same format required for postfix key/value pair databases`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := callPersistentPreRunE(cmd, args); err != nil {
			return err
		}
		return importRedirect(cmd, args)
	},
	PersistentPostRunE: importClose,
}

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export [table]",
	Short: "Export the specified table to a file or stdout",
	Long: `Export the specified table to a file using the expected format
used by postfix and/or dovecot.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := callPersistentPreRunE(cmd, args); err != nil {
			return err
		}
		return exportRedirect(cmd, args)
	},
	PersistentPostRunE: exportClose,
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
	Short: "Delete an entry in the specified table",
	Long:  `Delete an entry in the specified table.`,
}

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit [table]",
	Short: "Edit an database entry in this table",
	Long:  `Edit an entry in the specified table.`,
}

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show [table]",
	Short: "Show the contents of a table entry",
	Long: `Show the contents of a table or table entry in the database in a person
readable format`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

// callPersistentPreRunE
// hack to chain preruns for import/export/??
// otherwise, the db doesn't get opened which is a root prerun for everything!
func callPersistentPreRunE(cmd *cobra.Command, args []string) error {
	if parent := cmd.Parent(); parent != nil {
		if parent.PersistentPreRunE != nil {
			if err := parent.PersistentPreRunE(parent, args); err != nil {
				return err
			}
		}
	}
	return nil
}

// openDB persistent root pre-run on all commands
// make the DB open for business
func openDB(cmd *cobra.Command, args []string) error {
	var err error

	if mdb, err = maildb.NewMailDB(dbFile); err != nil {
		return err
	}
	return nil
}

// closeDB
// persistent post-run to clean up the DB
func closeDB(cmd *cobra.Command, args []string) {
	mdb.Close()
	mdb = nil
}

func init() {
	// dbfile points to a database file that is other than the system default
	rootCmd.PersistentFlags().StringVarP(&dbFile, "dbfile", "d",
		defaultDB,
		"Sqlite3 database file")

	// Report version
	rootCmd.PersistentFlags().BoolVarP(&reportVersion, "version", "v",
		false,
		"Report Postdove version and exit")

	// Create command and schema arg
	rootCmd.AddCommand(createCmd)

	// Import command and input file arg
	rootCmd.AddCommand(importCmd)
	importCmd.PersistentFlags().StringVarP(&inFilePath, "input", "i", "-",
		"Input file in postfix/dovecot format")

	// Export command and output file arg
	rootCmd.AddCommand(exportCmd)
	exportCmd.PersistentFlags().StringVarP(&outFilePath, "output", "o", "-",
		"Output file in postfix/dovecot format")

	// Add command
	rootCmd.AddCommand(addCmd)

	// Delete command
	rootCmd.AddCommand(deleteCmd)

	// Edit command
	rootCmd.AddCommand(editCmd)

	// Show command
	rootCmd.AddCommand(showCmd)
}

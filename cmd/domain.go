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

	"github.com/lieb/postdove/maildb"
	"github.com/spf13/cobra"
)

var (
	vUid   int
	vGid   int
	rClass string
)

// importDomain do import of an domaines file
var importDomain = &cobra.Command{
	Use:   "domain",
	Short: "Import a file containing a domain name and its attributes, one per line",
	Long:  `Import a domains file from the file named by the -i flag (default stdin '-').`,
	Args:  cobra.NoArgs,
	Run:   domainImport,
}

// exportDomain do export of an domaines file
var exportDomain = &cobra.Command{
	Use:   "domain",
	Short: "Export domains into file one per line",
	Long:  `Export domains to the file named by the -o flag (default stdout '-').`,
	Args:  cobra.NoArgs,
	Run:   domainExport,
}

// addDomain do add of an domaines file
var addDomain = &cobra.Command{
	Use:   "domain name class",
	Short: "Add an domain into the database",
	Long: `Add an domain into the database. The name is the FQDN for the domain.
The class defines what the domain is used for, i.e. for virtual mailboxes or local/my domain.`,
	Args: cobra.MinimumNArgs(1), // domain if no second, use DB default
	RunE: domainAdd,
}

// deleteDomain do delete of an domaines file
var deleteDomain = &cobra.Command{
	Use:   "domain ",
	Short: "Delete an domain from the database.",
	Long: `Delete an address domain from the database.
All of the recipients pointed to by this name will be also deleted`,
	Args: cobra.ExactArgs(1), // domain name
	Run:  domainDelete,
}

// editDomain do edit of an domaines file
var editDomain = &cobra.Command{
	Use:   "domain name",
	Short: "Edit the named domain and attributes in the database",
	Long:  `Edit a domain and its attributes.`,
	Args:  cobra.ExactArgs(1), // edit just this domain
	RunE:  domainEdit,
}

// linkage to top level commands
func init() {
	importCmd.AddCommand(importDomain)
	exportCmd.AddCommand(exportDomain)
	addCmd.AddCommand(addDomain)
	deleteCmd.AddCommand(deleteDomain)
	editCmd.AddCommand(editDomain)
	editDomain.Flags().IntVarP(&vUid, "uid", "u", 0,
		"Virtual user id for this domain")
	editDomain.Flags().IntVarP(&vGid, "gid", "g", 0,
		"Virtual group id for this domain")
	editDomain.Flags().StringVarP(&rClass, "rclass", "r", "",
		"Restriction class for this domain")
}

// domainImport the domains from inFile
func domainImport(cmd *cobra.Command, args []string) {
	fmt.Println("import domain called infile", dbFile, inFile)
}

// domainExport the domains to outFile
func domainExport(cmd *cobra.Command, args []string) {
	fmt.Println("export domain called outfile", dbFile, outFile)
}

// domainAdd the domain and its class
func domainAdd(cmd *cobra.Command, args []string) error {
	var class string = ""

	if len(args) > 1 {
		class = args[1]
	}
	if d, err := mdb.InsertDomain(args[0], class); err != nil {
		return err
	} else {
		d.Release()
		return nil
	}
}

// domainDelete the domain in the first arg
func domainDelete(cmd *cobra.Command, args []string) {
	fmt.Println("delete domain called")
}

// domainEdit the domain in the first arg
func domainEdit(cmd *cobra.Command, args []string) error {
	var (
		err error
		d   *maildb.Domain
	)

	if d, err = mdb.GetDomain(args[0]); err != nil {
		return err
	}
	if cmd.Flags().Changed("uid") {
		if err = d.SetVUid(vUid); err != nil {
			fmt.Printf("uid set, %s\n", err)
		}
	}
	if cmd.Flags().Changed("gid") {
		if err = d.SetVGid(vGid); err != nil {
			fmt.Printf("gid set, %s\n", err)
		}
	}
	if cmd.Flags().Changed("rclass") {
		if err = d.SetRclass(rClass); err != nil {
			fmt.Printf("rclass set, %s\n", err)
		}
	}
	d.Release()
	return nil
}

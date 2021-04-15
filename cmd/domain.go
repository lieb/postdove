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
	"github.com/lieb/postdove/maildb"
	"github.com/spf13/cobra"
)

var (
	vUid   int
	vGid   int
	rClass string
)

// importDomain do import of a domains file
var importDomain = &cobra.Command{
	Use:   "domain",
	Short: "Import a file containing a domain name and its attributes, one per line",
	Long:  `Import a domains file from the file named by the -i flag (default stdin '-').`,
	Args:  cobra.NoArgs,
	Run:   domainImport,
}

// exportDomain do export of a domains file
var exportDomain = &cobra.Command{
	Use:   "domain",
	Short: "Export domains into file one per line",
	Long:  `Export domains to the file named by the -o flag (default stdout '-').`,
	Args:  cobra.NoArgs,
	Run:   domainExport,
}

// addDomain do add of a domains file
var addDomain = &cobra.Command{
	Use:   "domain name class",
	Short: "Add an domain into the database",
	Long: `Add an domain into the database. The name is the FQDN for the domain.
The class defines what the domain is used for, i.e. for virtual mailboxes or local/my domain.`,
	Args: cobra.MinimumNArgs(1), // domain if no second, use DB default
	RunE: domainAdd,
}

// deleteDomain do delete of a domains file
var deleteDomain = &cobra.Command{
	Use:   "domain ",
	Short: "Delete an domain from the database.",
	Long: `Delete an address domain from the database.
All of the recipients pointed to by this name will be also deleted`,
	Args: cobra.ExactArgs(1), // domain name
	RunE: domainDelete,
}

// editDomain do edit of a domains file
var editDomain = &cobra.Command{
	Use:   "domain name",
	Short: "Edit the named domain and attributes in the database",
	Long:  `Edit a domain and its attributes.`,
	Args:  cobra.ExactArgs(1), // edit just this domain
	RunE:  domainEdit,
}

// showDomain display domain contents
var showDomain = &cobra.Command{
	Use:   "domain name",
	Short: "Display the contents of the named domain to the standard output",
	Long: `Show the contents of the named domain to the standard output
showing all its attributes`,
	Args: cobra.ExactArgs(1),
	RunE: domainShow,
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
	showCmd.AddCommand(showDomain)
}

// domainImport the domains from inFile
func domainImport(cmd *cobra.Command, args []string) {
	cmd.Println("import domain called infile", dbFile, inFile)
}

// domainExport the domains to outFile
func domainExport(cmd *cobra.Command, args []string) {
	cmd.Println("export domain called outfile", dbFile, outFile)
}

// domainAdd the domain and its class
func domainAdd(cmd *cobra.Command, args []string) error {
	var class string = ""

	if len(args) > 1 {
		class = args[1]
	}
	mdb.Begin()
	_, err := mdb.InsertDomain(args[0], class)
	mdb.End(err == nil)
	return err
}

// domainDelete the domain in the first arg
func domainDelete(cmd *cobra.Command, args []string) error {
	return mdb.DeleteDomain(args[0])
}

// domainEdit the domain in the first arg
func domainEdit(cmd *cobra.Command, args []string) error {
	var (
		err error
		d   *maildb.Domain
	)

	mdb.Begin()
	d, err = mdb.GetDomain(args[0])
	if err == nil && cmd.Flags().Changed("uid") {
		err = d.SetVUid(vUid)
	}
	if err == nil && cmd.Flags().Changed("gid") {
		err = d.SetVGid(vGid)
	}
	if err == nil && cmd.Flags().Changed("rclass") {
		err = d.SetRclass(rClass)
	}
	mdb.End(err == nil)
	return err
}

// domainShow
func domainShow(cmd *cobra.Command, args []string) error {
	var (
		err error
		d   *maildb.Domain
	)

	if d, err = mdb.LookupDomain(args[0]); err != nil {
		return err
	}
	cmd.Printf("Name:\t\t%s\nClass:\t\t%s\nTransport:\t%s\nAccess:\t\t%s\n",
		d.String(), d.Class(), d.Transport(), d.Access())
	cmd.Printf("UserID:\t\t%s\nGroup ID:\t%s\nRestrictions:\t%s\n",
		d.Vuid(), d.Vgid(), d.Rclass())
	return nil
}

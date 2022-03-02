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
	"strconv"
	"strings"

	"github.com/lieb/postdove/maildb"
	"github.com/spf13/cobra"
)

var (
	dClass       string
	vUid         int64
	noVUid       bool
	vGid         int64
	noVGid       bool
	rClass       string
	noRClass     bool
	dTransport   string
	noDTransport bool
)

// importDomain do import of a domains file
var importDomain = &cobra.Command{
	Use:   "domain",
	Short: "Import a file containing a domain name and its attributes, one per line",
	Long:  `Import a domains file from the file named by the -i flag (default stdin '-').`,
	Args:  cobra.NoArgs,
	RunE:  domainImport,
}

// exportDomain do export of a domains file
var exportDomain = &cobra.Command{
	Use:   "domain",
	Short: "Export domains into file one per line",
	Long:  `Export domains to the file named by the -o flag (default stdout '-').`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  domainExport,
}

// addDomain do add of a domains file
var addDomain = &cobra.Command{
	Use:   "domain name",
	Short: "Add an domain into the database",
	Long: `Add an domain into the database. The name is the FQDN for the domain.
The optional class flag defines what the domain is used for, i.e. for virtual mailboxes or local/my domain.`,
	Args: cobra.ExactArgs(1), // domain if no second, use DB default
	RunE: domainAdd,
}

// deleteDomain do delete of a domains file
var deleteDomain = &cobra.Command{
	Use:   "domain name",
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
	addDomain.Flags().StringVarP(&dClass, "class", "c", "",
		"Domain class (internet, local, relay, virtual, vmailbox) for this domain")
	addDomain.Flags().Int64VarP(&vUid, "uid", "u", 99, // nobody user (at least on RH/Fedora)
		"Virtual user id for this domain")
	addDomain.Flags().Int64VarP(&vGid, "gid", "g", 99, // nobody group (at least on RH/Fedora)
		"Virtual group id for this domain")
	addDomain.Flags().StringVarP(&rClass, "rclass", "r", "",
		"Restriction class for this domain")
	addDomain.Flags().StringVarP(&dTransport, "transport", "t", "",
		"Transport to use for this domain")
	deleteCmd.AddCommand(deleteDomain)
	editCmd.AddCommand(editDomain)
	editDomain.Flags().StringVarP(&dClass, "class", "c", "",
		"Domain class (internet, local, relay, virtual, vmailbox) for this domain")
	editDomain.Flags().Int64VarP(&vUid, "uid", "u", 99, // nobody user (at least on RH/Fedora)
		"Virtual user id for this domain")
	editDomain.Flags().BoolVarP(&noVUid, "no-uid", "U", false,
		"Clear virtual uid value for this domain")
	editDomain.Flags().Int64VarP(&vGid, "gid", "g", 99, // nobody group (at least on RH/Fedora)
		"Virtual group id for this domain")
	editDomain.Flags().BoolVarP(&noVGid, "no-gid", "G", false,
		"Clear virtual group id for this domain")
	editDomain.Flags().StringVarP(&rClass, "rclass", "r", "",
		"Restriction class for this domain")
	editDomain.Flags().BoolVarP(&noRClass, "no-rclass", "R", false,
		"Clear the restriction class for this domain")
	editDomain.Flags().StringVarP(&dTransport, "transport", "t", "",
		"Transport to use for this domain")
	editDomain.Flags().BoolVarP(&noDTransport, "no-transport", "T", false,
		"Clear the transport for this domain")
	showCmd.AddCommand(showDomain)
}

// domainImport the domains from inFile
func domainImport(cmd *cobra.Command, args []string) error {
	var err error

	mdb.Begin()
	defer mdb.End(&err)

	err = procImport(cmd, POSTFIX, procDomain)
	return err
}

// procDomain
// options are "option=XXX"
func procDomain(tokens []string) error {
	var (
		d   *maildb.Domain
		err error
		id  int64
	)

	if d, err = mdb.InsertDomain(tokens[0]); err != nil {
		return err
	}
	if len(tokens) > 1 {
		for _, opt := range tokens[1:] {
			kv := strings.Split(opt, "=")
			if len(kv) < 2 {
				return fmt.Errorf("domain import option %s is not a key=value pair", opt)
			}
			switch kv[0] {
			case "class":
				if kv[1] != "\"\"" { // "" implies default so skip on import
					err = d.SetClass(kv[1])
				}
			case "vuid":
				id, err = strconv.ParseInt(kv[1], 10, 64)
				if err == nil {
					err = d.SetVUid(id)
				}
			case "vgid":
				id, err = strconv.ParseInt(kv[1], 10, 64)
				if err == nil {
					err = d.SetVGid(id)
				}
			case "rclass":
				err = d.SetRclass(kv[1])
			default:
				return fmt.Errorf("Unknown domain import option %s", kv[0])
			}
			if err != nil {
				break
			}
		}
	}
	return err
}

// domainExport the domains to outFile
func domainExport(cmd *cobra.Command, args []string) error {
	var domain string

	switch len(args) {
	case 0: // all domains
		domain = "*"
	case 1:
		domain = args[0] // domains by wildcard
	default:
		return fmt.Errorf("Only one domain can be specified")
	}
	dl, err := mdb.FindDomain(domain)
	if err == nil {
		for _, d := range dl {
			cmd.Printf("%s\n", d.Export())
		}
	}
	return err
}

// domainAdd the domain and its class
func domainAdd(cmd *cobra.Command, args []string) error {
	var (
		d   *maildb.Domain
		err error
	)

	mdb.Begin()
	defer mdb.End(&err)

	d, err = mdb.InsertDomain(args[0])
	if err == nil && cmd.Flags().Changed("class") {
		err = d.SetClass(dClass)
	}
	if err == nil && cmd.Flags().Changed("uid") {
		err = d.SetVUid(vUid)
	}
	if err == nil && cmd.Flags().Changed("gid") {
		err = d.SetVGid(vGid)
	}
	if err == nil && cmd.Flags().Changed("rclass") {
		err = d.SetRclass(rClass)
	}
	if err == nil && cmd.Flags().Changed("transport") {
		err = d.SetTransport(dTransport)
	}
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
	defer mdb.End(&err)

	d, err = mdb.GetDomain(args[0])
	if err == nil && cmd.Flags().Changed("class") {
		err = d.SetClass(dClass)
	}
	if err == nil {
		if cmd.Flags().Changed("no-uid") {
			err = d.ClearVUid()
		} else if cmd.Flags().Changed("uid") {
			err = d.SetVUid(vUid)
		}
	}
	if err == nil {
		if cmd.Flags().Changed("no-gid") {
			err = d.ClearVGid()
		} else if cmd.Flags().Changed("gid") {
			err = d.SetVGid(vGid)
		}
	}
	if err == nil {
		if cmd.Flags().Changed("no-rclass") {
			err = d.ClearRclass()
		} else if cmd.Flags().Changed("rclass") {
			err = d.SetRclass(rClass)
		}
	}
	if err == nil {
		if cmd.Flags().Changed("no-transport") {
			err = d.ClearTransport()
		} else if cmd.Flags().Changed("transport") {
			err = d.SetTransport(dTransport)
		}
	}
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
	cmd.Printf("Name:\t\t%s\nClass:\t\t%s\nTransport:\t%s\n",
		d.Name(), d.Class(), d.Transport())
	cmd.Printf("UserID:\t\t%s\nGroup ID:\t%s\nRestrictions:\t%s\n",
		d.Vuid(), d.Vgid(), d.Rclass())
	return nil
}

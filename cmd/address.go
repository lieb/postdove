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
	"strings"

	"github.com/lieb/postdove/maildb"
	"github.com/spf13/cobra"
)

var (
	localPart    string
	arClass      string
	aNoRclass    bool
	aTransport   string
	aNoTransport bool
)

// importAddress do import of a addresss file
var importAddress = &cobra.Command{
	Use:   "address",
	Short: "Import a file containing an address and its attributes, one per line",
	Long:  `Import an addresses file from the file named by the -i flag (default stdin '-').`,
	Args:  cobra.NoArgs,
	RunE:  addressImport,
}

// exportAddress do export of a addresses file
var exportAddress = &cobra.Command{
	Use:   "address",
	Short: "Export addresses into file one per line",
	Long:  `Export addresses to the file named by the -o flag (default stdout '-').`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  addressExport,
}

// addAddress do add of a addresss file
var addAddress = &cobra.Command{
	Use:   "address name",
	Short: "Add an address into the database",
	Long: `Add an address into the database. The name is either local (just a name with no domain part) or an RFC2822 format address.
The optional restriction class defines how postfix processes this address.`,
	Args: cobra.ExactArgs(1),
	RunE: addressAdd,
}

// deleteAddress do delete of a addresss file
var deleteAddress = &cobra.Command{
	Use:   "address ",
	Short: "Delete an address from the database.",
	Long:  `Delete an address address from the database.`,
	Args:  cobra.ExactArgs(1),
	RunE:  addressDelete,
}

// editAddress do edit of a domains file
var editAddress = &cobra.Command{
	Use:   "address name",
	Short: "Edit the named address and attributes in the database",
	Long:  `Edit a address and its attributes.`,
	Args:  cobra.ExactArgs(1), // edit just this address
	RunE:  addressEdit,
}

// showAddress display address contents
var showAddress = &cobra.Command{
	Use:   "address name",
	Short: "Display the contents of the named address to the standard output",
	Long: `Show the contents of the named address to the standard output
showing all its attributes`,
	Args: cobra.ExactArgs(1),
	RunE: addressShow,
}

// linkage to top level commands
func init() {
	importCmd.AddCommand(importAddress)
	exportCmd.AddCommand(exportAddress)
	addCmd.AddCommand(addAddress)
	addAddress.Flags().StringVarP(&arClass, "rclass", "r", "",
		"Restriction class for this address")
	addAddress.Flags().BoolVarP(&aNoRclass, "no-rclass", "R", false,
		"Clear restriction class for this address")
	addAddress.Flags().StringVarP(&aTransport, "transport", "t", "",
		"Transport to be used for this address")
	addAddress.Flags().BoolVarP(&aNoTransport, "no-transport", "T", false,
		"Clear transport used by this address")
	deleteCmd.AddCommand(deleteAddress)
	editCmd.AddCommand(editAddress)
	editAddress.Flags().StringVarP(&arClass, "rclass", "r", "",
		"Restriction class for this address")
	editAddress.Flags().BoolVarP(&aNoRclass, "no-rclass", "R", false,
		"Clear restriction class for this address")
	editAddress.Flags().StringVarP(&aTransport, "transport", "t", "",
		"Transport to be used for this address")
	editAddress.Flags().BoolVarP(&aNoTransport, "no-transport", "T", false,
		"Clear transport used by this address")
	showCmd.AddCommand(showAddress)
}

// addressImport the addresss from inFile
func addressImport(cmd *cobra.Command, args []string) error {
	var err error

	mdb.Begin()
	defer mdb.End(&err)

	err = procImport(cmd, POSTFIX, procAddress)
	return err
}

// procAddress
// options are "option=XXX"
func procAddress(tokens []string) error {
	var (
		a   *maildb.Address
		err error
	)

	if a, err = mdb.InsertAddress(tokens[0]); err != nil {
		return err
	}
	if len(tokens) > 1 {
		for _, opt := range tokens[1:] {
			kv := strings.Split(opt, "=")
			if len(kv) < 2 {
				return fmt.Errorf("Address field %s is not a key=value pair", opt)
			}
			switch kv[0] {
			case "rclass":
				if kv[1] == "\"\"" { // fallback to domain rclass
					err = a.SetRclass(kv[1])
				}
			default:
				return fmt.Errorf("Unknown address field %s", kv[0])
			}
			if err != nil {
				break
			}
		}
	}
	return err
}

// addressExport the addresss to outFile
func addressExport(cmd *cobra.Command, args []string) error {
	var address string

	switch len(args) {
	case 0: // all addresss
		address = "*@*"
	case 1:
		address = args[0] // addresses by wildcard
	default:
		return fmt.Errorf("Only one address can be specified")
	}
	al, err := mdb.FindAddress(address)
	if err == nil {
		for _, a := range al {
			cmd.Printf("%s\n", a.Export())
		}
	}
	return err
}

// addressAdd the address and its class
func addressAdd(cmd *cobra.Command, args []string) error {
	var (
		a   *maildb.Address
		err error
	)

	mdb.Begin()
	defer mdb.End(&err)

	a, err = mdb.InsertAddress(args[0])
	if err == nil && cmd.Flags().Changed("rclass") {
		err = a.SetRclass(arClass)
	}
	if err == nil && cmd.Flags().Changed("transport") {
		err = a.SetTransport(aTransport)
	}
	return err
}

// addressDelete the address in the first arg
func addressDelete(cmd *cobra.Command, args []string) error {
	return mdb.DeleteAddress(args[0])
}

// addressEdit the address in the first arg
func addressEdit(cmd *cobra.Command, args []string) error {
	var (
		err error
		a   *maildb.Address
	)

	mdb.Begin()
	defer mdb.End(&err)

	a, err = mdb.GetAddress(args[0])
	if err == nil {
		if cmd.Flags().Changed("no-rclass") {
			err = a.ClearRclass()
		} else if cmd.Flags().Changed("rclass") {
			err = a.SetRclass(arClass)
		}
	}
	if err == nil {
		if cmd.Flags().Changed("no-transport") {
			err = a.ClearTransport()
		} else if cmd.Flags().Changed("transport") {
			err = a.SetTransport(aTransport)
		}
	}
	return err
}

// addressShow
func addressShow(cmd *cobra.Command, args []string) error {
	var (
		err error
		a   *maildb.Address
	)

	if a, err = mdb.LookupAddress(args[0]); err != nil {
		return err
	}
	cmd.Printf("Address:\t\t%s\nTransport:\t%s\nRestrictions:\t%s\n",
		a.Address(), a.Transport(), a.Rclass())
	return nil
}

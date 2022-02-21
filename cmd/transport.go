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
	transTransport string
	transNexthop   string
	noTransport    bool
	noNexthop      bool
)

// importTransport do import of an transport file
var importTransport = &cobra.Command{
	Use:   "transport",
	Short: "Import a file containing a transport and its name",
	Long:  "Import transports from a file named by the -i flag (default stdin '-').",
	Args:  cobra.NoArgs,
	RunE:  transportImport,
}

// exportTransport do export of an transport file
var exportTransport = &cobra.Command{
	Use:   "transport",
	Short: "Export transports to a file, one per line.",
	Long:  "Export transports to a file named by the -o flag (default stdout '-').",
	Args:  cobra.MaximumNArgs(1),
	RunE:  transportExport,
}

// addTransport do add of an transport rule
var addTransport = &cobra.Command{
	Use:   "transport name",
	Short: "Add a transport to the database.",
	Long:  "Add a named transport to the database with a transport matching transport(5) description.",
	Args:  cobra.ExactArgs(1),
	RunE:  transportAdd,
}

// deleteTransport delete an transport rule
var deleteTransport = &cobra.Command{
	Use:   "transport name",
	Short: "Delete a transport entry from the database.",
	Long:  "Delete the named transport entry from the database so long as no domain or address references it.",
	Args:  cobra.ExactArgs(1),
	RunE:  transportDelete,
}

// editTransport edit an transport rule
var editTransport = &cobra.Command{
	Use:   "transport name",
	Short: "Edit the named transport.",
	Long:  "Edit the transport and next hop attributes of the named transport.",
	Args:  cobra.ExactArgs(1),
	RunE:  transportEdit,
}

// showTransport diaplay rule contents
var showTransport = &cobra.Command{
	Use:   "transport name",
	Short: "Display the named transport.",
	Long:  "Display the contents of the named transport entry.",
	Args:  cobra.ExactArgs(1),
	RunE:  transportShow,
}

// linkage to top level
func init() {
	importCmd.AddCommand(importTransport)
	exportCmd.AddCommand(exportTransport)
	addCmd.AddCommand(addTransport)
	addTransport.Flags().StringVarP(&transTransport, "transport", "t", "",
		"Transport protoocol/method")
	addTransport.Flags().StringVarP(&transNexthop, "nexthop", "n", "",
		"Transport nexthop to send email")
	deleteCmd.AddCommand(deleteTransport)
	editCmd.AddCommand(editTransport)
	editTransport.Flags().StringVarP(&transTransport, "transport", "t", "",
		"Transport protoocol/method")
	editTransport.Flags().BoolVarP(&noTransport, "no-transport", "T", false,
		"Transport protoocol/method")
	editTransport.Flags().StringVarP(&transNexthop, "nexthop", "n", "",
		"Transport nexthop to send email")
	editTransport.Flags().BoolVarP(&noNexthop, "no-nexthop", "N", false,
		"Transport nexthop to send email")
	showCmd.AddCommand(showTransport)
}

// transportImport
func transportImport(cmd *cobra.Command, args []string) error {
	var err error

	mdb.Begin()
	defer mdb.End(&err)
	err = procImport(cmd, SIMPLE, procTransport)
	return err
}

// procTransport
func procTransport(tokens []string) error {
	var (
		tr  *maildb.Transport
		err error
	)

	if tr, err = mdb.InsertTransport(tokens[0]); err != nil {
		return err
	}
	kv := strings.SplitN(tokens[1], ":", 2)
	if len(kv) != 2 {
		err = fmt.Errorf("Transport value must include ':' to separate transport from nexthop, (%v)", kv)
	} else {
		if kv[0] != "" {
			err = tr.SetTransport(kv[0])
		}
		if err == nil && kv[1] != "" {
			err = tr.SetNexthop(kv[1])
		}
	}
	return err
}

// transportExport
func transportExport(cmd *cobra.Command, args []string) error {
	var name string

	switch len(args) {
	case 0: // All transports
		name = "*"
	case 1:
		name = args[0]
	default:
		return fmt.Errorf("Only one transport name can be specified")
	}
	tl, err := mdb.FindTransport(name)
	if err == nil {
		for _, tr := range tl {
			cmd.Printf("%s\n", tr.Export())
		}
	}
	return err
}

// transportAdd
func transportAdd(cmd *cobra.Command, args []string) error {
	var (
		err error
		tr  *maildb.Transport
	)

	mdb.Begin()
	defer mdb.End(&err)

	tr, err = mdb.InsertTransport(args[0])
	if err == nil && cmd.Flags().Changed("transport") {
		err = tr.SetTransport(transTransport)
	}
	if err == nil && cmd.Flags().Changed("nexthop") {
		err = tr.SetNexthop(transNexthop)
	}
	return err
}

// transportDelete
func transportDelete(cmd *cobra.Command, args []string) error {
	return mdb.DeleteTransport(args[0])
}

// transportEdit
func transportEdit(cmd *cobra.Command, args []string) error {
	var (
		err error
		tr  *maildb.Transport
	)

	mdb.Begin()
	defer mdb.End(&err)

	tr, err = mdb.GetTransport(args[0])
	if err == nil {
		if cmd.Flags().Changed("no-transport") {
			err = tr.ClearTransport()
		} else if cmd.Flags().Changed("transport") {
			err = tr.SetTransport(transTransport)
		}
	}
	if err == nil {
		if cmd.Flags().Changed("no-nexthop") {
			err = tr.ClearNextHop()
		} else if cmd.Flags().Changed("nexthop") {
			fmt.Printf("edit nexthop to %s\n", transNexthop)
			err = tr.SetNexthop(transNexthop)
		}
	}
	return err
}

// transportShow
func transportShow(cmd *cobra.Command, args []string) error {
	var (
		err error
		tr  *maildb.Transport
	)

	if tr, err = mdb.LookupTransport(args[0]); err != nil {
		return err
	}
	cmd.Printf("Name:\t\t%s\nTransport:\t%s\nNexthop:\t%s",
		tr.Name(), tr.Transport(), tr.Nexthop())
	return nil
}

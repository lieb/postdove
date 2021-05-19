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
	schemaFile string
)

// cmdCreate
func cmdCreate(cmd *cobra.Command, args []string) error {
	var (
		d   *maildb.Domain
		err error
	)
	if err = mdb.LoadSchema(schemaFile); err != nil {
		return err
	}
	mdb.Begin()
	defer mdb.End(&err)

	d, err = mdb.InsertDomain("localhost")
	if err == nil {
		err = d.SetClass("local")
	}
	if err == nil {
		d, err = mdb.InsertDomain("localhost.localdomain")
	}
	if err == nil {
		err = d.SetClass("local")
	}
	return err
}

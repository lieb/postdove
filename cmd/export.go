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
	"io"
	"os"

	"github.com/spf13/cobra"
)

var (
	outFilePath string
	outFile     *os.File
	outWriter   io.Writer
	savedOut    io.Writer
)

// exportRedirect
func exportRedirect(cmd *cobra.Command, args []string) error {
	var (
		err error
	)

	if cmd.Flags().Changed("output") {
		if outFilePath == "-" {
			outFile = os.Stdout
		} else {
			if outFile, err = os.Create(outFilePath); err != nil {
				cmd.PrintErrf("exportRedirect: error %s\n", err)
				return err
			}
		}
		outWriter = outFile
		savedOut = cmd.OutOrStdout()
		cmd.SetOut(outWriter)
	}
	return nil
}

// exportClose
func exportClose(cmd *cobra.Command, args []string) error {
	var (
		err error
	)

	if savedOut != nil {
		cmd.PrintErrf("exportClose: savedOut set\n")
		outWriter = cmd.OutOrStdout()
		if err = outFile.Close(); err != nil {
			return err
		}
		cmd.SetOut(savedOut)
		savedOut = nil
	}
	return nil
}

// export helpers

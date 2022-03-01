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
	pw_type    string
	password   string
	noPassword bool
	uid        int64
	noUid      bool
	gid        int64
	noGid      bool
	home       string
	noHome     bool
	quota      string
	enable     bool
)

// importMailbox do import of an mailboxes file
var importMailbox = &cobra.Command{
	Use:   "mailbox",
	Short: "Import a set of mailboxes from a file in a /etc/passwd format",
	Long: `Import a set of mailboxes into the database
from the file named by the -i flag (default stdin '-').`,
	Args: cobra.NoArgs,
	RunE: mailboxImport,
}

// exportMailbox do export of an mailboxes file
var exportMailbox = &cobra.Command{
	Use:   "mailbox",
	Short: "Export mailboxes in a /etc/passwd similar format",
	Long: `Export mailboxes in a /etc/passwd similar format to
the file named by the -o flag (default stdout '-').`,
	Args: cobra.MaximumNArgs(3),
	RunE: mailboxExport,
}

// addMailbox do add of an mailboxes file
var addMailbox = &cobra.Command{
	Use:   "mailbox address [ flags ]",
	Short: "Add an mailbox and its address into the database",
	Long: `Add an mailbox into the database. The address must be in an already
existing vmailbox domain. The flags set the various login parameters such as password and
quota.`,
	Args: cobra.ExactArgs(1), // mailbox recipient ...
	RunE: mailboxAdd,
}

// deleteMailbox do delete of an mailboxes file
var deleteMailbox = &cobra.Command{
	Use:   "mailbox address",
	Short: "Delete an mailbox and its address from the database.",
	Long: `Delete an address mailbox and its address from the database.
All of the aliases that point to it must be changed or deleted first`,
	Args: cobra.ExactArgs(1), // mailbox name
	RunE: mailboxDelete,
}

// editMailbox do edit of an mailboxes file
var editMailbox = &cobra.Command{
	Use:   "mailbox address [ flags ]",
	Short: "Edit the mailbox  for the address in the database",
	Long:  `Edit a mailbox to change attributes such as uid/gid, password, quota.`,
	Args:  cobra.MaximumNArgs(4), // mailbox to edit
	RunE:  mailboxEdit,
}

// showMailbox display the mailbox and its attributes
var showMailbox = &cobra.Command{
	Use:   "mailbox address",
	Short: "Display the mailbox",
	Long:  `Display the mailbox and its attributes to standard output`,
	Args:  cobra.ExactArgs(1), // can be wildcarded
	RunE:  mailboxShow,
}

// linkage to top level commands
func init() {
	importCmd.AddCommand(importMailbox)
	exportCmd.AddCommand(exportMailbox)
	addCmd.AddCommand(addMailbox)
	addMailbox.Flags().StringVarP(&pw_type, "type", "t", "PLAIN",
		"Password encoding type")
	addMailbox.Flags().StringVarP(&password, "password", "p", "",
		"Account password")
	addMailbox.Flags().Int64VarP(&uid, "uid", "u", 99, // nobody user
		"User ID for this mailbox")
	addMailbox.Flags().Int64VarP(&gid, "gid", "g", 99, // nobody group
		"User ID for this mailbox")
	addMailbox.Flags().StringVarP(&home, "mail-home", "m", "",
		"Home directory for mail")
	addMailbox.Flags().StringVarP(&quota, "quota", "q", "",
		"Storage quota")
	addMailbox.Flags().BoolVarP(&enable, "enable", "e", false,
		"Enable this mailbox for access")
	addMailbox.Flags().BoolVarP(&enable, "no-enable", "E", false,
		"Enable this mailbox for access")
	deleteCmd.AddCommand(deleteMailbox)
	editCmd.AddCommand(editMailbox)
	editMailbox.Flags().StringVarP(&pw_type, "type", "t", "PLAIN",
		"Password encoding type")
	editMailbox.Flags().StringVarP(&password, "password", "p", "",
		"Account password")
	editMailbox.Flags().BoolVarP(&noPassword, "no-password", "P", false,
		"Clear Account password")
	editMailbox.Flags().Int64VarP(&uid, "uid", "u", 99, // nobody user
		"User ID for this mailbox")
	editMailbox.Flags().BoolVarP(&noUid, "no-uid", "U", false, // nobody user
		"Clear User ID for this mailbox")
	editMailbox.Flags().Int64VarP(&gid, "gid", "g", 99, // nobody group
		"Group ID for this mailbox")
	editMailbox.Flags().BoolVarP(&noGid, "no-gid", "G", false, // nobody group
		"Clear Group ID for this mailbox")
	editMailbox.Flags().StringVarP(&home, "mail-home", "m", "",
		"Home directory for mail")
	editMailbox.Flags().BoolVarP(&noHome, "no-mail-home", "M", false,
		"Clear Home directory for mail")
	editMailbox.Flags().StringVarP(&quota, "quota", "q", "",
		"Storage quota")
	editMailbox.Flags().BoolVarP(&enable, "enable", "e", true,
		"Enable this mailbox for access")
	editMailbox.Flags().BoolVarP(&enable, "no-enable", "E", false,
		"Enable this mailbox for access")
	showCmd.AddCommand(showMailbox)
}

// mailboxImport the mailboxes from inFile
func mailboxImport(cmd *cobra.Command, args []string) error {
	var err error

	mdb.Begin()
	defer mdb.End(&err)

	err = procImport(cmd, PWFILE, procMailbox)
	return err
}

// procMailbox
// user:password:uid:gid:(gecos):home:(shell):extra_fields
// gecos and shell fields ignored (for now). quota is encoded in extra_fields
func procMailbox(tokens []string) error {
	var (
		mb               *maildb.VMailbox
		pwType, password string
		err              error
	)

	if len(tokens) < 2 {
		return fmt.Errorf("Must have at least a user field and a password field")
	}
	// tokens[0] account email address
	if mb, err = mdb.InsertVMailbox(tokens[0]); err != nil {
		return err
	}
	// tokens[1] password with possible "{type}" field
	if i := strings.Index(tokens[1], "{"); i >= 0 {
		j := strings.Index(tokens[1], "}")
		if j == -1 || j < i {
			return fmt.Errorf("Badly formed password type field")
		}
		pwType = strings.Trim(tokens[1][i+1:j], " \t")
		if err = mb.SetPwType(pwType); err != nil {
			return err
		}
		password = strings.Trim(tokens[1][j+1:], " \t")
		if err = mb.SetPassword(password); err != nil {
			return err
		}
	} else {
		if err = mb.SetPassword(strings.Trim(tokens[1], " \t")); err != nil {
			return err
		}
	}
	// optional tokens[2] uid (if not null field)
	if len(tokens) > 2 && len(tokens[2]) > 0 {
		if uid, err = strconv.ParseInt(tokens[2], 10, 64); err != nil {
			return fmt.Errorf("Imported uid field: %s", err)
		}
		if err = mb.SetUid(uid); err != nil {
			return err
		}
	}
	// optional tokens[3] gid (if not null field)
	if len(tokens) > 3 && len(tokens[3]) > 0 {
		if gid, err = strconv.ParseInt(tokens[3], 10, 64); err != nil {
			return fmt.Errorf("Imported gid field: %s", err)
		}
		if err = mb.SetGid(gid); err != nil {
			return err
		}
	}
	// optional tokens[5] home, skip tokens[4] gecos field (remember that?)
	if len(tokens) > 5 && len(tokens[5]) > 0 {
		if err = mb.SetHome(strings.Trim(tokens[5], " \t")); err != nil {
			return err
		}
	}
	// optional tokens[7] extra fields, skip tokens[6] shell field.
	if len(tokens) > 7 && len(tokens[7]) > 0 {
		tokens[7] = strings.Join(tokens[7:], ":")
		ef := strings.Fields(tokens[7])
		for _, f := range ef {
			kv := strings.Split(f, "=")
			if len(kv) < 2 {
				return fmt.Errorf("Extra field \"%s\" is not a key=value pair", f)
			}
			switch kv[0] {
			case "userdb_quota_rule":
				if len(kv) == 2 && strings.ToLower(kv[1]) == "none" {
					err = mb.ClearQuota()
				} else if len(kv) == 3 {
					err = mb.SetQuota(kv[1] + "=" + kv[2])
				} else {
					err = fmt.Errorf("Badly formatted quota rule")
				}
				if err != nil {
					return err
				}
			case "mbox_enabled":
				if e, err := strconv.ParseBool(kv[1]); err != nil {
					return err
				} else {
					if e {
						err = mb.Enable()
					} else {
						err = mb.Disable()
					}
				}
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("Unknown extra field")
			}
		}
	}
	return err
}

// mailboxExport the mailboxes to outFile
func mailboxExport(cmd *cobra.Command, args []string) error {
	var vMailbox string

	switch len(args) {
	case 0: // all vMailboxs
		vMailbox = "*@*"
	case 1:
		vMailbox = args[0] // vMailboxs by wildcard
	default:
		return fmt.Errorf("Only one vMailbox can be specified")
	}
	ml, err := mdb.FindVMailbox(vMailbox)
	if err == nil {
		for _, m := range ml {
			cmd.Printf("%s\n", m.Export())
		}
	}
	return err
}

// mailboxAdd the mailbox and its address
func mailboxAdd(cmd *cobra.Command, args []string) error {
	var (
		mb  *maildb.VMailbox
		err error
	)

	mdb.Begin()
	defer mdb.End(&err)

	mb, err = mdb.InsertVMailbox(args[0])
	// use flags to add stuff
	if err == nil && cmd.Flags().Changed("type") {
		err = mb.SetPwType(pw_type)
	}
	if err == nil && cmd.Flags().Changed("password") {
		err = mb.SetPassword(password)
	}
	if err == nil && cmd.Flags().Changed("uid") {
		err = mb.SetUid(uid)
	}
	if err == nil && cmd.Flags().Changed("gid") {
		err = mb.SetGid(gid)
	}
	if err == nil && cmd.Flags().Changed("mail-home") {
		err = mb.SetHome(home)
	}
	if err == nil && cmd.Flags().Changed("quota") {
		if quota == "none" {
			err = mb.ClearQuota()
		} else {
			err = mb.SetQuota(quota)
		}
	}
	// bools are a bit strange and require "=" to set as they
	// are expected to be toggles. This treats them as toggles
	// They will also only accept "true" or "false" unlike strconv.ParseBool()...
	if err == nil && cmd.Flags().Changed("enable") {
		err = mb.Enable()
	}
	if err == nil && cmd.Flags().Changed("no-enable") {
		err = mb.Disable()
	}
	return err
}

// mailboxDelete the mailbox and address in the first arg
func mailboxDelete(cmd *cobra.Command, args []string) error {
	return mdb.DeleteVMailbox(args[0])
}

// mailboxEdit the mailbox of the address in the first arg
func mailboxEdit(cmd *cobra.Command, args []string) error {
	var (
		mb  *maildb.VMailbox
		err error
	)

	mdb.Begin()
	defer mdb.End(&err)

	mb, err = mdb.GetVMailbox(args[0])
	// use flags to add stuff
	if err == nil && cmd.Flags().Changed("type") {
		err = mb.SetPwType(pw_type)
	}
	if err == nil {
		if cmd.Flags().Changed("no-password") {
			err = mb.ClearPassword()
		} else if cmd.Flags().Changed("password") {
			err = mb.SetPassword(password)
		}
	}
	if err == nil {
		if cmd.Flags().Changed("no-uid") {
			err = mb.ClearUid()
		} else if cmd.Flags().Changed("uid") {
			err = mb.SetUid(uid)
		}
	}
	if err == nil {
		if cmd.Flags().Changed("no-gid") {
			err = mb.ClearGid()
		} else if cmd.Flags().Changed("gid") {
			err = mb.SetGid(gid)
		}
	}
	if err == nil {
		if cmd.Flags().Changed("no-mail-home") {
			err = mb.ClearHome()
		} else if cmd.Flags().Changed("mail-home") {
			err = mb.SetHome(home)
		}
	}
	if err == nil && cmd.Flags().Changed("quota") {
		if strings.ToLower(quota) == "none" {
			err = mb.ClearQuota()
		} else if strings.ToLower(quota) == "reset" {
			err = mb.ResetQuota()
		} else {
			err = mb.SetQuota(quota)
		}
	}
	if err == nil {
		if cmd.Flags().Changed("enable") {
			err = mb.Enable()
		} else if cmd.Flags().Changed("no-enable") {
			err = mb.Disable()
		}
	}
	return err
}

// mailboxShow
func mailboxShow(cmd *cobra.Command, args []string) error {
	var (
		err         error
		ml          []*maildb.VMailbox
		MoreThanOne bool
	)

	if ml, err = mdb.FindVMailbox(args[0]); err != nil {
		return err
	}
	for _, m := range ml {
		if MoreThanOne {
			cmd.Printf("=====================\n")
		}
		cmd.Printf("Name:\t\t%s\nPassword Type:\t%s\nPassword:\t%s\n",
			m.User(), m.PwType(), m.Password())
		cmd.Printf("UserID:\t\t%s\nGroupID:\t%s\nHome:\t\t%s\nQuota:\t\t%s\n",
			m.Uid(), m.Gid(), m.Home(), m.Quota())
		if m.IsEnabled() {
			cmd.Printf("Enabled:\ttrue\n")
		} else {
			cmd.Printf("Enabled:\tfalse\n")
		}
		MoreThanOne = true
	}
	return nil
}

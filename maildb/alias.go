package maildb

/*
 * Copyright (C) 2020, Jim Lieb <lieb@sea-troll.net>
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public
 * License as published by the Free Software Foundation; either
 * version 3 of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public
 * License along with this library; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 *
 * -------------
 */

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3" // do I really need this here?
)

// Recipient in alias list
type Recipient struct {
	id  int64 // id of alias row
	ext sql.NullString
	t   *Address
}

// String
// beware! virtual aliases cannot have /etc/aliases attributes (pipes and stuff)
func (tg *Recipient) String() string {
	var (
		line strings.Builder
	)
	if tg.t != nil {
		if tg.ext.Valid {
			fmt.Fprintf(&line, "%s+%s", tg.t.localpart, tg.ext.String)
			if !tg.t.IsLocal() {
				fmt.Fprintf(&line, "@%s", tg.t.d.Name())
			}
		} else {
			fmt.Fprintf(&line, "%s", tg.t.String())
		}
	} else if tg.ext.Valid {
		fmt.Fprintf(&line, "%s", tg.ext.String)
	} // panic else here? CHECK constraint on alias should apply on insert/update
	return line.String()
}

// Alias
type Alias struct {
	addr   *Address
	recips []*Recipient
}

// Id
func (al *Alias) Id() int64 {
	return al.addr.id
}

// Targets
func (al *Alias) Targets() []*Recipient {
	return al.recips
}

// LookupAlias
// get either "local_user" or "mbox@domain" aliases
// name@domain returns that alias recipients for this address
// *           returns all local (/etc/aliases) aliases
// *@domain    returns all aliases in this domain
// name@*      returns all aliases of this name, e.g. abuse@foo.com, abuse@example.org
// *@*         returns all virtual aliases in the database
func (mdb *MailDB) LookupAlias(alias string) ([]*Alias, error) {
	var (
		al_list []*Alias
		a_list  []*Address
		err     error
		rowCnt  int
	)

	if a_list, err = mdb.FindAddress(alias); err != nil {
		return nil, err
	}
	for _, a := range a_list {
		al, err := a.Alias()
		if err != nil {
			if err == ErrMdbNotAlias {
				continue
			} else {
				break
			}
		}
		al_list = append(al_list, al)
		rowCnt++
	}
	if err == nil && rowCnt == 0 {
		err = ErrMdbNoAliases
	}
	return al_list, err
}

// String
// return a line for this alias
// Note that /etc/aliases is a different syntax from virtual(5)
func (al *Alias) String() string {
	var (
		line   strings.Builder
		commas int
	)

	if al.addr.IsLocal() {
		fmt.Fprintf(&line, "%s: ", al.addr.String())
	} else {
		fmt.Fprintf(&line, "%s ", al.addr.String())
	}
	for _, r := range al.recips {
		if commas > 0 {
			fmt.Fprintf(&line, ", ")
		}
		fmt.Fprintf(&line, "%s", r.String())
		commas++
	}
	return line.String()
}

// RemoveAlias and all its targets
// All we need to do here is delete the aliases that aliasAddr points to
// As the set of aliases disappear, their delete triggers clean up all the
// orphan targets (and the alias address itself) on the way out
func (mdb *MailDB) RemoveAlias(alias string) error {
	var (
		ap  *AddressParts
		err error
		c   int64
		res sql.Result
	)

	if ap, err = DecodeRFC822(alias); err != nil {
		return err
	}
	if ap.IsLocal() {
		qd := `
DELETE FROM alias WHERE address =
(SELECT a.id FROM address a  WHERE a.domain IS NULL AND a.localpart = ?)
`
		res, err = mdb.db.Exec(qd, ap.lpart)
	} else {
		qd := `
DELETE FROM alias WHERE address =
(SELECT a.id FROM address a, domain d
  WHERE a.domain = d.id AND a.localpart = ? AND d.name = ?)
`
		res, err = mdb.db.Exec(qd, ap.lpart, ap.domain)
	}
	if err == nil {
		c, err = res.RowsAffected()
		if err == nil {
			if c < 1 {
				err = ErrMdbNotAlias
			}
		}
	}
	return err
}

// RemoveRecipient. Remove the alias as well if this is the last target
func (mdb *MailDB) RemoveRecipient(alias string, recipient string) error {
	var (
		ap  *AddressParts
		err error
		c   int64
		rp  *AddressParts
		res sql.Result
	)

	if ap, err = DecodeRFC822(alias); err != nil {
		return err
	}
	if rp, err = DecodeTarget(recipient); err != nil {
		return err
	}
	if ap.IsLocal() {
		if rp.IsPipe() {
			qd := `
DELETE FROM alias WHERE target IS NULL AND extension IS ? AND address =
  (SELECT id FROM address WHERE localpart = ? AND domain IS NULL)
`
			res, err = mdb.db.Exec(qd, rp.extension, ap.lpart)
		} else {
			qd := `
DELETE FROM alias WHERE address =
  (SELECT id FROM address WHERE localpart = ? AND domain IS NULL)
`
			if rp.IsLocal() {
				qd += `
 AND target = (SELECT id from address WHERE localpart = ? AND domain IS NULL)
`
				res, err = mdb.db.Exec(qd, ap.lpart, rp.lpart)
			} else {
				qd += `
 AND target = (SELECT a.id from address a, domain d
   WHERE a.localpart = ? AND a.domain = d.id AND d.name = ?)
`
				res, err = mdb.db.Exec(qd, ap.lpart, rp.lpart, rp.domain)
			}
		}
	} else { // name@domain
		qd := `
DELETE FROM alias WHERE address =
  (SELECT a.id FROM address a, domain d WHERE a.localpart = ? AND a.domain = d.id AND d.name = ? )
`
		if !rp.IsPipe() {
			if rp.IsLocal() {
				qd += `
 AND target = (SELECT id from address WHERE localpart = ? AND domain IS NULL)
`
				res, err = mdb.db.Exec(qd, ap.lpart, ap.domain, rp.lpart)
			} else {
				qd += `
 AND target = (SELECT a.id from address a, domain d
   WHERE a.localpart = ? AND a.domain = d.id AND d.name = ?)
`
				res, err = mdb.db.Exec(qd, ap.lpart, ap.domain, rp.lpart, rp.domain)
			}
		} else {
			err = ErrMdbNoLocalPipe // for name@domain virtuals
		}
	}
	if err == nil {
		c, err = res.RowsAffected()
		if err == nil {
			if c < 1 {
				err = ErrMdbRecipientNotFound
			}
		}
	}
	return err
}

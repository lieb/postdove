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

// RowID - let's not have confusion over what Sqlite considers an INTEGER row id...
type RowID int64

type MailDB struct {
	db         *sql.DB
	transports map[int64]*Transport
}

func NewMailDB(db *sql.DB) *MailDB {
	mdb := &MailDB{
		db:         db,
		transports: make(map[int64]*Transport),
	}
	return mdb
}

// AddressParts
type AddressParts struct {
	lpart     string
	domain    string
	extension string
}

// DecodeRFC822 Decode an RFC822 address into its constituent parts
// Actually, we decode per RFC5322
func DecodeRFC822(addr string) (*AddressParts, error) {
	var (
		local  string = ""
		domain string = ""
		// extension is transparent here and embedded in local
	)
	a := strings.ToLower(strings.Trim(addr, " "))    // clean up and lower everything
	if strings.ContainsAny(a, "\n\r\t\f{}()[];\"") { // contains illegal cruft
		return nil, fmt.Errorf("DecodeRFC822: %s contains illegal characters", addr)
	}
	if strings.Contains(a, "@") { // local@fqdn
		at := strings.Index(a, "@")
		local = a[0:at]
		domain = a[at+1:]
	} else { // just local
		local = a
	}
	return &AddressParts{
		lpart:     local,
		domain:    domain,
		extension: "",
	}, nil
}

// DecodeTarget Decode an RFC822 address and the various options for extensions
func DecodeTarget(addr string) (*AddressParts, error) {
	var (
		loc string = ""
		dom string = ""
		ext string = ""
	)
	ap, err := DecodeRFC822(addr)
	if err != nil {
		return nil, fmt.Errorf("DecodeTarget: %s", err)
	}
	if strings.Contains(ap.lpart, "+") { // we have an address extension
		pl := strings.Index(ap.lpart, "+")
		loc = ap.lpart[0:pl]
		ext = ap.lpart[pl+1:]
		dom = ap.domain
	} else if ap.lpart[0] == '/' || ap.lpart[0] == '|' { // a local pipe or file redirect
		if ap.domain != "" { // can't have a domain for /etc/aliases targets
			return nil, fmt.Errorf("DecodeTarget: %s cannot have a domain for locals",
				addr)
		}
		ext = ap.lpart
	} else if ap.lpart[0] == ':' { // an include
		if len(ap.lpart) < 10 || ap.lpart[:9] != ":include:" { // bad include
			return nil, fmt.Errorf("DecodeTarget: \"%s\" is badly formed include",
				ap.lpart)
		}
		ext = ap.lpart
	} else { // target is a clean RFC822
		return ap, nil
	}
	return &AddressParts{
		lpart:     loc,
		domain:    dom,
		extension: ext,
	}, nil
}

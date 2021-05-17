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
	"fmt"
	"strings"
)

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
		ext           = ""
		// extension is transparent here and embedded in local
	)

	if addr == "" {
		return nil, ErrMdbTargetEmpty
	}
	a := strings.ToLower(strings.Trim(addr, " "))    // clean up and lower everything
	if strings.ContainsAny(a, "\n\r\t\f{}()[];\"") { // contains illegal cruft
		return nil, ErrMdbAddrIllegalChars
	}
	if strings.Contains(a, "@") { // local@fqdn
		at := strings.Index(a, "@")
		local = a[0:at]
		domain = a[at+1:]
	} else { // just local
		local = a
	}
	if strings.Contains(local, "+") { // we have an address extension
		pl := strings.Index(local, "+")
		loc := local[0:pl]
		if loc == "" {
			return nil, ErrMdbAddrNoAddr
		}
		ext = local[pl+1:]
		local = loc
	}
	return &AddressParts{
		lpart:     local,
		domain:    domain,
		extension: ext,
	}, nil
}

// DecodeTarget Decode an RFC822 address and the various options for extensions
func DecodeTarget(addr string) (*AddressParts, error) {
	ap := &AddressParts{
		lpart:     "",
		domain:    "",
		extension: addr,
	}
	if addr == "" {
		return nil, ErrMdbTargetEmpty
	} else if addr[0] == '/' { // a file redirect
		if len(addr) > 1 {
			return ap, nil
		} else {
			return nil, ErrMdbNoLocalPipe
		}
	} else if addr[0] == '|' { // a local unquoted pipe
		if len(addr) > 1 {
			if strings.ContainsAny(addr, " \t") {
				return nil, ErrMdbNoQuotedSpace
			} else {
				return ap, nil
			}
		} else {
			return nil, ErrMdbNoLocalPipe
		}
	} else if addr[0] == '"' { // a quoted local pipe
		if len(addr) > 1 {
			if addr[1] != '|' || !strings.HasSuffix(addr, "\"") {
				return nil, ErrMdbNoQuotedSpace
			} else {
				return ap, nil
			}
		} else {
			return nil, ErrMdbNoLocalPipe
		}
	} else if addr[0] == ':' {
		if len(addr) > 10 && addr[:9] == ":include:" { // an include
			return ap, nil
		} else {
			return nil, ErrMdbBadInclude
		}
	} else {
		return DecodeRFC822(addr)
	}
}

// IsPipe
func (ap *AddressParts) IsPipe() bool {
	return ap.lpart == "" && ap.domain == ""
}

// IsLocal
func (ap *AddressParts) IsLocal() bool {
	return ap.lpart != "" && ap.domain == ""
}

func (ap *AddressParts) String() string {
	var (
		line strings.Builder
	)
	if ap.lpart != "" {
		fmt.Fprintf(&line, "%s", ap.lpart)
		if ap.extension != "" {
			fmt.Fprintf(&line, "+%s", ap.extension)
		}
		if ap.domain != "" {
			fmt.Fprintf(&line, "@%s", ap.domain)
		}
	} else if ap.domain != "" {
		fmt.Fprintf(&line, "@%s", ap.domain)
	} else {
		fmt.Fprintf(&line, ap.extension)
	}
	return line.String()
}

func (ap *AddressParts) dump() string {
	var (
		line strings.Builder
	)
	fmt.Fprintf(&line, "lpart:%s, domain:%s, ext:%s.", ap.lpart, ap.domain, ap.extension)
	return line.String()
}

type TransportParts struct {
	transport string
	nexthop   string
}

// DecodeTransport
func DecodeTransport(trans string) (*TransportParts, error) {
	i := strings.Index(trans, ":")
	if i >= 0 {
		t := &TransportParts{
			transport: trans[0:i],
			nexthop:   trans[i+1:],
		}
		return t, nil
	} else {
		return nil, ErrMdbTransNoColon
	}
}

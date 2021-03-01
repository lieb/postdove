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
	//"database/sql"
	"fmt"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3" // do I really need this here?
)

type RFC822Res struct {
	addr   string
	lpart  string
	domain string
	ext    string
}

// TestDecode
func TestDecode(t *testing.T) {
	fmt.Printf("RFC822 Test\n")

	DecodeRes := []RFC822Res{
		{"foo", "foo", "", ""},
		{"foo@baz", "foo", "baz", ""},
		{"foo+bar@baz", "foo", "baz", "bar"},
		{"@baz", "", "baz", ""},
	}
	var (
		ap  *AddressParts
		err error
	)

	for _, r := range DecodeRes {
		ap, err = DecodeRFC822(r.addr)
		if err != nil {
			t.Errorf("Parsing \"%s\" throws an error %s", r.addr, err)
		}
		if ap.lpart != r.lpart || ap.domain != r.domain || ap.extension != r.ext {
			t.Errorf("%s: lpart = %s, domain = %s, extension = %s",
				r.addr, ap.lpart, ap.domain, ap.extension)
		}
		if r.addr != ap.String() {
			t.Errorf("%s != %s", r.addr, ap.String())
		}
	}
	ap, err = DecodeRFC822("foo{bar@baz")
	if err == nil {
		t.Errorf("foo/bar@baz: did not throw illegal char error")
	} else if err != ErrMdbAddrIllegalChars {
		t.Errorf("foo/bar@baz: err code, %s", err)
	}
	ap, err = DecodeRFC822("+bar@baz")
	if err == nil {
		t.Errorf("+bar@baz did not throw address extension without user part error")
	} else if err != ErrMdbAddrNoAddr {
		t.Errorf("+bar@baz: err code, %s", err)
	}
}

// TestTarget
func TestTarget(t *testing.T) {
	fmt.Printf("Target Test\n")

	TargetRes := []RFC822Res{
		{"foo", "foo", "", ""},
		{"foo@baz", "foo", "baz", ""},
		{"foo+bar@baz", "foo", "baz", "bar"},
		{"Foo+baR@bAz", "foo", "baz", "bar"},
		{"| cat foo", "", "", "| cat foo"},
		{"| cat Foo", "", "", "| cat Foo"},
		{"/dev/null", "", "", "/dev/null"},
		{":include:everybody.txt", "", "", ":include:everybody.txt"},
	}
	var (
		ap  *AddressParts
		err error
	)

	for _, r := range TargetRes {
		ap, err = DecodeTarget(r.addr)
		if err != nil {
			t.Errorf("Parsing \"%s\" throws an error %s", r.addr, err)
		}
		if ap.lpart != r.lpart || ap.domain != r.domain || ap.extension != r.ext {
			t.Errorf("%s: lpart = %s, domain = %s, extension = %s",
				r.addr, ap.lpart, ap.domain, ap.extension)
		}
		if r.addr != ap.String() {
			if strings.Contains(r.addr, "@") && strings.ToLower(strings.Trim(r.addr, " ")) != ap.String() {
				t.Errorf("%s != %s", r.addr, ap.String())
			}
		}
	}
	ap, err = DecodeTarget("foo{bar@baz")
	if err == nil {
		t.Errorf("foo/bar@baz: did not throw illegal char error")
	} else if err != ErrMdbAddrIllegalChars {
		t.Errorf("err code: %s", err)
	}
	ap, err = DecodeTarget(":bogus:")
	if err == nil {
		t.Errorf(":bogus: did not throw illegal char error")
	} else if err != ErrMdbBadInclude {
		t.Errorf("DecodeTarget: unexpected err code: %s", err)
	}
}

type TransportRes struct {
	trans     string
	transport string
	nexthop   string
}

// TestTransport
func TestTransport(t *testing.T) {
	fmt.Printf("Transport Decode Test\n")

	res := []TransportRes{
		{":", "", ""},
		{"smtp:", "smtp", ""},
		{":some.domain", "", "some.domain"},
		{"uucp:example.com", "uucp", "example.com"},
		{"relay:[gateway.com]", "relay", "[gateway.com]"},
		{"smtp:bar.example.com:25", "smtp", "bar.example.com:25"},
		{"error:mail for you bounces", "error", "mail for you bounces"},
	}
	var (
		tr  *TransportParts
		err error
	)

	for _, r := range res {
		tr, err = DecodeTransport(r.trans)
		if err != nil {
			t.Errorf("Parsing \"%s\" throws an error %s", r.trans, err)
		}
		if tr.transport != r.transport || tr.nexthop != r.nexthop {
			t.Errorf("%s, transport = %s, nexthop = %s",
				r.trans, tr.transport, tr.nexthop)
		}
	}
	tr, err = DecodeTransport("foo")
	if err == nil {
		t.Errorf("foo: did not throw a no separator error")
	}
}

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

	_ "github.com/mattn/go-sqlite3" // do I really need this?
)

// Transport DB table
type Transport struct {
	lpart     sql.NullString
	aname     sql.NullString
	id        int64
	transport sql.NullString
	nexthop   sql.NullInt64
	nextname  sql.NullString
	mx        sql.NullInt64 // default 1 (true)
	port      sql.NullInt64
}

// GetTransport
// The recip is in three forms:
// *    "" retrieves all transports (for dumping...)
// *    "user@domain" returns transport for this address with priority for user over domain
// *    "domain" returns transport for all users in domain
//
// There are special cases marginally handled at this point.
// *    ".domain" this is for subdomains. Requires a separate entry which messes with
//      the domain table. Untested YMMV
// *    "user+ext@domain" this is for local part extensions. Requires a separate entry
//      in address table that may mess with it. Untested YMMV

func (mdb *MailDB) GetTransport(recip string) ([]*Transport, error) {
	var (
		rows   *sql.Rows
		row    *sql.Row
		err    error
		ap     *AddressParts
		lpart  sql.NullString
		aname  sql.NullString
		t_list []*Transport
	)

	q := `
SELECT DISTINCT t.id AS id, t.transport AS trans, t.nexthop AS next,
       CASE WHEN t.nexthop IS NULL
       THEN NULL
       ELSE
         (SELECT name FROM domain WHERE id = t.nexthop)
       END AS nextname,
       t.mx AS mx, t.port AS port
FROM transport AS t
`
	if recip == "" { // Get all transports
		rows, err = mdb.db.Query(q)
	} else {
		var trans sql.NullInt64

		ap, err = DecodeRFC822(recip)
		if err != nil {
			return nil, fmt.Errorf("GetTransport: decode %s, %s",
				recip, err)
		}
		if len(ap.lpart) > 0 && len(ap.domain) == 0 { // assume lpart is @domain
			aname = sql.NullString{
				Valid:  true,
				String: ap.lpart,
			}
			qd := `SELECT transport FROM domain WHERE name = ?`
			row = mdb.db.QueryRow(qd, ap.lpart)
		} else { // it is user@domain
			lpart = sql.NullString{
				Valid:  true,
				String: ap.lpart,
			}
			aname = sql.NullString{
				Valid:  true,
				String: ap.domain,
			}
			qa := `
SELECT CASE WHEN a.transport IS NULL
       THEN d.transport ELSE NULL END AS trans
FROM address AS a, domain AS d
WHERE a.localpart = ? AND a.domain = d.id AND d.name = ?
`
			row = mdb.db.QueryRow(qa, ap.lpart, ap.domain)
		}
		switch err := row.Scan(&trans); err {
		case sql.ErrNoRows:
			return nil, fmt.Errorf("GetTransport: No %s", recip)
		case nil:
			break
		default:
			panic(err)
		}
		if !trans.Valid {
			return nil, fmt.Errorf("GetTransport: %s has no transport", recip)
		}
		q += `WHERE id = ?`
		rows, err = mdb.db.Query(q, trans.Int64)
	}
	if err != nil {
		return nil, fmt.Errorf("GetTransport: query = %s, %s", q, err)
	}
	for rows.Next() {
		var (
			id        int64
			transport sql.NullString
			nexthop   sql.NullInt64
			nextname  sql.NullString
			mx        sql.NullInt64
			port      sql.NullInt64
		)
		err = rows.Scan(&id, &transport, &nexthop, &nextname, &mx, &port)
		if err != nil {
			return nil, fmt.Errorf("GetTransport: scan, %s", err)
		}
		if ap == nil { // no recip, find one...
			lpart = sql.NullString{Valid: false}
			aname = sql.NullString{Valid: false}
			qa := `
SELECT a.localpart, d.name FROM address AS a, domain AS d
WHERE a.domain = d.id AND a.transport = ?
`
			row := mdb.db.QueryRow(qa, id)
			switch err = row.Scan(&lpart, &aname); err {
			case sql.ErrNoRows:
				qd := `SELECT name FROM domain WHERE transport = ?`
				row = mdb.db.QueryRow(qd, id)
				switch err = row.Scan(&aname); err {
				case sql.ErrNoRows:
				case nil:
					break
				default:
					panic(err)
				}
			case nil:
				break
			default:
				panic(err)
			}
		}
		t := &Transport{
			lpart:     lpart,
			aname:     aname,
			id:        id,
			transport: transport,
			nexthop:   nexthop,
			nextname:  nextname,
			mx:        mx,
			port:      port,
		}
		t_list = append(t_list, t)
	}
	return t_list, nil
}

// String
func (t *Transport) String() string {
	var line strings.Builder

	if t.lpart.Valid {
		fmt.Fprintf(&line, "%s@%s\t", t.lpart.String, t.aname.String)
	} else if t.aname.Valid {
		fmt.Fprintf(&line, "%s\t", t.aname.String)
	} else {
		fmt.Fprintf(&line, "\t")
	}
	if t.transport.Valid {
		fmt.Fprintf(&line, "%s:", t.transport.String)
	} else {
		fmt.Fprintf(&line, ":")
	}
	if t.nextname.Valid {
		if t.mx.Valid && t.mx.Int64 == 0 {
			fmt.Fprintf(&line, "[%s]", t.nextname.String)
		} else {
			fmt.Fprintf(&line, "%s", t.nextname.String)
		}
		if t.port.Valid {
			fmt.Fprintf(&line, ":%d", t.port.Int64)
		}
	}
	return line.String()
}
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
	nexthop   sql.NullString
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

	q := `SELECT DISTINCT id, transport, nexthop FROM transport `
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
			aname = sql.NullString{Valid: true, String: ap.lpart}
			qd := `SELECT transport FROM domain WHERE name = ?`
			row = mdb.db.QueryRow(qd, ap.lpart)
		} else { // it is user@domain
			lpart = sql.NullString{Valid: true, String: ap.lpart}
			aname = sql.NullString{Valid: true, String: ap.domain}
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
			nexthop   sql.NullString
		)
		err = rows.Scan(&id, &transport, &nexthop)
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
		}
		t_list = append(t_list, t)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("GetTransport: Next loop, %s", err)
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
	if t.nexthop.Valid {
		fmt.Fprintf(&line, "%s", t.nexthop.String)
	}
	return line.String()
}

// NewTransport
func (mdb *MailDB) NewTransport(t string) error {
	var (
		err     error
		tr      *TransportParts
		trans   sql.NullString
		nexthop sql.NullString
	)

	if tr, err = DecodeTransport(t); err != nil {
		return fmt.Errorf("NewTransport: %s", err)
	}
	if tr.transport == "" {
		trans = sql.NullString{Valid: false}
	} else {
		trans = sql.NullString{Valid: true, String: tr.transport}
	}
	if tr.nexthop == "" {
		nexthop = sql.NullString{Valid: false}
	} else {
		nexthop = sql.NullString{Valid: true, String: tr.nexthop}
	}
	// Enter a transaction for everything else
	if mdb.tx, err = mdb.db.Begin(); err != nil {
		return fmt.Errorf("MakeAlias: begin, %s", err)
	}
	defer func() {
		if err == nil {
			if err = mdb.tx.Commit(); err != nil {
				panic(fmt.Errorf("MakeAlias: commit, %s", err)) // we are screwed
			}
		} else {
			mdb.tx.Rollback()
		}
	}()
	_, err = mdb.tx.Exec("INSERT INTO transport (transport, nexthop) VALUES( ?, ?)",
		trans, nexthop)
	if err != nil {
		return fmt.Errorf("NewTransport: insert, %s", err)
	}
	return nil
}

// AttachTransport
// for smtp and lmtp transports, we can have an ordered list of destinations
// ex: smtp:dest1, dest2, ...
// where dest1 is tried first and destN if dest1 fails
func (mdb *MailDB) AttachTransport(addr string, t string) error {
	var (
		err     error
		ap      *AddressParts
		tr      *TransportParts
		row     *sql.Row
		trans   sql.NullString
		nexthop sql.NullString
		tID     int64
	)

	if ap, err = DecodeRFC822(addr); err != nil {
		return fmt.Errorf("AttachTransport: %s", err)
	}
	if tr, err = DecodeTransport(t); err != nil {
		return fmt.Errorf("AttachTransport: %s", err)
	}
	if tr.transport == "" {
		trans = sql.NullString{Valid: false}
	} else {
		trans = sql.NullString{Valid: true, String: tr.transport}
	}
	if tr.nexthop == "" {
		nexthop = sql.NullString{Valid: false}
	} else {
		nexthop = sql.NullString{Valid: true, String: tr.nexthop}
	}
	// Enter a transaction for everything else
	if mdb.tx, err = mdb.db.Begin(); err != nil {
		return fmt.Errorf("AttachTransport: begin, %s", err)
	}
	defer func() {
		if err == nil {
			if err = mdb.tx.Commit(); err != nil {
				panic(fmt.Errorf("AttachTransport: commit, %s", err))
				// we are screwed
			}
		} else {
			mdb.tx.Rollback()
		}
	}()
	row = mdb.db.QueryRow(
		"SELECT id FROM transport WHERE transport = ? and nexthop = ?",
		trans, nexthop)
	switch err = row.Scan(&tID); err {
	case sql.ErrNoRows:
		return fmt.Errorf("AttachTransport: transport does not exist")
	case nil:
		break
	default:
		return fmt.Errorf("AttachTransport: select transport, %s", err)
	}
	if ap.domain == "" { // lpart is really a domain in this context so attach to it
		var dID int64

		row = mdb.db.QueryRow("SELECT id FROM domain WHERE name = ?", ap.lpart)
		switch err = row.Scan(&dID); err {
		case sql.ErrNoRows:
			return fmt.Errorf("AttachTransport: domain does not exist")
		case nil:
			break
		default:
			fmt.Errorf("AttachTransport: select domain, %s", err)
		}
		_, err = mdb.tx.Exec("UPDATE domain SET transport = ? WHERE id = ?",
			tID, dID)
		if err != nil {
			return fmt.Errorf("AttachTransport: update domain, %s", err)
		}
	} else { // a user@domain specific transport
		var (
			a *Address
		)

		if a, err = mdb.GetAddress(addr); err != nil {
			return fmt.Errorf("AttachTransport: %s", err)
		}
		_, err = mdb.tx.Exec("UPDATE address SET transport = ? WHERE id = ?",
			tID, a.id)
		if err != nil {
			return fmt.Errorf("AttachTransport: update address, %s", err)
		}
	}
	return nil
}

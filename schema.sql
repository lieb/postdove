-- Mentioned in the old TODO file was that I wanted to do away with the
-- ugly unique column names and use plain words instead. Also mentioned
-- was the desire to improve the Transport and Alias tables. This is the
-- result: what the new schema should be. It surprised me!
--
-- New revision 2012-02-16 for those who are following along at home:
-- This removes the RClass table and foreign key constraints thereto,
-- because it is useless unless main.cf reads restriction classes from
-- the database.
--

-- remove active in most places. an inactive is simply removed...
PRAGMA foreign_keys=ON;
BEGIN TRANSACTION;
--
-- transport table
CREATE TABLE "Transport" (
       id INTEGER PRIMARY KEY,
       transport TEXT,  -- lmtp|smtp|relay|local|throttled|custom|...
       nexthop TEXT,	-- [domain]:port or domain:port
       UNIQUE (transport,nexthop)
       );
       -- managing a unique on all those columns and then doing runtime
       -- FU to coalesce stuff is a waste of effort. They don't move so
       -- don't move them...

-- domain table
CREATE TABLE "Domain" (
       id INTEGER PRIMARY KEY,
       name TEXT NOT NULL,
       class INTEGER DEFAULT 0, -- 1 == local, 2 == relay, 3 == valias,
       	     	     	     	-- >800 == vmbox, 0 == default, none
				-- this is bogus overloading. We don't need
				-- it. Local/relay can be sorted by transport
				-- not null. The vmbox is bogus. I don't know
				-- what '0' is for. There are 2 recs, 'alien'
				-- and 'martian' which will become external
				-- addresses that "will become aliases". Huh?
				-- is this a WIP step in some external process?
       transport INTEGER,
       rclass INTEGER DEFAULT 30, -- restriction class, RCxx, default RC30
       	      	      	      	  -- breaks w/ NULL. make NOT NULL and make it TEXT
       UNIQUE (name),
       FOREIGN KEY(transport) REFERENCES Transport(id)
       );
-- This revision removes the NOT NULL constraint from Domain.name; we
-- will insert a special record Domain.id=0 with Domain.name=NULL,
-- which thus maintains a defacto NOT NULL constraint for other rows.

-- Address table
CREATE TABLE "Address" (
       id INTEGER PRIMARY KEY,
       localpart TEXT NOT NULL,
       domain INTEGER,
       transport INTEGER,
       rclass INTEGER, -- restriction class, RCxx, default RC30
       	      	       -- if this is null, use domain rclass make TEXT
       FOREIGN KEY(domain) REFERENCES Domain(id),
       FOREIGN KEY(transport) REFERENCES Transport(id),
       UNIQUE (localpart, domain)
       );

-- address_transport
-- return transport for address/domain.
-- if address doesn't have one, use its domain's transport
DROP VIEW IF EXISTS "address_transport";
CREATE VIEW "address_transport" AS
   SELECT DISTINCT la.localpart as local, ld.name as dname,
       CASE WHEN at.id IS NOT NULL
          THEN coalesce (at.transport, '') || ':' ||
	       coalesce (at.nexthop, '')
	  ELSE coalesce (dt.transport, '') || ':' ||
	       coalesce (dt.nexthop, '')
	  END
       END AS trans
FROM domain as	ld
    LEFT JOIN address AS la ON	ld.id = la.domain
    LEFT JOIN transport AS dt ON ld.transport = dt.id
    LEFT JOIN transport AS at ON la.transport = at.id

-- Alias table
CREATE TABLE "Alias" (
       id INTEGER PRIMARY KEY,
       address INTEGER NOT NULL,
       target INTEGER,
       extension TEXT, 
       FOREIGN KEY(address) REFERENCES Address(id),
       FOREIGN KEY(target) REFERENCES Address(id),
       UNIQUE(address, target, extension));
-- This revision changes Alias.name to Alias.address, because INTEGER
-- fields should not be called "name" (to me that implies TEXT.) If
-- Alias.target=0, that tells us that the Alias.extension contains a
-- /file/name, or a |command, or an :include:/file/name. Otherwise,
-- Alias.extension contains an address extension.  Differentiation of
-- local(8) aliases(5) from virtual(5) is in the Address table, where
-- Address.domain=0, the special null record in the Domain table.

-- etc_aliases
DROP VIEW IF EXISTS "etc_aliases";
CREATE VIEW "etc_aliases" AS SELECT aa.id as id, aa.localpart as local,
				 CASE WHEN Alias.target = 0
                 THEN Alias.extension
				ELSE ta.localpart ||
				   (CASE WHEN Alias.extension is NOT NULL
				         THEN '-' || alias.extension
						 ELSE '' END) ||
				    (CASE WHEN td.id = 0
			              THEN '' ELSE '@' || td.name END)
				END as VALUE
		FROM Alias
		JOIN address as ta on alias.active != 0 and alias.target = ta.id
		JOIN domain as td on ta.domain = td.id
		JOIN address as aa on Alias.address = aa.id and aa.domain = 0

-- alias_recipient models the alias/valias file where a line is:
--   alias	   recipient, recipient
--
-- return one or more rows, one for each recipient
CREATE VIEW alias_recipient as
       select a.localpart as tlocal, d.name as tdom,
       	      aa.localpart as alocal, dd.name as adom, al.extension as ext
       from address as a
       	    join domain as d on (a.domain = d.id)
       	    join alias as al on (al.target = a.id)
       	    join address as aa on (aa.id = al.address)
       	    join domain as dd on (aa.domain = d.id);

-- virt_alias models the virtuals file where a line is
--   alias    recipient
--
DROP VIEW IF EXISTS "virt_alias";
CREATE VIEW "virt_alias" AS SELECT aa.localpart as lcl, ad.name as name, ta.localpart ||
		(CASE WHEN va.extension is not NULL
		      THEN '+' || va.extension
			  ELSE '' END) ||
			     (CASE WHEN td.id = 0
				       THEN '' ELSE '@' || td.name end) as valias
	FROM Alias as va
	JOIN address as ta on (va.target = ta.id)
	join domain as td on (ta.domain = td.id)
	JOIN address as aa on (va.address = aa.id)
	join domain as ad on (aa.domain != 0 and aa.domain = ad.id)

-- look at last_insert_rowid() function for insert/update trigger.
--
CREATE TABLE "VMailbox" (
       id INTEGER PRIMARY KEY,
       enable INTEGER DEFAULT 1, -- to disable imap+lmtp
       uid INTEGER,
       gid INTEGER,
       home TEXT,  -- just home part for dovecot config of mail_home
       password TEXT,
       FOREIGN KEY(id) REFERENCES Address(id));

-- user_mailbox is a combination of an address row and a vmailbox row
-- There are bits of this I do not like, namely the coalesce functions
-- with baked in constants. Uid and gid should be NOT NULL and map to
-- the common passwd entry from sssd. home should be the dir under
-- home_dir in dovecot's config.
CREATE VIEW "user_mailbox" AS
       select mb.id as id, a.localpart as user, d.name as dom,
       	      mb.password as pw, coalesce(mb.home, 'vmail') as home,
       	      coalesce(mb.uid, d.class) as uid, coalesce(mb.gid, 800) as gid,
       	      mb.active as inuse, a.active as active
       from VMailbox as mb
       	      join address as a on (a.id = mb.id)
	      join domain as d on (a.domain = d.id)

-- backscatter and catchall here are for example. I don't do it so
-- scratch this bit.
--
-- CREATE TABLE "BScat" (
--       id INTEGER PRIMARY KEY,
--       sender TEXT NOT NULL,
--       priority INTEGER,
--       target TEXT NOT NULL,
--       UNIQUE (sender, priority));
--
COMMIT;

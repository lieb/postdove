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
CREATE TABLE "Transport" (
       id INTEGER PRIMARY KEY,
       -- active INTEGER DEFAULT 1,
       transport TEXT,  -- lmtp|smtp|relay|local|throttled|custom|...
       -- nexthop INTEGER,	-- foreign key (nexthop) references Domain(id)
       -- mx INTEGER DEFAULT 1, -- 0 :[domain]:port, 1 :domain:port
       -- port INTEGER,
       -- UNIQUE (transport,nexthop,mx,port));
       ;
       -- rubbish removed. transport does nexthop etc. as simple text...
       -- scrap the unique as well. transport becomes an edit field and
       -- "select box" option for domain/address editing.
       -- managing a unique on all those columns and then doing runtime
       -- FU to coalesce stuff is a waste of effort. They don't move so
       -- don't move them...
-- Transport.nexthop would be a pointer to Domain.id, but without a
-- foreign key constraint because the Domain table has yet to be
-- created at this point.
--
CREATE TABLE "Domain" (
       id INTEGER PRIMARY KEY,
       name TEXT,
       -- active INTEGER DEFAULT 1,
       class INTEGER DEFAULT 0, -- 1 == local, 2 == relay, 3 == valias,
       	     	     	     	-- >800 == vmbox, 0 == default, none
				-- this is bogus overloading. We don't need
				-- it. Local/relay can be sorted by transport
				-- not null. The vmbox is bogus. I don't know
				-- what '0' is for. There are 2 recs, 'alien'
				-- and 'martian' which will become external
				-- addresses that "will become aliases". Huh?
				-- is this a WIP step in some external process?
       -- owner INTEGER DEFAULT 0, -- system UID of domain owner. not used
       transport INTEGER,
       rclass INTEGER DEFAULT 30, -- restriction class, RCxx, default RC30
       	      	      	      	  -- breaks w/ NULL. make NOT NULL
       UNIQUE (name),
       FOREIGN KEY(transport) REFERENCES Transport(id);
-- This revision removes the NOT NULL constraint from Domain.name; we
-- will insert a special record Domain.id=0 with Domain.name=NULL,
-- which thus maintains a defacto NOT NULL constraint for other rows.
INSERT INTO "Domain" VALUES(0,NULL,NULL,NULL,NULL,NULL,NULL);
--
CREATE TABLE "Address" (
       id INTEGER PRIMARY KEY,
       localpart TEXT NOT NULL,
       domain INTEGER NOT NULL,
       active INTEGER DEFAULT 1, -- boolean active/inactive, -1 == no imap login
       transport INTEGER,
       rclass INTEGER, -- restriction class, RCxx, default RC30
       	      	       -- if this is null, use domain rclass
       FOREIGN KEY(domain) REFERENCES Domain(id),
       FOREIGN KEY(transport) REFERENCES Transport(id),
       UNIQUE (localpart, domain));
-- We will insert a special record Address.id=0 with Address.domain=0
-- to differentiate aliases(5) commands, paths or includes from
-- address targets.
INSERT INTO "Address" VALUES(0,'root',0,NULL,NULL,NULL);
--
CREATE TABLE "Alias" (
       id INTEGER PRIMARY KEY,
       address INTEGER NOT NULL,
       active INTEGER DEFAULT 1,
       target INTEGER NOT NULL,
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

-- look at last_insert_rowid() function for insert/update trigger.
--
CREATE TABLE "VMailbox" (
       id INTEGER PRIMARY KEY,
       active INTEGER DEFAULT 1, -- keep to disable imap+lmtp
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

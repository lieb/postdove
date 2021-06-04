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
-- Access table
DROP TABLE IF EXISTS "Access";
CREATE TABLE "Access" (
       id INTEGER PRIMARY KEY,
       action TEXT NOT NULL
       );

-- transport table
DROP TABLE IF EXISTS "Transport";
CREATE TABLE "Transport" (
       id INTEGER PRIMARY KEY,
       transport TEXT,  -- lmtp|smtp|relay|local|throttled|custom|...
       nexthop TEXT,	-- [domain]:port or domain:port
       UNIQUE (transport,nexthop)
       );

-- domain table
DROP TABLE IF EXISTS "Domain";
CREATE TABLE "Domain" (
       id INTEGER PRIMARY KEY,
       name TEXT NOT NULL,
       class INTEGER DEFAULT 0, -- 1 == local, 2 == relay, 3 == valias,
       	     	     	     	-- 0 == default (internet), none
       transport INTEGER,
       access INTEGER,
       vuid INTEGER,		-- virtual UID for dovecot general mboxes
       vgid INTEGER,		-- virtual GID
       rclass TEXT DEFAULT "DEFAULT", -- recipient restriction class
       	      	      	      	  -- breaks w/ NULL. make NOT NULL and make it TEXT
       UNIQUE (name),
       CONSTRAINT dom_trans FOREIGN KEY(transport) REFERENCES Transport(id),
       CONSTRAINT dom_access FOREIGN KEY(access) REFERENCES Access(id)
       );

-- Address table
DROP TABLE IF EXISTS "Address";
CREATE TABLE "Address" (
       id INTEGER PRIMARY KEY,
       localpart TEXT NOT NULL,
       domain INTEGER,
       transport INTEGER,
       rclass TEXT,    -- recipient restriction class
       	      	       -- if this is null, use domain rclass
       access INTEGER,
       CONSTRAINT addr_domain FOREIGN KEY(domain) REFERENCES Domain(id),
       CONSTRAINT addr_trans FOREIGN KEY(transport) REFERENCES Transport(id),
       CONSTRAINT addr_access FOREIGN KEY(access) REFERENCES Access(id)
       UNIQUE (localpart, domain)
       );

-- Create triggers to enforce unique on domain column nulls
-- This is a ambiguity in the SQL92 spec that (most) everyone handles by making unique
-- with null columns break. So we explicitly check. For safety we trigger both INSERT
-- and UPDATE although there is no application of updating a domain.
DROP TRIGGER IF EXISTS addr_insert_null_check;
CREATE TRIGGER addr_insert_null_check BEFORE INSERT ON Address
 WHEN NEW.domain IS NULL
   BEGIN
     SELECT CASE WHEN (
       (SELECT 1 FROM Address WHERE localpart IS NEW.localpart AND domain IS NULL)
       NOTNULL) THEN RAISE(FAIL, "Duplicate insert with NULL") END; END;

DROP TRIGGER IF EXISTS addr_update_null_check;
CREATE TRIGGER addr_update_null_check BEFORE UPDATE ON Address
 WHEN NEW.domain IS NULL
   BEGIN
     SELECT CASE WHEN (
       (SELECT 1 FROM Address WHERE localpart is NEW.localpart AND domain IS NULL)
       NOTNULL) THEN RAISE(FAIL, "Duplicate update with NULL")
     END;  END;

-- create a trigger to delete the domain when addr refs are 0 meaning this is the only
-- one pointing to it and domain.class != vmailbox
DROP TRIGGER  IF EXISTS after_addr_del;
CREATE TRIGGER after_addr_del AFTER DELETE ON address
 WHEN OLD.domain IS NOT NULL
    AND (SELECT class FROM domain WHERE id = OLD.domain) != 4
    AND (SELECT count(*) FROM address WHERE domain = OLD.domain) < 1
 BEGIN
  DELETE FROM domain WHERE id = OLD.domain; END;

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
       AS trans
FROM domain as	ld
    LEFT JOIN address AS la ON	ld.id = la.domain
    LEFT JOIN transport AS dt ON ld.transport = dt.id
    LEFT JOIN transport AS at ON la.transport = at.id;

-- Alias table
DROP TABLE IF EXISTS "Alias";
CREATE TABLE "Alias" (
       id INTEGER PRIMARY KEY,
       address INTEGER NOT NULL,
       target INTEGER,
       extension TEXT, 
       CONSTRAINT alias_addr FOREIGN KEY(address) REFERENCES Address(id),
       CONSTRAINT alias_target FOREIGN KEY(target) REFERENCES Address(id),
       UNIQUE(address, target, extension)
       CHECK (target IS NOT NULL OR extension IS NOT NULL));

-- Create triggers to clean up the mess left behind when an alias is deleted
-- We need two of them. One for the recipient (target) and one for the alias
-- key (address) itself. Protect over-eager deletes by checking the reference
-- linkage. This can cascade via the after_addr_del trigger to a domain.

-- Delete addresses so long as no other alias target or a vmailbox references it
DROP TRIGGER IF EXISTS after_alias_del_recip;
CREATE TRIGGER after_alias_del_recip AFTER DELETE ON alias
 WHEN (SELECT count(*) FROM alias WHERE target = OLD.target) < 1
    AND (SELECT count(*) FROM vmailbox WHERE id = OLD.target) < 1
  BEGIN
    DELETE FROM address WHERE id = OLD.target; END;

-- Delete addresses so long as no other alias references it as a target
DROP TRIGGER IF EXISTS after_alias_del_addr;
CREATE TRIGGER after_alias_del_addr AFTER DELETE ON alias
 WHEN (SELECT count(*) FROM alias WHERE address = OLD.address) < 1
  BEGIN
    DELETE FROM address WHERE id = OLD.address; END;

-- etc_aliases (local aliases)
-- alias	recipient, recipient ...
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
		JOIN address as aa on Alias.address = aa.id and aa.domain = 0;

-- alias_recipient models the alias/valias file where a line is:
--   alias@dom	   recipient, recipient
--
-- return one or more rows, one for each recipient ...
DROP VIEW IF EXISTS alias_recipient;
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
	join domain as ad on (aa.domain != 0 and aa.domain = ad.id);

-- vmailbox, dovecot user database
DROP TABLE IF EXISTS "VMailbox";
CREATE TABLE "VMailbox" (
       id INTEGER PRIMARY KEY,
       pw_type TEXT NOT NULL DEFAULT 'PLAIN',
       password TEXT,
       uid INTEGER, -- if these are NULL, use domain values
       gid INTEGER,
       home TEXT,  -- just home part for dovecot config of mail_home
       quota TEXT DEFAULT '*:bytes=300M', -- in Dovecot form. NULL is no quota
       enable INTEGER NOT NULL DEFAULT 1, -- bool to disable imap+lmtp
       CONSTRAINT vmbox_addr FOREIGN KEY(id) REFERENCES Address(id));

-- An address can either be an alias or a mailbox but not both. Just imagine
-- the "both" case. The alias half would re-direct postfix off somewhere else
-- and orphan the mailbox. If someone wants to do such a a re-direct, there are other
-- ways to do that. This trigger and its matching one for vmailbox checks first
-- to see of the other is already in existence...
DROP TRIGGER IF EXISTS alias_insert_mailbox_check;
CREATE TRIGGER alias_insert_mailbox_check BEFORE INSERT ON alias
 WHEN (SELECT count(*) FROM vmailbox WHERE id = NEW.address) > 0
  BEGIN SELECT RAISE(FAIL, 'New alias already a mailbox'); END;

DROP TRIGGER IF EXISTS mailbox_insert_alias_check;
CREATE TRIGGER mailbox_insert_alias_check BEFORE INSERT ON vmailbox
 WHEN (SELECT count(*) FROM alias WHERE address = NEW.id) > 0
  BEGIN SELECT RAISE(FAIL, 'New mailbox already an alias'); END;

-- Create trigger to extend address constraint to vmailbox which shares its id
-- We return an error string naming the app err
DROP TRIGGER IF EXISTS before_del_mbox;
CREATE TRIGGER before_del_mbox BEFORE DELETE ON vmailbox
 WHEN (SELECT count(*) FROM alias WHERE target = OLD.id) > 0
  BEGIN
     SELECT RAISE(ABORT, 'ErrMdbMboxIsRecip'); END;

-- Create a trigger to clean up the address on delete
DROP TRIGGER IF EXISTS after_del_mbox;
CREATE TRIGGER after_del_mbox AFTER DELETE ON vmailbox
 WHEN (SELECT count(*) FROM alias WHERE target = OLD.id) < 1
  BEGIN
    DELETE FROM address WHERE id = OLD.id; END;

-- user_mailbox is a combination of an address row and a vmailbox row
-- There are bits of this I do not like, namely the coalesce functions
-- with baked in constants. Uid and gid should be NOT NULL and map to
-- the common passwd entry from sssd. home should be the dir under
-- home_dir in dovecot's config.
DROP VIEW IF EXISTS "user_mailbox";
CREATE VIEW "user_mailbox" AS
       select mb.id as id, a.localpart as user, d.name as dom,
       	      mb.password as pw, coalesce(mb.home, 'vmail') as home,
       	      coalesce(mb.uid, d.vuid) as uid, coalesce(mb.gid, d.vgid) as gid,
       	      mb.active as inuse, a.active as active
       from VMailbox as mb
       	      join address as a on (a.id = mb.id)
	      join domain as d on (a.domain = d.id);

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
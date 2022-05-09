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
-- Table rows match smtpd_restriction_classes list actually implemented
-- in postfix. The names are the set of acceptable choices in the UI and
-- we catch editing errors here rather than in the postfix runtime
DROP TABLE IF EXISTS "Access";
CREATE TABLE "Access" (
       id INTEGER PRIMARY KEY,
       name TEXT UNIQUE NOT NULL,
       action TEXT NOT NULL
       );

-- transport table
DROP TABLE IF EXISTS "Transport";
CREATE TABLE "Transport" (
       id INTEGER PRIMARY KEY,
       name TEXT UNIQUE NOT NULL,
       transport TEXT,  -- lmtp|smtp|relay|local|throttled|custom|...
       nexthop TEXT,	-- [domain]:port or domain:port
       UNIQUE (transport,nexthop)
       );

-- domain table
DROP INDEX IF EXISTS domain_name;
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
       CONSTRAINT dom_trans FOREIGN KEY(transport) REFERENCES Transport(id),
       CONSTRAINT dom_access FOREIGN KEY(access) REFERENCES Access(id)
       );

CREATE UNIQUE INDEX domain_name ON domain(name);

-- domain_access
DROP VIEW IF EXISTS domain_access;
CREATE VIEW domain_access AS
       SELECT d.name AS domain_name, ac.action AS access_key
       FROM domain AS d, access AS ac WHERE d.access IS ac.id;
       
-- domain_transport
DROP VIEW IF EXISTS domain_transport;
CREATE VIEW domain_transport AS
       SELECT d.name AS domain_name,
       	      COALESCE(tr.transport, '') || ':' || COALESCE(tr.nexthop, '') AS transport
       FROM domain AS d, transport AS tr WHERE d.transport IS tr.id;
       
-- internet_domain
DROP VIEW IF EXISTS internet_domain;
CREATE VIEW internet_domain AS
  SELECT name FROM domain WHERE class = 0;


-- local_domain
DROP VIEW IF EXISTS local_domain;
CREATE VIEW local_domain AS
  SELECT name FROM domain WHERE class = 1;


-- relay_domain
DROP VIEW IF EXISTS relay_domain;
CREATE VIEW relay_domain AS
  SELECT name FROM domain WHERE class = 2;


-- virtual_domain
DROP VIEW IF EXISTS virtual_domain;
CREATE VIEW virtual_domain AS
  SELECT name FROM domain WHERE class = 3;

-- vmailbox_domain
DROP VIEW IF EXISTS vmailbox_domain;
CREATE VIEW vmailbox_domain AS
  SELECT name FROM domain WHERE class = 4;



-- Address table
DROP INDEX IF EXISTS address_localpart;
DROP TABLE IF EXISTS "Address";
CREATE TABLE "Address" (
       id INTEGER PRIMARY KEY,
       localpart TEXT NOT NULL,
       domain INTEGER,
       transport INTEGER,
       access INTEGER,
       CONSTRAINT addr_domain FOREIGN KEY(domain) REFERENCES Domain(id),
       CONSTRAINT addr_trans FOREIGN KEY(transport) REFERENCES Transport(id),
       CONSTRAINT addr_access FOREIGN KEY(access) REFERENCES Access(id)
       UNIQUE (localpart, domain)
       );

CREATE UNIQUE INDEX address_localpart ON address(localpart, domain);

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

-- address_access
-- view to handle access(5) processing
DROP VIEW IF EXISTS address_access;
CREATE VIEW "address_access" AS
       SELECT a.localpart AS username, d.name AS domain_name,
          CASE WHEN a.access IS NOT NULL
	  THEN (SELECT action FROM access WHERE id=a.access)
	  ELSE (SELECT action FROM access WHERE id=d.access)
	  END AS access_key
       FROM "Address" AS a
          JOIN "Domain" AS d ON a.domain=d.id
       WHERE a.access IS NOT NULL OR d.access IS NOT NULL;

-- address_transport
-- return transport for address/domain.
-- if address doesn't have one, use its domain's transport
DROP VIEW IF EXISTS "address_transport";
CREATE VIEW "address_transport" AS
   SELECT a.localpart as username, d.name as domain_name,
       CASE WHEN a.transport IS NOT NULL
          THEN (SELECT coalesce (tr.transport, '') || ':' ||
	               coalesce (tr.nexthop, '') FROM transport AS tr
		WHERE a.transport IS tr.id)
	  ELSE (SELECT coalesce (tr.transport, '') || ':' ||
	               coalesce (tr.nexthop, '') FROM transport AS tr
		WHERE d.transport IS tr.id)
	  END AS transport
  FROM address AS a
     JOIN domain AS d ON a.domain IS d.id
  WHERE a.transport IS NOT NULL OR d.transport IS NOT NULL;

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
CREATE VIEW "etc_aliases" AS
  SELECT DISTINCT aa.localpart AS local_user,
        (CASE WHEN al.target IS NULL
              THEN al.extension
              ELSE
               (SELECT (CASE WHEN ta.domain IS NULL
	                     THEN ta.localpart
	                     ELSE ta.localpart || '@' ||
			          (SELECT name FROM domain WHERE id = ta.domain)
	                END)
	       FROM address ta WHERE ta.id = al.target)
         END) AS recipient
  FROM alias AS al, address AS aa
  WHERE al.address IS aa.id AND aa.domain IS NULL;

-- virt_alias models the virtuals file where a line is
--   alias    recipient
--
DROP VIEW IF EXISTS "virt_alias";
CREATE VIEW "virt_alias" AS
       SELECT aa.localpart AS mailbox, ad.name AS domain_name,
	      ta.localpart ||
	      (CASE WHEN va.extension IS NOT NULL
	      	    THEN '+' || va.extension
		    ELSE ''
	      END) ||
	      (SELECT CASE WHEN ta.domain IS NULL
		    THEN ''
		    ELSE '@' || (SELECT name FROM domain WHERE id IS ta.domain)
	      END) AS recipient
	FROM Alias AS va
	JOIN address AS aa ON (va.address = aa.id)
	JOIN domain AS ad ON (aa.domain = ad.id)
	JOIN address AS ta ON (va.target = ta.id);

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

-- user_mailbox is a combination of an address row and a vmailbox row.
-- Field names are chosen to match dovecot variables.
-- There are bits of this I do not like, namely the coalesce functions
-- with baked in constants. Uid and gid should be NOT NULL and map to
-- the common passwd entry from sssd. home should be the dir under
-- home_dir in dovecot's config.
-- This view brings together all this together. It is used as the base for
-- the specializes views and dovecot queries
DROP VIEW IF EXISTS "user_mailbox";
CREATE VIEW "user_mailbox" AS
       SELECT mb.id AS id, a.localpart AS username, d.name AS domain,
       	      '{' || mb.pw_type || '}' || COALESCE(mb.password, '*') AS password,
	      COALESCE(mb.uid,
	              COALESCE(d.vuid,
		              (SELECT vuid FROM domain WHERE name = 'localhost'))) AS uid,
	      COALESCE(mb.gid,
	             COALESCE(d.vgid,
		              (SELECT vgid FROM domain WHERE name = 'localhost'))) AS gid,
	      COALESCE(mb.home, '') AS home,
	      COALESCE(mb.quota, '*:bytes=0') AS quota_rule,
       	      mb.enable AS enable
       FROM VMailbox AS mb
       	      JOIN address AS a ON (a.id = mb.id)
	      JOIN domain AS d ON (a.domain = d.id);

-- user_deny
-- This query looks for denied (locked out) users
DROP VIEW IF EXISTS "user_deny";
CREATE VIEW "user_deny" AS
     SELECT username, domain, 'true' AS deny
     FROM user_mailbox WHERE enable = 0;
     
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

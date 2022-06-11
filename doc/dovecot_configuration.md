# Dovecot Configuration
All `dovecot` configuration is done on `pobox`.

All of the configuration for `dovecot` is done in the `/etc/dovecot` directory.
Its configuration follows a familiar style where global parameters are in a single
file that at its end includes specific module parameters from the set of files
located in the `/etc/dovecot/conf.d` directory.
Those files are named so that the order in which they are included is determined
by the lexical sort order of their names.
Note the names in `conf.d` below.
The files `*.conf` are read in the lexical order shown here.
The `*.conf.ext` files are included by other individual configuration files.

There are a lot of parameters and the files are well commented which means they
are big. They also can change from release to release. Therefore it does not make
sense to save and update these files in our own repository. Rather, we simply
document via `diff` hunks what we change and leave the rest alone. There are also
some files, namely query parameters that are not part of the distribution package.
Those files are located and maintained in our repository.

## dovecot.conf
As mentioned above, there is little to change in this file.

```bash
[root@pobox ~]# cd /etc/dovecot
[root@pobox dovecot]# diff -u ./dovecot.conf.orig ./dovecot.conf
--- ./dovecot.conf.orig 2021-03-04 00:38:06.000000000 -0800
+++ ./dovecot.conf      2021-06-11 11:13:13.528375974 -0700
@@ -22,6 +22,7 @@
 
 # Protocols we want to be serving.
 #protocols = imap pop3 lmtp submission
+protocols = imap lmtp
 
 # A comma separated list of IPs or hosts where to listen in for connections. 
 # "*" listens in all IPv4 interfaces, "::" listens in all IPv6 interfaces.
```

I only enable *imap* and *lmtp*. I'm not interested in *pop3* for a local server and
*submission* is not needed because my clients submit their outbound mail directly to 
the mail service at my ISP.

## conf.d/10-master.conf
We accomplish two things in the master file.
I break up the *diff* output for the discussion.

```bash
[root@pobox dovecot]# diff -u ./conf.d/10-master.conf.orig ./conf.d/10-master.conf
--- ./conf.d/10-master.conf.orig        2021-03-04 00:38:06.000000000 -0800
+++ ./conf.d/10-master.conf     2021-06-11 13:43:59.100174916 -0700
@@ -19,8 +19,8 @@
     #port = 143
   }
   inet_listener imaps {
-    #port = 993
-    #ssl = yes
+    port = 993
+    ssl = yes
   }
 
   # Number of connections to handle before starting a new process. Typically
@@ -57,11 +57,11 @@
   }
 ```

The first change is not really necessary because the commented lines with values
show the defaults.
This change enables *imaps* via its port and enables SSL for it.

```bash
   # Create inet listener only if you can't use the above UNIX socket
-  #inet_listener lmtp {
+  inet_listener lmtp {
     # Avoid making LMTP visible for the entire internet
-    #address =
-    #port = 
-  #}
+    address = *
+    port = 24
+  }
 }
 
 service imap {
```

This enables *lmtp* for the whole net.
However, my whole net is just my home network.
One should not do this on a larger net and especially not on a host that is directly
addressable on the public network unless they really want the *virtual vandals* pillaging their mailstore.
I may change this later back to the defaults.

## conf.d/10-mail.conf
The core of the running server is configured here.
As above, I have broken the changes into individual hunks in order to explain
each one.

```bash
root@pobox dovecot]# diff -u ./conf.d/10-mail.conf.orig ./conf.d/10-mail.conf
--- ./conf.d/10-mail.conf.orig  2021-01-07 09:38:00.000000000 -0800
+++ ./conf.d/10-mail.conf       2021-06-12 16:41:02.231454876 -0700
@@ -28,6 +28,9 @@
 # <doc/wiki/MailLocation.txt>
 #
 #mail_location = 
+mail_home = /srv/dovecot/%d/%n
+mail_location = maildir:~/Maildir:LAYOUT=fs
+
 # If you need to set multiple mailbox locations or want to change default
 # namespace settings, you can do it by defining namespace sections.
```

All the work setting up the filesystems is so this bit of configuration works.
The `mail_home` directive sets the *root* of the mailstore tree.
The `%d` is the domain name for the virtual domain, in this case, `example.com`.
The `%n` is the mailbox name, in this case, `bill`.
This is set up so that `bill@example.com` will have a home for his mail at
`/srv/dovecot/example.com/bill`. The `mail_location` directive places email in
`srv/dovecot/example.com/bill/Maildir` and its format will be *maildir* which
will place each email in a separate file. This works well in BTRFS because it
will store small files more efficiently than other, more traditional filesystems.

This does not prevent `dovecot` from using a different location for mail.
The user record in the database can override this if desired.

Next set up the default user and group IDs for when all else fails to identify one.

```bash 
@@ -105,8 +108,8 @@
 # System user and group used to access mails. If you use multiple, userdb
 # can override these by returning uid or gid fields. You can use either numbers
 # or names. <doc/wiki/UserIds.txt>
-#mail_uid =
-#mail_gid =
+mail_uid = 2000
+mail_gid = 2000
 
 # Group to enable temporarily for privileged operations. Currently this is
 # used only with INBOX when either its initial creation or dotlocking fails.
```

The `mail_uid` and `mail_gid` are the default values to be used for some
configurations of `dovecot` that do not want to use user unique uid/gid values.
We have uid/gid values specified in the database and have our own defaults
should the record not specify a number.
However, having our own in the database does not help because `dovecot` insists
that these values be set.
I have set them to `2000` arbitrarily to get them out of the way of any active numbers.
In this configuration, they should never be used by `dovecot`
other than to annoy the administrator.

There are a number of different methods to do file locking which
is necessary given that mail can come in while users are actively reading
already arrived mail. This chunk selects the method to use.
```
@@ -165,6 +168,12 @@
 # methods. NFS users: flock doesn't work, remember to change mmap_disable.
 #lock_method = fcntl
 
+# F_SETLKW does not work on virtiofs filesystem. Given the extremely
+# broken nature of POSIX locks, this is hard to make work on anything
+# other than local filesystems. BTW, flock works just fine in NFSv4.
+# OOPS. blocking flock doesn't work either. fall back to dotlock...
+lock_method = dotlock
+
 # Directory where mails can be temporarily stored. Usually it's used only for
 # mails larger than >= 128 kB. It's used by various parts of Dovecot, for
 # example LDA/LMTP while delivering large mails or zlib plugin for keeping
```
The `fcntl`, `flock` methods use system calls to gain exclusive access to files.
The `dotlock` method uses the *exclusive open* option for opening files to do the
same thing.
The comment above is from my experimentation.
I found that `flock`, which works just fine on *NFS* also worked with *VirtioFS*
until the developers discovered that blocking locks where the process waits for the
lock can *deadlock*, a very bad thing.
The `virtiofsd` process now returns a *not supported* error, eliminating this
efficient method.
Hence, the fallback to `dotlock`. Should the developers come up with a solution,
we can change it back.

The server also wants to know the legal ranges for user and group IDs.

```bash 
@@ -175,15 +178,15 @@
 # to make sure that users can't log in as daemons or other system users.
 # Note that denying root logins is hardcoded to dovecot binary and can't
 # be done even if first_valid_uid is set to 0.
-#first_valid_uid = 500
-#last_valid_uid = 0
+first_valid_uid = 1000
+#last_valid_uid = 65535
 
 # Valid GID range for users, defaults to non-root/wheel. Users having
 # non-valid GID as primary group ID aren't allowed to log in. If user
 # belongs to supplementary groups with non-valid GIDs, those groups are
 # not set.
-#first_valid_gid = 1
-#last_valid_gid = 0
+first_valid_gid = 1000
+#last_valid_gid = 65535
 
 # Maximum allowed length for mail keyword name. It's only forced when trying
 # to create new keywords.
```

These values are the authentication boundaries. I set the first valid value
to `1000` because **Fedora** sets its first non-system value to `1000`.
The maximum value is the highest 16 bit uid/gid. Linux has long since made
these numbers 32 bit but there is always some crufty bits of code that did not
get the memo.

## conf.d/10-ssl.conf

```bash
[root@pobox dovecot]# diff -u ./conf.d/10-ssl.conf.orig ./conf.d/10-ssl.conf
--- ./conf.d/10-ssl.conf.orig   2021-03-22 13:41:13.000000000 -0700
+++ ./conf.d/10-ssl.conf        2021-06-11 15:56:03.458388966 -0700
@@ -53,7 +53,7 @@
 # Generate new params with `openssl dhparam -out /etc/dovecot/dh.pem 4096`
 # Or migrate from old ssl-parameters.dat file with the command dovecot
 # gives on startup when ssl_dh is unset.
-#ssl_dh = </etc/dovecot/dh.pem
+ssl_dh = </etc/dovecot/dh.pem
 
 # Minimum SSL protocol version to use. Potentially recognized values are SSLv3,
 # TLSv1, TLSv1.1, and TLSv1.2, depending on the OpenSSL version used.
```

This is the SSL certificate setup.
Normally, one would not need to do anything here because the `dovecot`package
has a self-signed certificate.
However, the current version of *OpenSSL* has deprecated the use of keys smaller
than 4096 bits.
We could have used the packaged certificate except that it is now too weak
(only 3K bits).
Therefore, we need to install a new certificate with a 4K key.
The following command will make that happen.

```bash
[root@pobox dovecot]# openssl dhparam -out /etc/dovecot/dh.pem 4096
```

Since this is a changing topic and if your installation is more than a private mailserver, we refer you to the appropriate documentation.
One option is to install a **Let's Encrypt** certificate which I may do at some
point now that I have found out that my ISP has enrolled my public site with them.

## conf.d/10-auth.conf
Authorization is a bit complicated so we will break this up into individual hunks
and explain each one.

```bash
[root@pobox dovecot]# diff -u ./conf.d/10-auth.conf.orig ./conf.d/10-auth.conf
--- ./conf.d/10-auth.conf.orig  2020-12-22 05:26:52.000000000 -0800
+++ ./conf.d/10-auth.conf       2021-06-10 17:04:35.276673882 -0700
@@ -8,6 +8,7 @@
 # connection is considered secure and plaintext authentication is allowed.
 # See also ssl=required setting.
 #disable_plaintext_auth = yes
+disable_plaintext_auth = no
 
 # Authentication cache size (e.g. 10M). 0 means it's disabled. Note that
 # bsdauth and PAM require cache_key to be set for caching to be used.
 ```
 
 We can use *plaintext* authorization because we are a black box server (no one here but email service) and connections are SSL.
This does not prevent other forms of (encrypted) authorization which always work.
If you are ok with the passwords being recognizable in the database, this will
work just fine.

Now set up authorization methods, both those we want and those we do not want.

 ```bash
@@ -116,11 +117,11 @@
 #
 # <doc/wiki/UserDatabase.txt>
 
-#!include auth-deny.conf.ext
+!include auth-deny.conf.ext
 #!include auth-master.conf.ext
 
-!include auth-system.conf.ext
-#!include auth-sql.conf.ext
+#!include auth-system.conf.ext
+!include auth-sql.conf.ext
 #!include auth-ldap.conf.ext
 #!include auth-passwdfile.conf.ext
 #!include auth-checkpassword.conf.ext
```

The first change enables *deny* which allows us to control access via the database.
This is handy especially for dealing with compromised accounts.
It is not really necessary for my enclosed email environment but anything larger
would definitely need the ability to disable an IMAP account.
The second change switches authorization from using the system authorization via
`/etc/passwd` entries to using SQL queries.

The next two files are now included by `conf.d/10-auth.conf`.

## conf.d/auth-sql.conf.ext
This file defines how the authorization SQL queries will be processed.

```bash
[root@pobox dovecot]# diff -u ./conf.d/auth-sql.conf.ext.orig ./conf.d/auth-sql.conf.ext
--- ./conf.d/auth-sql.conf.ext.orig     2019-06-29 16:29:52.948197065 -0700
+++ ./conf.d/auth-sql.conf.ext  2021-02-13 17:41:40.006609404 -0800
@@ -12,9 +12,9 @@
 # "prefetch" user database means that the passdb already provided the
 # needed information and there's no need to do a separate userdb lookup.
 # <doc/wiki/UserDatabase.Prefetch.txt>
-#userdb {
-#  driver = prefetch
-#}
+userdb {
+  driver = prefetch
+}
 
 userdb {
   driver = sql

```

All we are doing here is enabling *prefetch* because this eliminates a second SQL
query to get the extra information needed for configuring a login after
authentication has been done. Note the comments below on how the query is
constructed.

## conf.d/auth-deny.conf.ext
We also change the method for denying access.

```bash
[root@pobox dovecot]# diff -u ./conf.d/auth-deny.conf.ext.orig ./conf.d/auth-deny.conf.ext
--- ./conf.d/auth-deny.conf.ext.orig    2020-05-12 08:44:05.000000000 -0700
+++ ./conf.d/auth-deny.conf.ext 2021-02-13 17:15:12.823314906 -0800
@@ -6,10 +6,17 @@
 # checked first.
 
 # Example deny passdb using passwd-file. You can use any passdb though.
+#passdb {
+#  driver = passwd-file
+#  deny = yes
+#
+#  # File contains a list of usernames, one per line
+#  args = /etc/dovecot/deny-users
+#}
 passdb {
-  driver = passwd-file
-  deny = yes
+   driver = sql
+   deny = yes
 
-  # File contains a list of usernames, one per line
-  args = /etc/dovecot/deny-users
+   args = /etc/dovecot/sql-deny.conf.ext
 }
+

```
These similar changes re-configure the *deny* operation to use an SQL query.

## SQL Queries
Everything in the previous sections changes the configuration to use SQL queries
instead of files to authorize and manage users and their mail stores.

We discuss the database schema in detail in the schema file itself as well as in the
build documentation.
There is one point, however, that is important to the queries in this section.
To avoid complex and/or obscure SQL queries, we use *views* in the
database to handle all the complexity in the database and just leave
a simple SQL `SELECT` statement for the `dovecot` configuration.

The following files can be found in the `config` directory of the source. They are copied
to the `/etc/dovecot` system directory.

### dovecot-sql.conf.ext

```bash
# cat /etc/dovecot/dovecot-sql.conf.ext 
driver = sqlite
connect = /etc/dovecot/private/postdove.sqlite
default_pass_scheme = PLAIN

password_query = SELECT username, domain, password, \
  uid as userdb_uid, gid as userdb_gid, home as userdb_home, \
  quota_rule AS userdb_quota_rule \
  FROM user_mailbox WHERE username = '%n' AND domain = '%d'

user_query = SELECT home, uid, gid, quota_rule \
  FROM user_mailbox WHERE username = '%n' AND domain = '%d'

# For using doveadm -A:
iterate_query = SELECT username, domain FROM user_mailbox

```

There are two queries here. One for the `password_query` and the other for the
`user_query`.
Note that from above, we have enabled *prefetch*.
Normally, a `password_query` would only have the user name and password and
leave the rest to the `user_query`.
In *prefetch* mode we can get that extra information fetched as well in the same query.
However, there is an obscure `dovecot` wrinkle ( I had a less kind word once
I figured it out...).
Arguments to the `WHERE` clause have substitution strings, `%n` for user name
and `%d` for domain etc. but the names of the returned fields to the `SELECT`
__must__ match the internal symbol names we intend to receive the values.
In the *user_query* we see the usual suspects `uid`, `gid`, and `home`.
But note carefully what is in the *password_query*.
Instead of `uid` it must be `userdb_uid`.
The same applies to `gid` as `userdb_gid` and `home` as `userdb_home`.
If you do not get this correct, the verbose debug output will show the query being
correct during authentication but the `mail_uid` and `mail_gid` values
being used instead for setting up the connection/session.
The other field of interest is the `user_db_quota_rule` in the *password_query* and
the `quota_rule` in the `user_query`.

### sql-deny.conf.ext

```bash
# cat /etc/dovecot/sql-deny.conf.ext 
driver = sqlite
connect = /etc/dovecot/private/postdove.sqlite

password_query = SELECT deny FROM user_deny \
WHERE username = '%n' AND domain = '%d'
```

This is a simple query that returns a result if the user has been disabled and nothing if it is active.
The authentication logic first checks for *deny* and then checks for an authenticated
user. This means that a user's account remains active and will receive mail but the
user cannot make a connection to the server.

With this, we are done with configuration of `dovecot`. If you do not intend to also
run a local SMTP server with it, we can move on to the
[Administrator Guide](admin.md).
Otherwise, it is time to configure `postfix` using
[Postfix Configuration](postfix_configuration.md).
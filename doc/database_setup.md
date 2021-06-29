# Database Creation
The database creation and management is all done on the mail server `pobox`.

The database file is shared by the `postfix` and `dovecot` servers.
Since neither server does any database updates, they only have read-only access.
The file is located in the `/etc/dovecot/private` directory simply for
convenience.
It is possible to only configure `dovecot` and either leave `postfix` to its own
devices or run with no MTA at all, simply an IMAP/POP3 mailstore.
The file is owned by `root` with group `mail`.
It is read/write for `root`, readable by both servers, and denied to everyone else.
No security labels other than the default ones for the `/etc` directory are applied.

**Fedora** creates *user* and *group* identities for both servers.
In addition to the default group `postfix`, another group, `mail`,
is created by the package for storing email files.
We use the `mail` group to link access for both servers.
This is typically safe because `postfix` only uses it to create mailboxes
and append mail in `/var/mail`.
If more security is required, there are no limitations built into `postdove` to
prevent changing to a new group.
All the program requires is read-write access for the administrator.

```bash
[root@pobox dovecot]# grep dovecot /etc/group
dovecot:x:97:
[root@pobox dovecot]# grep postfix /etc/group
mail:x:12:postfix
postfix:x:89:
```
The next step is to add `dovecot`.

```bash
[root@pobox dovecot]# usermod -a -G mail dovecot
[root@pobox dovecot]# grep dovecot /etc/group
mail:x:12:postfix,dovecot
dovecot:x:97:
```
The database file itself is located in the `/etc/dovecot/private` directory which
we now create. We also lock it down so ordinary users cannot see its contents.

```bash
[root@pobox dovecot]# mkdir /etc/dovecot/private
[root@pobox dovecot]# chmod 750 /etc/dovecot/private
[root@pobox dovecot]# chown root.mail /etc/dovecot/private
```
We now have place for out database so create it and see what we have so far.
```bash
[root@pobox dovecot]# postdove create
[root@pobox dovecot]# ls -l /etc/dovecot/private/
[root@pobox dovecot]# ls -la /etc/dovecot/private
total 56
drwxr-x---. 2 root mail    28 Jun 10 14:46 .
drwxr-xr-x. 4 root root   160 Mar 22 13:43 ..
-rw-r--r--. 1 root root 57344 Jun 10 14:23 dovecot.sqlite
```
The last thing we need to do is change its ownership and protection mode.
```bash
[root@pobox dovecot]# chmod 750 private/dovecot.sqlite 
[root@pobox dovecot]# chown root.mail /etc/dovecot/private/dovecot.sqlite 
[root@pobox dovecot]# ls -la /etc/dovecot/private
total 56
drwxr-x---. 2 root mail    28 Jun 10 14:46 .
drwxr-xr-x. 4 root root   160 Mar 22 13:43 ..
-rwxr-x---. 1 root mail 57344 Jun 10 14:23 dovecot.sqlite
```
The end result is that `root` is the only user that can run `postdove`
to modify the database and only `postfix` and `dovecot` can have read
access to use it in the running system.
We also have an (almost) empty database that just contains the following domains:
```bash
[root@pobox dovecot]# postdove show domain localhost
Name:           localhost
Class:          local
Transport:      --
Access:         --
UserID:         --
Group ID:       --
Restrictions:   DEFAULT
[root@pobox dovecot]# postdove show domain localhost.localdomain
Name:           localhost.localdomain
Class:          local
Transport:      --
Access:         --
UserID:         --
Group ID:       --
Restrictions:   DEFAULT
```
We will need a lot more than that to have a useful mail server.
We will cover those details in [Postdove Administration](admin.md) but first
we must configure the `dovecot` server.
The next step is [Dovecot Configuration](dovecot_configuration.md).
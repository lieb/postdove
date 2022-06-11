# Database Creation
The database creation and management is all done on the mail server `pobox`.

## Postdove Utility Installation
All of the mail system administration is done with a combination of `dovecot` administration
tools and the `postdove` utility from this package. The usual installation location
for the utility is `/root/bin`, the private commands directory of the system administrator.
Installing the utility in `/usr/local/bin` or `/usr/local/sbin` is typical but that is not recommended.
To protect it further, it should be owned by `root` and only `root` should be able to
execute it, i.e.
```
[root@pobox ~]# cd bin
[root@pobox bin]# chmod 700 postdove 
[root@pobox bin]# ls -l
total 8144
-rwx------. 1 root root 8338640 May  9 12:59 postdove
```

The database file is shared by the `postfix` and `dovecot` servers.
Since neither server does any database updates, they only have read-only access.
The file is located in the `/etc/postfix/private` directory simply for
convenience.
It is possible to only configure `dovecot` and either leave `postfix` to its own
devices or run with no MTA at all, simply an IMAP/POP3 mailstore.
The file is owned by `root` with group `mail`.
It is read/write for `root`, readable by both servers, and denied to everyone else.
No security labels other than the default ones for the `/etc/postfix` directory are applied.

**Fedora** creates *user* and *group* identities for both services.
In addition to the default group `postfix`, another group, `mail`,
is created by the package for storing email files.
We use the `mail` group to link access for both servers.
This is typically safe because `postfix` only uses it to create mailboxes
and append mail in `/var/mail`.
If more security is required, there are no limitations built into `postdove` to
prevent changing to a new group.
All the program requires is read-write access for the administrator.

This is how the groups are set up for both `postfix` and `dovecot`.
```bash
[root@pobox ~]# cd /etc/postfix
[root@pobox postfix]# grep dovecot /etc/group
dovecot:x:97:
[root@pobox postfix]# grep postfix /etc/group
mail:x:12:postfix
postfix:x:89:
```
The next step is to add `dovecot` to the `mail` group that `postfix` owns.

```bash
[root@pobox postfix]# usermod -a -G mail dovecot
[root@pobox postfix]# grep dovecot /etc/group
mail:x:12:postfix,dovecot
dovecot:x:97:
```
The database file itself is located in the `/etc/postfix/private` directory which
we now create. We also lock it down so ordinary users cannot see its contents.

```bash
[root@pobox postfix]# mkdir /etc/postfix/private
[root@pobox postfix]# chmod 750 /etc/postfix/private
[root@pobox postfix]# chown root.mail /etc/postfix/private
```
We now have place for our database so create it and see what we have so far.
```bash
[root@pobox postfix]# postdove create
[root@pobox postfix]# ls -la /etc/postfix/private
total 56
drwxr-x---. 2 root mail    28 Jun 10 14:46 .
drwxr-xr-x. 4 root root   160 Mar 22 13:43 ..
-rw-r--r--. 1 root root 57344 Jun 10 14:23 postdove.sqlite
```
The next thing we need to do is change its ownership and protection mode.
```bash
[root@pobox postfix]# chmod 750 private/postdove.sqlite 
[root@pobox postfix]# chown root.mail /etc/postfix/private/postdove.sqlite 
[root@pobox postfix]# ls -la /etc/postfix/private
total 56
drwxr-x---. 2 root mail    28 Jun 10 14:46 .
drwxr-xr-x. 4 root root   160 Mar 22 13:43 ..
-rwxr-x---. 1 root mail 57344 Jun 10 14:23 postdove.sqlite
```
The end result is that `root` is the only user that can run `postdove`
to modify the database and only `postfix` and `dovecot` can have read
access to use it in the running system.
We also have an (almost) empty database.
The `create` command also imports the local host names `localhost` and `localhost.localdomain`
and the standard RFC 2142 set of local aliases.
Later on during Postfix configuration we will add an alias for `root` that will
redirect all this standard system email traffic to someone who will actually read them.
It also adds the following domains:
```bash
[root@pobox dovecot]# postdove show domain localhost
Name:           localhost
Class:          local
Transport:      --
Access:         --
UserID:         99
Group ID:       99
Restrictions:   --
[root@pobox dovecot]# postdove show domain localhost.localdomain
Name:           localhost.localdomain
Class:          local
Transport:      --
Access:         --
UserID:         --
Group ID:       --
Restrictions:   --
```
Notice the `99` for the `UserID` and `Group ID` for `localhost`.
This is the system fallback default for users.
See [Postdove Administration](admin.md) for more information.

## SELinux Changes
We use **Fedora** which has SELinux enabled by default. Some other distributions, such as
**Ubuntu** use the **AppArmor** security system that enforces access controls using a different
method. Since our development base is **Fedora**, we only discuss SELinux. Consult the appropriate
documentation in your distribution if that is the case.

SELinux is a Mandatory Access Control (MAC) security system.
There are many HOWTOs out on the net that advise as first
thing to disable SELinux which could be understandable years ago or in a simplistic and
sheltered local network. However, those innocent days are long gone and this is an email
system, by definition, a lucrative target for mischief. Given today's open Internet environment,
one should not disable their MAC subsystem simply because it is too complex or unfamiliar.
The access protections outlined above are barely sufficient for a public server.

The default access configuration for `postfix` and `dovecot` is to have them isolated in their
own enviroment such that neither service can get out of their own environment and especially into
the other's environment. The original design had the database reside under the `/etc/dovecot`
directory. However, this proved messy because although `dovecot` has only a few processes,
`postfix` has a whole army of them, one for each isolated step of mail processing.
It is much simpler to make the database live in the `postfix` configuration directory and
grant access to it by `dovecot`.

For each of these SELinux changes what we do is run the server and then add access policies based
on the output of the SELinux utilities. Since the database is located inside the `/etc/postfix`
directory, there is nothing needed for `postfix`. In addition, since mail, either local or relay,
is passed out of `postfix` via the network or a UNIX socket typically,
everything has been already set up by the package installation.
The only service we must customize is for `dovecot` access.

The process is simple. The `dovecot` service will fail when it attempts to access the database file.
The easiest way to trigger this is to have a user log in to IMAP (or POP3). This will trigger an
error because the database file is not accessible. The system messages log will simply report that the file
is not accessible even though the UNIX mode bits and ownerships all seem in order as set up
above. It is SELinux that is denying access so it is necessary to ask its audit logs what the
problem is. Since you must run these commands as **root**, do this work in the **root** home
directory where you can save the files for future reference.

```
[root@pobox dovecot]# cd ~
[root@pobox ~]# sealert -a /var/log/audit/audit.log
```

This command will scan the audit log looking for trouble (denied accesses). If the system has
been running for a while, the log could be long. Eventually it will find something to report.
It will first display the details of the access violation and then recommend two commands to
run should the violation be what we are looking for. Be careful to verify the details to make
sure that the violation is for our subsystem, probably `dovecot` and not from something else.
There are a lot of other violations that occur on a busy system. 
We are ignoring them here. The first command it recommends is:

```
[root@pobox ~]# ausearch -c 'auth' --raw | audit2allow -M my-auth
```

This is a pipeline that does essentially a `grep` of the log looking for,
in this case `auth` which is the `dovecot` user authentication process.
Those records are fed into `audit2allow` which generates a policy module that looks like this:

```
[root@pobox ~]# cat my-auth.te 

module my-auth 1.0;

require {
        type postfix_etc_t;
        type dovecot_etc_t;
        type dovecot_auth_t;
        class file { getattr lock open read write };
        class dir search;
}

#============= dovecot_auth_t ==============

#!!!! This avc is allowed in the current policy
allow dovecot_auth_t dovecot_etc_t:file write;

#!!!! This avc is allowed in the current policy
allow dovecot_auth_t postfix_etc_t:dir search;

#!!!! This avc is allowed in the current policy
allow dovecot_auth_t postfix_etc_t:file { getattr open read write };
allow dovecot_auth_t postfix_etc_t:file lock;
```

There is also a `my-auth.pp` file which is binary and used in the next step.
The contents of the policy are simple. The `type` is the SELinux label attached to files
and executables. The `class` lines describe what is wanted. In this case, the `sqlite3`
library wants to find and then open the file for read/write/lock. The `getattr` is needed
to find out about the file while searching for it. The `allow` directives are what `dovecot`
needs to properly use it. Note that neither `dovecot` nor `postfix` ever write to the
database. The `sqlite3` library does not know what the application that calls it
would do and assumes it might do updates so it opens the file for full access anyway.

**NOTE:** Should either server somehow attempt a database query that would
attempt to write a record,
i.e. a `INSERT`, `DELETE`, or `UPDATE` the library would return an operation failed error because the application only has read-only access.

The next step is to install the new policy. This is done by the `semodule` command:

```
[root@pobox ~]# semodule -X 300 -i my-auth.pp
```

This command will take more than a few seconds to run because it must wait for things to
settle in the operating system before making the change.

The next step is to run the test again. This will most likely be a rinse-and-repeat cycle because
SELinux will trigger the denial at each step starting with the directory search and ending
up with the 'lock' denial as `sqlite3` starts to process queries.

### Other SELinux Access
If the configuration uses `virtiofs` for accessing the mail storage, there may be further
SELinux changes required. Should the `dovecot` logs report access errors, follow the procedure above
to check what SELinux is doing. There is separate access labeling for the mounted volume because
it is actually on the host system which has its own labeling and access controls. The following
module file is the result of testing as above:
```
[root@pobox ~]#  cat dove.te 

module dove 1.0;

require {
        type virtiofs_t;
        type dovecot_t;
        type dovecot_etc_t;
        type dovecot_auth_t;
        class file { append getattr read write };
        class dir { read write };
}

#============= dovecot_auth_t ==============
allow dovecot_auth_t dovecot_etc_t:file write;

#============= dovecot_t ==============
allow dovecot_t virtiofs_t:dir { read write };
allow dovecot_t virtiofs_t:file { append getattr read write };
```
Again, test first and testing may be a rinse-repeat cycle because later (deeper) access
issues will not appear until the ones above it are resolved.

This covers database setup.
We will need a lot more than that to have a useful mail server.
We will cover those details in [Postdove Administration](admin.md) but first
we must configure the `dovecot` server.
The next step is [Dovecot Configuration](dovecot_configuration.md).
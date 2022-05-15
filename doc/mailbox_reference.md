# Mailbox
The `mailbox` sub-command manages `dovecot` user accounts and their properties.

Mailbox properties are `dovecot` user password file entry values.
The *uid*, *gid*, *mail-home*, and *password* properties are the same as what is present
in the UNIX `/etc/passwd` file. The configuration documented here uses the *uid* and *gid* values
from the server. If `dovecot` uses LDAP or similar directory services, the `dovecot` usage would
be the same as for user logins and storage.
Since `postdove` is standalone, using the same numbers is optional although recommended.
The `mail-home` property is slightly different in that it is it can be relative to a user's
home directory or relative to another part of the filesystem outside a user's directory.
The `postdove` configuration has its email storage on a separate filesystem.

The password type is different from what is expected in `/etc/passwd`.
There are three types available in `postdove` although others are possible in `dovecot`.
* `plain` passwords are clear text.
This is the most common type since IMAPS and POP3S (using TLS encryption) are now the typical
and recommended configuration.
* `crypt` passwords are the same as `/etc/passwd` passwords.
* `sha256` passwords use the SHA256 digest algorithm.

See the `dovecot` documentation for more details, especially the advantages of each type.

The *quota* value is, where applicable, in three forms of interest here but see the documentation
for all the variations.
Multiple quota rules can be set on an account's storage, i.e. one for *Trash* and another for the rest.
* `none` No quota is set allowing unlimited storage.
* `reset` This is used for editing an entry to reset the value to the *default* defined in the database schema.
* `<mailbox name>:<limit configuration>` Sets the limits for the folder/mailbox.
	- `<mailbox name>` This is where this rule applies. `*` configures the default limit for everything.
	Using a folder name applies just to that folder. For example, for having extra space for *Trash*.
	- `<limit configuration>` The limit is `<limit name>=<size>` where the limit name is `backend`, `bytes`,
	`ignore`, `messages`, or `storage`. The size is either a number or a number with suffix `B`, `K`, `M`, or `G`.

The current quota rule defined in the schema is `*:bytes=300M` which means 300MB of storage is the
quota for the account.
All added accounts that do not specify a quota get this default value.
This is also the value that is stored for the account when `reset` is used as the quota rule.
The mailbox must be explicitly edited to set quota to `none` to remove quota limits.
See the documentation for all the variations.
	
## Add
Add a user to the `dovecot` email system.
The `dovecot` configuration is set up such that the first access, either by the user logging in
or by `postfix` posting new email, will create the necessary directories in the email storage.

Use the help option to show the command.
```
[root@pobox ~]# postdove add mailbox -h
Add an mailbox into the database. The address must be in an already
existing vmailbox domain. The flags set the various login parameters such as password and
quota.

Usage:
  postdove add mailbox address [ flags ] [flags]

Flags:
  -e, --enable             Enable this mailbox for access
  -g, --gid int            User ID for this mailbox (default 99)
  -h, --help               help for mailbox
  -m, --mail-home string   Home directory for mail
  -E, --no-enable          Enable this mailbox for access
  -p, --password string    Account password
  -q, --quota string       Storage quota
  -t, --type string        Password encoding type (default "PLAIN")
  -u, --uid int            User ID for this mailbox (default 99)

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit
```

### Options
The one required argument is the name, the email address to be added.

* `--type=<scheme>` This is the password encoding scheme.
Accepted types are `plain`, `crypt`, or `sha256`.
* `--password=<password string>` This is plain text for *plain* passwords.
See the `dovecot` documentation for how to create encrypted type passwords.
The result of that action is copied to here.
* `--uid=<number>` This is the *uid* used for all file operations including inter-user access control.
* `--gid=<number>` This is the *gid* used for all file operations.
These two fields typically copy the values in the `/etc/passwd` authorization on the server or network.
* `--mail-home=<path to home>` This is the *home* of the email store for this account.
If this option is not set, the `dovecot` configuration default is used.
If you wish to set this to something other than the default described in the `postdove` documentation,
carefully consult the `dovecot` documentation. It can be done but the "there be dragons" in the details.
* `--quota=<string>` This is the quota for the mailbox. If not set, use the database default.
* `--enable` This enables the mailbox for IMAP/POP3 login. If not set, the default is `true`.
* `--no-enable` This is equivalent to `--enable=false`.
IMAP/POP3 logins are denied but the mailbox can receive email.

Extra care should be taken with the *uid*, *gid* and *home* properties because
once email has been delivered or the user has logged in, file storage is created.
Changing these properties later will also require the changing of ownerships and/or locations in
the email storage. In other words, get it right the first time or expect to work late...

### Examples
Add a mailbox taking all the defaults.
The *password* will be blank, allowing logins without a password. The type will be set to `plain`.
The *uid* and *gid* will be the domain wide (or server wide) default and the *mail home* will be the
configuration default. The mailbox will be enabled with the default quota set.
```
[root@pobox ~]# postdove add mailbox test@example.com
```
A more complete and safer add of the user would be:
```
[root@pobox ~]# postdove add mailbox test@example.com -u 1003 -g 1003 -p ChangeMe
```
Note that you may have to use single quotes `'` if you use characters that the shell may want to expand.

## Delete
Delete a mailbox from the `dovecot` system.
This removes the user from the database and makes access by either `postfix` or via IMAP or POP3 return
an error to indicate that the mailbox does not exist.

This command does not delete any emails stored for the user.
In order to completely remove a user, the emails stored in the filesystem must be removed as well.
This process is more completely documented in the `dovecot` documentation.

The command will return an error if this mailbox is a target/recipient of any alias.
The alias must be edited to remove this mailbox first.
The address associated with this mailbox is also deleted.

The domain part of the address will not be deleted because it references the virtual domain served by `dovecot`.
In order to remove the whole virtual domain, first remove all of the mailboxes and then remove the domain.

Use the help option to show the command.
```
[root@pobox ~]# postdove delete mailbox -h
Delete an address mailbox and its address from the database.
All of the aliases that point to it must be changed or deleted first

Usage:
  postdove delete mailbox address [flags]

Flags:
  -h, --help   help for mailbox

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit 
```

### Options
The command requires one argument naming the mailbox to be deleted.

The command has no options.

### Examples
Delete a mailbox.
This command will not remove any of the user's emails.
See the `dovecot` documentation for that process.
```
[root@pobox ~]# postdove delete mailbox test@example.com
```

## Edit
Edit the properties of a mailbox.
As noted in the `add` command above, take care in editing the *uid*, *gid* and *home* properties.
If you must edit them, do so before the user logs in or the system stores email.
Note the disabling the mailbox only blocks logins but not incoming mail.

Use the help option to show the command.
```
[root@pobox ~]# postdove edit mailbox -h
Edit a mailbox to change attributes such as uid/gid, password, quota.

Usage:
  postdove edit mailbox address [ flags ] [flags]

Flags:
  -e, --enable             Enable this mailbox for access (default true)
  -g, --gid int            Group ID for this mailbox (default 99)
  -h, --help               help for mailbox
  -m, --mail-home string   Home directory for mail
  -E, --no-enable          Enable this mailbox for access
  -G, --no-gid             Clear Group ID for this mailbox
  -M, --no-mail-home       Clear Home directory for mail
  -P, --no-password        Clear Account password
  -U, --no-uid             Clear User ID for this mailbox
  -p, --password string    Account password
  -q, --quota string       Storage quota
  -t, --type string        Password encoding type (default "PLAIN")
  -u, --uid int            User ID for this mailbox (default 99)

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit
```

### Options
The command requires an email address argument.

* `--enable` Enable the account for login via IMAP or POP3.
* `--no-enable` Disable the account. This prevents logins but does not block incoming email.
* `--gid=<number>` Change the gid to this number.
This is typically the same *gid* number used for files and logins elsewhere.
* `--no-gid` Clear the group ID for this user.
This will result in the default ID for the domain or the installation to be used.
* `--uid=<number>` Change the uid to this number.
This is typically the same *uid* number used for files and logins elsewhere.
* `--no-uid` Clear the user ID for this user.
This will result in the default ID for the domain or the installation to be used.
* `--mail-home=<mail path>` Change the location for where email is stored or fetched
to the path specified in the option.
* `--no-mail-home` Clear the mail home property.
This will result in `dovecot` using the configuration default.
* `--password=<string>` Change the account password to the string.
* `--no-password` This clears the password for this account.
There is no password for this account.
Depending on how `dovecot` is configured this could open the account to the world.
* `--type=<password encoding>` Change the encoding type for the password to this value.
The accepted types are `plain`, `crypt`, and `sha256`.
The default is `plain` which is typical for most IMAPS accounts.
See the `dovecot` documentation for all the current encoding types.
`postdove` does not check the validity of this string in the `dovecot` configuration.
Any errors will be reported by `dovecot` to its logs.
* `--quota=<quota value>` Change the storage quota for this account.
The quota value is the string defined in the `dovecot` documents.
If the value is `none`, no quota is set, i.e. storage is not limited.
If the value is `reset`, the quota is set to the database schema default.
The validity of this string is not checked by `postdove` so any errors (typos)
will show up in the `dovecot` logs.

### Examples
Edit a mailbox to change the password. Note we leave the type unchanged.
```
[root@pobox ~]# postdove edit mailbox test@example.com --password=CamelC@se
```
Remove the quota on this mailbox.
```
[root@pobox ~]# postdove edit mailbox test@example.com --quota=none
```
Reset the quota back to the system default.
```
[root@pobox ~]# postdove edit mailbox test@example.com --quota=reset
```

## Export
Export mailbox definitions to a file.
The format of the file conforms to the user database definitions of `dovecot`.

Use the help option to show the command.
```
[root@pobox ~]# postdove export mailbox -h
Export mailboxes in a /etc/passwd similar format to
the file named by the -o flag (default stdout '-').

Usage:
  postdove export mailbox [flags]

Flags:
  -h, --help   help for mailbox

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -o, --output string   Output file in postfix/dovecot format (default "-")
  -v, --version         Report Postdove version and exit
```

### File Format
The export and import entry format follows the password file format for `dovecot`.
Each field is separated by a `:` except for the last (eighth) and optional *extra fields*.
For example:
```
test@sea-troll.net:{PLAIN}bogustest:1003:1003::::userdb_quota_rule=*:bytes=300M mbox_enabled=true
```

The fields are in order:
* `test@sea-troll.net` This is the account/mailbox name.
It is the user name for IMAP login and the address for email delivery.
* `{PLAIN}bogustest` This is the password field with the type enclosed in `{...}`.
* `1003` This is the *uid* used for all file access.
* `1003` This is the *gid* used for all file access.
* `<gecos>` This field, the `/etc/passwd` *gecos* field is unused. This is really ancient compatibility...
* `<home>` This is the *mail-home* location. Being blank means use the `dovecot` configuration default.
* `<shell>` This is the *shell field*, obviously not used in `dovecot`.
* `<extra fields>` This is an optional field. We use it for *quota* and the mailbox *enable* property.
The quota here is set to 300MB of total storage and the mailbox is enabled.


### Options
One optional argument can be supplied to define which mailboxes to export.
If a wildcard `*` is used, it will export all the mailboxes that match.
No argument implies export of all mailboxes.

* `-o <output file` Redirect the standard output to the output file.

### Examples
Export all the mailboxes to the file `users.txt`.
```
[root@pobox ~]# postdove export mailbox > users.txt
```
Export just the mailbox `test@example.com` to the standard output.
```
[root@pobox ~]# postdove export mailbox test@example.com
```
Export all the mailboxes in `example.com` to the file `mailboxes.list`.
We put single quotes `'` around the argument to keep the shell from expanding it.
```
[root@pobox ~]# postdove export mailbox '*@example.com' -o mailboxes.list
```

## Import
Import mailbox definitions from a file.

Use the help option to show the command.
```
[root@pobox ~]# postdove import mailbox -h
Import a set of mailboxes into the database
from the file named by the -i flag (default stdin '-').

Usage:
  postdove import mailbox [flags]

Flags:
  -h, --help   help for mailbox

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -i, --input string    Input file in postfix/dovecot format (default "-")
  -v, --version         Report Postdove version and exit
```

### Options
There are no arguments for this command.

* `-i <import file>` Redirect the standard input to the import file.

### Examples
Import mailbox definitions from the file `mailbox.list`.
```
[root@pobox ~]# postdove import mailbox < mailbox.list
```

## Show
Display a mailbox and its properties.

Use the help option to show the command.
```
[root@pobox ~]# postdove show mailbox -h
Display the mailbox and its attributes to standard output

Usage:
  postdove show mailbox address [flags]

Flags:
  -h, --help   help for mailbox

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit
```
### Options
The one required argument is the mailbox name.

There are no options.
### Examples
Show the properties of user `test@example.com`.
The user ID and group ID match what the server system's `/etc/passwd` file has.
The home directory is the `dovecot` system default.
The user has an allocated quota of 300 megabytes.
```
[root@pobox ~]# postdove show mailbox test@example.com
Name:           test@example.com
Password Type:  PLAIN
Password:       bogustest
UserID:         1003
GroupID:        1003
Home:           --
Quota:          *:bytes=300M
Enabled:        true

```




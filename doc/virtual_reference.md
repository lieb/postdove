# Virtual Alias
The `virtual` sub-command manages aliases that also have a domain part, e.g. `user@some.domain`
as opposed to an alias, e.g. `postmaster`.
Virtual aliases can either be for any domain whereas an *alias* is local system specific.
Although similar in use, there are differences in what types of recipients are allowed.
Whereas an *alias* is local and therefore can have file or pipeline recipients,
a virtual address can be addressed to any domain including the local one so its
recipients can only be email mailbox destinations.

## Add
Add a virtual alias to the system.
The virtual alias is created if it does not already exist.
Otherwise, the recipient(s) listed as arguments are added to the existing virtual alias.

Use the help option to show the command.
```
[root@pobox ~]# postdove add virtual -h
Add an virtual alias address into the database. The address is an RFC2822
email address. One or more recipients can be added. A recipient can either be a single local mailbox or
an RFC2822 format email address. See postfix virtual(5) man page for details.

Usage:
  postdove add virtual address recipient ... [flags]

Flags:
  -h, --help   help for virtual

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit
```

### Options
The command requires a minimum of two arguments.
The first is the virtual alias to be created.
The second and additional arguments are recipients targeted by the virtual alias.
Recipients can be of the form `user` or `user@domain`.

There are no options for this command.

### Examples
Add a virtual alias for a simple distribution list.
```
[root@pobox ~]# postdove add virtual gang@example.com mary@example.com dave@example.com
```

## Delete
Delete a virtual alias.
The commmand will also remove any recipients that would be *orphaned*,
namely they are not used anywhere else in the system.

This command is similar to `postdove delete alias`.
The difference is in the name of the alias to be deleted.
If the intent is to delete an alias with no domain part, such as `admin`,
use the `postdove delete alias` command instead.

Use the help option to show the command.
```
[root@pobox ~]# postdove delete virtual -h
Delete an virtual address alias from the database.
All of the recipients pointed to by this name will be also deleted

Usage:
  postdove delete virtual address [flags]

Flags:
  -h, --help   help for virtual

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit
```

### Options
This command requires a single argument naming the virtual alias to be removed.

There are no options for this command.

### Examples
Delete the virtual alias `marketing@example.com`.
This command will also delete all recipients that are not *busy*, namely, they are not recipients of another alias, aliases themselves, or are a `dovecot` mailbox.
```
[root@example.com]# postdove delete virtual marketing@example.com
```

## Edit
Edit a virtual alias to add or remove a recipient.

Use the help option to show the command.
```
[root@pobox ~]# postdove edit virtual -h
Edit a virtual alias address and its list of recipients. You can edit, add,
or delete recipients

Usage:
  postdove edit virtual address [flags]

Flags:
  -a, --add strings      Recipient to add to this virtual alias
  -h, --help             help for virtual
  -r, --remove strings   Recipient to remove from this virtual alias

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit
```

### Options
This command requires the virtual alias to be edited as a single argument.

* `--add=<recipient>` Add a recipient to the virtual alias.
* `--remove=<recipient>` Remove the named recipient from the virtual alias.

### Examples
Edit `admin@example.com` to add a new recipient and remove `bill@eng.example.com`.
```
[root@pobox ~]# postdove edit virtual admin@example.com -r bill@eng.example.com --add=dave@eng.example.com
```
These two commands which add a recipient to `admin@example.com` are equivalent.
```
[root@pobox ~]# postdove edit virtual admin@example.com -a mike@example.com
[root@pobox ~]# postdove add virtual admin@example.com mike@example.com
```

## Export
Export virtual aliases to the standard output or a file in `postfix` virtual alias format.

Use the help option to show the command.
```
[root@pobox ~]# postdove export virtual -h
Export virtual aliases in postfix virtual(5) format to
the file named by the -o flag (default stdout '-').
This is typically the file that maps various email virtual addresses to relay or IMAP/POP3 mailboxes

Usage:
  postdove export virtual [flags]

Flags:
  -h, --help   help for virtual

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -o, --output string   Output file in postfix/dovecot format (default "-")
  -v, --version         Report Postdove version and exit
```

### File Format
The file format for export of virtual aliases conforms to the format described in the
`postfix` documentation. See `virtual(5)` for the details.
The first field is the virtual alias.
The following comma(`,`) separated fields are its recipients.

An individual entry is of the following form:
```
dave@example.com dave@eng.example.com, dave@home.net
```

### Options
The command accepts one argument that defines which virtual aliases are exported.
If a wildcard, `*` is used, all the virtual aliases that match the expression
are exported.
If there is no argument, all virtual aliases are exported.

* `-o <output file>` Redirect the standard output to the output file.

### Examples
Export `root@example.com`.
```
[root@pobox ~]# postdove export virtual root@example.com
```
Export all of the virtual aliases to a file.
```
[root@pobox ~]# postdove export virtual -o backup_virtuals.txt
```
Export just the virtuals for `example.com`.
```
[root@pobox ~]# postdove export virtual '*@example.com' > example.com.aliases
```

## Import

Use the help option to show the command.
```
[root@pobox ~]# postdove import virtual -h
Import a local virtual alias file in the postfix virtual(5) format
from the file named by the -i flag (default stdin '-').
This is postfix file associated with the $virtual_aliases hash

Usage:
  postdove import virtual [flags]

Flags:
  -h, --help   help for virtual

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -i, --input string    Input file in postfix/dovecot format (default "-")
  -v, --version         Report Postdove version and exit
```

### Options
There are no arguments for this command.

* `-i <input file>` Redirect the standard input to this file.

### Examples
Import virtual aliases from the file `virtual.txt`.
Both examples are equivalent.
```
[root@pobox ~]# postdove import virtual < virtual.txt
[root@pobox ~]# postdove import virtual -i virtual.txt
```

## Show
Show a virtual alias and its targets/recipients.

Use the help option to show the command.
```
[root@pobox ~]# postdove show virtual -h
Display the contents of an virtual alias and all its recipients
to the standard output

Usage:
  postdove show virtual address [flags]

Flags:
  -h, --help   help for virtual

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit
```

### Options
One argument is required, the virtual alias

There are no command options.

### Examples
Show the virtual alias `root@example.com`.
```
[root@pobox ~]# postdove show virtual root@example.com
Virtual Alias:  root@example.com
Targets:        lieb@example.com
        	test@example.com
```

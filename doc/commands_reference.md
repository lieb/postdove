# Postdove Commands Reference
The `postdove` utility is a CLI tool for managing a `postfix` and `dovecot` mail server.
It resides on the mailserver and is accessed through an `ssh` session. There is no GUI or
web interface planned because such an interface opens up the security attack surface for the
server considerably. Using `ssh`, especially with its authorized keys and no password login
restrictions, administrative access would be limited to the already set up procedures in
place on the system. On systems that also use the `cockpit` web based administation tool,
`postdove` can be used in its **Terminal** page.

The user must have **root** privileges to access the database.

## Options and Arguments
Arguments are ordered command line strings that are passed to the command. Options or flags
are used to modify the behavior of the command. These follow the GNU command line arguments
style. For example, the following command adds an
access rule to the database:
```
[root@pobox ~]# postdove -d /tmp/play.sqlite add access frob frobby_action
```

The two arguments, `frob` and `frobby_action` are ordered. The first, `frob` is the name
argument and the second, `frobby_action` is the restriction or action key.

* `-d` is the short form flag and it has a value that modifies the behavior, specifically, it sets the database file
to be used by the command. There are two equivalent forms of a flag/option.
`-d <filename>` is a short form where the `-d` is followed by a space and then the
`<filename>` in this case, `/tmp/play.sqlite`.

* `--dbfile=<filename>` is a long form where the flag can be a string which is easier to
understand than a single letter. Notice that the space has been replaced by a `=` character.
There can be no spaces in this form.

If we use just the long form in this command, it looks like:
```
[root@pobox ~]# postdove --dbfile=/tmp/play.sqlite add access frob frobby_action
```

The simplest way to explore the tool is to take advantage of the extensive command line help
provided by the interface. For example:
```
[root@pobox ~]# postdove help
Postdove is a management tool to manage the sqlite database file that
is used by postfix to manage aliases, domains, and delivery and by dovecot to 
manage email user IMAP/POP3 email accounts

Usage:
  postdove [flags]
  postdove [command]

Available Commands:
  add         Add an entry into the specified table
  completion  Generate the autocompletion script for the specified shell
  create      Create the Sqlite database and initialize its tables
  delete      Delete an entry in the specified table
  edit        Edit an database entry in this table
  export      Export the specified table to a file or stdout
  help        Help about any command
  import      Import a file to the database
  show        Show the contents of a table entry

Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -h, --help            help for postdove
  -v, --version         Report Postdove version and exit

Use "postdove [command] --help" for more information about a command.
```
### Global Flags
While there are option flags that are specific to individual commands the three listed above
are global. They apply to all commands.

* `--dbfile` sets an alternate database file for the command. This is useful for testing
and experimentation. Administrator, i.e. `root`, privilege is only required for the system
database. Its value string is the path to the database file in the filesystem.

* `--help` option flag displays a description of all of the option flags,
subcommands and their meanings in the context of a particular command.
This display above is for the top level. It shows all of the available commmands which are each fully
described in the following sections.

* `--version` displays the version of the utility and exits. It displays the current `git` tag to identify
the source version used to build the utility.

```
[root@pobox ~] postdove --version
Version: V0.9-RC1-3-ga8b7420
```
## Create a Database
This is a one time command to create the database.
It must be run before any other command.
It will overwrite an existing database so it should be run with care, i.e. only once.

See [Create Command Reference](create_reference.md) for the details. This is the only command
that has no sub-commmands.

The rest of the commands are associated with the data files, usually hash indexes, used by `postfix` with
the exception of `mailbox` which manages the `dovecot` user database.

## Imports and Exports
Each of the following commands have an `export` and `import` command.
These commands are used for bulk transfer of database contents and the format
for both `export` and `import` is identical.
Where possible and relevant, the format used by other email systems is used.
For example, the format for *alias* import is identical to `/etc/aliases` which
has been around for a long time. The same applies to virtual aliases which are
used for hash files in `postfix`.
The import and export of user authorization data is identical to the format
used by `dovecot` for `/etc/passwd` style user authorization.

These formats are useful for editing files from other sources into a form
suitable for import by `postdove`.
The formats are also useful for exporting database contents to other processes.

Each management reference describes the detail for each import/export.
These files are formatted as one line per entry with different field definitions
depending on the context.

All files have common format conventions for parsing or generating the file.

* White space is either a space or tab character.
One or more white space characters are treated as a single white space.
* A file import can have comments which are stripped and ignored by the import.
A comment is any text starting with a `#` and ending at the end of the line.
The `#` can be anywhere on the line.
For parsing purposes the `#` is treated as the end of that line and is stripped
before any other processing.
* A line with a leading `#` is treated as a blank line.
It can have any amount of leading white space before the `#` which is ignored.
* Any trailing white space on a line is trimmed before any token processing is done.
* All new lines start at the first column unless it is a *continuation* line.
A *continuation* line starts with any amount of white space.
The leading white space is stripped to a single blank space and it is appended to
the previous line for processing.
* A line is split into tokens based on a specified format.
All leading and trailing white space is trimmed from the token before
further processing.

Depending on the file, tokens are parsed in different formats to match the legacy
use for the file contents. There are four line formats using common BNF notation:
* `SIMPLE` - This is `key WS+ text`.
The *key* is a single alphanumeric word.
It is followed by one or more white space characters.
The *text* is the remainder of the line.
* `POSTFIX` - This is used for `postfix` compatible files.
The syntax for this type of file is `key WS+ [token ',' WS*]+`.
The *key* is a single alphanumeric word.
It is followed by one or more white space characters.
The rest of the line is tokens separated by commas (`,`).
* `ALIASES` - This is the format of the `/etc/aliases` file which dates back to
antiquity when UNIX systems didn't even have networking yet.
The syntax for this type of file is `key ':' WS* [token ',' WS*]+`.
This is similar to `POSTFIX` with one exception.
The *key* is terminated by a colon (`:`) character.
* `PWFILE` - This is the familiar colon(`:`) separated fields format used for
`/etc/passwd.`
The individual fields are separated into tokens by the colon(`:`). Any internal white space in the token is untouched.

Some files have tokens of the form `<property>=<value>`.
These are used for importing and exporting the properties of an item.
For example, an *address* has *transport* and *rclass* properties.
The *rclass* property would be the token `rclass=restrict` which means the
*rclass* property has a value of `restrict`.

Since comments are stripped on import, no comments are added on export.

## Access Controls
`postfix` examines incoming email with the primary goal of rejecting spam and phishing attacks.
This is done by a series of filters that examine either the incoming connecting server or
the content of the email itself. It examines the server at connection time against databases of
known bad actors. This allows the connection (and email) to be rejected before the data transfer.
If the incoming email survives the connection, the next step is to examine the email itself.

There are a series of tests that are controlled by lists of filters specified
in the `main.cf` configuration file.
One option is to define classes of filters and then select them via lookups where the lookup key
is a domain or user at a domain. The access controls in `postdove` manage those lookups.
See [Access Commands Reference](access_reference.md) for the details. These are not mandatory, i.e.,
`postfix` filtering can be configured to treat all emails the same but using these
access controls can fine tune the behavior.

## Transport Management
Once an email has been checked by the filters, `postfix` must then do something with it.
The protocols require `postfix` to either forward the email to its destination or bounce
the undeliverable email back to its source.
The transport tables determine where to forward the email based on the recipient address.
See [Transport Management Reference](transport_reference.md) for the details.
At least one transport entry must be present. Namely, the one that sends email to the
`dovecot` server.

## Domain Management
Domains are used for a number of functions either as part of an address or for domain wide actions.
Depending on their use, domains have properties that control the actions.
See [Domain Management Reference](domain_reference.md) for details and use.

## Address Management
In addition to domains, addresses within a domain are the primary endpoints of email transmissions.
Addresses can be the name of an alias, one of the recipients associated with an alias, or a user
account in `dovecot`.
See [Address Management Reference](address_reference.md) for details and use.

## Alias Management
An alias does not have a domain component, e.g. `info` and it is associated only with the local
server. The domain is inferred to be `localhost` or the DNS name of the server itself.
See [Alias Management Reference](alias_reference.md) for details and use.

## Virtual Alias Management
A virtual alias is similar to a local alias except that its name always has a domain part,
i.e. `user@some.domain`. There are no restrictions on what the domain part is for
either the name or any in the list of recipients.
See [Virtual Alias Management Reference](virtual_reference.md) for details.

## Mailbox Management
Mailboxes are managed by `dovecot`. Each mailbox has a set of properties that are managed by `dovecot`.
See [Mailbox Management Reference](mailbox_reference.md) for details.

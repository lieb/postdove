# Alias
The `alias` sub-commmand manages local aliases. A local alias is a single term like `postmaster` or `admin` without a domain component.
Local aliases are handled differently by `postfix` and can have different types of
recipients given that their scope is the local email server system.

Aliases can have three forms for their recipients because all references are local.

* `user` or `user@domain` The first is a local user and the second is a full address.
The email to the alias is forwarded the mailbox of `user` or possibly relayed to
`user@domain` if `domain` is other than the local system.
* `/some/file` If the first character is a `/` the recipient is the file `/some/file` on
the local system. The email is appended to the file.
* `"| program --flag arguments"` The email is piped to the command `program` via the shell.
This is a pipeline and all the flags and arguments within the quotes (`"`) are set.
Note that the quotes must be present. Otherwise it would be an error both in `postdove` and `postfix`.

## Add
Add an alias and its recipients to the database.
The alias is created if it does not already exist.
Otherwise, the recipient(s) are added to the already existing alias.

Use the help option to show the command.
```
[root@pobox ~]# postdove add alias -h
Add an alias into the database. The address is a local
user or alias target without a "@domain" part, i.e. "postmaster" or "daemon".
One or more recipients can be added. A recipient can either be a single local mailbox,
i.e. "root" or "admin", an RFC2822 format email address, or a file or a pipe to a command.
 See aliases(5) man page for details.

Usage:
  postdove add alias address recipient ... [flags]

Flags:
  -h, --help   help for alias

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit

```
### Options
The command requires at least two arguments.
The first argument is the alias itself and the rest are recipients.
Recipients can be of the form:
* `name` which is either a local system user or another alias.
* `name@domain` which is a virtual alias or mailbox at `domain`.
* `/file/path` which is a named file that email will be appended to.
* `"| command <args and options>"`which is a shell command that the email
will be piped to. Note that the double quotes `"` and `|` are required.

There are no command options.

### Examples
Create a new alias `noc` with a two recipients, `dave@site.com` and `"| logger --tag=noc"`
```
[root@pobox ~]# postdove add alias noc dave@site.com '"| logger --tag=noc"'
```
Note the single quotes (`'`). This is necessary so the shell passes the '"' properly.

Add a recipient to `noc`.
```
[root@pobox ~]# postdove add alias noc bill@admin
```
The alias now has three recipients.

## Delete
Delete the alias and all of its recipients if they are not referenced for any other use.
This cleans up *orphans* that would otherwise clutter the database but leaves recipients that are used in other parts of the system from disappearing.
An example of this would be `root` which is the standard recipient for all the RFC 2142 required aliases found in the `/etc/aliases` file in most systems.
It is the recipient for all of these aliases and should not disappear in this case.

This command is similar to `postdove delete virtual`.
The difference between the two is an *alias* has no domain part as in `root` or `mailer-daemon`.
The command will return an error if the alias is of the form `user@domain` which is
a virtual alias. In that case, use `postdove delete virtual`.

Use the help option to show the command.
```
[root@pobox ~]# postdove delete alias -h
Delete an address alias from the database.
All of the recipients pointed to by this name will be also deleted

Usage:
  postdove delete alias address [flags]

Flags:
  -h, --help   help for alias

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit

```
### Options
The command requires one argument, the alias to be deleted.
If the intent is to delete a recipient, use the `edit` command.

There are no options.

### Examples
Delete the `noc` alias.
```
[root@pobox ~]# postdove delete noc
```
This will also delete the recipient `"| logger --tag=noc"` but may or may not delete the
others depending on whether they are referenced elsewhere.

## Edit
Edit the alias.
Use the help option to show the command.
```
[root@pobox ~]# postdove edit alias -h
Edit a local alias address and its list of recipients.

Usage:
  postdove edit alias address [flags]

Flags:
  -a, --add strings      Recipient to add to this alias
  -h, --help             help for alias
  -r, --remove strings   Recipient to remove from this alias

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit

```
### Options
The command takes a single argument.

* `--add=<recipient>` Add a recipient to the alias.
* `--remove=<recipient>` Remove the named recipient from the alias.

### Examples
Add more recipients and remove one from `noc`.
```
[root@pobox ~]# postdove edit alias noc -a mike@admin --add=jane@admin -r dave@site.com
```

## Export
Export aliases in `/etc/aliases` format.

Use the help option to show the command.
```
[root@pobox ~]# postdove export alias -h
Export a local aliases file in the aliases(5) format to
the file named by the -o flag (default stdout '-').
This is typically the /etc/aliases file that maps various system
users and email aliases to a specific user or site sysadmin mailbox

Usage:
  postdove export alias [flags]

Flags:
  -h, --help   help for alias

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -o, --output string   Output file in postfix/dovecot format (default "-")
  -v, --version         Report Postdove version and exit

```

### File Format
The alias file format conforms to the definition described in the `aliases(5)`
manual page in `postfix`.
The most familiar file with this format is `/etc/aliases`.
It is of the form:
```
alias: recipient1, recipient2, ...
```
where
* `alias` is a name without a domain part `@some.domain`.
* `recipient` can be of three forms, an alias, a file path, or a shell pipeline.
For example:
```
logs: root, /var/log/log_emails, "| logeater --flag"
```
which means that email sent to `logs` will be forwarded to `root`,
appended to `/var/log/log_emails`, and piped to the standard input of
of the utility `logeater` with the option flag `--flag`.

### Options
The command accepts no more than one argument.
If the argument is present, only aliases that match the argument are exported.
No arguments implies all aliases.

* `-o <aliases file>` Redirect the standard output to the named aliases file.

### Examples
Export all of the aliases to the file `aliases.txt`. Both these commands are equivalent.
```
[root@pobox ~]# postdove export aliases > aliases.txt
[root@pobox ~]# postdove export aliases -o aliases.txt
```
Export all the aliases starting with `ma`.
```
[root@pobox ~]# postdove export aliases 'ma*'
```
In a system with the embedded RFC 2142 aliases, this would export `mail`, `mailer-daemon`,
`mailnull`, `manager`, and `marketing`.

## Import
Import aliases from a file in `/etc/aliases` format.
The format allows comments but they are stripped by `import`.
This means that exporting after importing results creates a file without comments.

Use the help option to show the command.
```
[root@pobox ~]# postdove import alias -h
Import a local aliases file in the aliases(5) format
from the file named by the -i flag (default stdin '-').
This is typically the /etc/aliases file that maps various system
users and email aliases to a specific user or site sysadmin mailbox

Usage:
  postdove import alias [flags]

Flags:
  -h, --help   help for alias

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -i, --input string    Input file in postfix/dovecot format (default "-")
  -v, --version         Report Postdove version and exit
```
### Options
The command accepts no arguments.

* `-i <input file>` Redirect standard input to the named file.

### Examples
Import the RFC 2142 standard aliases.
```
[root@pobox ~]# postdove import alias -i /etc/aliases
```
This is equivalent to the initialization the `create` command does using its embedded aliases file.

## Show
Use the help option to show the command.
```
[root@pobox ~]# postdove show alias -h
Display the contents of an alias and all its recipients
to the standard output

Usage:
  postdove show alias address [flags]

Flags:
  -h, --help   help for alias

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit
```
### Options
The command requires one argument.

There are no options.

### Examples
Show the alias `postmaster` and its recipients.
These are named as *targets* since they can be more than an email address.
```
[root@pobox ~]# postdove show alias postmaster
Alias:  	postmaster
Targets:	root
```



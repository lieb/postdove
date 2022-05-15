# Domain
The `domain` sub-command manages domain names. A domain name refers to a host name discovered
from either Domain Name Service (DNS) queries or from local databases. A domain name always
resolves to an Internet address. However, the email system only deals with the names and leaves
the address resolution and use to the network connection services which means that a misspelling
in the database will only show errors in the `postfix` logs.

Domains for the most part are managed as part of other, mainly address, functions.
For example, if the address `jerry@golf.edu` is entered into the system and `golf.edu` is also new,
the domain `golf.edu` will be added as part of adding `jerry@golf.edu`.

There are exceptions.
The mailbox domain must be added before any mailboxes are added to that domain.
This is because the entry of `dovecot` mailboxes must check if the domain is for mailboxes.

Domains can also be automatically deleted when addresses are deleted.
If `jerry@golf.edu` is later deleted and there are no other users in the domain, the domain will
also be deleted. Of course, attempting to delete a domain that is *busy* will report an error.

Domains can be in any of five classes:
* `internet` - This is the default class and is used for most domains in the system.
* `local` - This class is for all the hostnames used by the server system.
The domains `localhost`, `localhost.localdomain` are set by the `create` command by importing
an embedded domains file.
The hostname of the server is added with a `local` class set so that `postfix` can know what
its name is. In addition, at some point `localdomain` part of `localhost.localdomain` gets
changed to to the domain of the server.
* `virtual` - A domain of this class is for virtual aliases.
* `relay` - A domain where email is relayed to uses this class.
* `vmailbox` - A domain in this class is part of the `dovecot` configuration.

The `uid` and `gid` properties are integer values that only have meaning for `vmailbox` domains.
Some configurations of `dovecot` use a common uid/gid pair for email ownership.
If a mailbox does not have a uid or gid set, the uid and gid from the mailbox's domain
are used instead.
There is also the case where the domains uid and/or gid are not set in which case, the
uid and gid properties of `localhost` are used.
This means that in an initial configuration not only does the `vmailbox` domain have to be
added, the `localhost` domain needs to be edited to set the uid/gid for the fallback case.
The default values for `localhost` are set by the embedded domains file to `99`, the user
and group of `nobody` on most Linux systems.

## Add
Add a domain to the database.

Use the help option to show the command.
```
[root@pobox ~]# postdove add domain -h
Add an domain into the database. The name is the FQDN for the domain.
The optional class flag defines what the domain is used for, i.e. for virtual mailboxes or local/my domain.

Usage:
  postdove add domain name [flags]

Flags:
  -c, --class string       Domain class (internet, local, relay, virtual, vmailbox) for this domain
  -g, --gid int            Virtual group id for this domain (default 99)
  -h, --help               help for domain
  -r, --rclass string      Restriction class for this domain
  -t, --transport string   Transport to use for this domain
  -u, --uid int            Virtual user id for this domain (default 99)

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit
```

### Options
The command expects one argument.

* `--class` Set the domain class.
* `--rclass` Set the restriction class (access rule).
An error will be returned if the class name does not already exist in the access rules.
* `--transport` Set the transport for the domain.
An error will be returned if the named transport does not exist in the transport table.
* `--uid` Set the default mailbox UID for this domain.
This is only applicable to `vmailbox` domains and if not set, the system will use the
value set for the `localhost` domain.
* `--gid` Set the default mailbox GID for this domain.
This is only applicable to `vmailbox` domains and if not set, the system will use the
value set for the `localhost` domain.

### Examples
Enter the domain that `dovecot` expects to use for IMAP services. Set the default uid/gid for
this particular domain.
```
[root@pobox ~]# postdove add domain my-domain.org --class=vmailbox -u 3000 -g 3000
```
Enter a relay domain where we have set up a transport `backroom` for it to be relayed to.
```
[root@pobox ~]# postdove add domain internal.my-domain.org --class=relay --transport=backroom
```

## Delete
Delete a domain.
Deleting a domain that is *busy*, is referenced by at least one `address` will return an error.

Use the help option to show the command.
```
[root@pobox ~]# postdove delete domain -h
Delete an address domain from the database.
All of the recipients pointed to by this name will be also deleted

Usage:
  postdove delete domain name [flags]

Flags:
  -h, --help   help for domain

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit
```
### Options
One argument must be specified naming the domain to be deleted.

There are no options.

### Examples
There is no way to edit a domain's name so the following makes the change for `localhost.localdomain`.
```
[root@pobox ~]# postdove add domain localhost@my-domain.org --class=local
[root@pobox ~]# postdove delete domain localhost.localdomain
```

## Edit
Edit the properties of a domain.

Use the help option to show the command.
```
[root@pobox ~]# postdove edit domain -h
Edit a domain and its attributes.

Usage:
  postdove edit domain name [flags]

Flags:
  -c, --class string       Domain class (internet, local, relay, virtual, vmailbox) for this domain
  -g, --gid int            Virtual group id for this domain (default 99)
  -h, --help               help for domain
  -G, --no-gid             Clear virtual group id for this domain
  -R, --no-rclass          Clear the restriction class for this domain
  -T, --no-transport       Clear the transport for this domain
  -U, --no-uid             Clear virtual uid value for this domain
  -r, --rclass string      Restriction class for this domain
  -t, --transport string   Transport to use for this domain
  -u, --uid int            Virtual user id for this domain (default 99)

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit
```
### Options
The command takes the name of the domain to be edited as its only argument.

* `--class=<class name>` Set the class to one of the classes named above.
* `--rclass=<access rule name>` Set the restriction class to the named access rule.
* `--no-rclass` Clear the restriction class property for this domain.
* `--transport=<transport name>` Set the transport for this domain to the named transport.
* `--no-transport` Clear the transport property of this domain.
This means that `postfix` will get no result and will then use the default transport set in its configuration.
* `--uid=<number>` Set the default uid to this value for mailboxes in this domain.
* `--no-uid` Clear the uid property for the domain.
If this is cleared, the uid property of `localhost` is used instead.
* `--gid=<number>` Set the default gid to this value for the mailboxes in this domain.
* `--no-gid` Clear the gid property for this domain.
If this is cleared, the gid property of `localhost` is used instead.
### Examples
Change the transport of `example.com` to `backend`.
```
[root@pobox ~]# postdove edit example.com --transport=backend
```

Clear the default uid for this domain to the system wide default in `localhost`.
```
[root@pobox ~]# postdove edit example.com --no-uid
```


## Export
Export domains and their properties to a file.
This is useful for backups as well as database transfers and reloads.
The format is similar to but does not match anything `postfix` uses.
This is because `postfix` typically only uses a domain name as a key in other tables.

Use the help option to show the command.
```
[root@pobox ~]# postdove export domain -h
Export domains to the file named by the -o flag (default stdout '-').

Usage:
  postdove export domain [flags]

Flags:
  -h, --help   help for domain

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -o, --output string   Output file in postfix/dovecot format (default "-")
  -v, --version         Report Postdove version and exit

```
### File Format
The file format has no similarity to file definitions in `postfix`.
This is because the hash files for `access(5)` and `transports(5)` use
an address or domain as their key.
The equivalent within `postdove` is to consolidate *access* and *transport*
entries into properties for addresses and domains.
The *class* property is internal to the database and is used to control the various queries and database operations. The remaining properties are used by `dovecot`
queries.

The format for the line defining a domain is:
```
domain class=<name> transport=<string> vuid=<number> vgid=<number> rclass=<string>
```
* `domain` is the domain name, either a subdomain or fully qualified host name.
* `class` is one of `internet`, `local`, `relay`, `virtual`, or `vmailbox`.
* `transport` is the name of the transport to be used for this domain.
* `vuid` is the user ID to be used for mailboxes in this domain if one is not
set for the mailbox itself.
* `vgid` is the group ID to be used for mailboxes in this domain if one is not
set for the mailbox itself.
* `rclass` string is the name of the access rule.

All domains have a class defined.
* `internet` This is the default class and most domains in the database have this class. It is mainly used to distinguish it as being not something else...
* `local` This is the local domain, either the DNS name of the host or `localhost`.
* `relay` This is a host used for relay operations by `postfix`.
* `virtual` This is a domain that has virtual aliases in it.
* `vmailbox` This is a domain that `dovecot` is configured for.

If any of these properties are cleared in the database, the property will not appear
in line for this domain in the export file.
For example:

```
example.com
```
has no properties set. If a query for a *uid* or *gid* is made,
the values set for the domain `localhost` are returned.
```
example.com rclass=allow
```
This is an `internet` class domain that has the `allow` access rule applied to it
by `postfix`. Defaults for all other properties are applied.
```
eng.example.com class=relay rclass=slow transport=relay
```
This domain gets relayed to transport `relay` but its access class indicates that
`postfix` will apply the restrictions defined by the `slow` access rule.
```
example.com class=vmailbox transport=dovecot vuid=3000 vgid=3000 rclass=allow
```
is a complete, explicit definition with all properties set.

### Options
This command accepts no more than one argument. If the argument is present, only domains that match the argument are exported. No arguments implies all domains.

* `-o <domains file>` Redirect the standard output to the named domains file.

### Examples
Export all the domains to the file `domains.txt`. Both of these commands are equivalent.
```
[root@pobox ~]# postdove export domain > domains.txt
[root@pobox ~]# postdove export domain -o domains.txt
```
Export all domains ending in `.org`.
Note that we single quote `'` the argument to prevent the shell from expanding the `*`.
```
[root@pobox ~]# postdove export domain '*.org'
```


## Import
Import domains and their properties from a file. The file format is the same as for `export`.

Use the help option to show the command.
```
[root@pobox ~]# postdove import domain -h
Import a domains file from the file named by the -i flag (default stdin '-').

Usage:
  postdove import domain [flags]

Flags:
  -h, --help   help for domain

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -i, --input string    Input file in postfix/dovecot format (default "-")
  -v, --version         Report Postdove version and exit
```
### Options
The command accepts no arguments.
* `-i <domains file>` Redirect standard input to the named file.
### Examples
The following commands import the domains in the file `domains.txt`.
Both commands are equivalent.
```
[root@pobox ~]# postdove import domain < domains.txt
[root@pobox ~]# postdove import domain -i domains.txt
```


## Show
This command displays a domain and its properties to the standard output.

Use the help option to show the command.
```
[root@pobox ~]# postdove show domain -h
Show the contents of the named domain to the standard output
showing all its attributes

Usage:
  postdove show domain name [flags]

Flags:
  -h, --help   help for domain

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit
```
### Options
The command requires one argument.

There are no options.
### Examples
Display the domain `example.com` and its properties.
Note that this is a `vmailbox` which means it is used for receiving email.
The transport `dovecot` has been set up to forward emails via `lmtp` to `dovecot`.
```
[root@pobox ~]# postdove show domain example.com
Name:           example.com
Class:          vmailbox
Transport:      dovecot
UserID:         --
Group ID:       --
Restrictions:   --
```



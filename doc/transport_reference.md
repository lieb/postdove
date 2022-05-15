# Transport
The `transport` sub-command manages the transport options that `postfix` uses to process email
either into mailboxes or to forward the email to another service. Addresses and domains refer
to transports defined here to direct `postfix` decisions.

We will use an example transport named `dovecot`.
This is used in a typical `postdove` configuration to link `postfix` and `dovecot` without adding any hard-wired configuration into the `postfix`.

## Add
Add a transport definition to the database.
This must be done before *addresses* or *domains* can use it as a property.

See `transports(5)` in the `postfix` documentation for how the transports and
nexthops are used.

Use the help option to show the command.
```
[root@pobox ~]# postdove add transport -h
Add a named transport to the database with a transport matching transport(5) description.

Usage:
  postdove add transport name [flags]

Flags:
  -h, --help               help for transport
  -n, --nexthop string     Transport nexthop to send email
  -t, --transport string   Transport protocol/method

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit
```

### Options
The command expects one argument that is the name of the transport.

* `--transport=<string>` The string is the transport type defined in `postfix`.
* `--nexthop=<string>` The string is the nexthop field defined in `postfix`.

If either of these properties are not set, they are cleared which causes
`postfix` to use its internal defaults.

### Examples
Add a transport named `dovecot` for forwarding email to `dovecot` via the LMTP
protocol to the host `localhost` at its well known socket `24`.
```
[root@pobox ~]# postdove add transport dovecot -t lmtp -n localhost:24
```
We will use this example in the commands below.

## Delete
Delete a transport definition.
If the transport is used by any *address* or *domain*, the command will return
an error.

Use the help option to show the command.
```
[root@pobox ~]# postdove delete transport -h
Delete the named transport entry from the database so long as no domain or address references it.

Usage:
  postdove delete transport name [flags]

Flags:
  -h, --help   help for transport

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit
```

### Options
The command accepts one argument as the name of the transport to delete.

The command has no options.

### Examples
Delete transport `dovecot`. This will return a error if `dovecot` is in use.
```
[root@pobox ~]# postdove delete transport dovecot
```

## Edit
Edit a transport to change its *transport* or *nexthop* properties.
See `transports(5)` in the `postfix` documentation for the how `postfix`
uses a transport.

Use the help option to show the command.
```
[root@pobox ~]# postdove edit transport -h
Edit the transport and next hop attributes of the named transport.

Usage:
  postdove edit transport name [flags]

Flags:
  -h, --help               help for transport
  -n, --nexthop string     Transport nexthop to send email
  -N, --no-nexthop         Clear transport nexthop to send email
  -T, --no-transport       Clear transport protocol/method
  -t, --transport string   Transport protocol/method

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit
```

### Options
The command accepts a single argument for the name of the transport to be edited.

* `-t <transport>` Set the *transport* property to `transport`.
The transport type must be defined in the `master.cf` configuration
file in `postfix`.
* `-T` Clear the *transport* property. This will cause `postfix` to use its default
transport as defined in `main.cf` of the `postfix` configuration.
* `--nexthop <string>`Set the *nexthop* property to the string.
* `--no-nexthop` Clear the nexthop property.
This will cause `postfix` to use its default transport parameter for forwarding the
email.

### Examples
Change the transport `dovecot` to use a named pipe at `/var/dovecot/lmtp-in` instead  of `localhost:24` for forwarding email via `lmtp`.
```
[root@pobox ~]# postdove edit transport dovecot --nexthop=unix:/var/dovecot/lmtp-in
```

## Export
Export defined transports to the standard output.

Use the help option to show the command.
```
[root@pobox ~]# postdove export transport -h
Export transports to a file named by the -o flag (default stdout '-').

Usage:
  postdove export transport [flags]

Flags:
  -h, --help   help for transport

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -o, --output string   Output file in postfix/dovecot format (default "-")
  -v, --version         Report Postdove version and exit

```
### File Format
The transports file format is similar to the `transports(5)`
manual page from `postfix` with a change to what `postfix` refers to as a
*pattern*. Whereas *pattern* is for matching an address and/or domain,
in `postdove` this is a key, the transport's name, that is used in the
*address* and *domain* commands to set their *transport* properties.
The first token is the name of the transport.

The second token is the *transport* and the *nexthop* separated by a `:`.
For example:
```
dovecot lmtp:localhost:24
```
is the transport named `dovecot` that uses the `lmtp` transport protocol of `postfix`
to forward email to *nexthop* at the address `localhost` using socket `24`.
See `transports(5)` in the `postfix` documentation for all the options available
for transports.

### Options
The command accepts no more than one argument naming the transports
to be exported.
If the name contains a wildcard, `*`, all the transports that match with the wild
card are exported. Otherwise just the named transport is output.
If there is no argument, all transports are exported.

* `-o <output file>` Redirect standard output to the output file.

### Examples
Export all transports with names starting with `r` such as `relay` or `router`
to the standard output.
```
[root@pobox ~]# postdove export transport 'r*'
```
Export all transports to the file `transports.bak`.
```
[rooto@pobox ~]# postdove export transport --output=transports.bak
```

## Import
Import transport definitions from the standard input or a file.

Use the help option to show the command.
```
[root@pobox ~]# postdove import transport -h
Import transports from a file named by the -i flag (default stdin '-').

Usage:
  postdove import transport [flags]

Flags:
  -h, --help   help for transport

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -i, --input string    Input file in postfix/dovecot format (default "-")
  -v, --version         Report Postdove version and exit
```

### Options
The command requires no arguments

* `-i <input file>` Redirect the standard input to the input file.
### Examples
The following two commands are equivalent to import a transports file.
```
[root@pobox ~]# postdove import transport < transports.list
[root@pobox ~]# postdove import transport --input=transports.list
```

## Show
Display a transport and its *transport* and *nexthop* properties.

Use the help option to show the command.
```
[root@pobox ~]# postdove show transport -h
Display the contents of the named transport entry.

Usage:
  postdove show transport name [flags]

Flags:
  -h, --help   help for transport

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit
```

### Options
### Examples
Show the transport `dovecot`.
```
[root@pobox ~]# postdove show transport dovecot
Name:           dovecot
Transport:      lmtp
Nexthop:        localhost:24
```
This is the linkage between `postfix` and `dovecot` for mail delivery.
Every mailbox domain that has `dovecot` as its *transport* property
will use the *lmtp* transport of `postfix` to forward email to the
*localhost*  address, socket *24*, the lmtp well known IP socket.

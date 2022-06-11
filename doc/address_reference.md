# Address
The `address` sub-command manages email addresses and their transport and access properties.
An address is the fundamental unit of the email transport system.
It can be either sender or recipient.
Aliases are composed of addresses.
A local address does not have a domain part which is implied as either `localhost`
or the DNS hostname of the machine.

An address has properties depending on the context in which it is used.
All of the properties can be edited and displayed but if the address is not used
in the context of its use, the property has no effect.

## Add
Addresses can be added to the database as part of adding or editing other tables in
the database.
For example, adding a mailbox to the `dovecot` server will add the address for it.

An error will be returned if the address is already in the database.

Use the help option to show the command.
```
[root@pobox ~]# postdove add address -h
Add an address into the database. The name is either local (just a name with no domain part) or an RFC2822 format address.
The optional restriction class defines how postfix processes this address.

Usage:
  postdove add address name [flags]

Flags:
  -h, --help               help for address
  -r, --rclass string      Restriction class for this address
  -t, --transport string   Transport to be used for this address

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit

```

### Options
The command requires one argument, the address itself.

There are two options to set address properties. If an option is not set, there is no default value
for the property.

* `--rclass` sets the restriction class. This would be the name of the access rule to apply.
* `--transport` sets the transport. This would be the name of the transport to be used for this address.

### Examples
In the first example we add an address `gramma@cottage` but set no properties for it.
```
[root@pobox ~]# postdove add address gramma@cottage
```

Add an address `bill@example.com` but set the restriction class to `dump` so that email would be rejected.
```
[root@pobox ~]# postdove add address bill@example.com -r dump
```

Add an address `info@example.com` that gets relayed to `moon` and uses the class `allow`.
```
[root@pobox ~]# postdove add address info@example.com -t moon -r allow
```

## Delete
Delete the named address.

If there are references to this address by other tables such as the mailboxes or aliases,
the command will return an error.

Use the help option to show the command.
```
[root@pobox ~]# postdove delete address -h
Delete an address address from the database.

Usage:
  postdove delete address  [flags]

Flags:
  -h, --help   help for address

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit

```

### Options
The name of the address to be deleted is the one required argument.

There are no options.

### Examples
Delete `info@example.com` from the database
```
[root@pobox ~]# postdove delete address info@example.com
```

## Edit
Edit the properties of an address.

Use the help option to show the command.
```
[root@pobox ~]# postdove edit address -h
Edit a address and its attributes.

Usage:
  postdove edit address name [flags]

Flags:
  -h, --help               help for address
  -R, --no-rclass          Clear restriction class for this address
  -T, --no-transport       Clear transport used by this address
  -r, --rclass string      Restriction class for this address
  -t, --transport string   Transport to be used for this address

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit

```

### Options
The address to be edited is the required argument.

* `--rclass=<access_rule>` Set the restriction class for this address to the named access rule.
* `--no-rclass` Clear the restriction class property for this address.
If the domain of this address has a restriction class property , the restriction class for the address
will default to the domain's class.
If the domain does not have a restriction class, the address no longer has any restriction class.
* `--transport=<transport name>` Set the transport for this address to the named transport.
* `--no-transport` Clear the transport property for this address.
If the domain for this address has a transport specified, this address inherits the domain's transport.
if the domain does not have a transport, the address no longer has one.


### Examples
Set the restriction class for `info@example.com` to `dump`.
```
[root@pobox ~]# postdove edit address info@example.com --rclass=dump
```

Set the transport for `info@example.com` to `inside_relay` and clear the restriction class.
```
[root@pobox ~]# postdove edit info@example.com -R -t inside_relay
```

## Export
Export the named addresses.
The export format is an extension to `postfix` hash file formats.
The following is an example of the format. Note the comments in the file are accepted by an
import but are not output by this command.
```
# addresses for address import testing

mary@little.lamb  rclass=DUMP# Need to fix this special case
wolf@forest rclass=STALL
gramma@cottage transport=relay
dave@wm.com
```
`dave@wm.com` has no properties but `gramma@cottage` has a transport property of `relay`.

Use the help option to show the command.
```
[root@pobox ~]# postdove export address -h
Export addresses to the file named by the -o flag (default stdout '-').

Usage:
  postdove export address [flags]

Flags:
  -h, --help   help for address

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -o, --output string   Output file in postfix/dovecot format (default "-")
```
### File Format
The file format has no similarity to file definitions in `postfix`.
This is because the hash files for `access(5)` and `transports(5)` use
an address or domain as their key.
The equivalent within `postdove` is to consolidate *access* and *transport*
entries into properties for addresses and domains.

The format for the line defining an address is:
```
address rclass=<string> transport=<string>
```
where `address` is of the form `user` or `user@domain`,
the `rclass` string is the name of the access rule, and
the `transport` string is the name of the transport.
If either of these properties are cleared, the property will not appear.
For example:

```
dave@example.com
```
has no properties set. If either property is set for `example.com`,
queries by `postfix` will return the domain's values.
```
mary@example.com rclass=allow
```
The `rclass` property is set to the `allow` access rule which will be
returned by `postfix` restriction class queries.
The transport property is processed in the same way as for `dave@example.com`.
```
charlie@example.com rclass=slow transport=relay
```
Both properties are set and therefore returned by queries.

### Options
If no argument is supplied, the export will include all addresses.
A subset of the addresses can be exported by using a wildcard in the address argument.

* `-o <output file>` redirect the standard output to the named file.

### Examples
Export all addresses.
```
[root@pobox ~]# postdove export address
```
Export the address `info@example.com`. This will export just one address to the file `info.txt`.
```
[root@pobox ~]# postdove export address info@example.com > ./info.txt
```
Export all the addresses in `example.com` to the file `addresses.txt`.
```
[root@pobox ~]# postdove export address -o addresses.txt *@example.com
```
Export all addresses `info` in any `.com` domain.
```
[root@pobox ~]# postdove export address info@*.com
```
These two are equivalent.
```
[root@pobox ~]# postdove export address *@*
[root@pobox ~]# postdove export address
```

## Import
Import addresses and their properties from a file.

The import will fail if any of the properties are not already set by importing or adding the property first.
Best practice is to import access rules and transports before importing anything else
to prevent this error.

Use the help option to show the command.
```
[root@pobox ~]# postdove import address -h
Import an addresses file from the file named by the -i flag (default stdin '-').

Usage:
  postdove import address [flags]

Flags:
  -h, --help   help for address

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -i, --input string    Input file in postfix/dovecot format (default "-")
  -v, --version         Report Postdove version and exit
```
### Options
There are no arguments.

* `-i <import file>` Redirect the standard input to the named file.

### Examples
Import addresses from the standard input.
```
[root@pobox ~]# cat addresses.txt | postdove import addresses
```

## Show
Show the details of an address and its properties to the standard output.

Use the help option to show the command.
```
[root@pobox ~]# postdove show address -h
Show the contents of the named address to the standard output
showing all its attributes

Usage:
  postdove show address name [flags]

Flags:
  -h, --help   help for address

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -v, --version         Report Postdove version and exit

```
### Options
An address argument is required. No wildcards, i.e. multiple addresses are recognized.

There are no options.
### Examples
Display `test@example.com` with its properties. The `--` indicated there are no properties set.

```
[root@pobox ~]# postdove show address test@example.com
Address:        test@example.com
Transport:      --
Restrictions:   --
```

Display `info@example.com`. It has a transport of `relay` but no restriction class.
```
[root@pobox ~]# postdove show address info@example.com
Address:        info@example.com
Transport:      relay
Restrictions:   --
```

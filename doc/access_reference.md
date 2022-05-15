# Access Control Management
The `access` sub-command manages the codes used to control the `postfix` restriction classes.
These classes in `postfix` determine the validation tests available in `postfix` to check
incoming email for validity, spam, and processing options. Usage of this feature
is for making email scan checks more or less stringent for either everyone in a domain or for specific
users.
For example, you may want to filter out all spam and phishing emails from Grandma but relax
checks for your honeypot user.

The classes are defined in the `postfix` configuration and attached to domains and addresses
by way of the access types defined here. There is a commented example of this use in the
`config/main.cf.diff` file. The first step would be to organize the labeled restriction classes
in `main.cf` and then populate this table with the restriction label. The name is used in
other commands to establish the link.

Structuring `postfix` access rules/constraints separate from the users they apply to is
useful because on the `postfix` side the filtering rules and their ordering sometimes need to
change. A labeled class can simply be modified without having to make a change to all of the
user accounts that use it. On the `dovecot` side, users can easily be added to a rule class
without messing with `postfix` internals. For example, there could be a case where a class
must be made more restrictive but only for some users. A new class could be created that
includes the original class and adds some more filters. A new access rule can then be
added in `postdove`.
Each of the users that needs the new rule could then be edited to use the new
rule. If the new rule gets further enhancements,
nothing further needs to be done on the `dovecot` side for those users.

Access rules are applied to either addresses or domains. Results are returned to `postfix`
in the following order:

1. If a rule is associated with `user@domain`, that rule is returned as the result.
2. If there is no rule associated with `user@domain` but there is one associated with `domain`, that rule is returned.
3. If there is no rule for either `user@domain` or `domain`, no result is returned to `postfix` which would cause `postfix` to skip to the next restriction in its list.

## Add
Add an access restriction.
```
[root@pobox ~] # postdove add access -h
Add the named access rule to the database. The the value is the key postfix uses to
select a set of recipient restrictions.

Usage:
  postdove add access name restriction [flags]

Flags:
  -h, --help   help for access

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
```

### Options
There are no sub-command options. The `name` and `restriction` argument strings are required.

### Examples
Create an access rule to be used for permitting (accepting) a message.
```
[root@pobox ~] # postdove add access permit x-allow
```
The name `permit` can be used in other commands to link an address or domain to the access
rule key `x-allow`. The `x-allow` will be used by `postfix` to select a set of filter/access
rules.

The regular process is to add access rules here first and then associate them later with
specific addresses or domains. 

## Delete
Delete an access restriction. If the restriction is in use by a domain or address, the command
will return a database error. All domains or addresses that refer to the access rule would
have to be edited to remove the reference before the rule can be deleted.

```
[root@pobox ~] # ./postdove delete access -h
Delete the named rule from the database so long as no address or domain references it.

Usage:
  postdove delete access name [flags]

Flags:
  -h, --help   help for access

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
```
### Options
There are no sub-command specific options but the `name` argument must be supplied.
### Examples
Delete and unused or unreferenced access rule
```
[root@pobox ~] # postdove delete access permit
```
The `permit` rule is deleted if it is not referenced.

## Edit
Edit the named access rule to change the action key.
```
[root@pobox ~] # postdove edit access -h
Edit the named access rule to change the postfix restriction class key.

Usage:
  postdove edit access name [flags]

Flags:
  -r, --action string   Access rule action value used by Postfix to process client access restrictions
  -h, --help            help for access

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
```

### Options
* `--action=<restriction>` edits the access rule to change the restriction to this new restriction tag.

### Examples
Assuming the label of the restriction class was changed in `main.cf`, edit the action here to 
match the new label.

```
[root@pobox ~] # postdove edit access permit --action=x-gofree
```
Note that if `x-gogree` is not defined as a class in `postfix`, any errors would be reported by `postfix`.
There are no checks in `postdove` because it has no access to the `postfix` configuration to do any checks.

## Export
Export the access table in a format similar to what `postfix` would use in its own databases.
What we mean by *similar* is that `postfix` uses a form for each rule as:

```
user@one.domain reject
another.domain permit
```
Note that this uses an address as the key and the next word as the value.
In a `postfix` only configuration, there would have to be an entry (line) for each address and/or domain.
This can get tedious really quickly. It would also be messy for `postdove` to make links between addresses
and rules. To handle access rules, `postdove` uses names for the rules and references them by name.
The result is that an exported format has the following form:
```
block reject
allow permit
```
Where `block` would be the name used as the access property of `user@one.domain` and `allow` would be
the name used for `another.domain`. Note that multiple domains and users can use a rule.

Here is the help display for the command:
```
[root@pobox ~] postdove export access -h
Export access rules into the file named by the -o flag (default stdout '-'

Usage:
  postdove export access [flags]

Flags:
  -h, --help   help for access

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -o, --output string   Output file in postfix/dovecot format (default "-")
  -v, --version         Report Postdove version and exit

```
### File Format
The file format has no equivalent in either `postfix` or `dovecot`.
Each access rule is a rule name followed by a value string that is returned
when `postfix` queries an address or domain for a restriction class.
The form is:
```
name key_value
```
where `name` is used to set the `rclass` property of an address or domain and
the `key_value` is what is returned on a `postfix` query.
The `key_value` must match the class definition in `main.cf`.

For example:
```
allow x-permit
```
The name `allow` is what is set as the `rclass` value and `x-permit` is
the restriction class name defined in `postfix`'s `main.cf` configuration file.

### Options
The command can take an optional argument.
If no rule argument is supplied on the command line, all rules are exported.
However, a specific set of rules can be exported based on the rule argument.
See the examples below.

* `-o <export_file>` redirects the export output from the standard output to the named file.

### Examples
Export all of the rules in the database with the following commands:
```
[root@pobox ~] postdove export access
```
This will export all the rules to the standard output.
Both of the following commands will export them to the file `exports.txt`.

```
[root@pobox ~] postdove export access > exports.txt
[root@pobox ~] postdove export access -o exports.txt
```

The command:
```
[root@pobox ~] postdove export access block
```
will only export the rule `block`.

Wildcards can also be used:
```
[root@pobox ~] postdove export access al*
```
will export all rules starting with `al` such as `allow` or `almost`.

## Import
Import is the inverse of `export`.
It is used to enter rules in bulk using the same file format as `export`.

Here is the help display for the command:
```
[root@pobox ~] postdove import access -h
Import access rules file from the file named by the -i flag (default stdin '-').

Usage:
  postdove import access [flags]

Flags:
  -h, --help   help for access

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
  -i, --input string    Input file in postfix/dovecot format (default "-")
  -v, --version         Report Postdove version and exit

```

### Options
There are no arguments.

* `-i <import_file>` The input for the command is redirected from the standard input to the named file

### Examples
The following commands accept the import of rules from the file `access.txt`.
```
[root@pobox ~] postdove import access < access.txt
[root@pobox ~] postdove import access -i access.txt
```

## Show
This command displays the contents of the named access rule.

Here is the help display for the command:
```
[root@pobox ~] # postdove show access -h
Display the contents of the named access rule to standard output.

Usage:
  postdove show access name [flags]

Flags:
  -h, --help   help for access

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")

```

### Options
The `name` argument is required.

There are no options.

### Examples
Display the contents of the access rule named `dump`.

```
[root@pobox ~] # postdove show access dump
Name:   dump
Action: x-reject

```

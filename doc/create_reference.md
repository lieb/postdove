# Create a database
The `create` command is the only command without sub-commands. The rest of them are specific to
a particular data item. It is also the first command that must be run given that all the others
depend on an existing database.

A **Sqlite** database is an regular, binary file in the filesystem that is accessed through
a library integrated into the utility. The `create` command initializes a new database file
with both the supported schema and some initial table contents. Starting with an new database file
is recommended because a new version of the utility may have a different schema which could cause
the schema execution to fail.

Again, use the help feature to show the command:
```
[root@pobox ~]# postdove help create
Create the Sqlite database file and initilize its tables.
You will also have to do some imports and adds to this otherwise empty database.

Usage:
  postdove create [flags]

Flags:
  -a, --alias string    RFC 2142 required aliases (default is built in")
  -h, --help            help for create
  -l, --local string    default local domains (localhost, localhost.localdomain
  -A, --no-aliases      Do not load RFC 2142 aliases
  -L, --no-locals       Do not load local domain hosts
  -s, --schema string   Schema file to define tables of database. Default is built in.

Global Flags:
  -d, --dbfile string   Sqlite3 database file (default "/etc/postfix/private/postdove.sqlite")
```

## Options
These options are local to the `create` command.

* `--alias=<file>` import **aliases** from a file. The RFC 2142 aliases are the common
aliases recommended for an Internet accessible email system. If this option is not set,
the built in aliases will be imported

 * `--no-aliases` will skip the loading of any aliases

 * `--local=<file>` import initial set of local domains, typically `localhost` and
    `localhost.localdomain`. The default set of domains is built in.

 * `--no-locals` will skip the loading of any local domains

 * `--schema=<file>` will select an alternate schema to load from the named file.
    Otherwise, if this option is not set, the command will used the built in schema.

Most usages should use the built in schema because the application logic of the
utility expects it, especially the defined triggers and views. Using an alternate
schema should be used with care.

**CAUTION**

	Use this command with care. If you apply it to a running database, it will
	drop all data and leave and initialized empty database. NEVER use this on an active
	system. If you must upgrade the database, first **export** all the the tables.
	
## Examples
Create a new database populated with aliases and local domains.
```
[root@pobox ~] # postdove create
```

Create a new, completely empty database with an alternate schema and no aliases or domains.
```
[root@pobox ~] # postdove create -s newschema.sql --no-aliases --no-locals
```

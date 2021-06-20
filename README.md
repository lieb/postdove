# postdove
A utility to manage the users and aliases on a mailserver comprised of `postfix` and `dovecot`.

> **NOTE:** This is a work in progress. The main functions are code complete and unit
tested but this will not be tagged until it has been tested in service.
***
The data used by both servers is stored in a `sqlite` database that is used at runtime by both servers.
`postdove` manages the data so that there is always a consistent view so changes can be made
without requiring either server to reload or restart.

## Building Postdove
`postdove` is an application for managing the database. It operates from the command line to add, delete,
and edit individual records. It also has import and export commands that use the file formats native to
both `postfix` and `dovecot` to make setup easier. It is built and installed following the instructions
in [Building and Installation](./doc/building.md).

## Configuration
In order to use the database both `dovecot` and `postfix` need configured.
This is done by manually editing their configuration files in the `/etc` directory
following the instructions in [Configuration](./doc/configure.md).

## Operation
Once we have all the parts up and running, we have to manage it.
Here is the [Administrator Guide](./doc/admin.md)
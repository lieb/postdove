# Postdove Administration

Nearly all email system administration is done via `postdove` command.
The `postfix` and `dovecot` configuration files are rarely changed after the installation is complete.
Consult the `postdove` [Commands Reference](commands_reference.md) for the details of all
the commands needed to administer the mail server.

## Customize the Installation
The database creation gave us a basic system.
It added the domains `localhost` and
`localhost.localdomain`.
It also added the RFC 2142 aliases, i.e. `postmaster`,
`abuse`, and their friends. For most installations these are effectively constants,
needed in a correct installation but rarely changed.

We will populate the database in the following order:

1. The first thing we need to do is add the glue between `postfix` and `dovecot`.
We do this by defining a transport for LMTP
```
[root@pobox ~]# postdove add transport dovecot --transport=lmtp --nexthop=lmtp:localhost:24
```

2. Add host names for this system. These will be used along with `localhost`
by postfix to know where to route incoming email. We also must change
'localhost.localdomain` to something sensible. Since we cannot edit the domain
name, we add the new one and delete the old.
```
[root@pobox ~]# postdove add domain mail.home.example.com --class=local
[root@pobox ~]# postdove add domain pobox.home.example.com --class=local
[root@pobox ~]# postdove add domain localhost.example.com --class=local
[root@pobox ~]# postdove delete domain localhost.localdomain
```

3. Add the domain name used for the virtual users. This type of domain
must be created before any users in that domain, i.e. the domain does
not automatically get added when a virtual mail user is created.
```
[root@pobox ~]# postdove add domain example.com --class=vmailbox --transport=dovecot
```

This sets up a basics for an installation. There are lots more to do in order
to get something useful.

## Add User Accounts
We can now add users. Note that this just adds the user to the database.
Other actions must be done before the account is usable for mail. This is
enough for `dovecot` to start serving mail.
If we are moving from another email system, an easy route to bulk load accounts
would be to dump the existing accounts in a format that matches the `dovecot` account
table. We could then `import` that into the database.

## Add Aliases
Add aliases. These will be used by `postfix` to process and deliver
email to `dovecot`. There are two types of aliases, `alias` and `virtual`.
The easiest way to enter them is to import using the file format that
`postfix` uses.

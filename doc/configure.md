# Configuration and Testing
Both *Postfix* and *Dovecot* use the same `Sqlite3` database for operation.
We have to make it available to both but deny access to anyone else and we
do this via the `mail` group which is set by the `postfix` package install.
To do this we add `dovecot` to the `mail` group.

> **Note:** This is the default setup in **Fedora**. This installation is current as of *Fedora 35*.
Debian based systems may have files in different locations.
This also assumes that `selinux` is enabled and it requires no additional security setting except as noted.
Users of `AppArmor`, such as **Ubuntu** will have to take this into consideration.
The text points out those places where MAC security controls are relevant.

>  I use BTRFS everywhere for local filesystems. It is kind to my SSDs and has properly integrated snapshots and subvolumes.
Yes, I know. Hysterics and ancient pronouncements (increasingly outdated but difficult to retract) say it is not ready for production
but I have found it to be solid for years and as close as I can get to the AdvFS I knew and loved on my many
OSF/1(Digital UNIX) based DEC Alphas all those years ago. BTW, tell Facebook how much better ZFS is ... and
see what happens. Except for some details in System Configuration, there are no BTRFS dependencies.

## System Configuration
My mail service runs as a VM hosted by my file server.
It is a "black box" system in that email users (actually any user other than an
administrator) does not have a login account.
User IDs for email storage do match the IDs assigned to other systems but that
is just a matter of convenience and a way to ensure that there are not ID overlaps.

System configuration involves steps on both the hosting server and
in the VM running my mail services. See [System Configuration](system_configuration.md) for the details.

## Database Creation
The `sqlite3` database is located in the `/etc/postfix/private` configuration directory and shared with `dovecot`.
The [Database Creation](database_setup.md) instructions set this up.

## Dovecot Configuration
Once the systems and the database are set up, `dovecot` is configured to access all these components.
See [Dovecot Configuration](dovecot_configuration.md) for the details once all the system "stuff" is complete.

## Postfix Configuration
The final big piece is `postfix`.

There are two general configurations to consider for setup.
The first is a public facing SMTP service.
This is the classic service that `postfix` and its older brother `sendmail`
were designed for.
In this configuration `postfix` is receiving email via inbound connections from the Internet
and forwarding outbound email directly to its destination on the net.
IMAP or POP3 service for email users is provided by `dovecot` with `postdove`
managing the details.
The `dovecot` services could either be private to the local network or made
accessible to the net as well as local users.
This option is getting less practical for a small business or personal email
service as spam bots and email based security attacks require ongoing
email administration work.

The second configuration option would be for a local network private installation.
In this case, neither `postfix` nor `dovecot` presents an SMTP/IMAP/POP3
service to bandit rich environment of the open Internet. Everything is private.
This could be an option for users/sites whose primary landing place for public
email is a managed email service provided by an ISP.
Users would configure the ISP managed email service as their primary
sending and receiving servers and configure a second receiving service that
points to the local `dovecot` service which simply becomes the
local IMAP email store.

That is not to say that a local SMTP doesn't have its uses. Although my server does not forward any email outside, it is
still useful for routing system management emails from the various "things" on my home network and combined with `fetchmail`
it is handy for feeding downloaded email from my ISP into LMTP and *Sieve* to pre-sort all the junk.

[Postfix Configuration](postfix_configuration.md) has all the gory details.

## Testing
Once both services are configured, it is time to test it.
There are a lot of moving parts to the system so system test will probably be
an iterative affair as typos are fixed, bad assumptions about configuration are disabused, and
database entries are changed to reflect reality.
In general, testing proceeds as follows with some rinse-repeat loops at various stages.

Both `postfix` and `dovecot` have been around for a while and there is a lot of online
documentation and tutorials directed towards installing and managing email systems based
on these and other tools.
Testing what we have configured assumes that the reader is experienced with system
and, in particular, email systems administration.
Our description here only focuses on the specifics of testing the integration of
`postfix`, `dovecot`, and `postdove`.

*Systemd* has gotten somewhat ubiquitous in Linux installations so we assume it here.
I will not show the individual command lines here because various distributions have
different names for the service files. **Fedora** also has `cockpit` which is a nice
web based admin tool that has a whole *services* page.
Both `postfix` and `dovecot` have mature integration with `systemd`.
Consult the system documentation for how to start/stop/enable these services.
We do not consider the legacy "Init scripts" system.

You will look for errors in two places, errors reported by your email client and errors
reported in the system logs of the email server.
All email server issues are fixed on the email server.
Ignore the fact that your mail store is on your host. You can see email files there but there
is "nothing to see here" for anything else.

In a **Fedora** installation, early testing will involve getting *Selinux* permissions correct.
Expect that this will be mostly a `dovecot` issue since `postfix` as configured can run
out-of-the-box.
All of the *Selinux* issues involve either the database file itself or, depending on how
you configure mail storage, access to the `/srv/dovecot` directory by email clients.
See [Database Creation](database_setup.md) for the details.
Testing starts with bringing up the newly configured service and inspecting the logs for errors.
Only enable the services once testing proves that they are functional together.

### Bring up `dovecot` 
This is the simpler of the services.
Besides the usual issues of typos in configuration files and missing bits,
you will be verifying three things:
* Access controls to the database file and the mailstore directory.
The service works when `dovecot` can access both without *Selinux* complaining.
* Check to see that `dovecot` is running all the services you are setting up
and it is listening on all the sockets you configured.
Especially check that the address is not `localhost`, otherwise it is only listening
to itself.
* Verify that a login session works by using `postdove` add a test user to the domain you configured,
i.e. `test@example.com`.
Remember, `postdove` only creates a user account in the database and
`dovecot` does nothing for an account until the user actually logs in for the first time.
Using your favorite mail reader, create an account for receiving,
using the password you configured for `test@example.com`.
Attempt to connect and watch what happens.
When all is done, what you should see on your client is its display of the service and *INBOX*.
On the server you should see a new directory `/srv/dovecot/example.com/test` with
some subdirectories and files that should be in the familiar *Maildir* format.
Check the ownerships and modes of the directories.
They should all have *read-write-execute* for the *owner* and no access for either *group* or *other*, i.e. `rwx------`.
They should also be owned by the user ID and group ID set for user `test`.
Note that if your mail service is a "black box", there will only be administrator accounts on it
and these IDs will be displayed as numbers, not names because `/etc/passwd` knows
nothing about them.

### Bring up `postfix`
There are lots of tutorials on `postfix` so this will not go into details beyond testing that
the queries configured in fact work.
The goal here is to inject email into the server and see it arrives in a mailbox.
A good tool for this is `msmtp` which is a stripped down SMTP client and server which is mainly
a client tool that simply forwards all email to a "real" SMTP server, namely our `postfix`
installation.
I also use it in production to forward all system emails such as the output of `logwatch` on
all my systems to my main service.
It can be configured with both a set of aliases and a forwarding, really a "relay" SMTP.
There are two basic tests:
* Send an email to `test@example.com`, the account you just set up with `postdove`.
This will test two paths. The first is whether the transport for `example.com` is set up
properly so the `dovecot` transport is used.
Otherwise, the message will bounce.
* Create an alias and virtual alias in `postdove` pointing to `test@example.com` and see
if alias and virtual alias queries are working.

Once this all works, it is a running system.
From now on, administration is done by `postdove`.
There is enough here to flesh out a running system.

**NOTE** We have not explicitly tested things like relays or access controls.
There could still be errors lurking in these places, primarily in `postfix`
configuration.
All of the functions of `postdove` are unit tested as part of the build and all
the views used by `postfix` queries are tested as part of the top-level
system test.
This means that the database will return appropriate results but
whether they work or not is a combination of `postfix` configuration and
correct data entered into the database relative to one's specific system
and network requirements. This is where Garbage In, Garbage Out applies.


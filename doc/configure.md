# Configuration
Both *Postfix* and *Dovecot* use the same `Sqlite3` database for operation.
We have to make it available to both but deny access to anyone else and
do this via the `mail` group which is set by the `postfix` package install.
To do this we add `dovecot` to the `mail` group.

> **Note:** This is the default setup in **Fedora**. This installation is current as of *Fedora 34*.
Debian based systems may have files in different locations.
This also assumes that `selinux` is enabled and it requires no additional security setting except as noted.
Users of `AppArmor`, such as **Ubuntu** will have to take this into consideration.
The text points out those places where MAC security controls are relevant.

>  I use BTRFS everywhere for local filesystems. It is kind to my SSDs and has properly integrated snapshots and subvolumes.
Yes, I know. Hysterics and ancient pronouncements (increasingly outdated but difficult to retract) say it is not ready for production
but I have found it to be solid for years and as close as I can get to the AdvFS I knew and loved on my many
OSF/1(Digital UNIX) based DEC Alphas all those years ago. BTW, tell Facebook how much better ZFS is ... and
see what happens. Except for some details in System Configuration, there is no BTRFS dependencies.

## System Configuration
My mail service runs as a VM hosted by my file server. This involves configuration steps on both the hosting file server and
in the VM running my mail services. See [System Configuration](system_configuration.md) for the details.

## Database Creation
The `sqlite3` database is located in the `/etc/dovecot` configuration directory and shared with `postfix`.
The [Database Creation](database_setup.md) instructions set this up.

## Dovecot Configuration
Once the systems and the database are set up, `dovecot` must be configured to access all these components.
See [Dovecot Configuration](dovecot_configuration.md) for the details once all the system "stuff" is complete.

## Postfix Configuration
The final big piece is `postfix`. This could be an option for users/sites whose primary landing place for public
email is an ISP. In that case, `dovecot` can simply be a local IMAP email store. With all of the security and email
access controls to fight off spam bots etc. A small business or personal email service directly connected to the
Net is getting less practical.

That is not to say that a local SMTP doesn't have its uses. Although my server does not forward any email outside, it is
still useful for routing system management emails from the various "things" on my home network and combined with `fetchmail`
it is handy for feeding downloaded email from my ISP into LMTP and *Sieve* to pre-sort all the junk.
[Postfix Configuration](postfix_configuration.md) has all the gory details.
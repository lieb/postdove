# Postfix Configuration

## Install Packages
Install postfix and the sqlite support packages. This is the base install.

```
# dnf install -y postfix postfix-sqlite
```

### SPF Policy Server
In addition to the basic postfix, we need to install spam filtering
if we are to keep the mailboxes free of dangerous dirt and litter.

Install the SPF policy server for Postfix

```
# dnf install pypolicyd-spf
```
### Clamav and Amavisd
Install the spam filters and their services.
This is the set used by **Fedora**.

```
# dnf install -y clamav clamav-update amavis clamd perl-Digest-SHA1 perl-IO-stringy
```
### Bad Guy Filtering Setup
Add the following lines to end of `master.cf` to integrate both packages into Postfix.
The first filter is Sender Policy Filter (SPF) which does its work when the sending server connects.
It checks the sender against an online database (actually a special purpose DNS server) to see
if the sender has been identified as a bad guy.

```
# SPF Policy checker
policyd-spf unix -     n       n       -       0       spawn
       user=nobody argv=/usr/libexec/postfix/policyd-spf
```
The companion piece to wire this into `postfix` is in `main.cf`.
See below for where this linkage is placed in `master.cf`.

The next filter actually filters the content of an email looking for bad stuff.
This is a more complex setup requiring a service to be periodically run which downloads
things like virus information databases.

Follow the instructions in these web pages to install the policy and filtering
servers. This also includes the wiring to integrate with Postfix.
That process is outside our scope but a good description of the process can be found here:
```
https://www.server-world.info/en/note?os=Fedora_35&p=clamav
https://www.server-world.info/en/note?os=Fedora_35&p=mail&f=7
```

The last filter that is usually set up for a internet connected mail server is
a Domain Keys Identified Mail (DKIM) service.
The package for this in Fedora is `opendkim`.
This service signs outgoing email with a private key and verifies inbound email
with the matching public key that has been stored as a DNS record in the sender's
domain zone file.
In my installation, this is done at my ISP since that is my ingress/egress point for
SMTP traffic.
I mention it here as another important filtering system but it too is outside scope.

These two additions end up being the following additions to `master.cf`.
```
[root@pobox postfix]# diff -u master.cf.orig master.cf
--- master.cf.orig      2022-01-20 15:45:15.000000000 -0800
+++ master.cf   2022-05-04 15:43:36.186315306 -0700
@@ -135,3 +135,29 @@
 #mailman   unix  -       n       n       -       -       pipe
 #  flags=FRX user=list argv=/usr/lib/mailman/bin/postfix-to-mailman.py
 #  ${nexthop} ${user}
+
+# SPF Policy checker
+policyd-spf unix -     n       n       -       0       spawn
+       user=nobody argv=/usr/libexec/postfix/policyd-spf
+
+
+# amavis milter
+smtp-amavis unix -    -    n    -    2 smtp
+    -o smtp_data_done_timeout=1200
+    -o smtp_send_xforward_command=yes
+    -o disable_dns_lookups=yes
+127.0.0.1:10025 inet n    -    n    -    - smtpd
+    -o content_filter=
+    -o local_recipient_maps=
+    -o relay_recipient_maps=
+    -o smtpd_restriction_classes=
+    -o smtpd_client_restrictions=
+    -o smtpd_helo_restrictions=
+    -o smtpd_sender_restrictions=
+    -o smtpd_recipient_restrictions=permit_mynetworks,reject
+    -o mynetworks=127.0.0.0/8
+    -o strict_rfc821_envelopes=yes
+    -o smtpd_error_sleep_time=0
+    -o smtpd_soft_error_limit=1001
+    -o smtpd_hard_error_limit=1000
+
```
Note these are at the end of the file.
The first chunk is the SPF filter.
This is the test during connection time to see if the sender is a bad guy before `postfix` even
thinks about letting the data in.
All of this is important but outside the scope of configuring `postfix` to work
with `postdove`.
It is very good and important to have but not fully relevant
to what this project is doing.

The second chunk is for filtering the data of the message.
It is sent to filter via a named pipe and returned back into the stream via socket `10025`.
This too is very important but parallel to what this project is doing.

It isn't until all this processing is done on a message that the controls
managed by `postdove`take effect.

### Query Setup
Email routing and filtering in `postfix` requires databases of domains and addresses.
In a conventional installation these use *hash* files, sets of key/value records.
This is what is documented in the `postfix` documentation.
One of the alternatives for *hash* files are SQL database queries.
This is what `postdove` manages.

The configuration comes in two parts. First, `postfix` itself is configured to use *Sqlite* queries for
its various lookups. Second, the database needs a number of additional records added so that `postfix` can
use these records to process the email.

The linkage to queries are in two places in the configuration.
Queries and their parameters are defined in various places in `main.cf`.
Anywhere a `hash:xxxx` parameter is valid, a `postdove` database query can be used.
Some queries can use comma separated items that are processed left to right in the parameter.
For example, a `maps-alias.query` may be preceded or followed by another
source for alias maps.

The query files are located in the `./config/postfix` directory in the repository.
They are copied to the server's `/etc/postfix/query` directory.
A copy would look something like:
```
[lieb@nighthawk postdove]$ cd config/postfix
[lieb@nighthawk postfix]$ scp *.query root@pobox:/etc/postfix/query
```
Note that the `/etc/postfix/query` directory must be created first since the package
installation knows nothing about it.

A query parameter looks like:
```
some_postfix_parameter = sqlite:$config_directory/some.query
```

Queries themselves are defined in separate files located in `/etc/postfix/query`. 
This is the actual SQL query itself.
These are all defined and do not need to be changed.
The query file, in this case, an alias lookup is:
```
[root@pobox postfix]# cat query/alias_maps.query
# local aliases

# open sqlite with foreign keys enabled to match postdove

dbpath = /etc/postfix/private/postdove.sqlite

query = SELECT recipient FROM etc_aliases WHERE local_user = '%u'
```
There are two things to note.
The query always returns a single result column, in this case `recipient`.
This matches the *key: value* pairs in *hash* queries.
There could be multiple results and if this were a *hash* entry,
it would look like: `key: value1, value2, ...`.
In an SQL query, there may also be multiple results returned but they would be returned
as multiple rows, each result `value` in a single row.
Some queries are just looking for the *key* so the actual result does not matter.
In other words, the query is really *Is it there?*.
If the result "is not there", no results are returned and `postfix` moves on to the
next step.

The second part is the `%u` in this query.
There are a set of these substitution parameters defined in `postfix`.
These are the *key* values corresponding to what a *hash* query would expect.
All of this magic is already set up in the supplied queries.

**NOTE:** The comment about foreign keys is not really correct.
For backwards compatibility, the *Sqlite* library itself does not enforce foreign key constraints unless the open of the database file specifies it.
The `postdove` utility does open the database with that *PRAGMA* enabled so database changes
are clean and proper but attempting the same thing in `postfix` does not work.
This is really not a problem because `postfix` only does queries and the library
applies a lock on the database file around queries.

One other point on queries.
Every query done by `postfix` references a database *View*.
We won't go into the details here so if one wants to know what an SQL view is, consult
the *Sqlite* documentation or any book on SQL.
The important thing is that the *view* is internal to the database library and schema.
It completely encapsulates all the query complexity into a simple `SELECT`
query returning one result.
An examination of some of the view definitions in the schema show just how messy things can be.

### The Details in `main.cf`
We make changes to the `main.cf` configuration file to manage everything else
after `master.cf` is modified to link routing of messages among new `postfix` processes.

These changes are presented here as chunks of a `diff -u` between the original
file and our working configuration.
We start at the top and go all the way through the changes one at a time.
Some changes are just basic `postfix` changes needed to make the package useful in a
configuration.
The rest, the interesting bits for `postdove` operation are the queries.

#### Query Setup
First off, we set up for queries.
```
--- main.cf.orig        2022-01-20 15:46:33.000000000 -0800
+++ main.cf     2022-05-12 11:33:47.097353337 -0700
@@ -66,6 +66,9 @@
 #
 data_directory = /var/lib/postfix
 
+# The query parameter sets up Sqlite database access
+query = sqlite:$config_directory/query
+
 # QUEUE AND PROCESS OWNERSHIP
 #
 # The mail_owner parameter specifies the owner of the Postfix queue
```
This is little more than a *macro* to keep the changes below simple.
The `$config_directory` is the default `postfix` setting `/etc/postfix`.
This is how `postfix` finds the query files copied into the system (See above).

#### When Queries Do Not Apply or Work
The next parameter, `myhostname` would be nice if it could be queried.
Unfortunately, this makes `postfix` very unhappy because it really wants a constant...
```
@@ -93,6 +96,7 @@
 #
 #myhostname = host.domain.tld
 #myhostname = virtual.domain.tld
+myhostname = pobox.home.example.com
 
 # The mydomain parameter specifies the local internet domain name.
 # The default is to use $myhostname minus the first component.
```
Set it to the hostname of the server.

The same applies to `mydomain`.
```
@@ -100,6 +104,9 @@
 # parameters.
 #
 #mydomain = domain.tld
+# any domain we recieve email for dovecot for is our domain.
+#mydomain = $query/vmailbox_domain.query
+mydomain = example.com
 
 # SENDING MAIL
 # 
```
You can see I tried here too...
I've left the line (commented out) to remember. Uncomment to break things.

The next bit is housekeeping to indicate that I've opened up the system.
I could have restricted to the local network but didn't bother.
The default for the package is to lock down to just `localhost`.
```
@@ -129,10 +136,10 @@
 #
 # Note: you need to stop/start Postfix when this parameter changes.
 #
-#inet_interfaces = all
+inet_interfaces = all
 #inet_interfaces = $myhostname
 #inet_interfaces = $myhostname, localhost
-inet_interfaces = localhost
+#inet_interfaces = localhost
 
 # Enable IPv4, and IPv6 if supported
 inet_protocols = all
```

We now get to do something interesting.
```
@@ -180,10 +187,12 @@
 #
 # See also below, section "REJECTING MAIL FOR UNKNOWN LOCAL USERS".
 #
-mydestination = $myhostname, localhost.$mydomain, localhost
+#mydestination = $myhostname, localhost.$mydomain, localhost
 #mydestination = $myhostname, localhost.$mydomain, localhost, $mydomain
 #mydestination = $myhostname, localhost.$mydomain, localhost, $mydomain,
 #      mail.$mydomain, www.$mydomain, ftp.$mydomain
+mydestination = $query/mydestination.query, $mydomain
+
 
 # REJECTING MAIL FOR UNKNOWN LOCAL USERS
 #
```
There is a subtle thing happening here.
The result setting is equivalent to the second line.
This is because we have these host names in the database.
Only `$mydomain` isn't set because of the reason above.

#### Mail Filtering
This is a big chunk, all of which is commented out here to show context.
All of this would apply to the *access rules* discussed in the
[Commands Reference](commands_reference.md) documentation.
I am currently using one rule for all inbound email at this time
which means that the first part would be uncommented.

**NOTE:** I have currently commented these out because my ISP is
already doing this at the border.
If we bounce email here, where does it go?
If there is ever another ingress from the wild to this server, it gets
activated.
```
@@ -241,6 +250,44 @@
 
 # TRUST AND RELAY CONTROL
 
+## in file sample-smtpd.cf to connect to greylister (sqlgrey).           
+#                                                                       
+#smtpd_recipient_restrictions =                                          
+#     reject_invalid_hostname
+#     reject_non_fqdn_recipient                                        
+#     reject_non_fqdn_sender                                           
+#     reject_unknown_sender_domain                                     
+#     reject_unknown_recipient_domain                                  
+#     reject_multi_recipient_bounce                                    
+#     reject_unauth_pipelining
+#     permit_mynetworks                                                
+#     reject_unauth_destination                                        
+#     reject_rbl_client zen.spamhaus.org
+#     check_policy_service unix:private/policyd-spf
+#     check_policy_service inet:127.0.0.1:2501
+#     reject_non_fqdn_hostname                                         
+#     reject_invalid_hostname                                          
+#     $query/recipient_access.query
+# uncomment below only for wildcard domain access. not a good idea...
+#     $query/domain_access.query
+#     permit                                                           
+#
+#policyd-spf_time_limit = 3600
+#
+## DKIM linkage for verifying and signing email
+## DMARC (sock 8893) runs last!
+#
+#smtpd_milters          = inet:127.0.0.1:8891, inet:127.0.0.1:8893
+#non_smtpd_milters      = $smtpd_milters
+#milter_default_action  = accept
+#
+ 
+#smtpd_restriction_classes = x-reject x-permit x-hold
+#
+#x-reject = reject-stuff
+#x-permit = permit-stuff
+#x-hold = hold-stuff
+
 # The mynetworks parameter specifies the list of "trusted" SMTP
 # clients that have more privileges than "strangers".
 #

```
The first parameter, the bulk of this chunk, defines what `smtpd_recipient_restrictions` does.
Each continuation like is a builtin filter in `postfix`.
The filters are applied in order.
You can see the SPF processing half way down in the `check_policy_service` parameter.
One is via a UNIX pipe and its alternative is via `localhost`.
A few lines further down, just before and after the "uncomment ... not a good idea..."
is the line:
```
+#     $query/recipient_access.query</p>
```
Checks access for *addresses* and 
```
+#     $query/domain_access.query</p>
```
checks for *domain* accesses.
These queries will return a `smtpd_restriction_classes` name *key* that may have been set for either
the *address* or its *domain*.
The changes at the bottom of the listing add a definition of `smtpd_restriction_classes`
followed by the definitions of each class.
These filters are imaginary examples not actual restrictions defined in `postfix`.
The way this works is as follows:
* Assume we set an *access rule* named `hold` with an action `x-hold`.
* Edit `my@address` to set its `rclass` property to `hold`.
* An email arrives for `my@address` and filtering progresses to the `recipient_access.query` line.
* The query returns `x-hold` as its result which is then processed to do something like "holding" the email.

See the `postfix` documentation for how these filters work in detail.
I use none of these in my local configuration at present because my ISP supplied SMTP server
handles all outside email and everything on my internal network is pretty tame system notifications.

The next chunk of `main.cf` is more general `postfix` configuration setting my "friendly" network.
```
@@ -281,6 +328,7 @@
 # (the value on the table right-hand side is not used).
 #
 #mynetworks = 168.100.3.0/28, 127.0.0.0/8
+mynetworks = 192.168.2.0/24, 127.0.0.0/8
 #mynetworks = $config_directory/mynetworks
 #mynetworks = hash:/etc/postfix/network_table

```
This is only my local LAN and, of course, the loopback.

#### Relay Processing
If we need relay support, we need to set things up so `postfix` can first decide
if an inbound email can be accepted because it can be relayed somewhere else
and second, it has a transport destination to forward it to.
Take for example, the case described in [Postfix Standard Configuration Examples](https://www.postfix.org/STANDARD_CONFIGURATION_README.html).
In this snippet, `postfix` gets the wiring to recognize addresses that can be
accepted for relay.
It also has the wiring to send it on its way to the next destination.
We have skipped access rules "stuff" to focus on the relay wiring.
```
 3     relay_domains = example.com
 
11     relay_recipient_maps = hash:/etc/postfix/relay_recipients
12     transport_maps = hash:/etc/postfix/transport
13 
14 /etc/postfix/relay_recipients:
15     user1@example.com   x
16     user2@example.com   x
17      . . .
18 
19 /etc/postfix/transport:
20     example.com   relay:[inside-gateway.example.com]

```
This shows both the parameters and the hash files that make it work.
The next chunk wires the relay processing using the database.

To make the discussion below work, we first must make entries into the database
with `postdove` that do what the *hash* files do above.
First, we create the transport needed by the relay.
```
[pobox ~]# postdove add transport inside --transport=relay --nexthop=[inside-gateway.example.com]
```

Next, we have to create the domain we will relay and add some addresses to relay.
```
[pobox ~]# postdove add domain example.com --class=relay --transport=inside
[pobox ~]# postdove add address user1@example.com
[pobox ~]# postdove add address user2@example.com
```
I do not have to add the transport for either user because they inherit
it from the domain. That is it.

I have it commented out because I'm not relaying anywhere.
A larger network would use this for distributed internal mail service.
```
@@ -313,6 +361,8 @@
 # permit_mx_backup restriction description in postconf(5).
 #
 #relay_domains = $mydestination
+#relay_domains = $query/relay_domain.query
+#relay_recipient_maps = $query/relay_recipients.query
 
 # INTERNET OR INTRANET
 
```
There are two queries. The first sorts out traffic for a whole domain.
The second is for individual recipients.
For example, if internal to `example.com` there were internal subdomains `eng.example.com` and
`office.example.com`, both of which have their own servers in different locations,
queries to the domain would determine the relay.
The same applies to the recipient queries.
If `bob@example.com` is actually served by the email service in `eng.example.com`, rather than
this local server, his email would be relayed based on the result of this query.

The next chunk manages *transports*.
Notice that other than the SPF and spam filtering transports we set above,
there are no others set.
In a conventional `postfix` installation, one of the transports at the end
of the `master.cf` file would be uncommented to provide the link to
whatever IMAP/POP3 server was used.
In my previous installation, I added a line to its LMTP server and made the default
transport set to it.
Here, we use the database.
```
@@ -384,6 +435,7 @@
 # TRANSPORT MAP
 #
 # See the discussion in the ADDRESS_REWRITING_README document.
+transport_maps = $query/transport_maps.query
 
 # ALIAS DATABASE
 #
```
This works similarly to a *relay*, in fact, it can be used for relays.
The *key* is an *address* for a specific recipient or a *domain* for all the
recipients in the domain.
The *value* returned is of the form `<transport type>:<next hop>`.
In the case of `example.com` the transport is to the local `dovecot` server.
We accomplish this by first adding a *transport* entry to the database called `dovecot` that has
a `transport` property of `lmtp` and a `nexthop` property of `localhost:24`.
We then set the `transport` property of the domain `example.com` to `dovecot`.
The end result is that when `postfix` looks up `example.com` to figure out where to send it next,
it gets back `lmtp:localhost:24` which will tell it to use its LMTP client engine to connect
to `localhost` and transfer the email.

#### Aliases and Virtual Aliases
This last chunk sets up all *alias* and *virtual alias* lookups to use database queries.
```
@@ -402,9 +454,16 @@
 # "postfix reload" to eliminate the delay.
 #
 #alias_maps = dbm:/etc/aliases
-alias_maps = hash:/etc/aliases
+#alias_maps = hash:/etc/aliases
 #alias_maps = hash:/etc/aliases, nis:mail.aliases
 #alias_maps = netinfo:/aliases
+alias_maps = $query/alias_maps.query
+
+# Virtual aliases
+#virtual_alias_domains = $query/virtual_domain.query
+virtual_alias_maps = $query/virtual_alias.query
+virtual_mailbox_domains = $query/virtual_domain.query
+#virtual_mailbox_maps = $query/virtual_mailbox.query
 
 # The alias_database parameter specifies the alias database(s) that
 # are built with "newaliases" or "sendmail -bi".  This is a separate
@@ -413,7 +472,7 @@
 #
 #alias_database = dbm:/etc/aliases
 #alias_database = dbm:/etc/mail/aliases
-alias_database = hash:/etc/aliases
+#alias_database = hash:/etc/aliases
 #alias_database = hash:/etc/aliases, hash:/opt/majordomo/aliases
 
 # ADDRESS EXTENSIONS (e.g., user+foo)
```
We turn off default maps and their databases and enable query driven maps.
I don't use `virtual_mailbox_maps` in my configuration but the database supports
such queries. See the `postfix` documentation for details.
If it fits your configuration, turn it on.

#### Spam Filter Wiring
This last chunk is just linkage for the *Amavisd* mail filter linkage.
This directive `content_filter` routes email to another process,
in this case the `amavis` filter that is listening on `localhost:10024`.
The companion to this was set up in `master.cf` to receive the email back
on `localhost:10025`.
It is part of the configuration but not involved with `postdove`.
```
@@ -736,3 +795,7 @@
 smtp_tls_security_level = may
 meta_directory = /etc/postfix
 shlib_directory = /usr/lib64/postfix
+
+# amavis linkage for content screening
+content_filter=smtp-amavis:[127.0.0.1]:10024
+
```

This finishes the `postfix` configuration.
The changes here reflect my specific configuration so use this as a guide, not a cookbook.
It depends on what your installation needs.
The important point is that it uses most of the query setups in the `postdove` project.
Note that there are query files that I do not use.
That does not mean they are not useful.

Once this step is done, we are finished with configuration and now can move on to testing
and operation. See [Administrator Guide](admin.md) to see what is next.
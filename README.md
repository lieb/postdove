# postdove
This project started off with the need to upgrade my private email service which was
based on an IMAP/POP3 server that was breaking in some rather unfriendly ways.
The project that my IMAP server was from didn't seem to have much active development
or support anymore so I had to move.
In the end, I chose `dovecot` which integrates nicely with `postfix` which
I had been using for years.
It is also actively being developed and has maintained **Fedora** packaging.

I looked around for something that would integrate `postfix` and `dovecot` without
either a big investment in something like *FreeIPA* or the endless hacking of
configuration files with `emacs`.
I have no problems with that editor. I use it all the time.
But system administration, especially email system configuration, via text editor
has never been my idea of a good time.
Hey, that was all I had when I was running a Bell Labs UNIX on a PDP-11/70
but that was a long time ago.
I have run `postfix` for years and I still have the scars from running
`sendmail` so when I read the docs for `dovecot` I was either faced with
`emacs` and/or `vi` sessions in the bowels of `/etc` or further research.
I decided that I needed to look around for something that would be more structured
but not needing heavyweight enterprise level "stuff" like LMAP,
even if it is wrapped in *FreeIPA*.
Both `postfix` and `dovecot` have clean integration with SQL databases so all
I needed was a "something" that would do consistently good SQL things in its basement
and let me do simple command line administration.

As usual in this space, I found lots of ideas and "tutorials" but all the SQL magic
was left as an exercise for the reader... I found more bits but nothing complete.

My research uncovered a "tutorial" referenced in the *Dovecot* wiki documentation.
The reference was an email post to `dovecot.org` back in 2012
on the [Dovecot project's pipermail archive](https://dovecot.org/pipermail/dovecot/2012-February/133734.html).
From what I could see, this was full of useful information and ideas but it was
just that, a writeup, nothing production ready.
Although useful as a starting point, I have read far too many SQL based ideas that
go the first few steps into something workable and stop,
usually with some pages of hand crafted queries and hand-stuffed table inserts.
For example, there is a populated database dump in the tarball that is full of half-eaten bits and dangling references.
I am grateful for the author's work as it gave me some pointers and ideas.
But that is only 5-10% of something real.

The result is `postdove` which is a self contained utility that uses a *Sqlite3* database to manage an email service comprised of `postfix` and `dovecot` servers.
Both servers make queries to the database in order to process email from incoming email filtering and forwarding to client access via IMAP/POP3.
 
The `postdove` utility application manages the database so that there
is always a consistent view of the data.
Administrative changes can be made in to the service in real-time
without requiring either server to reload or restart.
`postdove` operates from the command line to add, delete, and edit aliases, routing and delivery, and the individual email user accounts.
It supports *import* and *export* commands that use the file formats native to
both `postfix` and `dovecot` to make setup and migration easier. 


This project is a real management utility with ACID level SQL (as best as Sqlite can do...) and a commands interface that an ordinary admin can use.
And it is documented...

The schema was also changed in places to simplify queries and apply constraints
that make it more ACID compliant.
Triggers and views were added to move all of the query complexity and relational logic into one place, the SQL engine.
See the comments attached to the schema change sets in the git repository
for details.

For secure operation, the `postdove` utility uses the command line interface
rather than, say a REST API, and proper use would be within an SSH session.
This keeps the attack surface of the service restricted to port 22.
Don't get me wrong, REST is a good API architecture but most interfaces of this
type only provide the API, not the rest of the browser bits to make it useful.
Like the tutorial that started me on this project, it is like handing you a pair
of live wires straight off the power pole and saying, "Now plug it in."
There are plans and notes for a text based interface but it too would be expecting
to run out of an SSH session.

## Building Postdove
The whole of the project comes down to a single utility program written in GO that
uses a good CLI interface and *Sqlite3*.
Key files such as the database schema and early-on table contents are contained
within the utility using the *embedded* module that is part of GO.
The rest is some system coonfiguration and documented changes
required in `postfix` and `dovecot` configuration files.

Build the application and install it following the instructions
in [Building and Installation](./doc/building.md).

## Configuration and Testing
In order to use the database both `dovecot` and `postfix` need configured.
Given that my installation runs on a VM hosted on one of my servers,
there is also some system configuration work to be done prior
to configuring the mail service.
See the instructions in [Configuration](./doc/configure.md) for all the gory details.

## Operation
Once we have all the parts up and running, we have to manage it.
Here is the [Administrator Guide](./doc/admin.md) to get started.
There is also the `postdove` [Commands Reference](./doc/commands_reference.md)
where one can go for all the details.
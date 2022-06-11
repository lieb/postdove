# Building and Installation
`postdove` is 100% written in the [GO](https://go.dev/) language.
Its database module requires the *Sqlite3* libraries which are used
by the *go* SQL database package. Everything else is managed by the
language's package and module system.

Although *GO* applications can be built on Windows,
the email systems `postfix` and `dovecot` are usually not available
on Windows.
*GO* also supports BSD Unix and its variants so `postdove` could be built
and run on BSD based systems such as mac/OS, but that has not been tested.

These instructions assume a Linux distribution and in detail, *Fedora 35*
or above.

## Development Packages
Building `postdove` requires three packages in the development environment
besides the usual developer tools like *git* and editors that would be
familiar to any UNIX/Linux based developer.

**NOTE:**
	The command lines below assume *Fedora* or *RHEL* packaging.
	Use the appropriate packaging tools for other distributions such as *apt*
	for Debian or Debian based distributions such as *Ubuntu*.
	
The two primary tools and libraries are the *GO* toolchain and the *Sqlite3*
libraries and utilities.
```
# dnf install golang sqlite sqlite-libs libdbi-dbd-sqlite
```
The *GO* compiler uses its *cgo* package to build the linkage to
the *sqlite* libraries.
This requires the *C* compiler (gcc) and friends installed.
This document assumes the developer already has these tools installed
or can install them when needed.
There may be some loose bits that I have missed.
I have a well-stocked development environment and have not done a packaging
build where only the minimum sandbox is used.

The *GO* build toolchain will manage all its dependencies.

## Building
Once the environment is set up, it is time to *clone* the repository
and set things up for the build.

### Getting the Software
My *git* based development is set up in my `~/git` directory.
Everything in these instructions assume that is where we are working.

First clone the repository:
```
[starlight ~]$ cd git
[starlight git]$ git clone http://github.com/lieb/postdove
[starlight git]$ cd postdove
[starlight postdove]$
```
We are now ready to set things up.

### Updating Dependencies
The first thing to check is to see if your *GO* language tools are current
enough to do a successful build.
Since documentation often lags the actual state of the code,
check what the code expects.
This application has been written using *GO*'s module system which
embeds the tools version in the `go.mod` modules description file.
Checking the current minimum version requirement can be done as follows:
```
[starlight postdove]$ grep 'go ' go.mod
go 1.16
```
Note the single quotes (`'`) to include a space character.
This eliminated a lot of extraneous "stuff" in the output.
What we are looking for is the `1.16` which is the minimum version
of the toolchain required to build the code.
The *GO* developers have a very strong standard to always be backward
compatible within a major release, in this case `1`.
This means that if your compiler chain is at least at `1.16` in this
everything should work.
I will only update this version string in `go.mod` if I should use
newer language features. But check first! This document may age
but the source does not.

The software depends on a number of published *go* modules.
Cloning the `postdove` is not enough code for the build.
Before we can attempt a build we must import the necessary extra modules.
These modules are referenced by the `import` statements in source code using
a net addressible (URI) name the *GO* can resolve.
We do this by the command:
```
[starlight postdove]$ go get
```
We could also use any other command like `go build` but this just downloads/updates
the required modules without attempting to compile.
This step could be done later but this is a nice place to start
given that a `go build` would fail if the next step is not done first.

### Updating the Version
This is a *git* repository and I have used tags to mark stable code.
The tag is used by `postdove` to report its version
so one can readily determine the version used at build time.
See the [Commands Reference](commands_reference.md) for the details on
how to display the version.
To transfer the *git* version information into a form that the compiler
and runtime can find we have to use the *generate* feature of *GO*.
This is done by executing the following command at the top level of
the source tree:
```
[starlight postdove]$ go generate
```
What this does is run the shell script `./set_version.sh`.
Let's look at the script and the bit of code that runs it.
First is the bit of source code:
```
[starlight postdove]$ grep generate main.go
//go:generate bash set_version.sh
```
The shell script is:
```
[starlight postdove]$ cat ./set_version.sh 
#!/bin/bash
git describe  > cmd/version.txt
```
which runs `git describe` directing its output to the file `./cmd/version.txt`.
We make that happen by running:
```
[starlight postdove]$ go generate
```
We now have the file `./cmd/version.txt`
```
[starlight postdove]$ cat ./cmd/version.txt 
V0.9-RC1-4-gd8f26c3
```
This was the current version at the time of writing this document.
See the *git* documentation for the details but in this case,
what we see in the current version tag is the current state
of the repository, and therefore the build which is four commits after that
tag at the commit hash `gd8f26c3`.

Note that this file is purposely ignored by *git*.
It should _never_ be committed to the repository and
should be re-generated whenever changes to the code are committed,
especially if the utility is put into production use.

If you forget to do this step, you will get the following error:
```
[starlight postdove]$ go build
cmd/root.go:27:12: pattern version.txt: no matching files found
```
as a "reminder" that you forgot the `go generate` step.
I would like this to be more closely coupled and automatic
but the toolchain does not conveniently support it.
Developers familiar with `make` and/or `cmake` would notice that this
functionality could be (and frequently is) included in the
`Makefile` making this step always current and automatic.

### Building
Once all the work above is done, we can see if all the bits are in place
and a `postdove` executable is here:
```
[starlight postdove]* go build
[starlight postdove]* ls -l postdove
-rwxrwxr-x. 1 lieb lieb 8338992 May 30 15:23 postdove
```
We now have something to play with.

### Testing
Although we have just built `postdove`, that executable is not what
is used for the unit tests.
The `go test` command called in `test_all.sh` does its own builds
including the testing module as a test harness.

The source comes with unit tests for all the functions of every package.
Since `go test` runs tests in its own seemingly arbitrary
internal order,
each directory has a shell script `test_all.sh` which orders the
runs of `go test -run=<test name>` to control the order of testing.
There is a top level `test_all.sh` which calls the `test_all.sh`
in each sub-directory.
The ordering is from the bottom to the top so if any test fails,
one can terminate the test run anytime after an error is reported.
The scripts were not written to terminate on first error so that all
the tests can be run in order to see the extent of the damage followed
by focusing on the first test failures.

Run the tests. This is what you should see:
```
[starlight postdove]$ ./test_all.sh 
RFC822 Test
PASS
ok      github.com/lieb/postdove/maildb 0.002s
Target Test
PASS
ok      github.com/lieb/postdove/maildb 0.002s
Transport Decode Test
PASS
ok      github.com/lieb/postdove/maildb 0.002s
Database load test
PASS
ok      github.com/lieb/postdove/maildb 0.004s
Access Test
PASS
ok      github.com/lieb/postdove/maildb 0.005s
Transport Test
PASS
ok      github.com/lieb/postdove/maildb 0.005s
Domain Test
Bad class (jazz)
PASS
ok      github.com/lieb/postdove/maildb 0.006s
Address Test
PASS
ok      github.com/lieb/postdove/maildb 0.007s
Alias ops Test
PASS
ok      github.com/lieb/postdove/maildb 0.017s
Mailbox Test
PASS
ok      github.com/lieb/postdove/maildb 0.006s
Test_Import
Test_Simple
Test_Simple errors
Test_Postfix
Test_Postfix errors
Test_Aliases
Test_Password
PASS
ok      github.com/lieb/postdove/cmd    0.002s
Test_Cmds
PASS
ok      github.com/lieb/postdove/cmd    0.005s
TestAccess
PASS
ok      github.com/lieb/postdove/cmd    0.012s
TestTransport
PASS
ok      github.com/lieb/postdove/cmd    0.015s
TestTransportEdit
edit nexthop to somewhere.org
PASS
ok      github.com/lieb/postdove/cmd    0.007s
TestTransportAdd
PASS
ok      github.com/lieb/postdove/cmd    0.007s
TestTransportAddOne
PASS
ok      github.com/lieb/postdove/cmd    0.005s
Test_Domain
domains from stdin
domains from file
PASS
ok      github.com/lieb/postdove/cmd    0.017s
Test_Address
PASS
ok      github.com/lieb/postdove/cmd    0.019s
TestAliasCmds
PASS
ok      github.com/lieb/postdove/cmd    0.029s
Test_VMailboxCmd
PASS
ok      github.com/lieb/postdove/cmd    0.015s
Test_Create
PASS
ok      github.com/lieb/postdove/cmd    0.019s
TestCreateNoAliases
PASS
ok      github.com/lieb/postdove/cmd    0.012s
TestViews
Domain class types
Domain access
user@domain access
Transport lookups
Domain lookups
Local Alias lookups
Virtual alias lookups
Lookup all users
Lookup prefetch of jeff@pobox.org
Lookup user
Lookup bogus user
Deny allow
Deny deny
PASS
ok      github.com/lieb/postdove/cmd    0.022s
```
The times would be different but the rest should indicate all tests returned a `PASS`.
The specific tests are run in the order of simplest to dependent and more complex.
The list of tests and their names may change but the official list is that in
the files `test_all.sh` throughout the repository.

If a test fails, you would see an error message and a line number in the testing source file
where the error occurred.

## Installation
It is highly likely and recommended that the utility not be built on the
mail server.
There is no need to have any development tools already installed on a service
machine waiting for an exploit to be compiled.
Therefore, we assume that the utility will be copied to the email server from
somewhere else.
Copy the built binary to the private `bin` directory of `root` on the service system.
```
[starlight postdove]$ scp postdove root@pobox:~/bin
```
which is where the executable will be located on the server.
Except for the initial configuration process, there are no other files to be installed
because the utility is self-contained.

Using `~root/bin` as the destination for the install does three things.
First, it avoids clutter in `/sbin` or `/usr/sbin` which would only get in the way of
system (Linux distribution) updates and upgrades.
Second, `/usr/local/sbin` is avoided for the same reason.
Much use of `/usr/local` is no longer relevant now that distributions package most of
the stuff that used be be locally compiled and installed in the old days.
Other applications seem to gravitate toward distribution independent packaging systems
like *flatpak*, *appimage*, or *snap* which are really oriented toward non-privileged
user environments.
The `postdove` utility requires Superuser (administrator) privileges to manage the email service and as such runs in a locked box.
Hence, we put it in `~root/bin`, the private home of the administrator account.


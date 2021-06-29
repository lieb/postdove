# Mail Storage Over NFS
The first filesystem setup is done on `suntan` before we attempt anything via NFS.
The first thing to create is the base directory for IMAP accounts.

>`NOTE:` We will only show the configuration results here.
NFSv3 configuration is very different from NFSv4 and is not shown.
Consult the system administration documents for the details.

This is what the filesystem layout on `suntan` looks like
once we have made what we want to export available to the NFS server
:
```bash
[root@suntan ~]# cat /etc/fstab
# skip uninteresting system mounts like '/' and swap
UUID=xxxxxx         /srv/pgdata   btrfs   subvol=pgdata   0 0
UUID=xxxxxx         /srv/dovecot  btrfs   subvol=dovecot  0 0
# Home
LABEL=suntan_str    /home         btrfs   defaults,subvol=home    1 2

# nfs v4 shares
/home              /nfs4exports/home       none  bind    0 0
/srv/dovecot       /nfs4exports/dovecot    none  bind    0 0
```

You will notice that the *bind* mounts make `/home` and `/srv/dovecot` which are on different
BTRFS volume sets available in one place for the NFS server.
The NFS server uses `/etc/exports` to export filesystems.

```bash
[root@suntan ~]# cat /etc/exports
/nfs4exports 192.168.2.0/255.255.255.0(rw,insecure,no_subtree_check,nohide,fsid=0)
/nfs4exports/home 192.168.2.0/255.255.255.0(rw,insecure,no_subtree_check,nohide) 
/nfs4exports/dovecot pobox(rw,insecure,no_root_squash,no_subtree_check,nohide)
```

The first entry sets the boundaries for all exports to my local subnet.
The next entry exports `/home` to that same network in the way typical for
shared/linked home directories on the laptops and desktops on my home network.

The last entry, by contrast, is of interest to us. Note that `dovecot` is _only_ exported to `pobox`.
In addition, it has `no_root_squash` added to its options.
This allows `root` on `pobox` to create and write files as `root` and not `nobody` on `suntan`.
This is important because it allows the administrator on `pobox` to do filesystem
tasks on this filesystem. Otherwise, such work would have to be done by the
`suntan` administrator on `suntan`.

>*Note:* The use of `pobox` instead of an IP address could be problematic here.
The name, resolved by DNS, may not be present at the time this file is processed if
the network uses dynamic DNS that gets its dynamic entries from the DHCP server.
In that case, there are two choices, either hardwire an IP address here or have
the DNS server own the static mapping instead of the DHCP server.
This issue comes up because our `pobox` VM is coming up in parallel within the
booting of `suntan`, its host.

NFSv4 can only export from a single directory for security reasons.
This is the purpose of the *bind* mounts.
This is what we have on `suntan`.
```bash
[root@suntan ~]# ls -l /nfs4exports/
total 0
drwxr-xr-t. 1     97     97  66 Jun 18 08:44 dovecot
drwxr-xr-x. 1 root   root   174 Feb 11  2019 home
```
Note that the owner and group of `dovecot` is `97`.
This is the uid/gid pair assigned by the **Fedora** install of `dovecot` package on `pobox` and may/will be different on another distribution.
No matter, we use what the install assigns.
We do not see the user name here because we are looking at it from `suntan` not `pobox`.

We are now done with filesystems on `suntan`.

## Automounting
On the `pobox` side we have to properly mount the export.
We can either use a *hard* mount which is set up in `/etc/fstab` or a mount managed
by *autofs*. As far as the server (`suntan`) is concerned, there is no difference.
One way or the other the mount is managed by the client, in this case `pobox`.
Most facilities prefer *autofs* for this to provide flexibility.
Hard mounts tend to hang systems if not managed in the right order and *autofs* is more tolerant to network burps.

The first step is to make sure *autofs* package is installed and enabled.
The service runs using the configuration files `/etc/auto.master` and `/etc/auto.suntan`.
How these are arranged is my personal choice, not hard wired.
Use the configuration most comfortable so long as it gets the same result.

First `/etc/auto.master`:

```bash
[root@pobox etc]# diff -u auto.master.orig auto.master
--- auto.master.orig    2018-02-01 18:01:22.705755203 -0800
+++ auto.master 2018-02-01 18:03:35.897832923 -0800
@@ -25,4 +25,5 @@
 # same will not be seen as the first read key seen takes
 # precedence.
 #
-+auto.master
+#+auto.master
+/mnt/suntan    /etc/auto.suntan        --timeout=30
```

This is how I set up access via the `/etc/auto.suntan` file.
There are two things to note here. First, all automounts from `suntan` will happen
under `/mnt/suntan` and second, `--timeout=30` will keep a mount active for only
30 seconds of idle. This is useful should any of the players, the server, a client, or the network go a way. if set up this way, everything gets properly reset.
This line also references `/etc/auto.suntan` any time a process attempts to access
any files under `/mnt/suntan`. To see what those accesses are we look at
`/etc/auto.suntan`.

```bash
[root@pobox etc]# cat auto.suntan
dovecot        -fstype=nfs4,rw,hard,intr,nodev,nosuid          suntan.example.com:/dovecot
```
Notice this matches the *bind* mount under `/nfsv4exports` without the top directory
because the *root* of the NFS server's eligible exports is `/nfsv4exports`.
This is why we do the *bind* mounts.
Most of the options are standard but we force `fstype=nfsv4` to get the correct protocol.
On *Fedora* This has been the default for a long time.
However, there are some systems out in the wild that still have NFSv3 as the default.

We do not show how we get the mount to `/srv/dovecot`.
We do this by creating a symbolic link `/srv/dovecot` that points to `/mnt/suntan/dovecot`.

## Access Controls
Remember we have *selinux* enabled so we have to set up `pobox` to allow access to
`/srv/dovecot`.

>*NOTE: The following applies to *selinux* enabled systems. There is a similar
control for systems that run AppArmor, primarily **Ubuntu**. Consult their
documentation for how NFS mounts are treated by clients.
If you have neither and run plain old vanilla UNIX style you can skip this section.

First, let us look at the labels on `suntan`'s side:

```bash
[root@suntan srv]# ls -laZ /srv/dovecot/
total 2931064
drwxr-xr-t. 1   97   97 system_u:object_r:var_t:s0             66 Jun 18 08:44 .
drwxr-xr-x. 1 root root system_u:object_r:var_t:s0             26 Feb 11  2019 ..
drwxrwxrwt. 1   97   97 unconfined_u:object_r:var_t:s0         26 Jun 16 16:24 example.com
```

For those unfamiliar with *selinux* labels, the `-Z` option to `ls` is displaying
the label in the field following the familiar *group* field. It is four fields
separated by a `:`.
This configuration shows `var_t` as the label which is used for services that keep
their data files in `/var`.
What each means is not important here except to see how the label changes as displayed on `pobox`.

```bash
[root@pobox dovecot]# ls -laZ /srv/dovecot
total 2931064
drwxr-xr-t. 1 dovecot dovecot system_u:object_r:nfs_t:s0            66 Jun 18 08:44 .
drwxr-xr-x. 3 root    root    system_u:object_r:autofs_t:s0          0 Jun 11 19:01 ..
drwxrwxrwt. 1 dovecot dovecot system_u:object_r:nfs_t:s0            26 Jun 16 16:24 example.com

```

We can see that the 3rd field in the labels have changed.
They have gone from `var_t` which is the label for things that locate in `/var`
to `nfs_t` and `autofs_t`. This is because the file is now on the client side and
*selinux* requires tighter access controls for network access and denies
access to *home directories*. We solve this on `pobox` by telling *selinux* that
it is ok to access nfs mounted home directories.

```bash
[root@pobox dovecot]# setsebool -P use_nfs_home_dirs 1
```

Without going into too much detail, this sets a *boolean* in the `pobox` kernel that
enables (the 1 argument) the `use_nfs_home_dirs` capability.

We are now done with setting up the export of the mailstore to the VM. For the final
steps return to [System Configuration](system_configuration.md) and test what we
have set up.

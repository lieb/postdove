# System Configuration
My mail service runs in a VM hosted by my file server.
I chose a VM rather than a container because a VM has its own IP address and
network stack and therefore is independent of where it is hosted.
The VM only has the standard **Fedora**
system users and one administrator account. The `root` account has login
disabled and the administrator account (which can `sudo`) requires `ssh` authorized key for login.

Email has been traditionally stored in either `/var/mail` or `/var/spool/mail`.
Typically, local mail inboxes have been located here since the days before networking on UNIX systems.
However, few IMAP/POP3 servers do that anymore.
They either deposit incoming mail in the user's home directory or in a central place,
often not even on the same system as the user. Our server is a black box server that only provides access via IMAP/POP3.
There are a number of options for configuring the email store in the VM:
* The first and worst choice is to have the mail store within the VM image.
The only data kept in the VM image are configuration files and system management databases.
All user files are located somewhere else. In short, a hosted VM is not like a standalone host.
* The mail store can be a mounted remote filesystem.
The typical choice is either Samba or NFS.
The configuration of either is no different for a VM than it is for a typical networked system.
The only caveat is that if NFS is chosen, it should require NFSv4 or higher.
The version 3 protocol has been deprecated from **Fedora** and other distributions because of its well known weaknesses but there are still a few legacy setups around.
Our configuration is NFSv4.1.
* The third choice is a virtual filesystem provided by the hypervisor. **Fedora** uses KVM/QUMU so there are two choices.
A virtual filesystem is based within the hypervisor itself rather than traversing the network stack.
Up until recent kernels and QEMU emulator versions the only virtual filesystem, called *VirtFS*, was based on the 9P remote file protocol
from the Plan9 operating system.
Although the concept was cleaner and access confined unlike a network filesystem,
the protocol has proved to be not very efficient, in fact, much slower than NFSv4.
Recent kernels and QEMU have a replacement called *Virtio-FS* which uses the FUSE filesystem infrastructure. Its performance is almost as fast as a local filesystem.
More important, it is fully POSIX compliant which none of the network based protocols support.

Our configuration will be migrating to it soon. *TBD update*.

### The players
We refer to a number of systems throughout this document. To clarify which is which, these are the players:
* `example.com` This is the usual example substitute as a disguise for my own domain.
I have a site presence by that name at my ISP that for this discussion provides SMTP/IMAP for me.
* `home.example.com` is the subdomain for my private home network.
There is no formal DNS connection (delegation) from my public host.
There is also no routeable inbound access to my home network.
* `pobox` is the VM instance running the SMTP/IMAP/POP3 services.
It is also registered in my home DNS zone with a companion name of `mail`.
In normal DNS conditions, this name would be a CNAME to `pobox` but there are some services
that will not accept a CNAME so it is managed as an ordinary A record.
Its full name is `pobox.home.example.com`. We shorten the name(s) for simplicity.
* `suntan` is my home file/service server. It is the host for `pobox`.
* `nighthawk` is a client workstation which I will use as a standin for everything else from our smartphones to my desktop workstation.
* `/nfs4exports` is the directory on `suntan` that the NFSv4 server is restricted to for exports.
Each entry here is a *bind* mount of portions of the server's filesystem elsewhere.
* `/srv/dovecot` This is the mount point on both `suntan` and `pobox`.
On `suntan` it is mounted as a sub-volume of one of the RAID1 disk sets.
On `pobox` it is the *autofs* managed mountpoint of that sub-volume.
* `/srv/dovecot/example.com` is the base directory for the IMAP accounts for the domain on both `suntan` and `pobox`.
Individual home directories for users are only located on `suntan` and are not
associated with the email configuration. Users only use `suntan` itself for home
directories.

## Network Configuration
Network configuration is outside the scope of this discussion. It has simply set up
so all systems have DNS names, assigned IP addresses, and appropriate routing.
For this service, there are no changes required for either `suntan` or any of the clients.

>**NOTE: The tools for this section are *NetworkManager* and *firewall-cmd*. See
your distribution's documentation for the details.

**Fedora**'s default network setup for VM instances is to configure their network to a private bridge that is NAT routed through the hypervisor.
This is identical to how the typical home Internet is set up. The VM can make
connections to the outside world but the outside world cannot make connections
to the VM. However, `pobox` must be visible to the rest of the home network.
This requires the setup of a second bridge on `suntan` that bypasses the hypervisor.
The steps are straight forward but we skip the details because there are plenty of
HOWTOs and tutorials on the subject. In sum, the following steps create what
we are looking for.

1. Create a bridge on `suntan`. This is separate from the one managed by
the hypervisor.
1. Add `suntan`'s ethernet interface to the bridge. This means that `suntan`'s
IP address is now attached to the bridge.
1. When creating the VM for `pobox`, choose this new bridge for its
network connection.

The end result looks like your typical network switch with two hosts attached to
it, one called `suntan` and the other called `pobox`.

### The Firewall
**Fedora** locks down network access with the *firewalld* service.
The typical installation only allows SSH and system management connections.
We need to add inbound access to `pobox` for the mail services.
We use the following commands to add them:

```bash
[root@suntan ~]# firewall-cmd --add-service=imap --permanent
[root@suntan ~]# firewall-cmd --add-service=imaps --permanent
[root@suntan ~]# firewall-cmd --add-service=smtp --permanent
[root@suntan ~]# firewall-cmd --add-service=smtps --permanent
[root@suntan ~]# firewall-cmd --permanent --add-service=managesieve
```

Checking our work, we see:

```bash
[root@pobox dovecot]# firewall-cmd --list-all
FedoraServer (active)
  target: default
  icmp-block-inversion: no
  interfaces: ens3
  sources: 
  services: cockpit dhcpv6-client imap imaps managesieve smtp smtps ssh
  ports: 
  protocols: 
  forward: no
  masquerade: no
  forward-ports: 
  source-ports: 
  icmp-blocks: 
  rich rules: 
```

We do not add *lmtp* for security reasons. If we choose to use the network address
rather than a UNIX socket for it, we will route through the *loopback* address
(127.0.0.1) which has no firewall rules. I have added the *sieve* port here because
I will be configuring its support in `dovecot`.

### Filesystem Configuration
The first filesystem setup is done on `suntan` before we attempt anything via NFS.
The first thing to create is the base directory for IMAP accounts.

>`NOTE:` We will only show the configuration results here.
Consult the system administration documents for the details.

This is what the filesystem layout on `suntan` looks like:
```bash
[root@suntan ~]# cat /etc/fstab
...
UUID=xxxxxx         /srv/pgdata   btrfs   subvol=pgdata   0 0
UUID=xxxxxx         /srv/dovecot  btrfs   subvol=dovecot  0 0
...
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

NFSv4 can only export from a single directory for security reasons.
This is what we have on `suntan`.
```bash
[root@suntan ~]# ls -l /nfs4exports/
total 0
drwxr-xr-t. 1     97     97  66 Jun 18 08:44 dovecot
drwxr-xr-x. 1 root   root   174 Feb 11  2019 home
```
Note that the owner and group of dovecot is `97`.
This is the uid/gid pair assigned by the **Fedora** install of `dovecot` on `pobox`.
It may/will be different on another distribution. No matter, use what the install assigns.
We do not see the user name here because we are looking at it from `suntan` not `pobox`.

The last thing we need to do before moving over to the `pobox` side is to create some directories for the domains we will be serving.
We will only be creating one directory but if the service is supporting multiple virtual domains,
we would be creating additional directories using the same commands.

```bash
[root@suntan dovecot]# cd /srv/dovecot
[root@suntan dovecot]# chown 97.97 .
[root@suntan dovecot]# mkdir example.com
[root@suntan dovecot]# chmod 777 example.com
[root@suntan dovecot]# chown 97:97 example.com
[root@suntan dovecot]# chmod o+t example.com
```

Note that we are using `97` again. Remember, `suntan` does not have `dovecot` installed.
If we look closely, we see that the `example.com` directory is wide open.
This is because when `dovecot` processes the first email, either by LMTP or IMAP/POP3, it does so as the authenticated user in `dovecot`.
This is the same behavior as for `/tmp` using the *sticky bit* on the directory which limits the access rights
so a user can only rename, or worse, delete anything not owned by them.
This is not a security issue because ordinary mail users can only access this directory via `dovecot`
which enforces additional access controls.
The end result looks like this:

```bash
[root@suntan ~]# ls -la /srv/dovecot/
total 2931064
drwxr-xr-t. 1   97   97         66 Jun 18 08:44 .
drwxr-xr-x. 1 root root         26 Feb 11  2019 ..
drwxrwxrwt. 1   97   97         26 Jun 16 16:24 example.com
```

We are now done with filesystems on `suntan`.

On the `pobox` side we have to properly mount the export.
We use *autofs* for this to provide flexibility.
Hard mounts tend to hang systems if not managed in the right order and *autofs* is more tolerant to network burps.
This whole issue goes away with a virtual filesystem.

#### Access Controls
Remember we have *selinux* enabled so we have to set up `pobox` to allow access to
`/srv/dovecot`. 
>*NOTE: The following applies to *selinux* enabled systems. There is a similar
control for systems that run AppArmor, primarily **Ubuntu**. Consult their
documentation for how NFS mounts are treated by clients.
If you have neither and run plain old vanilla UNIX style you can skip this section.

First, let us look at the labels on `suntan`'s side:

```bash
[root@suntan srv]# ls -laZ dovecot/
total 2931064
drwxr-xr-t. 1   97   97 system_u:object_r:var_t:s0             66 Jun 18 08:44 .
drwxr-xr-x. 1 root root system_u:object_r:var_t:s0             26 Feb 11  2019 ..
drwxrwxrwt. 1   97   97 unconfined_u:object_r:var_t:s0         26 Jun 16 16:24 example.com
```

For those unfamiliar with *selinux* labels, the `-Z` option to `ls` is displaying
the label in the field following the familiar *group* field. It is four fields
separated by a `:`. What each means is not important here except to see how the
label changes as displayed on `pobox`.

```bash
[root@pobox dovecot]# ls -laZ 
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
enables (the 1) the `use_nfs_home_dirs` capability.

We check our work by adding and removing files and directories in the
`/srv/dovecot/example.com` directory.
They should be owned by the ordinary user we are logged in as. If we change to
a different user, we should see the same behavior for the files created by the
new user but the new user should not be able to remove any other user's files.

At this point we are done with system configuration and it is time to move on to
[Database Creation](database_setup.md).
# System Configuration
My mail service runs in a VM hosted by my file server.
I chose a VM rather than a container because a VM has its own IP address and
network stack and therefore is independent of where it is hosted.
The VM only has the standard **Fedora**
system users and one administrator account. The `root` account has login
disabled and the administrator account (which can `sudo`) requires `ssh` authorized keys for login.

Email has been traditionally stored in either `/var/mail` or `/var/spool/mail`.
Local mail inboxes have been located here since the days before networking on UNIX systems.
However, few IMAP/POP3 servers do that anymore.
They either deposit incoming mail in the user's home directory or in a central place,
often not even on the same system as the user. Our server is a black box server that only provides access via IMAP/POP3.
There are a number of options for configuring the email store in the VM:
* The first and worst choice is to have the mail store within the VM image.
The only data that should be kept in the VM image are configuration files and system management databases.
All user files should be located somewhere else. In short, a hosted VM is not like a standalone host.
* The mail store can be a mounted remote filesystem.
The typical choice is either Samba or NFS.
The configuration of either is no different for a VM than it is for a typical networked system.
The only caveat is that if NFS is chosen, it should require NFSv4 or higher.
The version 3 protocol has been deprecated from **Fedora** and other distributions because of its well known weaknesses
but there are still a few legacy setups around.
Our initial configuration was NFSv4.1.
* The third choice is a virtual filesystem provided by the hypervisor. **Fedora** uses KVM/QEMU so there are two choices.
A virtual filesystem is based within the hypervisor itself rather than traversing the network stack.
Up until recent kernels and QEMU emulator versions, the only virtual filesystem, called *VirtFS*, was based on the 9P remote file protocol
from the Plan9 operating system.
Although the concept was cleaner and access confined, unlike a network filesystem,
the protocol has proved to be not very efficient, in fact, much slower than NFSv4 which requires the network
stack.
Recent kernels and QEMU have a replacement called *Virtio-FS* which uses the FUSE filesystem infrastructure.
Its performance is almost as fast as a local filesystem.
More important, it is fully POSIX compliant which none of the network based protocols support.

Our configuration started over NFSv4 but has since migrated to *Virtio-FS*.
We document both because every installation is different.

## The Players
We refer to a number of systems and directories throughout this document.
To clarify which is which, these are the players:
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
* `starlight` is a client workstation that I use for development and testing. I will use it as a standin for everything else from our smartphones to my desktop workstation.
* `/nfs4exports` is the directory on `suntan` that the NFSv4 server is restricted to for exports.
Each entry here is a *bind* mount of portions of the server's filesystem elsewhere.
* `/srv/dovecot` This is the mount point on both `suntan` and `pobox` for the
mail store.
On `suntan` it is mounted as a sub-volume of one of the RAID1 filesystems.
On `pobox` it is either the *autofs* managed mountpoint or the *virtiofs* mount of that sub-volume.
* `/srv/dovecot/example.com` is the base directory for the IMAP accounts for the `example.com` virtual mailbox domain on both `suntan` and `pobox`.
There are individual home directories for users located on `suntan` but they are not
associated with the email configuration. Ordinary login users only use `suntan` itself for shared directories.

## Network Configuration
Network configuration is outside the scope of this discussion. It has simply set up
so all systems have DNS names, assigned IP addresses, and appropriate routing.
For this service, there are no changes required for either `suntan` or any of the clients.

>**NOTE**: The tools for this section are *NetworkManager* and *firewall-cmd*. See
your distribution's documentation for the details.

**Fedora**'s default network setup for VM instances is to configure their network to a private bridge that is NAT routed through the hypervisor.
This is identical to how the typical home Internet router is set up by the ISP.
The VM can make connections to the outside world but the outside world cannot make
connections to the VM.
We have to do something else because `pobox` must be visible to the rest of the home network.
This requires the setup of a second bridge on `suntan` that bypasses the hypervisor.
The steps are straight forward but we skip the details because there are plenty of
HOWTOs and tutorials on the subject. In sum, the following steps create what
we are looking for.

1. Create a bridge on `suntan`. This is separate from the one managed by
the hypervisor.
1. Attach `suntan`'s ethernet interface to the bridge.
This means that `suntan`'s ethernet interface address is now attached to the bridge.
1. When creating the VM for `pobox`, choose this new bridge for its
network connection instead of the hypervisor supplied one.

The end result looks like your typical network switch with two hosts attached to
it, one called `suntan` and the other called `pobox`.

### The Firewall
**Fedora** locks down network access with the *firewalld* service.
The typical installation only allows SSH and system management connections.
We need to add inbound access to `pobox` for the mail services.
We use the following commands to add them:

```
[root@pobox ~]# firewall-cmd --add-service=imap --permanent
[root@pobox ~]# firewall-cmd --add-service=imaps --permanent
[root@pobox ~]# firewall-cmd --add-service=smtp --permanent
[root@pobox ~]# firewall-cmd --add-service=smtps --permanent
[root@pobox ~]# firewall-cmd --permanent --add-service=managesieve
```

**Note:** We do not need to do this for `suntan`, the VM host
because its own interface was "given" to the bridge so there is no
packet routing through the `suntan` network stack and therefore
subject to the firewall rules set for `suntan`.
Think of the bridge as logically like a network switch sitting on
the floor outside the cabinet and connected to both `suntan` and `pobox`.

Checking our work, we see:

```bash
[root@pobox ~]# firewall-cmd --list-all
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

## Filesystem Configuration
We have two choices for storing user's mail. The first is using NFS.
With this choice, the mail store can be anywhere reachable by NFS.

Before we export a filesystem to our mail server, we have to create it.
Allocating disk space and formatting it into a filesystem and then
mounting it on `/srv/dovecot` is an exercise left to the reader.
This is the key part of `suntan`'s local filesystem that we will be working with.

```bash
[root@suntan ~]# cat /etc/fstab
...
UUID=xxxxxx         /srv/pgdata   btrfs   subvol=pgdata   0 0
UUID=xxxxxx         /srv/dovecot  btrfs   subvol=dovecot  0 0
...
LABEL=suntan_str    /home         btrfs   defaults,subvol=home    1 2
```

Our filesystem is a BTRFS *subvolume*. That is just my practice. It is
convenient. For the sake of the discussion, whatever it ends up being,
the root of that tree resides at `/srv/dovecot`.

Before we export anything, we can first get the basic skeleton of a mailstore
in place.
Note that the owner and group of dovecot is `97` in the commands below.
This is the uid/gid pair assigned by the **Fedora** install of `dovecot` on `pobox`.
It may be different on another distribution.
No matter, use the IDs that the install assigns.
We do not see the user name here because we are looking at it from `suntan` not `pobox`.

### Create Virtual Domain
This installation can support multiple *virtual domains* which are the destination in `dovecot`
where email arrives.
There are two tasks required to set one up.
First, a home directory for the domain is created.
That is what this section describes.
It would be nice to have this a part of what `postdove` does but that would be
messier than simply doing these manual steps given that setting up these domains
is not an every day task.
The second part is creating the domain in the database and then the mailboxes within that domain.
That is described in [Postdove Administration](admin.md).

At this point, `/srv/dovecot` is empty.
A directory for each served *virtual domain* will be a sub-directory under `/srv/dovecot`.
We will only be creating one directory `/srv/dovecot/example.com` for the domain `example.com`.
If the service is supporting multiple virtual domains,
we would be creating them using the same commands, i.e. `/srv/dovecot/homey.org` for domain `homey.org`.

```
[root@suntan ~]# cd /srv/dovecot
[root@suntan dovecot]# chown 97.97 .
[root@suntan dovecot]# mkdir example.com
[root@suntan dovecot]# chmod 777 example.com
[root@suntan dovecot]# chown 97:97 example.com
[root@suntan dovecot]# chmod o+t example.com
```

If we look closely, we see that the `example.com` directory is wide open.
This is because when `dovecot` processes the first email, either by LMTP or IMAP/POP3,
it does so as the authenticated user in `dovecot`.
This is the same behavior as for `/tmp` using the *sticky bit* on the directory.
It limits the access rights so a user can create, write, rename, or delete
files and directories only if they owned by that user.
The top level directory for an individual user's account will be created
by `dovecot` on behalf of the
authenticated user using the user's uid/gid credentials.
That directory, for example,
for user `bob@example.com` would be `/srv/dovecot/example.com/bob`.
It would also be created with mode 0700 which only allows access to user
`bob@example.com`.
This is not a security issue because ordinary mail users can only access this directory via `dovecot` which
enforces additional access controls and they cannot access any user mail directories other than their own because the directory's mode is `drwx------`.
The end result looks like this:

```bash
[root@suntan ~]# ls -la /srv/dovecot/
total 2931064
drwxr-xr-t. 1   97   97         66 Jun 18 08:44 .
drwxr-xr-x. 1 root root         26 Feb 11  2019 ..
drwxrwxrwt. 1   97   97         26 Jun 16 16:24 example.com
```
### Exporting The Mail Storage 
The next step is to setup the export of this filesystem to the VM.
There are two choices, use the traditional method of exporting the filesystem
via NFS. This is how one would do it when using a separate server instead of a VM.
The second choice is to use a *passthrough* filesystem method supported by the
hypervisor's emulator function. Either configuration will give an identical
filesystem layout from the mail server's point of view.

If you want to use NFS, move next to [Mail Storage Over NFS](nfs_storage.md).
If your distribution is recent enough so you can use *VirtioFS*,
proceed to [Mail Storage Over VirtioFS](virtiofs_storage.md).
Once you have completed the setup, return here to test everything.

## Testing the System Configuration
We check our work by adding and removing files and directories in the
`/srv/dovecot/example.com` directory.
They should be owned by the ordinary user we are logged in as. If we change to
a different user, we should see the same behavior for the files created by the
new user but the new user should not be able to remove any other user's files.

At this point we are done with system configuration and it is time to move on to
[Database Creation](database_setup.md).
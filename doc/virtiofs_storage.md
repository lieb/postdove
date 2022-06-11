# Mail Storage Over VirtioFS
This is the configuration for a system based on *Virtio-FS*.
This is a relatively new addition to Linux virtualization.
Nearly all virtual machines (VM) used for services need storage *somewhere* in addition
to the filesystem images containing the system software.
One choice is a network filesystem like NFS or CIFS.
They are configured with time tested tools and procedures developed long before VM
technologies were available.
However, what is fine when the storage is only accessible via the
network is a lot of extra work and overhead when everything is within the same
physical box.
A memory bus is orders of magnitude faster than a network link and its enabling
network protocols.
The *VirtioFS* filesystem solves that problem.
It is based on the Filesystem in User Space (FUSE) technology which connects the
VM virtualization directly to the server filesystem.

Our examples will show the bits from the NFSv4 installation commented out to
highlight what we do not need from NFS.
This does not imply that one must do one before the other.
One can safely ignore them or consider them simply comments that show
the NFS bits that are not necessary.

>*NOTE:* This feature is only available on distributions with a Linux kernel since
V5.4, QEMU 5.5 and libvirt 6.2. Anything older does not have all the parts (yet).

## Setting up the Virtual Machine
The virtual machine subsystem, like so much of software is layered.
It is also true that those layers evolve new releases on their own time usually
in the order bottom to top.

Setting up the initial VM is an exercise left to the reader.
Using whatever preferred tool, the VM should have sufficient memory, virtual
CPUs and image filesystems to support the load. The key configuration issue
in this base configuration is that the networking has been set up to use
the bridge that allows the VM to be directly visible to the local network.

This functionality is relatively new although most distributions now fully
support it.
We assume here that not all the pieces are in place in order to show what
really goes on under the covers in the fancy GUI virtualization tools.
Although the kernel, QEMU, and libvirt have been updated enough to support
*VirtioFS*, we will skip the GUI interfaces which could simplify
a few configuration details.
This means some extra manual labor on our part.

## Creating a Virtual Filesystem
We make changes in three places. First, we do not need the *bind* mount in
`/etc/fstab`. It is the last line in the display of the mounts
shown in [Mail Storage Over NFS](nfs_storage.md).
The subvolume mount of `/srv/dovecot` is enough.
If you had the *bind* mount, remove it.
We also no longer need the companion line in `/etc/exports`.
Remove that as well. We do not need to export things into the ether where they
can be exploited.
This is one advantage of *VirtioFS*. Nothing is network visible.

We do, however have to make this filesystem visible to the VM running `pobox`.
To do that, we do the second step in the virtual machine management.
We will do the following manual steps to edit to the `pobox`
configuration file.
Depending on which GUI you have available, it ends up doing the same thing
under the covers to make the filesystem available to the VM client.

Everything we need to configure is in the XML file that describes `pobox`.
As with other configuration parameters, we break this up into two hunks
for the discussion.
Everything we will edit is in `/etc/libvirt/qemu/pobox.xml`.
We use the configuration we ran previously with NFSv4 and simply add what
we need. Note that we did not have to do anything special at the VM level
for NFS. That was all conventional system administration.
On to the changes:

The first task is to enable a `memfd` capability in the VM emulation itself.

```bash
[root@suntan qemu]# diff -u pobox.xml.orig pobox.xml
--- pobox.xml.orig      2021-06-27 14:11:01.638419310 -0700
+++ pobox.xml   2021-06-28 12:57:13.271272354 -0700
@@ -9,6 +9,10 @@
   <name>pobox</name>
   <uuid>c6604419-2a94-4438-b610-ef2874b4db0b</uuid>
   <memory unit='KiB'>1048576</memory>
+  <memoryBacking>
+    <source type="memfd"/>
+    <access mode="shared"/>
+  </memoryBacking>
   <currentMemory unit='KiB'>1048576</currentMemory>
   <vcpu placement='static'>2</vcpu>
   <os>
```
This first change is near the top of the file.
Look for the tags in your file to find where this addition is placed.

We set up memory backing store because *VirtioFS* supports `mmap` and file sharing
and it needs somewhere to put the bits.
There are multiple options in the documentation but we chose `memfd` because this
uses the hypervisor's kernel memory (a memory file) that has backing store on the
system's *swap*. In the case of current *Fedora*, that is first `zram` followed
by disk based swap. It is fast and simple and requires less configuration tweaking.
This is optional but useful.

The second task is to make the filesystem available from the server to the VM, a task not all that
different than the work we would do for an NFS export.

   ```bash
@@ -110,5 +114,14 @@
     <memballoon model='virtio'>
       <address type='pci' domain='0x0000' bus='0x00' slot='0x07' function='0x0'/>
     </memballoon>
+    <filesystem type='mount' accessmode='passthrough'>
+      <driver type='virtiofs' queue='1024'/>
+      <binary path='/usr/libexec/virtiofsd' xattr='on'>
+        <cache mode='always'/>
+        <lock posix='on' flock='on'/>
+      </binary>
+      <source dir='/srv/dovecot'/>
+      <target dir='mailstore'/>
+    </filesystem>
     </devices>
 </domain>

```

This change is at or near the bottom of the file.

This is where we open up the virtual system to the real system. We will look at
the parts one by one.
* It is all contained within a `filesystem` that is `mount`ed and is a
`passthrough`, i.e. it goes straight from the VM kernel to the host kernel by
way of the emulator, not a filesystem+network stack.
* The type of the filesystem driver is `virtiofs`.
A request `queue` is needed by the FUSE side to handle the requested filesystem operations.
This should be big enough at `1024` slots.
* The `binary` has some important properties:
  * `xattr` is enabled. This allows passing *selinux* labels back and forth.
  * `cache` mode is `always` for performance. Some sharing environments cannot use
  this but we can. See the `virtiofsd` documentation for details.
  * Both `posix` and `flock` (BSD) locks are enabled.
* The `source` is `/srv/dovecot`. This is the the equivalent to an *export* in
NFS terms. The *VirtioFS* service contains all references to this to just
`/srv/dovecot` so there is no leakage of anything else on `suntan` just in
case the VM decided to go rogue.
* The `target` is the name/handle, analogous to a disk name or UUID, that the VM references.
We will see `mailstore` on the VM side where we configure filesystem mounts.

File locking can be problematic on non-local filesystems.
This includes not only *NFS* or *CIFS* but, as it turns out, *VirtioFS* as well.
In the *VirtioFS* case, there are potential deadlocks with blocking locks.
See [Dovecot Configuration](dovecot_configuration.md) for how to handle this
situation in `dovecot`.
One choice here is to leave the `<lock ... />` directive out resulting in the
default being `off` for both types of locks.
We have left it in because *some* of the locking API does work and the
system will return a *not supported* error for the ones that do not.

**NOTE:** This may change in future versions of `virtiofsd`, the QEMU server that
manages the filesystem.

This sets up everything for the hypervisor on `suntan` to launch the `pobox` VM.
The next step is done on `pobox` to mount the filesystem.
Let us check out the client side. We look at `/etc/fstab`.

```bash
root@pobox etc]# cat /etc/fstab
/dev/mapper/fedora-root /                       xfs     defaults        0 0
UUID=43b2a048-5b62-4648-a67c-95dac4f84075 /boot                   ext4    defaults        1 2
/dev/mapper/fedora-swap swap                    swap    defaults        0 0

# mailstore via virtif-fs
mailstore               /srv/dovecot            virtiofs defaults       0 0

```

The first three lines are not of interest to us.
They are the standard layout for a VM on *Fedora* with a *COW2* image for the VM.

The interesting line is the last one which mounts `/srv/dovecot` from `suntan`.
Here, `mailstore`, the `target` tag in our configuration is mounted on
`/srv/dovecot` as filesystem type `virtiofs`. Nothing else is special.

>NOTE: If you have migrated from the NFSv4 configuration, this would have been
a symbolic link. Remove the link and create a directory in its place.

This solves a big startup problem in the NFSv4 configuration.
Given that the host `suntan` is bringing up the `pobox` VM at the same time
everything else is settling down, there is a race where the NFS service
cannot find the DNS record for `pobox` because its VM hasn't come up far
enough to contact the DHCP service for an address. I use dynamic DNS so I
can have all network configuration in one place which caused this race
that can only really be solved by using `pobox`'s IP address in
`suntan`'s `/etc/exports`, something I am loath to do. *Virtio-FS*
completely bypasses this problem - and it is blazingly fast and fully
POSIX unlike NFS.

This is all well and good except for access controls. *Selinux* is still
in charge.

### Access Controls
In the NFS configuration, we had to enable the `use_nfs_home_dirs` boolean
in order for the VM to use NFS.
We no longer need this and a prudent migration would disable it to the
default again.

We provide access via the `xaddr` option we set up in configuring
`virtiofsd`.
Whatever labels present in `/srv/dovecot` at the time the BTRFS subvolume
was created and mounted are passed to the kernel on `pobox`.
Therefore, we need to ensure that the labels on `suntan` are correct.
This is done by *selinux* tools.
The `chcon` command changes the *Selinux* security context of a file.
The security context is identified by the label, in this case `mail_home_rw_t`.
On `suntan`, do the following:

```bash
[root@suntan srv]# chcon -R -t mail_home_rw_t dovecot
[root@suntan srv]# ls -lZa
total 0
drwxr-xr-x. 1 root     root     system_u:object_r:var_t:s0            26 Jan 25 22:05 .
dr-xr-xr-x. 1 root     root     system_u:object_r:root_t:s0          162 Jun 26 17:39 ..
drwxr-xr-t. 1       97       97 system_u:object_r:mail_home_rw_t:s0   66 Jun 18 08:44 dovecot
drwx------. 1 postgres postgres system_u:object_r:postgresql_db_t:s0 300 Jun 26 19:43 pgdata
```
This recursively changes the label for `dovecot` and all its children
from `var_t` to `mail_home_rw_t`.
There are a few things to note. First, since we are doing these commands while in
`/srv`, we are seeing multiple mount points. One is ours and the other, another
subvolume, is used for data storage by the *postgresql* server.
As you can see, `.` is also `/srv`, the point in the root filesystem where all these
filesystems are mounted. It's label is `var_t` which would get inherited by anything
created below it. We've already made the change for *postgresql*.
We do the same for `dovecot` (remember, it doesn't run on `suntan`).
The `mail_home_rw_t` label is used by any mail service to access a user's email files.
The *selinux* configuration for `dovecot` can access files with this label.
If you forget this step, you will get complaints in the logs similar to the ones
you would get if you forgot the boolean in the NFS configuration.

See [Database Creation](database_setup.md) for instructions for granting access
to the mail storage if SELinux is denying access to it from the VM.

We are now done with setting up the mounting of the mailstore to the VM. For the final
steps return to [System Configuration](system_configuration.md) and test what we
have set up.

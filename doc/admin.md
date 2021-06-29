# Postdove Administration

You wouldn't believe what you've got to do...
We will populate the database in the following order:

1. The domain name used for the virtual users. This type of domain
must be created before any users in that domain, i.e. the domain does
not automatically get added when a virtual mail user is created.
1. We can now add users. Note that this just adds the user to the database.
Other actions must be done before the account is usable for mail. This is
enough for ```dovecot``` to start serving mail.
1. Add aliases. These will be used by ```postfix``` to process and deliver
email to ```dovecot```. There are two types of aliases, ```alias``` and ```virtual```.
The easiest way to enter them is to ```import``` using the file format that
```postfix``` uses.

Single binary to associate an Elastic IP address with an instance
-----------------------------------------------------------------

Fed up with using the web console to change an Elastic IP?

If you have a machine pointed to by an Elastic IP address, and you want to
automatically update that Elastic IP address with minimal fuss, this program
can help you achieve it.

The goal is to associate the Elastic IP address with the machine in a secure
manner without any human contact, just through the configuration of the
machine. This is useful for making it possible to replace the machine trivially.

## Usage

AWS has the notion of "user data" associated with an instance. Operating systems
ship with [`cloud-init`](http://cloudinit.readthedocs.org/en/latest/) which
read the user data and take action.

AWS also has a notion of a "machine role", which allowed a machine to take
specific actions. In this case, the machine role needs to have the capability
to associate an address with itself.

Given these facts, one can

```
curl https://github.com/
associate-address --ip 54.12.34.56
```

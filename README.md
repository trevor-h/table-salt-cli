# tableSALT CLI

A ready to serve GO based remote client for Saltstack using SSH with support for bastion/jump hosts. Run the 'salt' command like on the salt-master from your workstation.

# Features
  - Support Linux and Windows
  - Multiple authentication support: agent (Linux only), private key and password
  - Built-in support to punch through bastion/jump servers automatically
  - 100% native salt command. All arguments passed directly to the salt command on the salt-master
  - No client side dependencies or setup. A single binary executable contains all that's needed
  - Supports any version of Saltstack without patching or other complex changes. Absolutely nothing to install on salt-master

# Who's This For?
Saltstack ships with external authentication support as well as a REST API via netapi so you might be wondering the use case. While supported many organizations are hesitant (and rightly so) to open their salt-master(s) to traditionally less secure protocols (SSH vs HTTP).

For reasons of security, old habits, etc this leaves many organizations relying on standard SSH to salt-master and run the salt command directly as root, via sudo, eauth, publisher ACL. Adding bastion/jump hosts to the mix leaves administrators hopping from shell to shell (usually) or coming up with their own cobbled solution of SSH forwarding, remote execution, etc to make their lives easier.

This tool is for anyone who:
  - Isn't particular comfortable with, or not allowed due to policy to enable HTTP in Saltstack
  - Already has a cobbled together solution with separate tunnels, etc to "get to the salt-master"
  - Is a Windows administrator wanting to work in a familiar environment, or a Linux administrator more comfortable with personal workstation setup/preferences than often generic profiles found on bastion and service hosts.
  - You want the benefits of Saltstack with similar operation to tools like Ansible

# How Does It Work?

The utility works identically to the salt command and syntax is 100% the same. In fact, all that is being done is really remotely executing the 'salt' command over SSH and appending all the arguments passed to this CLI tool.

On the salt-master you might run:

```sh
$ salt '*' cmd.run 'some command' --output=json --async
```

On your workstation with table-salt this would be:
```sh
$ table-salt '*' cmd.run 'some command' --output=json --async
```

Support for any and all arguments and parameters of the salt command now and in the future is automatic. Everything is done over SSH using the same connection methods and credentials you are likely using already.

For authentication SSH agents (Linux only currently), unencrypted private keys and passwords are supported. There is also automatic support to traverse through a jump/bastion host to reach the salt-master. Just enable the option in the configuration file. Different users and/or credentials can be used between bastion hosts and salt-masters if your situation requires it.

In testing with the extra hop of a bastion hosts between the user workstation and salt-master there is about 180-240ms overhead vs running locally on the salt-master. This generally makes it acceptable to use even for quick executions. All the overhead is SSH and is comparable roughly to execution RTT with Ansible.

# Specific Requirements

The supported Salt-master setups include:
  * [Using eauth with either PAM or LDAP](https://docs.saltstack.com/en/latest/topics/eauth/index.html)
  * [Using publisher ACL system](https://docs.saltstack.com/en/latest/ref/publisheracl.html)
  * Sudo to root with no password
  * [Run salt as non-root user](https://docs.saltstack.com/en/latest/ref/configuration/nonroot.html) with approriate privileges for authorized users

**If you need to sudo to root with a password in your environment then unfortunately this utility will not work for you currently.**

The reason for this 'limitation' is currently the utility has no ability to do interactive sudo. Due to the undesirable behavior of all executions being done by the root (or custom salt user) there is no current plan to support this type of account sharing environment setup. If someone submits a pull request to add this it will be accepted with pleasure, it's just not a priority.

If you are using a bastion host to connect to the salt-master you could run into issues with a restritive SSH configuration preventing you from reaching the salt-master. If you've used SSH port forwarding before then you should not have a problem as this utility uses the same principle.

# Installation

You can clone/download this repo and build it yourself. Or if you prefer you can always find [binaries for Linux and Windows available here](https://github.com/trevor-h/table-salt-cli-bin). Setup is very simple. Place the executable anywhere you like (ideally somewher in your system or user path).

# Configuration

The configuration file must be located in the same directory as the command, or in the user's home directory as a hidden file (e.g. /home/jsmith/.table-salt-conf.json). If you have chosen the binary version there is nothing more to setup. **Just make sure to always have all configuration fields present, even if they are not applicable and are are empty strings!**

Below is a commented example configuration file:
```sh
{
  "UseJump": false, // Use a bastion/jump host defined with Jump* below
  "HostKeyCheck": false, // Do a host key check
  "Auth": "agent", // Authentication type: agent, key, password
  "JumpUsername": "", // Optional. Required if UseJump:true User to use on bastion/jump host
  "JumpPassword": "", // Optional. Leave an empty field if doing 'agent' or 'key' auth
  "JumpPrivateKey": "", // Optional. Leave an empty field if doing 'agent' or 'password' auth
  "JumpServer": "", // Optional. Leave en ampty field if not using bastion/jump host. Specified as <host>:<port>
  "RemoteEndpoint": "192.168.1.202:22", // Required. This is the host/ip of the salt-master. Specified as <host>:<port>
  "RemoteUsername": "trevor", // Required. This is the user to login to salt-master with
  "RemotePassword": "", // Optional. Leave an empty field if doing 'agent' or 'key' auth
  "RemotePrivateKey": "" // Optional. Leave an empty field if doing 'agent' or 'password' auth
}
```


# Other Plans

A few more features are planned currently including:

  * Support for Windows SSH agent (Pageant)
  * Support for decrypting private keys at execution via configuration or keyboard-interative
  * Environment variable support for configuration path to allow both custom config locations as well as the ability to quickly and easily work on different salt-masters
  * Support for other Saltstack commands like salt-run, salt-key

# Known Issues

  - On this initial release error handling is just functional, and needs improvement
  - SSH host key checks were quickly thrown in and not fully tested

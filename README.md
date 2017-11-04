# tableSALT CLI

A ready to serve GO based remote client for Saltstack using SSH. Run the 'salt' command like on the salt-master from your workstation.

# Features
  - Support Linux and Windows
  - Multiple authentication support: agent (Linux only), private key and password
  - Support for sudo (password or passwordless)
  - Built-in support to punch through bastion/jump servers automatically
  - 100% native salt command compatible. All arguments passed directly to the salt command on the salt-master
  - No client side dependencies or setup. A single binary executable contains all that's needed
  - Supports any version of Saltstack without patching or configuration changes. Absolutely nothing to install on salt-master

# Who's This For?
Saltstack ships with external authentication support as well as a REST API via netapi so you might be wondering the use case. While supported many organizations are hesitant (and rightly so) to open their salt-master(s) to traditionally less secure protocols (SSH vs HTTP).

For reasons of security, old habits, etc this leaves many organizations relying on standard SSH to salt-master and run the salt command directly as root, via sudo, eauth, publisher ACL. Adding bastion/jump hosts to the mix leaves administrators hopping from shell to shell (usually) or coming up with their own cobbled solution of SSH forwarding, remote execution, etc to make their lives easier.

This tool is for anyone who:
  - Isn't particular comfortable with, or not allowed due to policy to enable HTTP in Saltstack
  - Already has a cobbled together solution with separate tunnels, etc to "get to the salt-master"
  - Is a Windows administrator wanting to work in a familiar environment, or a Linux administrator more comfortable with personal workstation setup/preferences than often generic profiles found on bastion and service hosts.
  - You want the benefits of Saltstack with similar operation to tools like Ansible, Fabric or plain old SSH remote executions

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

If you are using a bastion host to connect to the salt-master you could run into issues with a restritive SSH configuration preventing you from reaching the salt-master. If you've used SSH port forwarding before then you should not have a problem as this utility uses the same principle.

# Installation

You can clone/download this repo and build it yourself. Or if you prefer you can always find [binaries for Linux and Windows available here](https://github.com/trevor-h/table-salt-cli-bin). Setup is very simple. Place the executable anywhere you like (ideally somewhere in your system or user path).

Use an example configuration from the 'configExamples' directory to get started.

# Configuration

The configuration file must be located in the same directory as the command or specified via an environment variable TABLESALTCONF. If you have chosen the binary version there is nothing more to setup. **Just make sure to always have all configuration fields present, even if they are not applicable and are are empty strings!**

Below is a commented example configuration file:
```sh
{
  "UseJump": false, // Use a bastion/jump host defined with Jump* below
  "UseSudo": false, // Use sudo when executing salt on RemoteEndpoint
  "SudoType": "password", // SudoType: password, nopassword. Used with UseSudo. Password for sudo taken from RemotePassword entry
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
Use one of the example configurations found in 'configuration_examples' directory to help illustrate the configuration better.

# Other Plans

A few more features are planned currently including:

  * Support for Windows SSH agent (Pageant)
  * Support for decrypting private keys at execution via configuration or keyboard-interative
  * Support for other Saltstack commands like salt-run, salt-key

# Known Issues

  - On this initial release error handling is just functional, and needs improvement
  - SSH host key checks were quickly thrown in and not fully tested

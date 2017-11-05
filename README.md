# tableSALT CLI

A ready to serve GO based remote client for Saltstack using SSH. Run the 'salt' command like on the salt-master from your workstation.

# Features
  - Supports Linux and Windows with both source and binary releases. Make Saltstack accessible for everyone
  - Multiple authentication support: agent (Linux only), private key and password
  - Support for sudo (password or passwordless)
  - Built-in support to punch through bastion/jump servers automatically
  - 100% native salt command compatible. All arguments passed directly to the salt command on the salt-master
  - No client side dependencies or setup. A single binary executable contains all that's needed
  - Works with any version of Saltstack without patching or configuration changes. Absolutely nothing to install on salt-master
  - Support for salt-key, salt-run and also any arbitrary command execution on salt-master
  - Virtual wrapper execution modules provide enhanced functionality like copy a local file directly to multiple minion targets in one step

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
Want to run salt-key:
```sh
$ table-salt --tsk --list-all
```
The **--tsk** positional argument informs table-salt to use salt-key instead of salt. After the parameter include any normal salt-key execution parameters.

Want to run salt-run:
```sh
$ table-salt --tsr jobs.print_job 20171105114119937597
```
The **--tsr** positional argument informs table-salt to use salt-run instead of salt. After the parameter include any normal salt-run execution parameters.

Want to run <ANYTHING> on salt-master:
```sh
$ table-salt --tse echo 'this is running on salt-master'
```
The **--tse** positional argument informs table-salt to pass all proceeding arguments directly as execution. Baically this is just remote execution over SSH on the salt-master.

Support for any and all arguments and parameters of the salt (and other) command(s) now and in the future is automatic. Everything is done over SSH using the same connection methods and credentials you are likely using already.

For authentication SSH agents (Linux only currently), unencrypted private keys and passwords are supported. There is also automatic support to traverse through a jump/bastion host to reach the salt-master. Just enable the option in the configuration file. Different users and/or credentials can be used between bastion hosts and salt-masters if your situation requires it.

In testing with the extra hop of a bastion hosts between the user workstation and salt-master there is about 180-240ms overhead vs running locally on the salt-master. This generally makes it acceptable to use even for quick executions. All the overhead is SSH and is comparable roughly to execution RTT with Ansible.

#### Special 'Virtual Module Wrappers'

Ever had a text file sitting locally on your workstation you wanted to move to one or more remote hosts?  More than one file to more than one host gets ugly pretty quickly!

Enter "virtual module wrappers". This is just a fancy way of taking locally sourced things (like files) and "doing stuff" with them through existing Saltstack execution modules.

**tablesalt.cp Virutal Module**
This virtual module is simply a wrapper for hashutil.base64_decodefile. When executed against a target before running the salt command the source file is base64 encoded and becomes input for the real Saltstack execution module.

For example:
```sh
$ table-salt '*' tablesalt.cp /some/local/filename /destination/path/filename
```
This will take the local file (if it exists) and encode it's contents and the actual execution on the salt-master will be:
```sh
$ salt '*' hashutil.base64_decodefile instr="<base64 encoded string of /some/local/filename>"  outfile="/destination/path/filename"
```
This module is still in a preview state and may have some issues with larger binary files. Use should be restricted to small text files, configs and other headaches to distribute as a one-off.

# Specific Requirements

If you are using a bastion host to connect to the salt-master you could run into issues with a restritive SSH configuration preventing you from reaching the salt-master. If you've used SSH port forwarding before then you should not have a problem as this utility uses the same principle.

# Installation

You can clone/download this repo and build it yourself. Or if you prefer you can always find [binaries for Linux and Windows available here](https://github.com/trevor-h/table-salt-cli-bin/releases). Setup is very simple. Place the executable anywhere you like (ideally somewhere in your system or user path).

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

# Known Issues

  - SSH host key checks were quickly thrown in and not fully tested

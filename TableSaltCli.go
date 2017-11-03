package main

import (
    "fmt"
    "os"
    "log"
    "net"
    "strings"
    "bytes"
    "runtime"
    "encoding/json"
    "golang.org/x/crypto/ssh"
    "golang.org/x/crypto/ssh/agent"
)

var configuration = Configuration{}
var bsshClientConnection *ssh.Client = nil
var sshConfig *ssh.ClientConfig = nil

type Configuration struct {
    Auth string
    UseJump bool
    JumpUsername string
    JumpPassword string
    JumpPrivateKey string
    JumpServer string
    LocalEndpoint string
    RemoteEndpoint string
    RemoteUsername string
    RemotePassword string
    RemotePrivateKey string
}

func SSHAgent() ssh.AuthMethod {
    if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
        return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
    }
    return nil
}

func main() {

    // Open configuration JSON
    file, _ := os.Open("conf.json")
    decoder := json.NewDecoder(file)
    configuration = Configuration{}
    err := decoder.Decode(&configuration)
    if err != nil {
      fmt.Println("error:", err)
    }

    // Parse salt command args
    args := os.Args[1:]

    for i := 0; i < len(args); i++ {
        args[i] = "\""+args[i]+"\""
    }

    // Format salt command
    saltCommand := "salt " + strings.Join(args, " ")

    // Use bastion/jump server?
    if configuration.UseJump {

        bsshAuthMethod := []ssh.AuthMethod{SSHAgent()}

        if configuration.Auth == "key" && len(configuration.JumpPrivateKey) > 0 {
            jumpKey, err := ssh.ParsePrivateKey([]byte(configuration.JumpPrivateKey))
            if err != nil {
                fmt.Println("Could not parse private key file. Check the path and ensure it is not encrypted.")
                os.Exit(1)
            }
            bsshAuthMethod[0] = ssh.PublicKeys(jumpKey)
        } else if configuration.Auth == "agent" && runtime.GOOS != "windows" {
            bsshAuthMethod[0] = SSHAgent()
        } else if configuration.Auth == "password" && len(configuration.JumpPassword) > 0 {
            bsshAuthMethod[0] = ssh.Password(configuration.JumpPassword)
        } else {
            fmt.Println("No supported authentication modes available/supported. Double check your configuration.")
            os.Exit(1)
        }

        bsshConfig := &ssh.ClientConfig{
            User: configuration.JumpUsername,
            Auth: bsshAuthMethod,
            HostKeyCallback: ssh.InsecureIgnoreHostKey(),
        }

        bsshClientConnection, err = ssh.Dial("tcp", configuration.JumpServer, bsshConfig)
        if err != nil {
            log.Fatal(err)
            os.Exit(1)
        }

    }

    sshAuthMethod := []ssh.AuthMethod{SSHAgent()}

    if configuration.Auth == "key" && len(configuration.RemotePrivateKey) > 0 {
        remoteKey, err := ssh.ParsePrivateKey([]byte(configuration.RemotePrivateKey))
        if err != nil {
            fmt.Println("Could not parse private key file. Check the path and ensure it is not encrypted.")
            os.Exit(1)
        }
        sshAuthMethod[0] = ssh.PublicKeys(remoteKey)
    } else if configuration.Auth == "agent" && runtime.GOOS != "windows" {
        sshAuthMethod[0] = SSHAgent()
    } else if configuration.Auth == "password" && len(configuration.RemotePassword) > 0 {
        sshAuthMethod[0] = ssh.Password(configuration.RemotePassword)
    } else {
        fmt.Println("No supported authentication modes available/supported. Double check your configuration.")
        os.Exit(1)
    }

    sshConfig = &ssh.ClientConfig{
        User: configuration.RemoteUsername,
        Auth: sshAuthMethod,
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }

    if configuration.UseJump {

        jumpConnection, err := bsshClientConnection.Dial("tcp", configuration.RemoteEndpoint)
        if err != nil {
            log.Fatal(err)
            os.Exit(1)
        }

        ncc, chans, reqs, err := ssh.NewClientConn(jumpConnection, configuration.RemoteEndpoint, sshConfig)
        if err != nil {
            log.Fatal(err)
            os.Exit(1)
        }

        sshClientConnection := ssh.NewClient(ncc, chans, reqs)

        session, err := sshClientConnection.NewSession()
        if err != nil {
            fmt.Println(err.Error())
        }
        defer session.Close()
        var b bytes.Buffer
        session.Stdout = &b
        err = session.Run(saltCommand)
        fmt.Println(b.String())


    } else {

        sshClientConnection, err := ssh.Dial("tcp", configuration.RemoteEndpoint, sshConfig)
        if err != nil {
            log.Fatal(err)
            os.Exit(1)
        }

        session, err := sshClientConnection.NewSession()
        if err != nil {
            fmt.Println(err.Error())
        }
        defer session.Close()
        var b bytes.Buffer
        session.Stdout = &b
        err = session.Run(saltCommand)
        fmt.Println(b.String())

    }


}
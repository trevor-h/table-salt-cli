package main

import (
    "fmt"
    "os"
    "log"
    "net"
    "strings"
    "bytes"
    "runtime"
    "bufio"
    "io"
    "regexp"
    "path/filepath"
    "encoding/json"
    "golang.org/x/crypto/ssh"
    "golang.org/x/crypto/ssh/agent"
)

var saltCommand string
var configuration = Configuration{}
var bsshClientConnection *ssh.Client = nil
var hostKeyCallBackConfig ssh.HostKeyCallback = nil
var sshConfig *ssh.ClientConfig = nil

type Configuration struct {
    Auth string
    UseJump bool
    UseSudo bool
    SudoType string
    HostKeyCheck bool
    JumpUsername string
    JumpPassword string
    JumpPrivateKey string
    JumpServer string
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

func HostKeyCheck(remoteHost string) (ssh.HostKeyCallback) {

    host := remoteHost
    file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
    if err != nil {
        log.Fatal(err)
        os.Exit(1)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    var hostKey ssh.PublicKey
    for scanner.Scan() {
        fields := strings.Split(scanner.Text(), " ")
        if len(fields) != 3 {
            continue
        }
        if strings.Contains(fields[0], host) {
            var err error
            hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
            if err != nil {
                log.Fatalf("Error parsing %q: %v", fields[2], err)
            }
            break
        }
    }

    if hostKey == nil {
        log.Fatalf("No hostkey for %s. You can disable checks in the config by setting HostKeyCheck to false.", host)
        os.Exit(1)
    }

    return ssh.FixedHostKey(hostKey)

}

func setupJump() {

    err := error(nil)

    // Set SSH configuration
    bsshConfig := generateSshConfig("jump")

    bsshClientConnection, err = ssh.Dial("tcp", configuration.JumpServer, bsshConfig)
    if err != nil {
        log.Fatal(err)
        os.Exit(1)
    }

}

func generateSshConfig(configType string) (*ssh.ClientConfig) {

    var sshConfigUsername string
    var sshConfigPassword string
    var sshConfigPrivateKey string
    var sshConfigEndpoint string

    sshAuthMethod := []ssh.AuthMethod{SSHAgent()}

    if configType == "jump" {
        sshConfigUsername = configuration.JumpUsername
        sshConfigPassword = configuration.JumpPassword
        sshConfigPrivateKey = configuration.JumpPrivateKey
        sshConfigEndpoint = configuration.JumpServer
    } else {
        sshConfigUsername = configuration.RemoteUsername
        sshConfigPassword = configuration.RemotePassword
        sshConfigPrivateKey = configuration.RemotePrivateKey
        sshConfigEndpoint = configuration.RemoteEndpoint
    }

    if configuration.Auth == "key" && len(sshConfigPrivateKey) > 0 {
        remoteKey, err := ssh.ParsePrivateKey([]byte(sshConfigPrivateKey))
        if err != nil {
            fmt.Println("Could not parse private key file. Check the path and ensure it is not encrypted.")
            os.Exit(1)
        }
        sshAuthMethod[0] = ssh.PublicKeys(remoteKey)
    } else if configuration.Auth == "agent" && runtime.GOOS != "windows" {
        sshAuthMethod[0] = SSHAgent()
    } else if configuration.Auth == "password" && len(sshConfigPassword) > 0 {
        sshAuthMethod[0] = ssh.Password(sshConfigPassword)
    } else {
        fmt.Println("No supported authentication modes available/supported. Double check your configuration.")
        os.Exit(1)
    }

    if configuration.HostKeyCheck {
        hostSplit := strings.Split(sshConfigEndpoint, ":")
        hostKeyCallBackConfig = HostKeyCheck(hostSplit[0])
    } else {
        hostKeyCallBackConfig = ssh.InsecureIgnoreHostKey()
    }

    sshConfig := &ssh.ClientConfig{
        User: sshConfigUsername,
        Auth: sshAuthMethod,
        HostKeyCallback: hostKeyCallBackConfig,
    }

    return sshConfig

}

func generateSaltCommand() (string) {

    var passedArgs string
    execCommand := "salt"

    args := os.Args[1:]

    // Advanced feature handling
    for i := 0; i < len(args); i++ {

        // Check if alt exec + build command
        if args[i] == "--tsr" {
            execCommand = "salt-run"
        } else if args[i] == "--tsk" {
            execCommand = "salt-key"
        } else if args[i] == "--tse" {
            execCommand = ""
        } else {
            args[i] = "\""+args[i]+"\""
            passedArgs = passedArgs + " " + args[i]
        }

    }

    runCommand := execCommand + " " + passedArgs

    // Handle sudo if necessary
    if configuration.UseSudo {
        if configuration.SudoType == "nopassword" {
            runCommand = "sudo " + runCommand
        } else {
            if len(configuration.RemotePassword) > 0 {
                runCommand = "sudo " + runCommand + "\n"
            }
        }

    }

    return runCommand

}


func useJump() (string) {

    var commandResult string

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
    commandResult = executePtySession(session)

    return commandResult

}

func goDirect() (string) {

    var commandResult string

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
    commandResult = executePtySession(session)

    return commandResult

}

func executePtySession(sshSession *ssh.Session) (string) {

    var commandResult string

    modes := ssh.TerminalModes{
        ssh.ECHO:  0,
        ssh.TTY_OP_ISPEED: 14400,
        ssh.TTY_OP_OSPEED: 14400,
        ssh.IGNCR: 1,
    }

    if err := sshSession.RequestPty("vt100", 80, 40, modes); err != nil {
        log.Fatalf("request for pseudo terminal failed: %s", err)
    }

    if configuration.UseSudo && configuration.SudoType == "password" {

        sshOut, err := sshSession.StdoutPipe()
        handleError(err)
        sshIn, err := sshSession.StdinPipe()
        handleError(err)

        if err := sshSession.Shell(); err != nil {
            log.Fatalf("failed to start shell: %s", err)
        }

        // send sudo salt command
        writeSession(saltCommand, sshIn)
        // wait for password prompt. will break loop to return
        readBuffForString(sshOut, false)
        // send password when prompted. will break loop on command prompt
        writeSession(configuration.RemotePassword + "\n", sshIn)
        rawCommandResult := readBuffForString(sshOut, true)
        outRegex := regexp.MustCompile(`(.*)\n.*` + configuration.RemoteUsername + `.*(\$|#|>)`)
        commandResult = outRegex.ReplaceAllString(rawCommandResult, " ${1}")

    } else {

        var b bytes.Buffer
        sshSession.Stdout = &b
        sshSession.Run(saltCommand)
        outRegex := regexp.MustCompile(`^.*: (.*)`)
        commandResult = outRegex.ReplaceAllString(b.String(), "${1}")

    }

    return strings.TrimSpace(commandResult)

}

func readBuffForString(sshOut io.Reader, checkPrompt bool) string {

    buf := make([]byte, 1000)
    n, err := sshOut.Read(buf)
    waitingString := ""

    if err == nil {
        waitingString = string(buf[:n])
    }
    for err == nil {
        n, err = sshOut.Read(buf)
        waitingString += string(buf[:n])
        if err != nil {
            fmt.Println(err)
        }

        var sudoPromptRegex = regexp.MustCompile(`.*password for.*`)
        var shellPromptRegex = regexp.MustCompile(configuration.RemoteUsername + `.*\$`)

        // use regexes to determine when to break from receiving output
        if configuration.UseSudo && configuration.SudoType == "password" {
            if sudoPromptRegex.MatchString(waitingString) {
                break
            } else if checkPrompt && shellPromptRegex.MatchString(waitingString) {
                break
            }
        }

    }

    return waitingString
}

func writeSession(cmd string, sshIn io.WriteCloser) {
    _, err := sshIn.Write([]byte(cmd + "\r"))
    handleError(err)
}

func stripEmptyLines(inString string) (string) {
        StripRegex := regexp.MustCompile(`\n\n`)
        return StripRegex.ReplaceAllString(inString, "\n")
}

func handleError(err error) {
    if err != nil {
        panic(err)
    }
}

func main() {

    // Open configuration JSON
    configPath := "ts_conf.json"
    if len(os.Getenv("TABLESALTCONF")) > 0 {
        configPath = os.Getenv("TABLESALTCONF")
    }
    file, _ := os.Open(configPath)
    decoder := json.NewDecoder(file)
    configuration = Configuration{}
    err := decoder.Decode(&configuration)
    if err != nil {
      fmt.Println("Error: Invalid or missing configuration. ", err)
    }

    // Parse salt command args
    saltCommand = generateSaltCommand()

    // Connect to bastion/jump server if necessary
    if configuration.UseJump {
        setupJump()
    }

    // Set SSH configuration
    sshConfig = generateSshConfig("remote")

    // Execute salt command
    var saltOutput string
    if configuration.UseJump {
        saltOutput = useJump()
    } else {
        saltOutput = goDirect()
    }

    fmt.Println(saltOutput)

}
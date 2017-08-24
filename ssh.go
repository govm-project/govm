package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

type password string

func (p password) Password(user string) (password string, err error) {
	return string(p), nil
}

// TODO: Reduce cyclomatic complexity
func getNewSSHConn(username, hostname, key string) { // nolint: gocyclo
	//var hostKey ssh.PublicKey

	privateKeyBytes, err := ioutil.ReadFile(key)
	if err != nil {
		log.Fatalf("Error on reading private key file: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	// Create client config
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			//ssh.Password("password"),
			ssh.PublicKeys(signer),
		},
		//HostKeyCallback: ssh.FixedHostKey(hostKey),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	config.SetDefaults()
	// Connect to ssh server
	conn, err := ssh.Dial("tcp", hostname+":22", config)
	if err != nil {
		log.Fatal("unable to connect: ", err)
	}
	defer func() {
		err := conn.Close()
		// TODO: Change to warning when log package is changed
		log.Println(err)
	}()

	// Create a session
	session, err := conn.NewSession()
	if err != nil {
		log.Fatal("unable to create session: ", err)
	}
	defer func() {
		err := session.Close()
		// TODO: Change to warning when log package is changed
		log.Println(err)
	}()

	// Set IO
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	in, err := session.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// Request pseudo terminal
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		log.Fatalf("request for pseudo terminal failed: %s", err)
	}

	// Start remote shell
	if err := session.Shell(); err != nil {
		log.Fatalf("failed to start shell: %s", err)
	}

	// Accepting commands
	for {
		reader := bufio.NewReader(os.Stdin)
		str, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading command")
		}
		_, err = fmt.Fprint(in, str)
		if err != nil {
			log.Println("Error reading command")
		}
	}
}

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

type password string

func (p password) Password(user string) (password string, err error) {
	return string(p), nil
}

//PublicKeyFile generate authentication method from a given public ssh key
func PublicKeyFile(file string) (ssh.AuthMethod, error) {
	file, _ = filepath.Abs(file)
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(key), nil
}

func getNewSSHConn(username, hostname, key string) {
	//var hostKey ssh.PublicKey

	privateKeyBytes, err := ioutil.ReadFile(key)
	if err != nil {
		log.Fatal("Error reading private key file. %v", err)
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
	defer conn.Close()
	// Create a session
	session, err := conn.NewSession()
	if err != nil {
		log.Fatal("unable to create session: ", err)
	}
	defer session.Close()

	// Set IO
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	in, _ := session.StdinPipe()

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
		str, _ := reader.ReadString('\n')
		fmt.Fprint(in, str)
	}
}

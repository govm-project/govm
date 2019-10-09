package docker

import (
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"github.com/govm-project/govm/internal"
	"github.com/govm-project/govm/pkg/homedir"
	"github.com/govm-project/govm/pkg/termutil"
)

//SSHVM initializes the SSH bits for the vm ssh connection
// nolint: funlen
func (e *Engine) SSHVM(namespace, id, user, key string, term *termutil.Terminal) error {
	container, err := e.docker.Inspect(id)
	if err != nil {
		fullName := internal.GenerateContainerName(namespace, id)

		container, err = e.docker.Inspect(fullName)
		if err != nil {
			return err
		}
	}

	ip := container.NetworkSettings.IPAddress
	keyPath := homedir.ExpandPath(key)

	privateKey, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return err
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return err
	}

	config := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // nolint: gosec
	}
	config.SetDefaults()

	conn, err := ssh.Dial("tcp", ip+":22", &config)
	if err != nil {
		return err
	}

	defer conn.Close()

	sess, err := conn.NewSession()
	if err != nil {
		return err
	}

	defer sess.Close()

	sess.Stdin = term.In()
	sess.Stdout = term.Out()
	sess.Stderr = term.Err()

	sz, err := term.GetWinsize()
	if err != nil {
		return err
	}

	err = term.MakeRaw()
	if err != nil {
		return err
	}

	defer handleError(term.Restore())

	err = sess.RequestPty(os.Getenv("TERM"), int(sz.Height), int(sz.Width), nil)
	if err != nil {
		return err
	}

	err = sess.Shell()
	if err != nil {
		return err
	}

	// If our terminal window changes, signal the ssh connection
	stopch := make(chan struct{})
	defer close(stopch)

	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, syscall.SIGWINCH)
		defer signal.Stop(sigch)
		defer close(sigch)
	outer:
		for {
			select {
			case <-sigch:
				sz, err := term.GetWinsize()
				if err == nil {
					handleError(sess.WindowChange(int(sz.Height), int(sz.Width)))
				}
			case <-stopch:
				break outer
			}
		}
	}()

	return sess.Wait()
}

func handleError(err error) {
	if err != nil {
		log.Error(err)
	}
}

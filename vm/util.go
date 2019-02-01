package vm

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"strings"

	log "github.com/sirupsen/logrus"
)

// helper function to find a tcp port
// TODO: Find a better way to do this
func findPort() int {

	log.Debug("Looking for an available port for VNC")
	address, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	if err != nil {
		panic(err)
	}

	listen, err := net.ListenTCP("tcp", address)
	if err != nil {
		log.WithField("error", err.Error).Fatal("Cannot find port")
	}

	defer func() {
		err = listen.Close()
		if err != nil {
			log.WithField("error",
				err.Error).Warn("Failed to close port lister")
		}
	}()

	return listen.Addr().(*net.TCPAddr).Port
}

//SaneImage is a helper function to perform various checks to the provided image.
func SaneImage(path string) error {

	if strings.HasPrefix(path, "~/") {
		path = strings.Replace(path, "~", getUserHomePath(), 1)
	}

	// Test if the image file exists
	imgArg, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("Image %v does not exist", path)
	}

	// Test if the image is valid or has a valid path
	mode := imgArg.Mode()
	if !mode.IsRegular() {
		return fmt.Errorf("%v is not a regular file", path)
	}
	return nil
}

func getUserHomePath() string {

	currentUser, err := user.Current()
	if err != nil {
		log.Warn("Unable to determine $HOME")
		log.Error("Please specify -workdir and -pubkey")
	}
	return currentUser.HomeDir
}

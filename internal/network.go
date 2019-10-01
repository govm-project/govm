package internal

import (
	"net"

	log "github.com/sirupsen/logrus"
)

// FindAvailablePort helps to find a tcp port
func FindAvailablePort() int {
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

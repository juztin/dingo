// +build !appengine

package dingo

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
)

func HttpHandler(ip string, port int) (net.Listener, error) {
	addr := fmt.Sprint(ip, ":", port)
	return net.Listen("tcp", addr)
}

func TLSHandler(ip string, port int, certFile, keyFile string) (net.Listener, error) {
	// this func is based off of Go source `net/http - server.go`
	addr := fmt.Sprint(ip, ":", port)
	config := &tls.Config{NextProtos: []string{"http/1.1"}}

	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	conn, err := net.Listen("tcp", addr)
	if err != nil {
		return conn, err
	}

	return tls.NewListener(conn, config), nil
}

func SOCKHandler(sockFile string, mode os.FileMode) (net.Listener, error) {
	// delete stale sock
	// TODO check errors other than file doesn't exist
	os.Remove(sockFile)

	// create UNIX sock
	sock, err := net.ResolveUnixAddr("unix", sockFile)
	if err != nil {
		return nil, err
	}
	if l, err := net.ListenUnix("unix", sock); err == nil {
		err = os.Chmod(sockFile, mode)
		return l, err
	}
	return nil, err
}

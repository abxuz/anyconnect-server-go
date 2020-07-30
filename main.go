package main

import (
	"acserv/server"
	"crypto/tls"
	"errors"
	"log"
	"net"
	"os"
	"strings"
)

var (
	config    *server.Config
	tlsConfig *tls.Config
	listeners []net.Listener
	logFile   *os.File
)

func main() {

	var err error

	// 1. 读取配置
	config, err = server.NewConfig("config.ini")
	if err != nil {
		log.Fatal(err)
	}

	// 2. 根据配置，初始化
	err = initTLS()
	if err != nil {
		log.Fatal(err)
	}

	err = initListeners()
	if err != nil {
		cleanListener()
		log.Fatal(err)
	}

	err = initLogFile()
	if err != nil {
		cleanListener()
		cleanLogFile()
		log.Fatal(err)
	}

	server := &server.Server{
		Listeners: listeners,
		PublicDir: config.Main.PublicDir,
		Loger:     log.New(logFile, "", log.LstdFlags),
	}
	err = server.Serve()
	if err != nil {
		cleanListener()
		cleanLogFile()
		log.Fatal(err)
	}
}

func initTLS() error {
	if len(config.Cert) == 0 {
		return errors.New("no cert configured")
	}

	certs := make([]tls.Certificate, 0)
	addrToCertificate := make(map[string]*tls.Certificate)
	for _, c := range config.Cert {
		cert, err := tls.LoadX509KeyPair(c.Public, c.Private)
		if err != nil {
			return err
		}
		certs = append(certs, cert)
		if c.Addr != "" {
			_, exists := addrToCertificate[c.Addr]
			if exists {
				return errors.New("duplicate addr configured for cert")
			}
			addrToCertificate[c.Addr] = &cert
		}
	}

	tlsConfig = &tls.Config{Certificates: certs}
	tlsConfig.BuildNameToCertificate()
	nameToCertificate := tlsConfig.NameToCertificate

	tlsConfig = &tls.Config{
		GetCertificate: func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			if len(certs) == 1 {
				return &certs[0], nil
			}
			if nameToCertificate != nil {
				name := strings.ToLower(clientHello.ServerName)
				if cert, ok := nameToCertificate[name]; ok {
					return cert, nil
				}
				if len(name) > 0 {
					labels := strings.Split(name, ".")
					labels[0] = "*"
					wildcardName := strings.Join(labels, ".")
					if cert, ok := nameToCertificate[wildcardName]; ok {
						return cert, nil
					}
				}
			}

			addr := clientHello.Conn.LocalAddr().String()
			if cert, ok := addrToCertificate[addr]; ok {
				return cert, nil
			}

			for _, cert := range certs {
				if err := clientHello.SupportsCertificate(&cert); err == nil {
					return &cert, nil
				}
			}
			return &certs[0], nil
		},
	}
	return nil
}

func initListeners() error {
	if len(config.Main.Listen) == 0 {
		return errors.New("no listen address configured")
	}

	listeners = make([]net.Listener, 0)
	for _, listenAddr := range config.Main.Listen {
		l, err := tls.Listen("tcp", listenAddr, tlsConfig)
		if err != nil {
			return err
		}
		listeners = append(listeners, l)
	}
	return nil
}

func cleanListener() {
	if listeners == nil {
		return
	}
	for _, l := range listeners {
		l.Close()
	}
	listeners = nil
}

func initLogFile() error {
	var err error
	if config.Main.LogFile == "" {
		return nil
	}
	logFile, err = os.OpenFile(config.Main.LogFile, os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	return nil
}

func cleanLogFile() {
	if logFile == nil {
		return
	}
	logFile.Close()
	logFile = nil
}

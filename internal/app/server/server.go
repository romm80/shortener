package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"
)

// Server - http server
type Server struct {
	httpServer     *http.Server
	certFile       string
	privateKeyFile string
}

func New() (*Server, error) {

	if err := initConfig(); err != nil {
		return nil, err
	}

	srv := new(Server)
	if !Cfg.EnableHTTPS {
		return srv, nil
	}

	srv.certFile = "cert.pem"
	srv.privateKeyFile = "privateKey.pem"

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"qwerty"},
			Country:      []string{"RU"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal(err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	var certPEM bytes.Buffer
	pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err := writeToFile(certPEM, srv.certFile); err != nil {
		return nil, err
	}

	var privateKeyPEM bytes.Buffer
	pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	if err := writeToFile(privateKeyPEM, srv.privateKeyFile); err != nil {
		return nil, err
	}
	return srv, nil
}

// Run starts http server
func (s *Server) Run(handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:    Cfg.SrvAddr,
		Handler: handler,
	}

	if Cfg.EnableHTTPS {
		return s.httpServer.ListenAndServeTLS(s.certFile, s.privateKeyFile)
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed:%+v", err)
	}
	log.Println("Server shutdowned")
}

func writeToFile(data bytes.Buffer, fileName string) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = data.WriteTo(file); err != nil {
		return err
	}

	return nil
}

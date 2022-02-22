package sutils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"sync"
	"time"
)

var (
	srvTLSDummyCerts     []tls.Certificate
	srvTLSDummyCertsInit sync.Once
)

// GetDummyTLSCert returns the singleton TLS certificate with a fresh
// private key and a dummy certificate.
func GetDummyTLSCerts() []tls.Certificate {
	var initErr error
	srvTLSDummyCertsInit.Do(func() {
		cert, err := generateKeyAndCert()
		if err != nil {
			initErr = fmt.Errorf("appquic: Unable to generate dummy TLS cert/key: %v", err)
		}
		srvTLSDummyCerts = []tls.Certificate{*cert}
	})
	if initErr != nil {
		panic(initErr)
	}
	return srvTLSDummyCerts
}

// generateKeyAndCert generates a private key and a self-signed dummy
// certificate usable for quic TLS with "InsecureSkipVerify==true"
func generateKeyAndCert() (*tls.Certificate, error) {
	priv, err := rsaGenerateKey()
	if err != nil {
		return nil, nil
	}
	return createCertificate(priv)
}

func rsaGenerateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// createCertificate creates a self-signed dummy certificate for the given key
// Inspired/copy pasted from crypto/tls/generate_cert.go
func createCertificate(priv *rsa.PrivateKey) (*tls.Certificate, error) {
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"scionlab"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"dummy"},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	certPEMBuf := &bytes.Buffer{}
	if err := pem.Encode(certPEMBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return nil, err
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal private key: %v", err)
	}

	keyPEMBuf := &bytes.Buffer{}
	if err := pem.Encode(keyPEMBuf, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(certPEMBuf.Bytes(), keyPEMBuf.Bytes())
	return &cert, err
}

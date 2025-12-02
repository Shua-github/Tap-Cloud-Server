package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		os.Exit(1)
	}

	domain := os.Args[1]

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"My Company"},
			Country:      []string{"CN"},
			Province:     []string{"Beijing"},
			Locality:     []string{"Beijing"},
			CommonName:   "My CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(50, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		panic(err)
	}

	caKeyFile, err := os.Create("ca.key")
	if err != nil {
		panic(err)
	}
	defer caKeyFile.Close()

	err = pem.Encode(caKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caKey),
	})
	if err != nil {
		panic(err)
	}

	caCertFile, err := os.Create("ca.crt")
	if err != nil {
		panic(err)
	}
	defer caCertFile.Close()

	err = pem.Encode(caCertFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertDER,
	})
	if err != nil {
		panic(err)
	}

	domainKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	domainTemplate := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"My Company"},
			Country:      []string{"CN"},
			Province:     []string{"Beijing"},
			Locality:     []string{"Beijing"},
			CommonName:   domain,
		},
		DNSNames:              []string{domain, "*." + domain, "localhost"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(50, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		panic(err)
	}

	domainCertDER, err := x509.CreateCertificate(rand.Reader, &domainTemplate, caCert, &domainKey.PublicKey, caKey)
	if err != nil {
		panic(err)
	}

	domainKeyFile, err := os.Create(domain + ".key")
	if err != nil {
		panic(err)
	}
	defer domainKeyFile.Close()

	err = pem.Encode(domainKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(domainKey),
	})
	if err != nil {
		panic(err)
	}

	domainCertFile, err := os.Create(domain + ".crt")
	if err != nil {
		panic(err)
	}
	defer domainCertFile.Close()

	err = pem.Encode(domainCertFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: domainCertDER,
	})
	if err != nil {
		panic(err)
	}
}

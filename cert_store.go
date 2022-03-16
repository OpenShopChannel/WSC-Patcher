package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
)

// generateSerial generates a random serial number for our issued certificates.
// It is taken from golang std: src/crypto/tls/generate_cert.go
// Direct permalink on GitHub: https://git.io/JyyDw
func generateSerial() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	check(err)

	return serialNumber
}

func createCertificates() []byte {
	////////////////////////////////////
	//        Generate root CA        //
	////////////////////////////////////
	rootCAFormat := x509.Certificate{
		SignatureAlgorithm: x509.SHA1WithRSA,
		SerialNumber:       generateSerial(),
		Subject: pkix.Name{
			CommonName: "Open Shop Channel CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	rootPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	check(err)

	rootCert, err := x509.CreateCertificate(rand.Reader, &rootCAFormat, &rootCAFormat, &rootPriv.PublicKey, rootPriv)
	check(err)

	////////////////////////////////////
	//  Issue server TLS certificate  //
	////////////////////////////////////
	serverCertFormat := x509.Certificate{
		SignatureAlgorithm: x509.SHA1WithRSA,
		SerialNumber:       generateSerial(),
		// We'll issue with a primary common name for our base domain.
		Subject: pkix.Name{
			CommonName: baseDomain,
		},
		// The SAN will be a wildcard for our base domain, as it cannot be the CN.
		DNSNames: []string{
			"*." + baseDomain,
		},
		NotBefore:      time.Now(),
		NotAfter:       time.Now().AddDate(10, 0, 0),
		KeyUsage:       x509.KeyUsageKeyAgreement | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:           false,
		MaxPathLenZero: true,
	}

	serverPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	check(err)

	serverCert, err := x509.CreateCertificate(rand.Reader, &serverCertFormat, &rootCAFormat, &serverPriv.PublicKey, rootPriv)
	check(err)

	////////////////////////////
	//  Persist certificates  //
	////////////////////////////
	rootCertPem := pemEncode("CERTIFICATE", rootCert)
	rootKeyPem := pemEncode("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(rootPriv))
	serverCertPem := pemEncode("CERTIFICATE", serverCert)
	serverKeyPem := pemEncode("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(serverPriv))

	writeOut("root.pem", rootCertPem)
	writeOut("root.cer", rootCert)
	writeOut("root.key", rootKeyPem)
	writeOut("server.pem", serverCertPem)
	writeOut("server.key", serverKeyPem)

	return rootCert
}

func pemEncode(typeName string, bytes []byte) []byte {
	block := pem.Block{Type: typeName, Bytes: bytes}
	return pem.EncodeToMemory(&block)
}

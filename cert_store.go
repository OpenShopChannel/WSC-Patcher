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

// YearIssueTime is an issuance of this year's date on January 1 at midnight.
var YearIssueTime = time.Date(time.Now().Year(), time.January, 1, 0, 0, 0, 0, time.UTC)

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
	rootCert := &x509.Certificate{
		SignatureAlgorithm: x509.SHA1WithRSA,
		SerialNumber:       generateSerial(),
		Subject: pkix.Name{
			CommonName: "Open Shop Channel CA",
		},
		NotBefore:             YearIssueTime,
		NotAfter:              YearIssueTime.AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	rootPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	check(err)

	rootPublic, err := x509.CreateCertificate(rand.Reader, rootCert, rootCert, &rootPriv.PublicKey, rootPriv)
	check(err)

	////////////////////////////////////
	//  Issue server TLS certificate  //
	////////////////////////////////////
	// We'll issue a wildcard for our CN and SANs.
	// Is this recommended? Absolutely not, but who's to stop us?
	issueName := "*." + baseDomain
	serverCert := x509.Certificate{
		SignatureAlgorithm: x509.SHA1WithRSA,
		SerialNumber:       generateSerial(),
		Subject: pkix.Name{
			CommonName: issueName,
		},
		DNSNames: []string{
			issueName,
		},
		NotBefore:             YearIssueTime,
		NotAfter:              YearIssueTime.AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageKeyAgreement | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	serverPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	check(err)

	serverPublic, err := x509.CreateCertificate(rand.Reader, &serverCert, rootCert, &serverPriv.PublicKey, rootPriv)
	check(err)

	////////////////////////////
	//  Persist certificates  //
	////////////////////////////
	rootCertPem := pemEncode("CERTIFICATE", rootPublic)
	rootKeyPem := pemEncode("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(rootPriv))
	serverCertPem := pemEncode("CERTIFICATE", serverPublic)
	serverKeyPem := pemEncode("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(serverPriv))

	writeOut("root.pem", rootCertPem)
	writeOut("root.cer", rootPublic)
	writeOut("root.key", rootKeyPem)
	writeOut("server.pem", serverCertPem)
	writeOut("server.key", serverKeyPem)

	return rootPublic
}

func pemEncode(typeName string, bytes []byte) []byte {
	block := pem.Block{Type: typeName, Bytes: bytes}
	return pem.EncodeToMemory(&block)
}

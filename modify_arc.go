package main

import (
	"bytes"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"io/ioutil"
)

// modifyAllowList patches the Opera filter to include our custom base domain.
func modifyAllowList() {
	file, err := mainArc.OpenFile("arc/opera/myfilter.ini")
	check(err)

	// TODO(spotlightishere): Find an INI parser that handles reading an array from a section
	// As I could not - and I spent a good while looking - and do not want to implement my own parser,
	// no matter how rudimentary - I've copied the original file as a template verbatim. We only make one edit,
	// adding the base domain.
	filter := fmt.Sprintf(`[prefs]
prioritize excludelist=0

[include]
file:/cnt/*
https://*.%s/*
miip:*

[exclude]
*`, baseDomain)

	// Replace UNIX line (LR) returns with that of Windows (CRLF).
	output := bytes.ReplaceAll([]byte(filter), []byte("\n"), []byte("\r\n"))
	file.Write(output)
}

// Tag represents a single byte representing a tag's ID.
type Tag byte

const (
	TagSSLCertType     = 0x20
	TagSSLCertName     = 0x21
	TagSSLCertSubject  = 0x22
	TagSSLCertContents = 0x23

	TagCACertificate   = 0x02
	TagUserCertificate = 0x03
	TagUserPassword    = 0x04
)

// generateTag generates a byte representation of a tag and contents.
func generateTag(tag Tag, tagContents []byte) []byte {
	// Tag ID
	contents := []byte{
		byte(tag),
	}
	// Tag length
	contents = append(contents, fourByte(uint32(len(tagContents)))...)
	// Tag contents
	contents = append(contents, tagContents...)

	return contents
}

// generateOperaCertStore creates our own custom Opera cert store for the given certificate.
func generateOperaCertStore() {
	file, err := mainArc.OpenFile("arc/opera/opcacrt6.dat")
	check(err)

	// Load our existing root certificate in DER form.
	rootCertContents, err := ioutil.ReadFile("./output/root.cer")
	check(err)
	rootCert, err := x509.ParseCertificate(rootCertContents)
	check(err)

	// The following array was done manually after several hours of tinkering.
	// Please refer to docs/opcacrt6.yml for more about the structure of this file.
	// TODO(spotlightishere): Is it possible to somehow generate a structure for easier access?
	header := []byte{
		// File version number
		0x00, 0x00, 0x10, 0x00,
		// App version number
		0x05, 0x05, 0x00, 0x23,
		// ID tag "length" - always one byte
		0x00, 0x01,
		// Length field byte length - always four bytes
		0x00, 0x04,
	}

	// It's unclear on what 0x01 is supposed to represent,
	// but it must be a CA certificate.
	certTypeTag := generateTag(TagSSLCertType, []byte{
		0x0, 0x0, 0x0, 0x1,
	})

	// We can obtain the name and subject from the root certificate.
	certNameTag := generateTag(TagSSLCertName, []byte(rootCert.Subject.CommonName))
	certSubjectTag := generateTag(TagSSLCertSubject, rootCert.RawSubject)

	// Finally, our actual certificate.
	certContentsTag := generateTag(TagSSLCertContents, rootCertContents)

	// We must enclose our type, name, subject and contents tag in a CA certificate tag.
	bundledContents := append(certTypeTag, certNameTag...)
	bundledContents = append(bundledContents, certSubjectTag...)
	bundledContents = append(bundledContents, certContentsTag...)

	caCertTag := generateTag(TagCACertificate, bundledContents)

	// Thankfully, that is all.
	// In the end, we have a structure similar to the following:
	// - header (file/app version, id/length byte length)
	// - ca certificate
	//   - id
	//   - length
	//   - value:
	//     - type tag (id, length, value)
	//     - type name (id, length, value)
	//     - type subject (id, length, value)
	//     - type contents (id, length, value)
	file.Write(append(header, caCertTag...))
}

// fourByte returns 4 bytes, suitable for the given length.
func fourByte(value uint32) []byte {
	holder := make([]byte, 4)
	binary.BigEndian.PutUint32(holder, value)
	return holder
}

package main

// LoadCustomCA loads our custom certificate, either generated or loaded,
// into the IOS trust store for EC usage.
// It is assumed that rootCertificate has been loaded upon invoking this patchset.
// See docs/patch_custom_ca_ios.md for more information.
func LoadCustomCA() PatchSet {
	return PatchSet{
		Patch{
			Name:     "Insert custom CA into free space",
			AtOffset: 3037368,

			Before: emptyBytes(len(rootCertificate)),
			After:  rootCertificate,
		},
		Patch{
			Name:     "Modify NHTTPi_SocSSLConnect to load cert",
			AtOffset: 644624,

			Before: Instructions{
				// Check whether internals->ca_cert is null
				LWZ(R4, 0xc0, R28),
				// cmpwi r4, 0
				CMPWI(R4, 0),

				// If it is, load the built-in root certificate.
				// beq LOAD_BUILTIN_ROOT_CA
				Instruction{0x41, 0x82, 0x00, 0x20},

				// ---

				// It seems we are loading a custom certificate.
				// r3 -> ssl_fd
				// r4 -> ca_cert, loaded previously
				// r5 -> cert_length
				LWZ(R3, 0xac, R28),
				LWZ(R5, 0xc4, R28),
				// SSLSetRootCA(ssl_fd, ca_cert, cert_index)
				BL(0x800acae4, 0x800c242c),

				// Check if successful
				CMPWI(R3, 0),
				// beq CONTINUE_CONNECTING
				Instruction{0x41, 0x82, 0x00, 0x28},

				// Return error -1004 if failed
				LI(R3, 0xfc14),
				// b FUNCTION_PROLOG
				B(0x800acaf4, 0x800acbb0),

				// ----

				// It seems we are loading the built-in root CA.
				// r3 -> ssl_fd
				// r4 -> cert_length
				LWZ(R3, 0xac, R28),
				LWZ(R4, 0xd8, R28),
				// SSLSetBuiltinRootCA(ssl_fd, cert_index)
				BL(0x800acb00, 0x800c2574),

				// Check if successful
				CMPWI(R3, 0),
				// beq CONTINUE_CONNECTING
				Instruction{0x41, 0x82, 0x00, 0x0c},

				// Return error -1004 if failed
				LI(R3, 0xfc14),
				// b FUNCTION_PROLOG
				B(0x800acb10, 0x800acbb0),
			}.toBytes(),
			After: Instructions{
				// Our certificate is present at 0x802e97b8.
				// r4 is the second parameter of SSLSetRootCA, the ca_cert pointer.
				LIS(R4, 0x802e),
				ORI(R4, R4, 0x97b8),

				// r5 is the third parameter of SSLSetRootCA, the cert_length field.
				// xor r5, r5, r5
				Instruction{0x7c, 0xa5, 0x2a, 0x78},
				ADDI(R5, R5, uint16(len(rootCertificate))),

				// r3 is the first parameter of SSLSetRootCA, the ssl_fd.
				// We load it exactly as Nintendo does.
				LWZ(R3, 0xac, R28),

				// SSLSetRootCA(ssl_fd, ca_cert, cert_index)
				BL(0x800acae4, 0x800c242c),

				// Check for errors
				CMPWI(R3, 0),
				// beq CONTINUE_CONNECTING
				Instruction{0x41, 0x82, 0x00, 0x28},

				// Return error -1004 if failed
				LI(R3, 0xfc14),
				// b FUNCTION_PROLOG
				B(0x800acaf4, 0x800acbb0),

				// NOP the rest in order to allow execution to continue.
				NOP(), NOP(), NOP(), NOP(), NOP(), NOP(), NOP(),
			}.toBytes(),
		},
	}
}

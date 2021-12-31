package main

// Instruction represents a 4-byte PowerPC instruction.
type Instruction [4]byte

// Instructions represents a group of PowerPC instructions.
type Instructions []Instruction

// toBytes returns the bytes necessary to represent these instructions.
func (i Instructions) toBytes() []byte {
	var contents []byte

	for _, instruction := range i {
		contents = append(contents, instruction[:]...)
	}

	return contents
}

// padding is not an actual instruction - it represents 4 zeros.
var padding Instruction = [4]byte{0x00, 0x00, 0x00, 0x00}

// BLR represents the blr mnemonic on PowerPC.
func BLR() Instruction {
	return [4]byte{0x4E, 0x80, 0x00, 0x20}
}

// ADDI represents the addi PowerPC instruction.
func ADDI(rT Register, rA Register, value uint16) Instruction {
	return EncodeInstrDForm(14, rT, rA, value)
}

// LI represents the li mnemonic on PowerPC.
func LI(rT Register, value uint16) Instruction {
	return ADDI(rT, 0, value)
}

// SUBI represents the subi mnemonic on PowerPC.
func SUBI(rT Register, rA Register, value uint16) Instruction {
	return ADDI(rT, 0, -value)
}

// ADDIS represents the addis PowerPC instruction.
func ADDIS(rT Register, rA Register, value uint16) Instruction {
	return EncodeInstrDForm(15, rT, rA, value)
}

// LIS represents the lis mnemonic on PowerPC.
func LIS(rT Register, value uint16) Instruction {
	return ADDIS(rT, 0, value)
}

// ORI represents the ori PowerPC instruction.
func ORI(rS Register, rA Register, value uint16) Instruction {
	return EncodeInstrDForm(24, rS, rA, value)
}

// STH represents the sth PowerPC instruction.
func STH(rS Register, offset uint16, rA Register) Instruction {
	return EncodeInstrDForm(44, rS, rA, offset)
}

// EIEIO represents the eieio PowerPC instruction.
func EIEIO() Instruction {
	return [4]byte{0x7C, 0x00, 0x06, 0xAC}
}

// STW represents the stw PowerPC instruction.
func STW(rS Register, offset uint16, rA Register) Instruction {
	return EncodeInstrDForm(36, rS, rA, offset)
}

// LWZ represents the lwz PowerPC instruction.
func LWZ(rT Register, offset uint16, rA Register) Instruction {
	return EncodeInstrDForm(32, rT, rA, offset)
}

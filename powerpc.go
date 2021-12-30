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

// BLR represents the blr PowerPC instruction.
func BLR() Instruction {
	return [4]byte{0x4E, 0x80, 0x00, 0x20}
}

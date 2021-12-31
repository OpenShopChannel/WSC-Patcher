package main

import (
	"encoding/binary"
	"encoding/hex"
	"log"
)

// Register represents a value for a PowerPC register.
type Register byte

const (
	R0 = iota
	R1
	R2
	R3
	R4
	R5
	R6
	R7
	R8
	R9
	R10
	R11
	R12
	R13
	R14
	R15
	R16
	R17
	R18
	R19
	R20
	R21
	R22
	R23
	R24
	R25
	R26
	R27
	R28
	R29
	R30
	R31
)

// Bits represents bits for a byte.
// If set, considered as 1. If not, 0.
type Bits [8]bool

// getBits returns a usable array of bits for the given byte.
func getBits(in byte) Bits {
	return [8]bool{
		(in>>7)&1 == 1,
		(in>>6)&1 == 1,
		(in>>5)&1 == 1,
		(in>>4)&1 == 1,
		(in>>3)&1 == 1,
		(in>>2)&1 == 1,
		(in>>1)&1 == 1,
		in&1 == 1,
	}
}

// getByte returns the byte represented by these bits.
func (b Bits) getByte() byte {
	var result byte
	for idx, truthy := range b {
		if truthy {
			result |= 1 << (7 - idx)
		}
	}

	return result
}

// EncodeInstrDForm handles encoding a given opcode, RT, RA and SI.
// D-form assumes:
//  - 6 bits for the opcode
//  - 5 for rT
//  - 5 for rA
//  - 16 for SI
func EncodeInstrDForm(opcode byte, rT Register, rA Register, value uint16) Instruction {
	var instr [2]Bits
	opBits := getBits(opcode)
	rTBits := getBits(byte(rT))
	rABits := getBits(byte(rA))

	instr[0] = Bits{
		// We need the upper six bits for our opcode.
		opBits[2],
		opBits[3],
		opBits[4],
		opBits[5],
		opBits[6],
		opBits[7],
		// Next, the lower two bits for rT.
		rTBits[3],
		rTBits[4],
	}
	instr[1] = Bits{
		// Third, the lower three bits for rT.
		rTBits[5],
		rTBits[6],
		rTBits[7],
		// Finally, all five lowest bits for rA.
		rABits[3],
		rABits[4],
		rABits[5],
		rABits[6],
		rABits[7],
	}

	firstInstr := instr[0].getByte()
	secondInstr := instr[1].getByte()
	valByte := twoByte(value)

	log.Println(hex.EncodeToString([]byte{
		firstInstr, secondInstr, valByte[0], valByte[1],
	}))

	return Instruction{firstInstr, secondInstr, valByte[0], valByte[1]}
}

// twoByte converts a uint16 to two big-endian bytes.
func twoByte(passed uint16) [2]byte {
	result := make([]byte, 2)
	binary.BigEndian.PutUint16(result, passed)
	return [2]byte{result[0], result[1]}
}

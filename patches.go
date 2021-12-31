package main

// OverwriteIOSPatch effectively nullifies IOSC_VerifyPublicKeySign.
// See docs/patch_overwrite_ios.md for more information.
var OverwriteIOSPatch = PatchSet{
	Patch{
		Name:     "Clear textinput::EventObserver functions",
		AtOffset: 20320,

		Before: Instructions{
			// Function: textinput::EventObserver::onSE
			BLR(),
			padding,
			padding,
			padding,
			// Function: textinput::EventObserver::onEvent
			BLR(),
			padding,
			padding,
			padding,
			// Function: textinput::EventObserver::onCommand
			BLR(),
			padding,
			padding,
			padding,
			// Function: textinput::EventObserver::onInput
			BLR(),
			padding,
			padding,
			padding,
		}.toBytes(),

		// We wish to clear extraneous blrs so that our custom overwriteIOSMemory
		// function does not somehow conflict. We only preserve onSE.
		After: append(Instructions{
			BLR(),
		}.toBytes(), emptyBytes(60)...),
	},
	Patch{
		Name:     "Repair textinput::EventObserver vtable",
		AtOffset: 3095452,

		Before: []byte{
			0x80, 0x01, 0x44, 0x50, // onSE
			0x80, 0x01, 0x44, 0x40, // onEvent
			0x80, 0x01, 0x44, 0x30, // onCommand
			0x80, 0x01, 0x44, 0x20, // onInput
		},
		After: []byte{
			// These are all pointers to our so-called doNothing.
			0x80, 0x01, 0x44, 0x20,
			0x80, 0x01, 0x44, 0x20,
			0x80, 0x01, 0x44, 0x20,
			0x80, 0x01, 0x44, 0x20,
		},
	},
	Patch{
		Name:     "Repair ipl::keyboard::EventObserver vtable",
		AtOffset: 3097888,

		Before: []byte{
			0x80, 0x01, 0x44, 0x50, // onSE
			0x80, 0x01, 0x84, 0xE0, // ipl::keyboard::EventObserver::onCommand - not patched
			0x80, 0x01, 0x44, 0x30, // onCommand
		},
		After: []byte{
			0x80, 0x01, 0x44, 0x20, // doNothing
			0x80, 0x01, 0x84, 0xE0, // ipl::keyboard::EventObserver::onCommand - not patched
			0x80, 0x01, 0x44, 0x20, // doNothing
		},
	},
	Patch{
		Name:     "Insert overwriteIOSMemory",
		AtOffset: 20328,

		// This area should be cleared.
		Before: emptyBytes(48),
		After: Instructions{
			// We want r9 to store the location of MEM_PROT at 0x0d8b420a.
			// For us, this is mapped to 0xcd8b420a.
			LIS(R9, 0xcd8b),
			ORI(R9, R9, 0x420a),

			// We want to write 0x2 and unlock everything.
			LI(R10, 0x02),

			// Write!
			STH(R10, 0x0, R9),
			// Flush memory
			EIEIO(),

			// Location of IOSC_VerifyPublicKeySign
			LIS(R9, 0xd3a7),
			ORI(R9, R9, 0x3ad4),

			// Write our custom THUMB.
			// 0x20004770 is equivalent to:
			//    mov r0, #0x0
			//    bx lr
			LIS(R10, 0x2000),
			ORI(R10, R10, 0x4770),

			// Write!
			STW(R10, 0x0, R9),
			// Possibly clear cache
			// TODO(spotlightishere): Is this needed?
			// dcbi 0, r10
			Instruction{0x7C, 0x00, 0x53, 0xAC},
			// And finish.
			BLR(),
		}.toBytes(),
	},
	Patch{
		Name:     "Modify ES_InitLib",
		AtOffset: 2399844,

		// We inject in the epilog of the function.
		Before: Instructions{
			LWZ(R0, 0x14, R1),
			// mtspr LR, r0
			Instruction{0x7C, 0x08, 0x03, 0xA6},
			ADDI(R1, R1, 0x10),
			BLR(),
			padding,
		}.toBytes(),
		After: Instructions{
			LWZ(R0, 0x14, R1),
			// bl overwriteIOSMemory @ 0x80014428
			Instruction{0x4B, 0xDB, 0xB1, 0x01},
			// mtspr LR, r0
			Instruction{0x7C, 0x08, 0x03, 0xA6},
			ADDI(R1, R1, 0x10),
			BLR(),
		}.toBytes(),
	},
}

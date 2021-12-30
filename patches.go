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

		// We wish to clear these so that our custom overwriteIOSMemory
		// function does not somehow conflict.
		After: emptyBytes(64),
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
		After: []byte{
			// We want r9 to store the location of MEM_PROT at 0x0d8b420a.
			// For us, this is mapped to 0xcd8b420a.
			// lis r9, 0xcd8b
			0x3D, 0x20, 0xCD, 0x8B,
			// ori r9, r9, 0x420a
			0x61, 0x29, 0x42, 0x0A,

			// We want to write 0x2 and unlock everything.
			// li r10, 0x02
			0x39, 0x40, 0x00, 0x02,

			// Write!
			// sth r10, 0x0(r9)
			0xB1, 0x49, 0x00, 0x00,
			// Flush memory
			// eieio
			0x7C, 0x00, 0x06, 0xAC,

			// Location of IOSC_VerifyPublicKeySign
			// lis r9, 0xd3a7
			0x3D, 0x20, 0xD3, 0xA7,
			// ori r9, r9, 0x3ad4
			0x61, 0x29, 0x3A, 0xD4,

			// Write our custom THUMB.
			// 0x20004770 is equivalent to:
			//    mov r0, #0x0
			//    bx lr
			// lis r10, 0x2000
			0x3D, 0x40, 0x20, 0x00,
			// ori r10, r10, 0x4770
			0x61, 0x4A, 0x47, 0x70,

			// Write!
			// stw r10, 0x0(r9)
			0x91, 0x49, 0x00, 0x00,
			// Possibly clear cache
			// TODO(spotlightishere): Is this needed?
			// dcbi 0, r10
			0x7C, 0x00, 0x53, 0xAC,
			// And finish.
			// blr
			0x4E, 0x80, 0x00, 0x20,
		},
	},
	Patch{
		Name:     "Modify ES_InitLib",
		AtOffset: 2399844,

		// We inject in the epilog of the function.
		Before: []byte{
			0x80, 0x01, 0x00, 0x14, // lwz r0, local_res4(r1)
			0x7C, 0x08, 0x03, 0xA6, // mtspr LR, r0
			0x38, 0x21, 0x00, 0x10, // addi r1, r1, 0x10
			0x4E, 0x80, 0x00, 0x20, // blr
			0x00, 0x00, 0x00, 0x00, // ; empty space following function
		},
		After: []byte{
			0x80, 0x01, 0x00, 0x14, // lwz r0, local_res4(r1)
			0x4B, 0xDB, 0xB1, 0x01, // bl overwriteIOSMemory @ 0x80014428
			0x7C, 0x08, 0x03, 0xA6, // mtspr LR, r0
			0x38, 0x21, 0x00, 0x10, // addi r1, r1, 0x10
			0x4E, 0x80, 0x00, 0x20, // blr
		},
	},
}

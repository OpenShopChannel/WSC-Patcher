package main

// OverwriteIOSPatch effectively nullifies IOSC_VerifyPublicKeySign.
// See docs/patch_overwrite_ios.md for more information.
var OverwriteIOSPatch = PatchSet{
	Patch{
		Name:     "Clear extraneous functions",
		AtOffset: 20272,

		Before: Instructions{
			// Function: textinput::EventObserver::onOutOfLength
			LIS(R3, 0x802f),
			ADDI(R3, R3, 0x7ac4),
			CRXOR(),
			// b printf
			Instruction{0x48, 0x2a, 0x8b, 0x78},

			// Function: textinput::EventObserver::onCancel
			LIS(R3, 0x802f),
			ADDI(R3, R3, 0x7ab8),
			CRXOR(),
			// b printf
			Instruction{0x48, 0x2a, 0x8b, 0x68},

			// Function: textinput::EventObserver::onOK
			// subi r3, r3, 0x7fe0
			Instruction{0x38, 0x6d, 0x80, 0x20},
			CRXOR(),
			// b printf
			Instruction{0x48, 0x2a, 0x8b, 0x5c},

			padding,

			// Function: textinput::EventObserver::onSE
			BLR(),
			padding, padding, padding,
			// Function: textinput::EventObserver::onEvent
			BLR(),
			padding, padding, padding,
			// Function: textinput::EventObserver::onCommand
			BLR(),
			padding, padding, padding,
			// Function: textinput::EventObserver::onInput
			BLR(),
			padding, padding, padding,
		}.toBytes(),

		// We wish to clear extraneous blrs so that our custom overwriteIOSMemory
		// function does not somehow conflict.
		// We only preserve onSE, which is this immediate BLR.
		After: append(Instructions{
			BLR(),
		}.toBytes(), emptyBytes(108)...),
	},
	Patch{
		Name:     "Repair textinput::EventObserver vtable",
		AtOffset: 3095452,

		Before: []byte{
			0x80, 0x01, 0x44, 0x50, // onSE
			0x80, 0x01, 0x44, 0x40, // onEvent
			0x80, 0x01, 0x44, 0x30, // onCommand
			0x80, 0x01, 0x44, 0x20, // onInput
			0x80, 0x01, 0x44, 0x10, // onOK
			0x80, 0x01, 0x44, 0x00, // onCancel
			0x80, 0x01, 0x43, 0xf0, // onOutOfLength
		},
		After: []byte{
			// These are all pointers to our so-called doNothing.
			0x80, 0x01, 0x43, 0xf0,
			0x80, 0x01, 0x43, 0xf0,
			0x80, 0x01, 0x43, 0xf0,
			0x80, 0x01, 0x43, 0xf0,
			0x80, 0x01, 0x43, 0xf0,
			0x80, 0x01, 0x43, 0xf0,
			0x80, 0x01, 0x43, 0xf0,
		},
	},
	Patch{
		Name:     "Repair ipl::keyboard::EventObserver vtable",
		AtOffset: 3097888,

		Before: []byte{
			0x80, 0x01, 0x44, 0x50, // textinput::EventObserver::onSE
			0x80, 0x01, 0x84, 0xE0, // onCommand - not patched
			0x80, 0x01, 0x44, 0x30, // textinput::EventObserver::onCommand
			0x80, 0x01, 0x85, 0x20, // onSE - not patching
			0x80, 0x01, 0x87, 0x40, // onOK - not patching
			0x80, 0x01, 0x87, 0x60, // onCancel - not patching
			0x80, 0x01, 0x43, 0xF0, // textinput::EventObserver::onOutOfLength
		},
		After: []byte{
			0x80, 0x01, 0x43, 0xf0, // textinput::EventObserver::doNothing
			0x80, 0x01, 0x84, 0xE0, // onCommand - not patched
			0x80, 0x01, 0x43, 0xf0, // textinput::EventObserver::doNothing
			0x80, 0x01, 0x85, 0x20, // onSE - not patching
			0x80, 0x01, 0x87, 0x40, // onOK - not patching
			0x80, 0x01, 0x87, 0x60, // onCancel - not patching
			0x80, 0x01, 0x43, 0xf0, // textinput::EventObserver::doNothing
		},
	},
	Patch{
		Name:     "Insert patch table",
		AtOffset: 3205088,

		Before: emptyBytes(40),
		After: []byte{
			//////////////
			// PATCH #1 //
			//////////////
			// We want to write to MEM_PROT at 0x0d8b420a.
			// For us, this is mapped to 0xcd8b420a.
			0xcd, 0x8b, 0x42, 0x0a,
			// We are going to write the value 0x2 to unlock everything.
			0x00, 0x00, 0x00, 0x02,

			//////////////
			// PATCH #2 //
			//////////////
			// We want to write to IOSC_VerifyPublicKeySign at 0x13a73ad4.
			// For us, this is mapped to 0xd3a73ad4.
			0xd3, 0xa7, 0x3a, 0xd4,
			// 0x20004770 is equivalent in ARM THUMB to:
			//    mov r0, #0x0
			//    bx lr
			0x20, 0x00, 0x47, 0x70,

			//////////////////////////
			// PATCH #3 - vWii only //
			//////////////////////////
			// Patch location:
			// We want to write at 0x20102100, aka "ES_AddTicket".
			// (To us, this is mapped at 0x939f2100.)
			0x93, 0x9f, 0x21, 0x00,
			// The original code has a few conditionals preventing system title usage.
			// We simply branch off past these.
			// 0x681a2a01 is equivalent in ARM THUMB to:
			//    ldr r2,[r3,#0x0]     ; original code we wish to preserve
			//                         ; so we can write 32 bits
			//    b +0x14              ; branch past conditionals
			0x68, 0x1a, 0x2a, 0x01,

			//////////////////////////
			// PATCH #4 - vWii only //
			//////////////////////////
			// We want to write to 0x20103240, aka "ES_AddTitleStart".
			// (For us, this is 0x939f3240.)
			0x93, 0x9f, 0x32, 0x40,
			// The original code has a few conditionals preventing system title usage.
			// 0xe00846c0 is equivalent in ARM THUMB to:
			//    b +0x8       ; branch past conditionals
			//    mov r8, r8   ; recommended THUMB nop
			0xe0, 0x08, 0x46, 0xc0,

			//////////////////////////
			// PATCH #5 - vWii only //
			//////////////////////////
			// Lastly, we want to write to 0x20103564, aka "ES_AddContentStart".
			// (We utilize memory cache and write at 0x939f3564.)
			0x93, 0x9f, 0x35, 0x64,
			// The original code has a few conditionals preventing system title usage.
			// We simply branch off past these.
			// 0xe00c46c0 is equivalent in ARM THUMB to:
			//    b +0xc       ; branch past conditionals
			//    mov r8, r8   ; recommended THUMB nop
			0xe0, 0x0c, 0x46, 0xc0,
		},
	},
	Patch{
		Name:     "Insert overwriteIOSMemory",
		AtOffset: 20276,

		// This area should be cleared in the patch
		// "Clear extraneous functions".
		Before: emptyBytes(108),
		After: Instructions{
			// Our patch table is available at 0x803126e0.
			LIS(R8, 0x8031),
			ORI(R8, R8, 0x26e0),

			// Load address/value pair for MEM_PROT
			LWZ(R9, 0x0, R8),
			LWZ(R10, 0x4, R8),
			// Apply lower half
			STH(R10, 0x0, R9),

			// Load address/value pair for IOSC_VerifyPublicKeySign
			LWZ(R9, 0x8, R8),
			LWZ(R10, 0xc, R8),

			// Apply!
			STW(R10, 0x0, R9),

			// The remainder of our patches will determine if we are on a Wii U.
			// Even in vWii mode, 0x0d8005a0 (LT_CHIPREVID) will have its upper
			// 16 bits set to 0xCAFE. We can compare against this.
			// See also: https://wiiubrew.org/wiki/Hardware/Latte_registers
			// (However, we must access the cached version at 0xcd8005a0.)
			LIS(R9, 0xcd80),
			ORI(R9, R9, 0x05a0),
			LWZ(R9, 0, R9),

			// Shift this value 16 bits to the right
			// in order to compare its higher value.
			// srawi r9, r9, 16
			Instruction{0x7d, 0x29, 0x86, 0x70},
			CMPWI(R9, 0xCAFE),
			// If we're not a Wii U, carry on until the end.
			// bne (last blr)
			ORI(R0, R0, 0),
			//Instruction{0x40, 0x82, 0x00, 0x28},

			// Apply ES_AddTicket
			LWZ(R9, 0x10, R8),
			LWZ(R10, 0x14, R8),
			STW(R10, 0x0, R9),

			// Apply ES_AddTicket
			LWZ(R9, 0x18, R8),
			LWZ(R10, 0x1c, R8),
			STW(R10, 0x0, R9),

			// Apply ES_AddTicket
			LWZ(R9, 0x20, R8),
			LWZ(R10, 0x24, R8),
			STW(R10, 0x0, R9),

			// We're finished patching!
			BLR(),

			BLR(), BLR(), BLR(),
		}.toBytes(),
	},
	Patch{
		Name:     "Modify ES_InitLib",
		AtOffset: 2399844,

		// We inject in the epilog of the function.
		Before: Instructions{
			LWZ(R0, 0x14, R1),
			MTSPR(),
			ADDI(R1, R1, 0x10),
			BLR(),
			padding,
		}.toBytes(),
		After: Instructions{
			LWZ(R0, 0x14, R1),
			// bl overwriteIOSMemory @ 0x80014428
			Instruction{0x4b, 0xdb, 0xb0, 0xcd},
			MTSPR(),
			ADDI(R1, R1, 0x10),
			BLR(),
		}.toBytes(),
	},
}

package main

var NegateECTitle = PatchSet{
	Patch{
		Name:     "Permit downloading all titles",
		AtOffset: 619648,

		// Generic function prolog
		Before: Instructions{
			STWU(R1, R1, 0xffe0),
			MFSPR(),
		}.toBytes(),

		// Immediately return true
		After: Instructions{
			LI(R3, 1),
			BLR(),
		}.toBytes(),
	},
	Patch{
		Name:     "Mark all titles as managed",
		AtOffset: 620656,

		Before: Instructions{
			STWU(R1, R1, 0xfff0),
			MFSPR(),
		}.toBytes(),
		After: Instructions{
			LI(R3, 1),
			BLR(),
		}.toBytes(),
	},
	Patch{
		Name:     "Mark all tickets as managed",
		AtOffset: 619904,
		Before: Instructions{
			STWU(R1, R1, 0xfff0),
			MFSPR(),
		}.toBytes(),
		After: Instructions{
			LI(R3, 1),
			BLR(),
		}.toBytes(),
	},
	Patch{
		Name:     "Nullify ec::removeAllTitles",
		AtOffset: 588368,
		Before: Instructions{
			STWU(R1, R1, 0xffc0),
			MFSPR(),
		}.toBytes(),
		After: Instructions{
			LI(R3, 0),
			BLR(),
		}.toBytes(),
	},
}

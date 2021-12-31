package main

var NegateECTitle = PatchSet{
	Patch{
		Name:     "Allow all titles",
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
}

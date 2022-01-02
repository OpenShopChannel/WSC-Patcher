package main

var PatchECCfgPath = PatchSet{
	Patch{
		Name:     "Change EC Configuration Path",
		AtOffset: 3319968,

		Before: []byte("ec.cfg\x00\x00"),
		After:  []byte("osc.cfg\x00"),
	},
}

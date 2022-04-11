package main

import (
	. "github.com/wii-tools/powerpc"
)

var PatchECCfgPath = PatchSet{
	Name: "Change EC Configuration Path",
	Patches: []Patch{
		{
			AtOffset: 3319968,

			Before: []byte("ec.cfg\x00\x00"),
			After:  []byte("osc.cfg\x00"),
		},
	},
}

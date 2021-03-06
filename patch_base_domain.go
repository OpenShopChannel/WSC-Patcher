package main

import (
	"strings"

	. "github.com/wii-tools/powerpc"
)

const (
	NintendoBaseDomain = "shop.wii.com"
	ShowManualURL      = "https://oss-auth.shop.wii.com/startup?initpage=showManual&titleId="
	GetLogURL          = "https://oss-auth.shop.wii.com/oss/getLog"
	TrustedDomain      = ".shop.wii.com"
	ECommerceBaseURL   = "https://ecs.shop.wii.com/ecs/services/ECommerceSOAP"
)

// PatchBaseDomain replaces all Nintendo domains to be the user's
// specified base domain.
// See docs/patch_base_domain.md for more information.
func PatchBaseDomain() PatchSet {
	return PatchSet{
		Name: "Change Base Domain",
		Patches: []Patch{
			{
				Name: "Modify /startup domain",

				Before: []byte(ShowManualURL),
				After:  padReplace(ShowManualURL),
			},
			{
				Name:     "Modify oss-auth URL",
				AtOffset: 3180692,

				Before: []byte(GetLogURL),
				After:  padReplace(GetLogURL),
			},
			{
				Name:     "Modify trusted base domain prefix",
				AtOffset: 3323432,

				Before: []byte(TrustedDomain),
				After:  padReplace(TrustedDomain),
			},
			{
				Name:     "Modify ECS SOAP endpoint URL",
				AtOffset: 3268896,

				Before: []byte(ECommerceBaseURL),
				After:  padReplace(ECommerceBaseURL),
			},
			{
				Name: "Wildcard replace other instances",

				Before: []byte(NintendoBaseDomain),
				After:  padReplace(baseDomain),
			},
		},
	}
}

func padReplace(url string) []byte {
	replaced := strings.ReplaceAll(url, NintendoBaseDomain, baseDomain)

	// See if we truly need to pad.
	if len(url) == len(replaced) {
		return []byte(replaced)
	}

	padding := len(url) - len(replaced)
	return append([]byte(replaced), EmptyBytes(padding)...)
}

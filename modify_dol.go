package main

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/logrusorgru/aurora/v3"
)

var (
	ErrInconsistentPatch = errors.New("before and after data present within file are not the same size")
	ErrPatchOutOfRange   = errors.New("patch cannot be applied past binary size")
	ErrInvalidPatch      = errors.New("before data present within patch did not exist in file")
)

// Patch represents a patch applied to the main binary.
type Patch struct {
	// Name is an optional name for this patch.
	// If present, its name will be logged upon application.
	Name string

	// AtOffset is the offset within the file this patch should be applied at.
	// If not present, the patch will be recursively applied across the entire file.
	// Relying on this behavior is highly discouraged, as it may damage other parts of the binary
	// if gone unchecked.
	AtOffset int

	// Before is an array of the bytes to find for, i.e. present within the original file.
	Before []byte

	// After is an array of the bytes to replace with.
	After []byte
}

// PatchSet represents multiple patches available to be applied.
type PatchSet []Patch

// applyPatch applies the given patch to the main DOL.
func applyPatch(patch Patch) error {
	// Print name if present
	if patch.Name != "" {
		fmt.Println(" + Applying patch", aurora.Cyan(patch.Name))
	}

	// Ensure consistency
	if len(patch.Before) != len(patch.After) {
		return ErrInconsistentPatch
	}
	if patch.AtOffset != 0 && patch.AtOffset > len(mainDol) {
		return ErrPatchOutOfRange
	}

	// Either Before or After should return the same length.
	patchLen := len(patch.Before)

	// Determine our patching behavior.
	if patch.AtOffset != 0 {
		// Ensure original bytes are present
		originalBytes := mainDol[patch.AtOffset : patch.AtOffset+patchLen]
		if !bytes.Equal(originalBytes, patch.Before) {
			return ErrInvalidPatch
		}

		// Apply patch at the specified offset
		copy(mainDol[patch.AtOffset:], patch.After)
	} else {
		// Recursively apply this patch.
		// We cannot verify if the original contents are present via this.
		mainDol = bytes.ReplaceAll(mainDol, patch.Before, patch.After)
	}

	return nil
}

// applyPatchSet iterates through all possible patches, noting their name.
func applyPatchSet(setName string, set PatchSet) {
	fmt.Printf("Handling patch set \"%s\":\n", aurora.Yellow(setName))

	for _, patch := range set {
		err := applyPatch(patch)
		check(err)
	}
}

// emptyBytes returns an empty byte array of the given length.
func emptyBytes(length int) []byte {
	return bytes.Repeat([]byte{0x00}, length)
}

// applyDefaultPatches iterates through a list of default patches.
func applyDefaultPatches() {
	applyPatchSet("Overwrite IOS Syscall for ES", OverwriteIOSPatch)
	applyPatchSet("Load Custom CA within IOS", LoadCustomCA())
	applyPatchSet("Change Base Domain", PatchBaseDomain())
	applyPatchSet("Negate EC Title Check", NegateECTitle)
}

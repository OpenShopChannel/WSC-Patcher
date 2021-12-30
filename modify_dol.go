package main

import (
	"bytes"
	"errors"
	"log"
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
		log.Println("Applying patch " + patch.Name)
	}

	// Ensure consistency
	if len(patch.Before) != len(patch.After) {
		return ErrInconsistentPatch
	}
	if patch.AtOffset > len(mainDol) {
		return ErrPatchOutOfRange
	}

	// Either Before or After should return the same length.
	patchLen := len(patch.Before)

	// Ensure original bytes are present
	originalBytes := mainDol[patch.AtOffset : patch.AtOffset+patchLen]
	if !bytes.Equal(originalBytes, patch.Before) {
		return ErrInvalidPatch
	}

	// Apply patch
	copy(mainDol[patch.AtOffset:], patch.After)

	return nil
}

// applyPatchSet iterates through all possible patches, noting their name.
func applyPatchSet(setName string, set PatchSet) {
	log.Printf("Handling patch set \"%s\":", setName)

	for _, patch := range set {
		err := applyPatch(patch)
		check(err)
	}
}

// applyDefaultPatches iterates through a list of default patches.
func applyDefaultPatches() {
	applyPatchSet("Overwrite IOS Syscall for ES", OverwriteIOSPatch)
}

// emptyBytes returns an empty byte array of the given length.
func emptyBytes(length int) []byte {
	return bytes.Repeat([]byte{0x00}, length)
}

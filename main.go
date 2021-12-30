package main

import (
	"errors"
	"fmt"
	"github.com/wii-tools/GoNUSD"
	"github.com/wii-tools/arclib"
	"github.com/wii-tools/wadlib"
	"io/fs"
	"log"
	"os"
)

// baseDomain holds our needed base domain.
var baseDomain string

// mainDol holds the main DOL - our content at index 1.
var mainDol []byte

// mainArc holds the main ARC - our content at index 2.
var mainArc *arclib.ARC

// filePresent returns whether the specified path is present on disk.
func filePresent(path string) bool {
	_, err := os.Stat(path)
	return errors.Is(err, fs.ErrNotExist) == false
}

// createDir creates a directory at the given path if it is not already present.
func createDir(path string) {
	if !filePresent(path) {
		os.Mkdir(path, 0755)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <base domain>\n", os.Args[0])
		fmt.Println("For more information, please refer to the README.")
		os.Exit(-1)
	}

	baseDomain = os.Args[1]
	if len(baseDomain) > 12 {
		fmt.Println("The given base domain must not exceed 12 characters.")
		fmt.Println("For more information, please refer to the README.")
		os.Exit(-1)
	}

	fmt.Println("===========================")
	fmt.Println("=       WSC-Patcher       =")
	fmt.Println("===========================")

	// Create directories we may need later.
	createDir("./output")
	createDir("./cache")

	var originalWad *wadlib.WAD
	var err error

	// Determine whether the Wii Shop Channel is cached.
	if !filePresent("./cache/original.wad") {
		log.Println("Downloading a copy of the original Wii Shop Channel, please wait...")
		originalWad, err = GoNUSD.Download(0x00010002_48414241, 21, true)
		check(err)

		// Cache this downloaded WAD to disk.
		contents, err := originalWad.GetWAD(wadlib.WADTypeCommon)
		check(err)

		os.WriteFile("./cache/original.wad", contents, 0755)
	} else {
		originalWad, err = wadlib.LoadWADFromFile("./cache/original.wad")
		check(err)
	}

	// Determine whether a certificate authority was provided, or generated previously.
	if !filePresent("./output/root.cer") {
		log.Println("Generating root certificates...")
		createCertificates()
	}

	// Load main DOL
	mainDol, err = originalWad.GetContent(1)
	check(err)

	// Load main ARC
	arcData, err := originalWad.GetContent(2)
	check(err)
	mainArc, err = arclib.Load(arcData)
	check(err)

	// Generate filter list and certificate store
	log.Println("Applying Opera patches...")
	modifyAllowList()
	generateOperaCertStore()

	// Save main ARC
	updated, err := mainArc.Save()
	check(err)
	err = originalWad.UpdateContent(2, updated)
	check(err)

	// Generate a patched WAD with our changes
	output, err := originalWad.GetWAD(wadlib.WADTypeCommon)
	check(err)

	log.Println("Done! Install ./output/patched.wad, sit back, and enjoy.")
	writeOut("patched.wad", output)
}

// check has an anxiety attack if things go awry.
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// writeOut writes a file with the given name and contents to the output folder.
func writeOut(filename string, contents []byte) {
	os.WriteFile("./output/"+filename, contents, 0755)
}

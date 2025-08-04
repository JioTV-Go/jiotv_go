package cmd

import (
	"fmt"
	"os"

	"github.com/jiotv-go/jiotv_go/v3/pkg/epg"
)

// GenEPG generates a new epg.xml.gz file with updated EPG data by first deleting any existing epg.xml.gz file.
// calls epg.GenXMLGz() to generate the XML, and returns any errors.
func GenEPG() error {

	fmt.Println("Deleting existing EPG file if exists")

	err := os.Remove("epg.xml.gz")
	if err != nil {
		// If file does not exist, ignore error
		if !os.IsNotExist(err) {
			return err
		}
	}

	fmt.Println("Generating new EPG file")

	err = epg.GenXMLGz("epg.xml.gz")
	return err
}

// DeleteEPG deletes the existing epg.xml.gz file if it exists.
// It logs status messages about deleting or not finding the file.
// Returns any errors encountered except os.ErrNotExist.
func DeleteEPG() error {

	fmt.Println("Deleting existing EPG file if exists")

	err := os.Remove("epg.xml.gz")

	if err != nil {
		if err == os.ErrNotExist || os.IsNotExist(err) { // Added os.IsNotExist for robustness
			fmt.Println("EPG file does not exist")
		} else {
			return err
		}
	} else {
		fmt.Println("EPG file deleted")
	}

	return nil
}

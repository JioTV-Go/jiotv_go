package cmd

import (
	"os"

	"github.com/jiotv-go/jiotv_go/v3/pkg/epg"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

// GenEPG generates a new epg.xml.gz file with updated EPG data by first deleting any existing epg.xml.gz file.
// It initializes the utils.Log global logger, calls epg.GenXMLGz() to generate the XML, and returns any errors.
func GenEPG() error {
	// Initialize the logger object as it is used in epg.GenXMLGz()
	// Do not remove this line, it will result in nil pointer dereference panic
	utils.Log = utils.GetLogger()

	utils.Log.Println("Deleting existing EPG file if exists")

	err := os.Remove("epg.xml.gz")
	if err != nil {
		// If file does not exist, ignore error
		if !os.IsNotExist(err) {
			return err
		}
	}

	utils.Log.Println("Generating new EPG file")

	err = epg.GenXMLGz("epg.xml.gz")
	return err
}

// DeleteEPG deletes the existing epg.xml.gz file if it exists.
// It logs status messages about deleting or not finding the file.
// Returns any errors encountered except os.ErrNotExist.
func DeleteEPG() error {
	utils.Log = utils.GetLogger() // Initialize logger

	utils.Log.Println("Deleting existing EPG file if exists")

	err := os.Remove("epg.xml.gz")

	if err != nil {
		if err == os.ErrNotExist || os.IsNotExist(err) { // Added os.IsNotExist for robustness
			utils.Log.Println("EPG file does not exist")
		} else {
			return err
		}
	} else {
		utils.Log.Println("EPG file deleted")
	}

	return nil
}

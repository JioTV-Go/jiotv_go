package cmd

import (
	"log"
	"os"

	"github.com/rabilrbl/jiotv_go/v3/pkg/epg"
	"github.com/rabilrbl/jiotv_go/v3/pkg/utils"
)

func GenEPG() error {
	// Initialize the logger object as it is used in epg.GenXMLGz()
	// Do not remove this line, it will result in nil pointer dereference panic
	utils.Log = utils.GetLogger()
	
	log.Println("Deleting existing EPG file if exists")

	err := os.Remove("epg.xml.gz")
	if err != nil {
		// If file does not exist, ignore error
		if !os.IsNotExist(err) {
			return err
		}
	}

	log.Println("Generating new EPG file")

	err = epg.GenXMLGz("epg.xml.gz")
	return err
}

func DeleteEPG() error {
	log.Println("Deleting existing EPG file if exists")

	err := os.Remove("epg.xml.gz")

	if err != nil {
		if err == os.ErrNotExist {
			log.Println("EPG file does not exist")
		} else {
			return err
		}
	} else {
		log.Println("EPG file deleted")
	}

	return nil
}
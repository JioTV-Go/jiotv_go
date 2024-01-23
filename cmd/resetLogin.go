package cmd

import (
	"log"
	"os"

	"github.com/rabilrbl/jiotv_go/v3/pkg/utils"
)

func ResetLogin() error {
	// Initialize the logger object as it is used in epg.GenXMLGz()
	// Do not remove this line, it will result in nil pointer dereference panic
	utils.Log = utils.GetLogger()
	
	log.Println("Deleting existing login file if exists")

	login_path := utils.GetCredentialsPath()

	err := os.Remove(login_path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Login file does not exist. Are you logged in? Please login using Web Interface")
			return nil
		} else {
			return err
		}
	}

	log.Println("Please navigate to web interface and login again")

	return nil
}
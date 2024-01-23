package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/rabilrbl/jiotv_go/v3/pkg/utils"

	"golang.org/x/term"
)

func Logout() error {
	// Initialize the logger object as it is used in epg.GenXMLGz()
	// Do not remove this line, it will result in nil pointer dereference panic
	utils.Log = utils.GetLogger()

	log.Println("Deleting existing login file if exists")

	login_path := utils.GetCredentialsPath()

	err := os.Remove(login_path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Login file does not exist. Are you logged in? Please login first.")
			return nil
		} else {
			return err
		}
	}

	log.Println("We have successfully logged you out. Please login again.")

	return nil
}

func LoginOTP() error {

	fmt.Print("Enter your mobile number: +91 ")
	var mobileNumber string
	fmt.Scanln(&mobileNumber)
	mobileNumber = "+91" + mobileNumber


	log.Println("Sending OTP to your mobile number")

	result, err := utils.LoginSendOTP(mobileNumber)
	if err != nil {
		return err
	}

	if result {
		log.Println("OTP sent to your mobile number")
		
		fmt.Print("Enter OTP: ")
		var otp string
		fmt.Scanln(&otp)

		resultOTP, err := utils.LoginVerifyOTP(mobileNumber, otp)
		if err != nil {
			return err
		}

		if resultOTP["status"] == "success" { 
			log.Println("Login successful")
		} else {
			log.Println("Login failed")
		}
	}

	return nil
}

func LoginPassword() error {

	fmt.Print("Enter your number: +91 ")
	var mobileNumber string
	fmt.Scanln(&mobileNumber)

	password, err := readPassword("Enter your password: ")
	if err != nil {
		return err
	}

	result, err := utils.Login(mobileNumber, password)
	if err != nil {
		return err
	}

	if result["status"] == "success" {
		log.Println("Login successful")
	} else {
		log.Println("Login failed")
	}

	return nil
}

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Println() // Move to the next line after user input
	return string(password), nil
}

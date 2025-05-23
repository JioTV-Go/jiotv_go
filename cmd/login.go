package cmd

import (
	"fmt"
	"os"

	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"

	"golang.org/x/term"
)

// Logout logs the user out by removing the saved login credentials file.
// It checks if the file exists before removing to avoid errors.
// Logs messages to provide feedback to the user.
// Returns any errors encountered.
func Logout() error {
	fmt.Println("Deleting existing login file if exists")

	err := utils.Logout()
	if err != nil {
		return err
	}

	fmt.Println("We have successfully logged you out. Please login again.")

	return nil
}

// LoginOTP handles the login flow using OTP.
// It takes the mobile number as input, sends an OTP,
// verifies the entered OTP by the user and logs in the user.
// Returns any error encountered.
func LoginOTP() error {
	fmt.Print("Enter your mobile number: +91 ")
	var mobileNumber string
	fmt.Scanln(&mobileNumber)
	mobileNumber = "+91" + mobileNumber

	fmt.Println("Sending OTP to your mobile number")

	result, err := utils.LoginSendOTP(mobileNumber)
	if err != nil {
		return err
	}

	if result {
		fmt.Println("OTP sent to your mobile number")

		fmt.Print("Enter OTP: ")
		var otp string
		fmt.Scanln(&otp)

		resultOTP, err := utils.LoginVerifyOTP(mobileNumber, otp)
		if err != nil {
			return err
		}

		if resultOTP["status"] == "success" {
			fmt.Println("Login successful")
		} else {
			fmt.Println("Login failed")
		}
	}

	return nil
}

// LoginPassword handles the login flow using password.
// It takes the mobile number and password as input,
// verifies the credentials by calling the Login API
// and logs in the user if successful.
// Returns any error encountered.
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
		fmt.Println("Login successful")
	} else {
		fmt.Println("Login failed")
	}

	return nil
}

// readPassword prompts the user for a password input from stdin.
// It prints the given prompt, reads the password while masking the input,
// and returns the password as a string along with any error.
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Println() // Move to the next line after user input
	return string(password), nil
}

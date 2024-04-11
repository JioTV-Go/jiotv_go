package handlers

// LoginRequestBodyData represents Request body for password based login request
type LoginRequestBodyData struct {
	// Username or mobile number of Jio account
	Username string `json:"username" xml:"username" form:"username"`
	// Password of Jio account
	Password string `json:"password" xml:"password" form:"password"`
}

// LoginSendOTPRequestBodyData represents Request body for OTP based login request
type LoginSendOTPRequestBodyData struct {
	// Mobile number of Jio account
	MobileNumber string `json:"number" xml:"number" form:"number"`
}

// LoginVerifyOTPRequestBodyData  represents Request body for OTP verification request
type LoginVerifyOTPRequestBodyData struct {
	// Mobile number of Jio account
	MobileNumber string `json:"number" xml:"number" form:"number"`
	// OTP received on mobile number
	OTP string `json:"otp" xml:"otp" form:"otp"`
}

// RefreshTokenResponse represents Response body for refresh token request
type RefreshTokenResponse struct {
	// Access token for JioTV API
	AccessToken string `json:"authToken"`
}

// RefreshSSOTokenResponse represents Response body for refresh token request
type RefreshSSOTokenResponse struct {
	// Access token for JioTV API
	SSOToken string `json:"ssoToken"`
}

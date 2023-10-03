package handlers

type LoginRequestBodyData struct {
	Username string `json:"username" xml:"username" form:"username"`
	Password string `json:"password" xml:"password" form:"password"`
}

type LoginSendOTPRequestBodyData struct {
	MobileNumber string `json:"number" xml:"number" form:"number"`
}

type LoginVerifyOTPRequestBodyData struct {
	MobileNumber string `json:"number" xml:"number" form:"number"`
	OTP          string `json:"otp" xml:"otp" form:"otp"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"authToken"`
}

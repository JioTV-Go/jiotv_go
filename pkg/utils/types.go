package utils

// JIOTV_CREDENTIALS represents JioTV Go credentials.
// Includes some mandatory header values required for making requests to JioTV API
type JIOTV_CREDENTIALS struct {
	SSOToken                string `json:"ssoToken"`
	UniqueID                string `json:"uniqueId"`
	CRM                     string `json:"crm"`
	AccessToken             string `json:"accessToken"`
	RefreshToken            string `json:"refreshToken"`
	LastTokenRefreshTime    string `json:"lastTokenRefreshTime"`
	LastSSOTokenRefreshTime string `json:"lastSSOTokenRefreshTime"`
}

// LoginPasswordPayload represents Request payload for password based login
type LoginPasswordPayload struct {
	// Username or phone number of the Jio account
	Identifier string `json:"identifier"`
	// Password of the Jio account
	Password string `json:"password"`
	// Remember user after login
	RememberUser         string                 `json:"rememberUser"`
	UpgradeAuth          string                 `json:"upgradeAuth"`
	ReturnSessionDetails string                 `json:"returnSessionDetails"`
	DeviceInfo           LoginPayloadDeviceInfo `json:"deviceInfo"`
}

// LoginOTPPayload represents Request payload for OTP based login
type LoginOTPPayload struct {
	Number     string                 `json:"number"`
	OTP        string                 `json:"otp"`
	DeviceInfo LoginPayloadDeviceInfo `json:"deviceInfo"`
}

// LoginPayloadDeviceInfo represents Device info for the login API
type LoginPayloadDeviceInfo struct {
	ConsumptionDeviceName string                     `json:"consumptionDeviceName"`
	Info                  LoginPayloadDeviceInfoInfo `json:"info"`
}

// LoginPayloadDeviceInfoInfo represents Info for the login API
type LoginPayloadDeviceInfoInfo struct {
	Type      string                             `json:"type"`
	Platform  LoginPayloadDeviceInfoInfoPlatform `json:"platform"`
	AndroidID string                             `json:"androidId"`
}

// LoginPayloadDeviceInfoInfoPlatform represents Platform info for the login API
type LoginPayloadDeviceInfoInfoPlatform struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// LoginResponse represents Response from the JioTV login API
type LoginResponse struct {
	AuthToken         string `json:"authToken"`
	RefreshToken      string `json:"refreshToken"`
	SSOToken          string `json:"ssoToken"`
	SessionAttributes struct {
		User struct {
			SubscriberID string `json:"subscriberId"`
			Unique       string `json:"unique"`
		} `json:"user"`
	} `json:"sessionAttributes"`
}

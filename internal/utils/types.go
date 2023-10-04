package utils

type JIOTV_CREDENTIALS struct {
	SSOToken             string `json:"ssoToken"`
	UniqueID             string `json:"uniqueId"`
	CRM                  string `json:"crm"`
	AccessToken          string `json:"accessToken"`
	RefreshToken         string `json:"refreshToken"`
	LastTokenRefreshTime string `json:"lastTokenRefreshTime"`
}

type LoginPayload struct {
	Identifier           string                 `json:"identifier"`
	Password             string                 `json:"password"`
	RememberUser         string                 `json:"rememberUser"`
	UpgradeAuth          string                 `json:"upgradeAuth"`
	ReturnSessionDetails string                 `json:"returnSessionDetails"`
	DeviceInfo           LoginPayloadDeviceInfo `json:"deviceInfo"`
}

type LoginPayloadDeviceInfo struct {
	ConsumptionDeviceName string                     `json:"consumptionDeviceName"`
	Info                  LoginPayloadDeviceInfoInfo `json:"info"`
}

type LoginPayloadDeviceInfoInfo struct {
	Type     string `json:"type"`
	Platform LoginPayloadDeviceInfoInfoPlatform `json:"platform"`
	AndroidID string `json:"androidId"`
}

type LoginPayloadDeviceInfoInfoPlatform struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

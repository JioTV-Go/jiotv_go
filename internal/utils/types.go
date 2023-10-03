package utils

type JIOTV_CREDENTIALS struct {
	SSOToken             string `json:"ssoToken"`
	UniqueID             string `json:"uniqueId"`
	CRM                  string `json:"crm"`
	AccessToken          string `json:"accessToken"`
	RefreshToken         string `json:"refreshToken"`
	LastTokenRefreshTime string `json:"lastTokenRefreshTime"`
}

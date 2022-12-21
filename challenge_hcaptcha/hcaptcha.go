package challenge_hcaptcha

import (
	"go.jolheiser.com/hcaptcha"
	"validator/config"
)

var client *hcaptcha.Client

//验证
func Verify(token string) bool {
	resp, err := client.Verify(token, hcaptcha.PostOptions{
		"",
		config.HCAPTCHA_SITEKEY,
	})
	if err != nil {
		return false
	}
	if !resp.Success {
		return false
	}
	return true
}

func Init() (err error) {
	if client, err = hcaptcha.New(config.HCAPTCHA_SECRETKEY); err != nil {
		return
	}
	return
}

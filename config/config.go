package config

import (
	"liberty-town/node/validator/validation/validation_type"
)

const (
	VERSION = "0.1"
)

var (
	CHALLENGE_TYPE     validation_type.ValidatorChallengeType = 0
	CHALLENGE_0        []byte                                 = nil
	HCAPTCHA_SITEKEY                                          = ""
	HCAPTCHA_SECRETKEY                                        = ""
	IMAGES_COUNT                                              = 0
)

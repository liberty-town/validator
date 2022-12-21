package arguments

import (
	"io/ioutil"
	libraryArguments "liberty-town/node/config/arguments"
	"liberty-town/node/validator/validation/validation_type"
	"strconv"
	"strings"
	"validator/config"
)

var Arguments = map[string]any{}

func InitArguments(argv []string) (err error) {

	libraryArguments.Text = text
	if err = libraryArguments.InitArguments(argv); err != nil {
		return
	}
	Arguments = libraryArguments.Arguments

	if Arguments["--hcaptcha-sitekey"] != nil {
		config.HCAPTCHA_SITEKEY = Arguments["--hcaptcha-sitekey"].(string)
		config.HCAPTCHA_SECRETKEY = Arguments["--hcaptcha-secretkey"].(string)
	}

	if Arguments["--challenge-type"] != nil {
		var value int
		if value, err = strconv.Atoi(Arguments["--challenge-type"].(string)); err != nil {
			return
		}
		config.CHALLENGE_TYPE = validation_type.ValidatorChallengeType(value)

		if config.CHALLENGE_TYPE == validation_type.VALIDATOR_CHALLENGE_CUSTOM {

			var content []byte
			if content, err = ioutil.ReadFile("./challenge_rotation/count.txt"); err != nil {
				return
			}

			if config.IMAGES_COUNT, err = strconv.Atoi(strings.TrimSpace(string(content))); err != nil {
				return
			}
		}
	}

	return
}

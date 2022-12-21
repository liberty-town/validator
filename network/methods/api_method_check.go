package methods

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/cryptography"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/validator/validation/validation_type"
	"net/http"
	"validator/config"
	"validator/config/arguments"
	"validator/validator_store"
)

func MethodCheck(r *http.Request, args *api_types.ValidatorCheckRequest, reply *api_types.ValidatorCheckResult) error {

	switch args.Version {
	case 0:
	default:
		return errors.New("invalid request version ")
	}

	//验证字符串长度
	if len(args.Message) != cryptography.HashSize {
		return errors.New("invalid Message size")
	}

	wr := advanced_buffers.NewBufferWriter()
	wr.Write(args.Message)
	wr.WriteUvarint(args.Size)
	addr, err := addresses.CreateAddrFromSignature(wr.Bytes(), args.Signature)
	if err != nil {
		return err
	}

	switch config.CHALLENGE_TYPE {
	//没有验证码
	case validation_type.VALIDATOR_CHALLENGE_NO_CAPTCHA:
		reply.Challenge = validation_type.VALIDATOR_CHALLENGE_NO_CAPTCHA
		reply.Required = true
		reply.Data = config.CHALLENGE_0
	case validation_type.VALIDATOR_CHALLENGE_HCAPTCHA, validation_type.VALIDATOR_CHALLENGE_CUSTOM:
		required, err := validator_store.IsAbuseRequired(addr, args.Size)
		if err != nil {
			return err
		}

		reply.Challenge = config.CHALLENGE_TYPE
		reply.Required = required
		reply.ChallengeUri = arguments.Arguments["--challenge-uri"].(string)

		if config.CHALLENGE_TYPE == validation_type.VALIDATOR_CHALLENGE_HCAPTCHA {
			//hCaptcha, reCaptcha
			reply.Data = []byte(config.HCAPTCHA_SITEKEY)
		} else {
			//自定义生成验证码
			reply.Data = nil
		}
	default:
		return errors.New("invalid challenge type ")
	}

	return nil

}

package methods

import (
	"bytes"
	"encoding/json"
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/cryptography"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/settings"
	"liberty-town/node/validator/validation/validation_type"
	"net/http"
	"time"
	"validator/challenge_custom"
	"validator/challenge_hcaptcha"
	"validator/config"
	"validator/validator_store"
)

func MethodSolution(r *http.Request, args *api_types.ValidatorSolutionRequest, reply *api_types.ValidatorSolutionResult) error {

	switch args.Version {
	case 0:
	default:
		return errors.New("invalid request version ")
	}

	//验证字符串长度
	if len(args.Message) != cryptography.HashSize {
		return errors.New("invalid Message size")
	}

	//验证消息部件
	wr := advanced_buffers.NewBufferWriter()
	wr.Write(args.Message)
	wr.WriteUvarint(args.Size)
	addr, err := addresses.CreateAddrFromSignature(wr.Bytes(), args.Signature)
	if err != nil {
		return err
	}

	challengeResolved := false

	switch config.CHALLENGE_TYPE {
	//没有验证码
	case validation_type.VALIDATOR_CHALLENGE_NO_CAPTCHA:

		if !bytes.Equal(args.Solution, config.CHALLENGE_0) {
			return errors.New("invalid Solution for Challenge_0")
		}
		challengeResolved = true

	//hCaptcha, reCaptcha
	//中文验证码 https://github.com/lingd3/Captcha
	//EasyCaptcha https://yufeixuan.github.io/easycaptcha/
	case validation_type.VALIDATOR_CHALLENGE_HCAPTCHA:
		if len(args.Solution) > 0 {
			if challenge_hcaptcha.Verify(string(args.Solution)) {
				challengeResolved = true
			} else {
				return errors.New("invalid Solution for Challenge_1")
			}
		}
	//自定义生成验证码
	case validation_type.VALIDATOR_CHALLENGE_CUSTOM:

		//验证
		if len(args.Solution) > 0 {

			//使用验证规则
			value := &struct {
				Id      []byte `json:"id"`
				Answers []int  `json:"answers"`
			}{}

			if err = json.Unmarshal(args.Solution, value); err != nil {
				return err
			}

			challenge, _ := challenge_custom.Challenges.Load(string(value.Id))
			if challenge == nil {
				return errors.New("invalid challenge id")
			}

			defer challenge_custom.Challenges.Delete(string(value.Id))

			//引用验证码
			if challenge.Expired() {
				return errors.New("invalid challenge - expired")
			}
			if challenge.Address != addr.Encoded {
				return errors.New("invalid challenge - address")
			}

			if len(challenge.Solutions) != len(value.Answers) {
				return errors.New("invalid number of answers")
			}

			challengeSolved := true
			for i, sol := range challenge.Solutions {
				if (sol*45-value.Answers[i])%360 != 0 {
					challengeSolved = false
					break
				}
			}

			if err = validator_store.ChallengeIncreaseAbuse(addr, challengeSolved); err != nil {
				return err
			}

			if !challengeSolved {
				return errors.New("invalid captcha")
			}

			challengeResolved = true
		}
	default:
		//如果没有验证码会怎样？
		return errors.New("invalid captcha - unknown")
	}

	abuse, err := validator_store.IncreaseAbuse(addr, args.Size, challengeResolved)
	if err != nil {
		return err
	}
	if abuse {
		return errors.New("challenge is required")
	}

	nonce := cryptography.RandomBytes(validation_type.VALIDATOR_NONCE_SIZE)
	timestamp := uint64(time.Now().Unix())

	wr = advanced_buffers.NewBufferWriter()
	wr.WriteUvarint(uint64(validation_type.VALIDATION_VERSION_V0))
	wr.Write(args.Message)
	wr.WriteUvarint(args.Size)
	wr.Write(nonce)
	wr.WriteUvarint(timestamp)

	signature, err := settings.Settings.Load().Account.PrivateKey.Sign(wr.Bytes())
	if err != nil {
		return err
	}

	reply.Nonce = nonce
	reply.Timestamp = timestamp
	reply.Signature = signature

	return nil
}

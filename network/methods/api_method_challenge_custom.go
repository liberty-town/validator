package methods

import (
	"io/ioutil"
	"liberty-town/node/addresses"
	"liberty-town/node/cryptography"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"math/rand"
	"net/http"
	"strconv"
	"time"
	"validator/challenge_custom"
	"validator/config"
	"validator/validator_store"
)

//自定义验证码
type APIChallengeRequest struct {
	Message   []byte `json:"message" msgpack:"message"`
	Size      uint64 `json:"size" msgpack:"size"`
	Signature []byte `json:"signature" msgpack:"signature"`
}

type APIChallengeCustomReply struct {
	Id     []byte `json:"id" msgpack:"id"`
	Count  int    `json:"count" msgpack:"count"`
	Image0 []byte `json:"im0,omitempty" msgpack:"im0,omitempty"`
	Image1 []byte `json:"im1,omitempty" msgpack:"im1,omitempty"`
	Image2 []byte `json:"im2,omitempty" msgpack:"im2,omitempty"`
	Image3 []byte `json:"im3,omitempty" msgpack:"im3,omitempty"`
	Image4 []byte `json:"im4,omitempty" msgpack:"im4,omitempty"`
	Image5 []byte `json:"im5,omitempty" msgpack:"im5,omitempty"`
}

func MethodChallengeCustom(r *http.Request, args *APIChallengeRequest, reply *APIChallengeCustomReply) (err error) {

	wr := advanced_buffers.NewBufferWriter()
	wr.Write(args.Message)
	wr.WriteUvarint(args.Size)
	addr, err := addresses.CreateAddrFromSignature(wr.Bytes(), args.Signature)
	if err != nil {
		return err
	}

	//困难
	registration, hard, extreme, err := validator_store.ChallengeDifficulty(addr)
	if err != nil {
		return
	}

	if registration || extreme {
		reply.Count = 6
	} else if hard {
		reply.Count = 3
	} else {
		reply.Count = 2
	}

	id := cryptography.RandomBytes(32)

	challenge := &challenge_custom.ChallengeCustom{
		//在收到验证码240秒内输入
		time.Now().Add(4 * time.Minute),
		make([]int, reply.Count),
		addr.Encoded,
	}

	for i := 0; i < reply.Count; i++ {
		challenge.Solutions[i] = rand.Intn(8)
	}

	challenge_custom.Challenges.Store(string(id), challenge)
	reply.Id = id

	dict := make(map[int]bool)
	for i := 0; i < reply.Count; i++ {

		n := rand.Intn(config.IMAGES_COUNT)
		for dict[n] {
			n = rand.Intn(config.IMAGES_COUNT)
		}
		dict[n] = true

		//直接从文件读取到[]byte中
		var x []byte
		if x, err = ioutil.ReadFile("../../../challenge_rotation/" + strconv.Itoa(n) + "_" + strconv.Itoa(challenge.Solutions[i]*45) + ".jpg"); err != nil {
			return err
		}

		switch i {
		case 0:
			reply.Image0 = x
		case 1:
			reply.Image1 = x
		case 2:
			reply.Image2 = x
		case 3:
			reply.Image3 = x
		case 4:
			reply.Image4 = x
		case 5:
			reply.Image5 = x
		}

	}

	return nil

}

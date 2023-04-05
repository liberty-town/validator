package challenge_custom

import (
	"liberty-town/node/pandora-pay/helpers/generics"
	"time"
)

type ChallengeCustom struct {
	Deadline  time.Time //限期
	Solutions []int
	Address   string
}

func (this *ChallengeCustom) Expired() bool {
	return time.Now().After(this.Deadline)
}

var Challenges *generics.Map[string, *ChallengeCustom]

func init() {
	Challenges = &generics.Map[string, *ChallengeCustom]{}

	//确定日期是否过期。
	go func() {
		for {
			c := 0
			Challenges.Range(func(key string, value *ChallengeCustom) bool {

				if value.Expired() {
					Challenges.Delete(key)
				}

				if c > 50 {
					time.Sleep(100 * time.Millisecond)
				}
				return true
			})

			time.Sleep(250 * time.Millisecond)
		}
	}()

}

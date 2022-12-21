package spam_abuse

import (
	"github.com/vmihailenco/msgpack"
	"liberty-town/node/store/store_db/store_db_interface"
	"time"
)

//限制KEY访问频率，防止暴力攻击

type SpamAbuse struct {
	Count int    `json:"count" msgpack:"count"`
	Date  uint64 `json:"date" msgpack:"date"`
}

func IsAbused(key string, max int, secondsReset float64, tx store_db_interface.StoreDBTransactionInterface) (bool, error) {

	b := tx.Get("spamAbuse:" + key)
	if b == nil {
		return false, nil
	}

	spamAbuse := &SpamAbuse{}
	if err := msgpack.Unmarshal(b, spamAbuse); err != nil {
		return false, err
	}

	if time.Now().Sub(time.Unix(int64(spamAbuse.Date), 0)).Seconds() > secondsReset {
		return false, nil
	}

	if spamAbuse.Count >= max {
		return true, nil
	}

	return false, nil
}

func IncreaseAbusedCounter(key string, max int, secondsReset float64, challengeSolved bool, tx store_db_interface.StoreDBTransactionInterface) (bool, error) {

	spamAbuse := &SpamAbuse{}

	b := tx.Get("spamAbuse:" + key)
	if b != nil {
		if err := msgpack.Unmarshal(b, spamAbuse); err != nil {
			return false, err
		}
	}

	if challengeSolved || time.Now().Sub(time.Unix(int64(spamAbuse.Date), 0)).Seconds() > secondsReset {
		spamAbuse.Date = uint64(time.Now().Unix())
		spamAbuse.Count = 0
	}

	if spamAbuse.Count >= max {
		return true, nil
	}

	spamAbuse.Count++
	b, err := msgpack.Marshal(spamAbuse)
	if err != nil {
		return false, err
	}

	tx.Put("spamAbuse:"+key, b)
	return false, nil
}

func IsRegistered(key string, secondsReset float64, tx store_db_interface.StoreDBTransactionInterface) (bool, error) {

	b := tx.Get("spamAbuse:" + key)
	if b == nil {
		return true, nil
	}

	spamAbuse := &SpamAbuse{}
	if err := msgpack.Unmarshal(b, spamAbuse); err != nil {
		return false, err
	}

	if spamAbuse.Count == 0 || (secondsReset != 0 && time.Now().Sub(time.Unix(int64(spamAbuse.Date), 0)).Seconds() > secondsReset) {
		return true, nil
	}

	return false, nil
}

func RequireRegistration(key string, secondsReset float64, challengeSolved bool, tx store_db_interface.StoreDBTransactionInterface) (bool, error) {

	spamAbuse := &SpamAbuse{}

	b := tx.Get("spamAbuse:" + key)
	if b != nil {
		if err := msgpack.Unmarshal(b, spamAbuse); err != nil {
			return false, err
		}
	}

	if challengeSolved && time.Now().Sub(time.Unix(int64(spamAbuse.Date), 0)).Seconds() > secondsReset {
		spamAbuse.Date = uint64(time.Now().Unix())
		spamAbuse.Count = 0
	}

	if spamAbuse.Count == 0 {

		if !challengeSolved {
			return true, nil
		} else if challengeSolved {
			spamAbuse.Count++

			b, err := msgpack.Marshal(spamAbuse)
			if err != nil {
				return false, err
			}

			tx.Put("spamAbuse:"+key, b)
		}

	}

	return false, nil
}

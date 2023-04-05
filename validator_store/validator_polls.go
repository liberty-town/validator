package validator_store

import (
	"github.com/vmihailenco/msgpack"
	"liberty-town/node/addresses"
	"liberty-town/node/store/store_db/store_db_interface"
)

type StoredPoll struct {
	Upvotes   uint64 `json:"up" msgpack:"up"`
	Downvotes uint64 `json:"down" msgpack:"down"`
}

func CheckVotedAlready(addr, identity *addresses.Address) (result bool, err error) {
	err = Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {

		b := tx.Get("voted_already:" + addr.Encoded + ":" + identity.Encoded)
		result = b != nil

		return
	})

	return
}

func ProcessVote(addr, identity *addresses.Address, vote int) (poll *StoredPoll, err error) {

	poll = &StoredPoll{}

	err = Store.DB.Update(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {

		if b := tx.Get("polls:" + identity.Encoded); b != nil {
			if err = msgpack.Unmarshal(b, &poll); err != nil {
				return
			}
		}

		if vote > 0 {
			poll.Upvotes += uint64(vote)
		} else if vote < 0 {
			poll.Downvotes += uint64(-vote)
		}

		b, err := msgpack.Marshal(poll)
		if err != nil {
			return
		}

		tx.Put("validator_polls:"+identity.Encoded, b)
		tx.Put("voted_already:"+addr.Encoded+":"+identity.Encoded, []byte{1})

		return
	})

	return
}

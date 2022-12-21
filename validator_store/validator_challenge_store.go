package validator_store

import (
	"liberty-town/node/addresses"
	"liberty-town/node/store/store_db/store_db_interface"
	"validator/validator_store/spam_abuse"
)

func ChallengeDifficulty(address *addresses.Address) (registration, hard, extreme bool, err error) {
	err = Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {

		if registration, err = spam_abuse.IsRegistered("challenge:reg:"+address.Encoded, 180*24*60*60, tx); registration || err != nil {
			return
		}

		if hard, err = spam_abuse.IsAbused("challenge:hard:"+address.Encoded, 5, 180*24*60*60, tx); hard || err != nil {
			return
		}

		if extreme, err = spam_abuse.IsAbused("challenge:extreme:"+address.Encoded, 20, 180*24*60*60, tx); hard || err != nil {
			return
		}

		return nil
	})

	return
}

func ChallengeIncreaseAbuse(address *addresses.Address, challengeSolved bool) (err error) {
	return Store.DB.Update(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {

		var abuse bool
		if abuse, err = spam_abuse.RequireRegistration("challenge:reg:"+address.Encoded, 180*24*60*60, challengeSolved, tx); abuse || err != nil {
			return
		}

		if abuse, err = spam_abuse.IncreaseAbusedCounter("challenge:hard:"+address.Encoded, 5, 180*24*60*60, challengeSolved, tx); abuse || err != nil {
			return
		}

		if abuse, err = spam_abuse.IncreaseAbusedCounter("challenge:extreme:"+address.Encoded, 20, 180*24*60*60, challengeSolved, tx); abuse || err != nil {
			return
		}

		return nil
	})
}

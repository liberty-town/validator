package validator_store

import (
	"liberty-town/node/addresses"
	"liberty-town/node/store"
	"liberty-town/node/store/store_db/store_db_interface"
	"validator/config/arguments"
	"validator/validator_store/spam_abuse"
)

var Store *store.Store

func IsAbuseRequired(address *addresses.Address, size uint64) (abuse bool, err error) {

	err = Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {

		if abuse, err = spam_abuse.IsRegistered("required:5days:"+address.Encoded, 5*24*60*60, tx); abuse || err != nil {
			return
		}

		if size <= 1500 {
			if abuse, err = spam_abuse.IsAbused("small:10min:"+address.Encoded, 5, 10*60, tx); abuse || err != nil {
				return
			}
			if abuse, err = spam_abuse.IsAbused("small:60min:"+address.Encoded, 20, 60*60, tx); abuse || err != nil {
				return
			}
			if abuse, err = spam_abuse.IsAbused("small:1day:"+address.Encoded, 50, 24*60*60, tx); abuse || err != nil {
				return
			}
			if abuse, err = spam_abuse.IsAbused("small:7days:"+address.Encoded, 100, 7*24*60*60, tx); abuse || err != nil {
				return
			}
			return
		} else {
			abuse = true
			return nil
		}

	})

	return
}

//超出此限制时，服务器将返回abuse错误.
func IncreaseAbuse(address *addresses.Address, size uint64, challengeSolved bool) (abuse bool, err error) {
	err = Store.DB.Update(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {

		if abuse, err = spam_abuse.RequireRegistration("required:5days:"+address.Encoded, 5*24*60*60, challengeSolved, tx); abuse || err != nil {
			return
		}

		if size <= 1500 {
			if abuse, err = spam_abuse.IncreaseAbusedCounter("small:10min:"+address.Encoded, 5, 10*60, challengeSolved, tx); abuse || err != nil {
				return
			}
			if abuse, err = spam_abuse.IncreaseAbusedCounter("small:60min:"+address.Encoded, 20, 60*60, challengeSolved, tx); abuse || err != nil {
				return
			}
			if abuse, err = spam_abuse.IncreaseAbusedCounter("small:1day:"+address.Encoded, 50, 24*60*60, challengeSolved, tx); abuse || err != nil {
				return
			}
			if abuse, err = spam_abuse.IncreaseAbusedCounter("small:7days:"+address.Encoded, 100, 7*24*60*60, challengeSolved, tx); abuse || err != nil {
				return
			}
		} else {
			if !challengeSolved {
				abuse = true
				return
			}
		}

		return
	})
	return
}

func InitializeStore() (err error) {

	if Store, err = store.CreateStoreNow("validator", store.GetStoreType(arguments.Arguments["--store-data-type"].(string), store.AllowedStores)); err != nil {
		return
	}

	return
}

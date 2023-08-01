package db

import (
	"fmt"

	"github.com/Prajjawalk/zond-indexer/types"
	"github.com/lib/pq"
)

type WatchlistFilter struct {
	Tag            types.Tag
	UserId         uint64
	Validators     *pq.ByteaArray
	JoinValidators bool
	Network        string
}

// GetTaggedValidators returns validators that were tagged by a user
func GetTaggedValidators(filter WatchlistFilter) ([]*types.TaggedValidators, error) {
	list := []*types.TaggedValidators{}
	args := make([]interface{}, 0)

	// var userId uint64
	// SELECT users_validators_tags.user_id, users_validators_tags.validator_publickey, event_name
	// FROM users_validators_tags inner join users_subscriptions
	// ON users_validators_tags.user_id = users_subscriptions.user_id and ENCODE(users_validators_tags.validator_publickey::bytea, 'hex') = users_subscriptions.event_filter;
	tag := filter.Network + ":" + string(filter.Tag)
	args = append(args, tag)
	args = append(args, filter.UserId)
	qry := `
		SELECT user_id, validator_publickey, tag
		FROM users_validators_tags
		WHERE tag = $1 AND user_id = $2`

	if filter.Validators != nil {
		args = append(args, *filter.Validators)
		qry += " AND "
		qry += fmt.Sprintf("validator_publickey = ANY($%d)", len(args))
	}

	qry += " ORDER BY validator_publickey desc "
	err := FrontendWriterDB.Select(&list, qry, args...)
	if err != nil {
		return nil, err
	}
	if filter.JoinValidators && filter.Validators == nil {
		pubkeys := make([][]byte, 0, len(list))
		for _, li := range list {
			pubkeys = append(pubkeys, li.ValidatorPublickey)
		}
		pubBytea := pq.ByteaArray(pubkeys)
		filter.Validators = &pubBytea
	}

	validators := make([]*types.Validator, 0, len(list))
	if filter.JoinValidators {
		err := ReaderDb.Select(&validators, `SELECT balance, pubkey, validatorindex FROM validators WHERE pubkey = ANY($1) ORDER BY pubkey desc`, *filter.Validators)
		if err != nil {
			return nil, err
		}
		if len(list) != len(validators) {
			logger.Errorf("error could not get validators for watchlist. Expected to retrieve %v validators but got %v", len(list), len(validators))
			for i, li := range list {
				if li == nil {
					logger.Errorf("empty validator entry %v", list[i])
				} else {
					li.Validator = &types.Validator{}
				}
			}
			return list, nil
		}
		for i, li := range list {
			if li == nil {
				logger.Errorf("empty validator entry %v", list[i])
			} else {
				li.Validator = validators[i]
			}
		}
	}
	return list, nil
}

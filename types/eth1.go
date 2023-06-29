package types

import (
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type GetBlockTimings struct {
	Headers  time.Duration
	Receipts time.Duration
	Traces   time.Duration
}

type BulkMutations struct {
	Keys  []string
	Model []mongo.WriteModel
}

package services

import "database/sql"

type MachineEvents struct {
	SubscriptionID  uint64         `db:"id"`
	UserID          uint64         `db:"user_id"`
	MachineName     string         `db:"machine"`
	UnsubscribeHash sql.NullString `db:"unsubscribe_hash"`
	EventThreshold  float64        `db:"event_threshold"`
}

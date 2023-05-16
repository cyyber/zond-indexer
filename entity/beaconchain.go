package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

type AttestationFamily struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ChainId      string
	Type         string
	Epoch        uint64
	ValidatorId  uint64
	AttestorSlot uint64
}

type IncomeDetailsColumnFamily struct {
	ID                                 primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ValidatorId                        uint64
	AttestationSourceReward            uint64
	AttestationSourcePenalty           uint64
	AttestationTargetReward            uint64
	AttestationTargetPenalty           uint64
	AttestationHeadReward              uint64
	FinalityDelayPenalty               uint64
	ProposerSlashingInclusionReward    uint64
	ProposerAttestationInclusionReward uint64
	ProposerSyncInclusionReward        uint64
	SyncCommitteeReward                uint64
	SyncCommitteePenalty               uint64
	SlashingReward                     uint64
	SlashingPenalty                    uint64
	TxFeeRewardWei                     []byte
	ProposalsMissed                    uint64
}

type ProposalsFamily struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ChainId     string
	ValidatorId uint64
	Type        string
	Status      uint64
	Slot        uint64
	Epoch       uint64
}

type SyncCommitteesFamily struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ChainId     string
	ValidatorId uint64
	Type        string
	Slot        uint64
	Status      uint64
	Epoch       uint64
}

type Stats struct {
	ID                                 primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	AttestationSourceReward            uint64
	AttestationSourcePenalty           uint64
	AttestationTargetReward            uint64
	AttestationTargetPenalty           uint64
	AttestationHeadReward              uint64
	FinalityDelayPenalty               uint64
	ProposerSlashingInclusionReward    uint64
	ProposerAttestationInclusionReward uint64
	ProposerSyncInclusionReward        uint64
	SyncCommitteeReward                uint64
	SyncCommitteePenalty               uint64
	SlashingReward                     uint64
	SlashingPenalty                    uint64
	TxFeeRewardWei                     []byte
	ProposalsMissed                    uint64
}

type ValidatorBalancesFamily struct {
	ID               primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ChainId          string
	Epoch            uint64
	Type             string
	ValidatorId      uint64
	Balance          uint64
	EffectiveBalance uint64
}

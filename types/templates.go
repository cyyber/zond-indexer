package types

import (
	"html/template"
	"math/big"
	"time"
)

// DataTableResponse is a struct to hold data for data table responses
type DataTableResponse struct {
	Draw            uint64          `json:"draw"`
	RecordsTotal    uint64          `json:"recordsTotal"`
	RecordsFiltered uint64          `json:"recordsFiltered"`
	Data            [][]interface{} `json:"data"`
	PageLength      uint64          `json:"pageLength"`
	DisplayStart    uint64          `json:"displayStart"`
	PagingToken     string          `json:"pagingToken"`
}

type Transfer struct {
	From   template.HTML
	To     template.HTML
	Amount template.HTML
	Token  template.HTML
}

type AddressBalance struct {
	Address  []byte
	Token    []byte
	Balance  []byte
	Metadata *ERC20Metadata
}

type ERC20TokenPrice struct {
	Token       []byte
	Price       []byte
	TotalSupply []byte
}

type ERC20Metadata struct {
	Decimals     []byte
	Symbol       string
	Name         string
	Description  string
	Logo         []byte
	LogoFormat   string
	TotalSupply  []byte
	OfficialSite string
	Price        []byte
}

type AddressMetadata struct {
	Balances   []*AddressBalance
	ERC20      *ERC20Metadata
	Name       string
	Tags       []template.HTML
	EthBalance *AddressBalance
}

type AddressSearchItem struct {
	Address string
	Name    string
	Token   string
}

type GasNowHistory struct {
	Ts       time.Time
	Slow     *big.Int
	Standard *big.Int
	Fast     *big.Int
	Rapid    *big.Int
}

// ValidatorBalance is a struct for the validator balance data
type ValidatorBalance struct {
	Epoch            uint64 `db:"epoch"`
	Balance          uint64 `db:"balance"`
	EffectiveBalance uint64 `db:"effectivebalance"`
	Index            uint64 `db:"validatorindex"`
	PublicKey        []byte `db:"pubkey"`
}

// ValidatorPerformance is a struct for the validator performance data
type ValidatorPerformance struct {
	Rank            uint64 `db:"rank"`
	Index           uint64 `db:"validatorindex"`
	PublicKey       []byte `db:"pubkey"`
	Name            string `db:"name"`
	Balance         uint64 `db:"balance"`
	Performance1d   int64  `db:"performance1d"`
	Performance7d   int64  `db:"performance7d"`
	Performance31d  int64  `db:"performance31d"`
	Performance365d int64  `db:"performance365d"`
	Rank7d          int64  `db:"rank7d"`
	TotalCount      uint64 `db:"total_count"`
}

// ValidatorAttestation is a struct for the validators attestations data
type ValidatorAttestation struct {
	Index          uint64
	Epoch          uint64 `db:"epoch"`
	AttesterSlot   uint64 `db:"attesterslot"`
	CommitteeIndex uint64 `db:"committeeindex"`
	Status         uint64 `db:"status"`
	InclusionSlot  uint64 `db:"inclusionslot"`
	Delay          int64  `db:"delay"`
	// EarliestInclusionSlot uint64 `db:"earliestinclusionslot"`
}

// ValidatorSyncParticipation hold information about sync-participation of a validator
type ValidatorSyncParticipation struct {
	Period uint64 `db:"period"`
	Slot   uint64 `db:"slot"`
	Status uint64 `db:"status"`
}

type ValidatorBalanceStatistic struct {
	Index                 uint64
	MinEffectiveBalance   uint64
	MaxEffectiveBalance   uint64
	MinBalance            uint64
	MaxBalance            uint64
	StartEffectiveBalance uint64
	EndEffectiveBalance   uint64
	StartBalance          uint64
	EndBalance            uint64
}

type ValidatorMissedAttestationsStatistic struct {
	Index              uint64
	MissedAttestations uint64
}

type ValidatorSyncDutiesStatistic struct {
	Index            uint64
	ParticipatedSync uint64
	MissedSync       uint64
}

type ValidatorWithdrawal struct {
	Index  uint64
	Epoch  uint64
	Slot   uint64
	Amount uint64
}

type ValidatorProposal struct {
	Index  uint64
	Slot   uint64
	Status uint64
}

type ValidatorEffectiveness struct {
	Validatorindex        uint64  `json:"validatorindex"`
	AttestationEfficiency float64 `json:"attestation_efficiency"`
}

type Eth1AddressBalance struct {
	Address  []byte
	Token    []byte
	Balance  []byte
	Metadata *ERC20Metadata
}

type Eth1AddressMetadata struct {
	Balances   []*Eth1AddressBalance
	ERC20      *ERC20Metadata
	Name       string
	Tags       []template.HTML
	EthBalance *Eth1AddressBalance
}

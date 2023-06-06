package types

import (
	"html/template"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
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

type ContractMetadata struct {
	Name    string
	ABI     *abi.ABI `msgpack:"-"`
	ABIJson []byte
}

type EtherscanContractMetadata struct {
	Message string `json:"message"`
	Result  []struct {
		Abi                  string `json:"ABI"`
		CompilerVersion      string `json:"CompilerVersion"`
		ConstructorArguments string `json:"ConstructorArguments"`
		ContractName         string `json:"ContractName"`
		EVMVersion           string `json:"EVMVersion"`
		Implementation       string `json:"Implementation"`
		Library              string `json:"Library"`
		LicenseType          string `json:"LicenseType"`
		OptimizationUsed     string `json:"OptimizationUsed"`
		Proxy                string `json:"Proxy"`
		Runs                 string `json:"Runs"`
		SourceCode           string `json:"SourceCode"`
		SwarmSource          string `json:"SwarmSource"`
	} `json:"result"`
	Status string `json:"status"`
}

// EpochsPageData is a struct to hold epoch data for the epochs page
type EthOneDepositsData struct {
	TxHash                []byte    `db:"tx_hash"`
	TxInput               []byte    `db:"tx_input"`
	TxIndex               uint64    `db:"tx_index"`
	BlockNumber           uint64    `db:"block_number"`
	BlockTs               time.Time `db:"block_ts"`
	FromAddress           []byte    `db:"from_address"`
	PublicKey             []byte    `db:"publickey"`
	WithdrawalCredentials []byte    `db:"withdrawal_credentials"`
	Amount                uint64    `db:"amount"`
	Signature             []byte    `db:"signature"`
	MerkletreeIndex       []byte    `db:"merkletree_index"`
	State                 string    `db:"state"`
	ValidSignature        bool      `db:"valid_signature"`
}

type EthOneDepositLeaderboardData struct {
	FromAddress        []byte `db:"from_address"`
	Amount             uint64 `db:"amount"`
	ValidCount         uint64 `db:"validcount"`
	InvalidCount       uint64 `db:"invalidcount"`
	TotalCount         uint64 `db:"totalcount"`
	PendingCount       uint64 `db:"pendingcount"`
	SlashedCount       uint64 `db:"slashedcount"`
	ActiveCount        uint64 `db:"activecount"`
	VoluntaryExitCount uint64 `db:"voluntary_exit_count"`
}

type EthTwoDepositData struct {
	BlockSlot             uint64 `db:"block_slot"`
	BlockIndex            uint64 `db:"block_index"`
	Proof                 []byte `db:"proof"`
	Publickey             []byte `db:"publickey"`
	ValidatorIndex        uint64 `db:"validatorindex"`
	Withdrawalcredentials []byte `db:"withdrawalcredentials"`
	Amount                uint64 `db:"amount"`
	Signature             []byte `db:"signature"`
}

type ValidatorDeposits struct {
	Eth1Deposits      []Eth1Deposit
	LastEth1DepositTs int64
	Eth2Deposits      []Eth2Deposit
}

// Eth1Deposit is a struct to hold eth1-deposit data
type Eth1Deposit struct {
	TxHash                []byte `db:"tx_hash"`
	TxInput               []byte `db:"tx_input"`
	TxIndex               uint64 `db:"tx_index"`
	BlockNumber           uint64 `db:"block_number"`
	BlockTs               int64  `db:"block_ts"`
	FromAddress           []byte `db:"from_address"`
	PublicKey             []byte `db:"publickey"`
	WithdrawalCredentials []byte `db:"withdrawal_credentials"`
	Amount                uint64 `db:"amount"`
	Signature             []byte `db:"signature"`
	MerkletreeIndex       []byte `db:"merkletree_index"`
	Removed               bool   `db:"removed"`
	ValidSignature        bool   `db:"valid_signature"`
}

// Eth2Deposit is a struct to hold eth2-deposit data
type Eth2Deposit struct {
	BlockSlot             uint64 `db:"block_slot"`
	BlockIndex            uint64 `db:"block_index"`
	BlockRoot             []byte `db:"block_root"`
	Proof                 []byte `db:"proof"`
	Publickey             []byte `db:"publickey"`
	Withdrawalcredentials []byte `db:"withdrawalcredentials"`
	Amount                uint64 `db:"amount"`
	Signature             []byte `db:"signature"`
}

type SlotVizSlots struct {
	BlockRoot []byte
	Epoch     uint64
	Slot      uint64
	Status    string `json:"status"`
	Active    bool   `json:"active"`
}
type SlotVizEpochs struct {
	Epoch          uint64          `json:"epoch"`
	Finalized      bool            `json:"finalized"`
	Justified      bool            `json:"justified"`
	Justifying     bool            `json:"justifying"`
	Particicpation float64         `json:"participation"`
	Slots          []*SlotVizSlots `json:"slots"`
}

// AdConfig is a struct to hold the configuration for one specific ad banner placement
type AdConfig struct {
	Id              string `db:"id"`
	TemplateId      string `db:"template_id"`
	JQuerySelector  string `db:"jquery_selector"`
	InsertMode      string `db:"insert_mode"`
	RefreshInterval uint64 `db:"refresh_interval"`
	Enabled         bool   `db:"enabled"`
	ForAllUsers     bool   `db:"for_all_users"`
	BannerId        uint64 `db:"banner_id"`
	HtmlContent     string `db:"html_content"`
}

type BLSChange struct {
	Slot           uint64 `db:"slot" json:"slot,omitempty"`
	BlockRoot      []byte `db:"block_rot" json:"blockroot,omitempty"`
	Validatorindex uint64 `db:"validatorindex" json:"validatorindex,omitempty"`
	BlsPubkey      []byte `db:"pubkey" json:"pubkey,omitempty"`
	Address        []byte `db:"address" json:"address,omitempty"`
	Signature      []byte `db:"signature" json:"signature,omitempty"`
}

type ExplorerConfigurationCategory string
type ExplorerConfigurationKey string
type ExplorerConfigValue struct {
	Value    string `db:"value"`
	DataType string `db:"data_type"`
}
type ExplorerConfig struct {
	Category ExplorerConfigurationCategory `db:"category"`
	Key      ExplorerConfigurationKey      `db:"key"`
	ExplorerConfigValue
}

type ValidatorsBLSChange struct {
	Slot                     uint64 `db:"slot" json:"slot,omitempty"`
	BlockRoot                []byte `db:"block_root" json:"blockroot,omitempty"`
	Validatorindex           uint64 `db:"validatorindex" json:"validatorindex,omitempty"`
	BlsPubkey                []byte `db:"pubkey" json:"pubkey,omitempty"`
	Address                  []byte `db:"address" json:"address,omitempty"`
	Signature                []byte `db:"signature" json:"signature,omitempty"`
	WithdrawalCredentialsOld []byte `db:"withdrawalcredentials" json:"withdrawalcredentials,omitempty"`
}

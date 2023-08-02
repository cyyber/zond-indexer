package types

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/lib/pq"
)

// PageData is a struct to hold web page data
type PageData struct {
	Active                string
	AdConfigurations      []*AdConfig
	Meta                  *Meta
	ShowSyncingMessage    bool
	User                  *User
	Data                  interface{}
	Version               string
	Year                  int
	ChainSlotsPerEpoch    uint64
	ChainSecondsPerSlot   uint64
	ChainGenesisTimestamp uint64
	CurrentEpoch          uint64
	LatestFinalizedEpoch  uint64
	CurrentSlot           uint64
	FinalizationDelay     uint64
	Mainnet               bool
	DepositContract       string
	Rates                 PageRates
	InfoBanner            *template.HTML
	ClientsUpdated        bool
	// IsUserClientUpdated   func(uint64) bool
	ChainConfig         ChainConfig
	Lang                string
	NoAds               bool
	Debug               bool
	DebugTemplates      []string
	DebugSession        map[string]interface{}
	GasNow              *GasNowPageData
	GlobalNotification  template.HTML
	AvailableCurrencies []string
	MainMenuItems       []MainMenuItem
}

type PageRates struct {
	EthPrice               float64
	EthRoundPrice          uint64
	EthTruncPrice          template.HTML
	UsdRoundPrice          uint64
	UsdTruncPrice          template.HTML
	EurRoundPrice          uint64
	EurTruncPrice          template.HTML
	GbpRoundPrice          uint64
	GbpTruncPrice          template.HTML
	CnyRoundPrice          uint64
	CnyTruncPrice          template.HTML
	RubRoundPrice          uint64
	RubTruncPrice          template.HTML
	CadRoundPrice          uint64
	CadTruncPrice          template.HTML
	AudRoundPrice          uint64
	AudTruncPrice          template.HTML
	JpyRoundPrice          uint64
	JpyTruncPrice          template.HTML
	Currency               string
	CurrentPriceFormatted  template.HTML
	CurrentPriceKFormatted template.HTML
	CurrentSymbol          string
	ExchangeRate           float64
}

// Meta is a struct to hold metadata about the page
type Meta struct {
	Title       string
	Description string
	Path        string
	Tlabel1     string
	Tdata1      string
	Tlabel2     string
	Tdata2      string
	GATag       string
	NoTrack     bool
	Templates   string
}

type User struct {
	UserID        uint64 `json:"user_id"`
	Authenticated bool   `json:"authenticated"`
	Subscription  string `json:"subscription"`
	UserGroup     string `json:"user_group"`
}

type MainMenuItem struct {
	Label        string
	Path         string
	IsActive     bool
	HasBigGroups bool // if HasBigGroups is set to true then the NavigationGroups will be ordered horizontally and their Label will be shown
	Groups       []NavigationGroup
}

type NavigationGroup struct {
	Label string // only used for "BigGroups"
	Links []NavigationLink
}

type NavigationLink struct {
	Label         string
	Path          string
	CustomIcon    string
	Icon          string
	IsHidden      bool
	IsHighlighted bool
}

type Empty struct {
}

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

type ChartDataPoint struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Color string  `json:"color"`
}

// DashboardValidatorBalanceHistory is a struct to hold data for the balance-history on the dashboard-page
type DashboardValidatorBalanceHistory struct {
	Epoch            uint64  `db:"epoch"`
	Balance          uint64  `db:"balance"`
	EffectiveBalance uint64  `db:"effectivebalance"`
	ValidatorCount   float64 `db:"validatorcount"`
}

// ValidatorBalance is a struct for the validator balance data
type ValidatorBalance struct {
	Epoch            uint64 `db:"epoch"`
	Balance          uint64 `db:"balance"`
	EffectiveBalance uint64 `db:"effectivebalance"`
	Index            uint64 `db:"validatorindex"`
	PublicKey        []byte `db:"pubkey"`
}

// ValidatorBalanceHistory is a struct for the validator income history data
type ValidatorIncomeHistory struct {
	Day              int64         `db:"day"` // day can be -1 which is pre-genesis
	ClRewards        int64         `db:"cl_rewards_gwei"`
	EndBalance       sql.NullInt64 `db:"end_balance"`
	StartBalance     sql.NullInt64 `db:"start_balance"`
	DepositAmount    sql.NullInt64 `db:"deposits_amount"`
	WithdrawalAmount sql.NullInt64 `db:"withdrawals_amount"`
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

type ValidatorsBLSChange struct {
	Slot                     uint64 `db:"slot" json:"slot,omitempty"`
	BlockRoot                []byte `db:"block_root" json:"blockroot,omitempty"`
	Validatorindex           uint64 `db:"validatorindex" json:"validatorindex,omitempty"`
	BlsPubkey                []byte `db:"pubkey" json:"pubkey,omitempty"`
	Address                  []byte `db:"address" json:"address,omitempty"`
	Signature                []byte `db:"signature" json:"signature,omitempty"`
	WithdrawalCredentialsOld []byte `db:"withdrawalcredentials" json:"withdrawalcredentials,omitempty"`
}

type DataTableSaveStateSearch struct {
	Search          string `json:"search"`          // Search term
	Regex           bool   `json:"regex"`           // Indicate if the search term should be treated as regex or not
	Smart           bool   `json:"smart"`           // Flag to enable DataTables smart search
	CaseInsensitive bool   `json:"caseInsensitive"` // Case insensitive flag
}

type DataTableSaveStateColumns struct {
	Visible bool                     `json:"visible"`
	Search  DataTableSaveStateSearch `json:"search"`
}

type DataTableSaveState struct {
	Key     string                      `json:"key"`
	Time    uint64                      `json:"time"`   // Time stamp of when the object was created
	Start   uint64                      `json:"start"`  // Display start point
	Length  uint64                      `json:"length"` // Page length
	Order   [][]string                  `json:"order"`  // 2D array of column ordering information (see `order` option)
	Search  DataTableSaveStateSearch    `json:"search"`
	Columns []DataTableSaveStateColumns `json:"columns"`
}

// GenericChartData is a struct to hold chart data
type GenericChartData struct {
	IsNormalChart                   bool
	ShowGapHider                    bool
	XAxisLabelsFormatter            template.JS
	TooltipFormatter                template.JS
	TooltipShared                   bool
	TooltipUseHTML                  bool
	TooltipSplit                    bool
	TooltipFollowPointer            bool
	PlotOptionsSeriesEventsClick    template.JS
	PlotOptionsPie                  template.JS
	DataLabelsEnabled               bool
	DataLabelsFormatter             template.JS
	PlotOptionsSeriesCursor         string
	Title                           string                    `json:"title"`
	Subtitle                        string                    `json:"subtitle"`
	XAxisTitle                      string                    `json:"x_axis_title"`
	YAxisTitle                      string                    `json:"y_axis_title"`
	Type                            string                    `json:"type"`
	StackingMode                    string                    `json:"stacking_mode"`
	ColumnDataGroupingApproximation string                    // "average", "averages", "open", "high", "low", "close" and "sum"
	Series                          []*GenericChartDataSeries `json:"series"`
	Drilldown                       interface{}               `json:"drilldown"`
}

// GenericChartDataSeries is a struct to hold chart series data
type GenericChartDataSeries struct {
	Name  string      `json:"name"`
	Data  interface{} `json:"data"`
	Stack string      `json:"stack,omitempty"`
	Type  string      `json:"type,omitempty"`
	Color string      `json:"color,omitempty"`
}

// ChartsPageDataChart is a struct to hold a chart for the charts-page
type ChartsPageDataChart struct {
	Order  int
	Path   string
	Data   *GenericChartData
	Height int
}

type Stats struct {
	TopDepositors                  *[]StatsTopDepositors
	InvalidDepositCount            *uint64 `db:"count"`
	UniqueValidatorCount           *uint64 `db:"count"`
	TotalValidatorCount            *uint64 `db:"count"`
	ActiveValidatorCount           *uint64 `db:"count"`
	PendingValidatorCount          *uint64 `db:"count"`
	ValidatorChurnLimit            *uint64
	LatestValidatorWithdrawalIndex *uint64 `db:"index"`
	WithdrawableValidatorCount     *uint64 `db:"count"`
	// WithdrawableAmount             *uint64 `db:"amount"`
	PendingBLSChangeValidatorCount *uint64 `db:"count"`
	NonWithdrawableCount           *uint64 `db:"count"`
	TotalAmountWithdrawn           *uint64 `db:"amount"`
	WithdrawalCount                *uint64 `db:"count"`
	TotalAmountDeposited           *uint64 `db:"amount"`
	BLSChangeCount                 *uint64 `db:"count"`
}

type StatsTopDepositors struct {
	Address      string `db:"from_address"`
	DepositCount uint64 `db:"count"`
}

type PoolsResp struct {
	PoolsDistribution       ChartsPageDataChart
	HistoricPoolPerformance ChartsPageDataChart
	PoolInfos               []*PoolInfo
}

type PoolInfo struct {
	Name                   string  `db:"name"`
	Count                  int64   `db:"count"`
	AvgPerformance31d      float64 `db:"avg_performance_31d"`
	AvgPerformance7d       float64 `db:"avg_performance_7d"`
	AvgPerformance1d       float64 `db:"avg_performance_1d"`
	EthstoreCompoarison1d  float64
	EthstoreCompoarison7d  float64
	EthstoreCompoarison31d float64
}

// IndexPageData is a struct to hold info for the main web page
type IndexPageData struct {
	NetworkName               string `json:"networkName"`
	DepositContract           string `json:"depositContract"`
	ShowSyncingMessage        bool
	CurrentEpoch              uint64                 `json:"current_epoch"`
	CurrentFinalizedEpoch     uint64                 `json:"current_finalized_epoch"`
	CurrentSlot               uint64                 `json:"current_slot"`
	ScheduledCount            uint8                  `json:"scheduled_count"`
	FinalityDelay             uint64                 `json:"finality_delay"`
	ActiveValidators          uint64                 `json:"active_validators"`
	EnteringValidators        uint64                 `json:"entering_validators"`
	ExitingValidators         uint64                 `json:"exiting_validators"`
	StakedEther               string                 `json:"staked_ether"`
	AverageBalance            string                 `json:"average_balance"`
	DepositedTotal            float64                `json:"deposit_total"`
	DepositThreshold          float64                `json:"deposit_threshold"`
	ValidatorsRemaining       float64                `json:"validators_remaining"`
	NetworkStartTs            int64                  `json:"network_start_ts"`
	MinGenesisTime            int64                  `json:"minGenesisTime"`
	Blocks                    []*IndexPageDataBlocks `json:"blocks"`
	Epochs                    []*IndexPageDataEpochs `json:"epochs"`
	StakedEtherChartData      [][]float64            `json:"staked_ether_chart_data"`
	ActiveValidatorsChartData [][]float64            `json:"active_validators_chart_data"`
	Subtitle                  template.HTML          `json:"subtitle"`
	Genesis                   bool                   `json:"genesis"`
	GenesisPeriod             bool                   `json:"genesis_period"`
	Mainnet                   bool                   `json:"mainnet"`
	DepositChart              *ChartsPageDataChart
	DepositDistribution       *ChartsPageDataChart
	Countdown                 interface{}
	SlotVizData               *SlotVizPageData `json:"slotVizData"`
}

type SlotVizPageData struct {
	Epochs   []*SlotVizEpochs
	Selector string
	Config   ExplorerConfigurationKeyMap
}

type IndexPageDataEpochs struct {
	Epoch                            uint64        `json:"epoch"`
	Ts                               time.Time     `json:"ts"`
	Finalized                        bool          `json:"finalized"`
	FinalizedFormatted               template.HTML `json:"finalized_formatted"`
	EligibleEther                    uint64        `json:"eligibleether"`
	EligibleEtherFormatted           template.HTML `json:"eligibleether_formatted"`
	GlobalParticipationRate          float64       `json:"globalparticipationrate"`
	GlobalParticipationRateFormatted template.HTML `json:"globalparticipationrate_formatted"`
	VotedEther                       uint64        `json:"votedether"`
	VotedEtherFormatted              template.HTML `json:"votedether_formatted"`
}

// IndexPageDataBlocks is a struct to hold detail data for the main web page
type IndexPageDataBlocks struct {
	Epoch                uint64        `json:"epoch"`
	Slot                 uint64        `json:"slot"`
	Ts                   time.Time     `json:"ts"`
	Proposer             uint64        `db:"proposer" json:"proposer"`
	ProposerFormatted    template.HTML `json:"proposer_formatted"`
	BlockRoot            []byte        `db:"blockroot" json:"block_root"`
	BlockRootFormatted   string        `json:"block_root_formatted"`
	ParentRoot           []byte        `db:"parentroot" json:"parent_root"`
	Attestations         uint64        `db:"attestationscount" json:"attestations"`
	Deposits             uint64        `db:"depositscount" json:"deposits"`
	Withdrawals          uint64        `db:"withdrawalcount" json:"withdrawals"`
	Exits                uint64        `db:"voluntaryexitscount" json:"exits"`
	Proposerslashings    uint64        `db:"proposerslashingscount" json:"proposerslashings"`
	Attesterslashings    uint64        `db:"attesterslashingscount" json:"attesterslashings"`
	SyncAggParticipation float64       `db:"syncaggregate_participation" json:"sync_aggregate_participation"`
	Status               uint64        `db:"status" json:"status"`
	StatusFormatted      template.HTML `json:"status_formatted"`
	Votes                uint64        `db:"votes" json:"votes"`
	Graffiti             []byte        `db:"graffiti"`
	ProposerName         string        `db:"name"`
	ExecutionBlockNumber int           `db:"exec_block_number" json:"exec_block_number"`
}

// IndexPageEpochHistory is a struct to hold the epoch history for the main web page
type IndexPageEpochHistory struct {
	Epoch                   uint64 `db:"epoch"`
	ValidatorsCount         uint64 `db:"validatorscount"`
	EligibleEther           uint64 `db:"eligibleether"`
	Finalized               bool   `db:"finalized"`
	AverageValidatorBalance uint64 `db:"averagevalidatorbalance"`
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
type ExplorerConfigurationKeyMap map[ExplorerConfigurationKey]ExplorerConfigValue
type ExplorerConfigurationMap map[ExplorerConfigurationCategory]ExplorerConfigurationKeyMap

func (configMap ExplorerConfigurationMap) GetConfigValue(category ExplorerConfigurationCategory, configKey ExplorerConfigurationKey) (ExplorerConfigValue, error) {
	configValue := ExplorerConfigValue{}
	keyMap, ok := configMap[category]
	if ok {
		configValue, ok = keyMap[configKey]
		if ok {
			return configValue, nil
		}
	}
	return configValue, fmt.Errorf("config value for %s %s not found", category, configKey)
}

func (configMap ExplorerConfigurationMap) GetUInt64Value(category ExplorerConfigurationCategory, configKey ExplorerConfigurationKey) (uint64, error) {
	configValue, err := configMap.GetConfigValue(category, configKey)
	if err == nil {
		if configValue.DataType != "int" {
			return 0, fmt.Errorf("wrong data type for %s %s, got %s, expected %s", category, configKey, configValue.DataType, "int")
		} else {
			return strconv.ParseUint(configValue.Value, 10, 64)
		}
	}
	return 0, err
}

func (configMap ExplorerConfigurationMap) GetStringValue(category ExplorerConfigurationCategory, configKey ExplorerConfigurationKey) (string, error) {
	configValue, err := configMap.GetConfigValue(category, configKey)
	return configValue.Value, err
}

type EthStoreStatistics struct {
	EffectiveBalances         [][]float64
	TotalRewards              [][]float64
	APRs                      [][]float64
	YesterdayRewards          float64
	YesterdayEffectiveBalance float64
	ProjectedAPR              float64
	YesterdayTs               int64
	StartEpoch                uint64
}

type BurnPageDataBlock struct {
	Number        int64     `json:"number"`
	Hash          string    `json:"hash"`
	GasTarget     int64     `json:"gas_target" db:"gaslimit"`
	GasUsed       int64     `json:"gas_used" db:"gasused"`
	Rewards       float64   `json:"mining_reward" db:"miningreward"`
	Txn           int       `json:"tx_count" db:"tx_count"`
	Age           time.Time `json:"time" db:"time"`
	BaseFeePerGas float64   `json:"base_fee_per_gas" db:"basefeepergas"`
	BurnedFees    float64   `json:"burned_fees" db:"burnedfees"`
}

type BurnPageData struct {
	TotalBurned      float64              `json:"total_burned"`
	Blocks           []*BurnPageDataBlock `json:"blocks"`
	BaseFeeTrend     int                  `json:"base_fee_trend"`
	BurnRate1h       float64              `json:"burn_rate_1_h"`
	BurnRate24h      float64              `json:"burn_rate_24_h"`
	BlockUtilization float64              `json:"block_utilization"`
	Emission         float64              `json:"emission"`
	Price            float64              `json:"price_usd"`
	Currency         string               `json:"currency"`
}

type RelaysResp struct {
	RelaysInfoContainers [3]RelayInfoContainer
	RecentBlocks         []*RelaysRespBlock
	TopBlocks            []*RelaysRespBlock
	LastUpdated          time.Time
	TopBuilders          []*struct {
		Tags       TagMetadataSlice `db:"tags"`
		Builder    []byte           `db:"builder_pubkey"`
		BlockCount uint64           `db:"c"`
		LatestSlot uint64           `db:"latest_slot"`
		BlockPerc  float64
	}
}

type RelaysRespBlock struct {
	Tags                 TagMetadataSlice `db:"tags"`
	Value                WeiString        `db:"value"`
	Slot                 uint64           `db:"slot"`
	Builder              []byte           `db:"builder_pubkey"`
	ProposerFeeRecipient []byte           `db:"proposer_fee_recipient"`
	Proposer             uint64           `db:"proposer"`
	BlockExtraData       string           `db:"block_extra_data"`
}

type RelayInfoContainer struct {
	Days                 uint64
	IsFirst              bool
	RelaysInfo           []*RelayInfo
	NetworkParticipation float64
}

type RelayInfo struct {
	RelayID        string         `db:"relay_id"`
	Name           sql.NullString `db:"name"`
	Link           sql.NullString `db:"link"`
	Censors        sql.NullBool   `db:"censors"`
	Ethical        sql.NullBool   `db:"ethical"`
	BlockCount     uint64         `db:"block_count"`
	UniqueBuilders uint64         `db:"unique_builders"`
	NetworkUsage   float64        `db:"network_usage"`
	TotalValue     WeiString      `db:"total_value"`
	AverageValue   WeiString      `db:"avg_value"`
	MaxValue       WeiString      `db:"max_value"`
	MaxValueSlot   uint64         `db:"max_value_slot"`
}

// LatestState is a struct to hold data for the banner
type LatestState struct {
	LastProposedSlot      uint64        `json:"lastProposedSlot"`
	CurrentSlot           uint64        `json:"currentSlot"`
	CurrentEpoch          uint64        `json:"currentEpoch"`
	CurrentFinalizedEpoch uint64        `json:"currentFinalizedEpoch"`
	FinalityDelay         uint64        `json:"finalityDelay"`
	IsSyncing             bool          `json:"syncing"`
	EthPrice              float64       `json:"ethPrice"`
	EthRoundPrice         uint64        `json:"ethRoundPrice"`
	EthTruncPrice         template.HTML `json:"ethTruncPrice"`
	UsdRoundPrice         uint64        `json:"usdRoundPrice"`
	UsdTruncPrice         template.HTML `json:"usdTruncPrice"`
	EurRoundPrice         uint64        `json:"eurRoundPrice"`
	EurTruncPrice         template.HTML `json:"eurTruncPrice"`
	GbpRoundPrice         uint64        `json:"gbpRoundPrice"`
	GbpTruncPrice         template.HTML `json:"gbpTruncPrice"`
	CnyRoundPrice         uint64        `json:"cnyRoundPrice"`
	CnyTruncPrice         template.HTML `json:"cnyTruncPrice"`
	RubRoundPrice         uint64        `json:"rubRoundPrice"`
	RubTruncPrice         template.HTML `json:"rubTruncPrice"`
	CadRoundPrice         uint64        `json:"cadRoundPrice"`
	CadTruncPrice         template.HTML `json:"cadTruncPrice"`
	AudRoundPrice         uint64        `json:"audRoundPrice"`
	AudTruncPrice         template.HTML `json:"audTruncPrice"`
	JpyRoundPrice         uint64        `json:"jpyRoundPrice"`
	JpyTruncPrice         template.HTML `json:"jpyTruncPrice"`
	Currency              string        `json:"currency"`
}

// EpochPageData is a struct to hold detailed epoch data for the epoch page
type EpochPageData struct {
	Epoch                   uint64        `db:"epoch"`
	BlocksCount             uint64        `db:"blockscount"`
	ProposerSlashingsCount  uint64        `db:"proposerslashingscount"`
	AttesterSlashingsCount  uint64        `db:"attesterslashingscount"`
	AttestationsCount       uint64        `db:"attestationscount"`
	DepositsCount           uint64        `db:"depositscount"`
	WithdrawalCount         uint64        `db:"withdrawalcount"`
	DepositTotal            uint64        `db:"deposittotal"`
	WithdrawalTotal         template.HTML `db:"withdrawaltotal"`
	VoluntaryExitsCount     uint64        `db:"voluntaryexitscount"`
	ValidatorsCount         uint64        `db:"validatorscount"`
	AverageValidatorBalance uint64        `db:"averagevalidatorbalance"`
	Finalized               bool          `db:"finalized"`
	EligibleEther           uint64        `db:"eligibleether"`
	GlobalParticipationRate float64       `db:"globalparticipationrate"`
	VotedEther              uint64        `db:"votedether"`

	Blocks []*IndexPageDataBlocks

	SyncParticipationRate float64
	Ts                    time.Time
	NextEpoch             uint64
	PreviousEpoch         uint64
	ProposedCount         uint64
	MissedCount           uint64
	ScheduledCount        uint64
	OrphanedCount         uint64
}

// EpochsPageData is a struct to hold epoch data for the epochs page
type EpochsPageData struct {
	Epoch                   uint64  `db:"epoch"`
	BlocksCount             uint64  `db:"blockscount"`
	ProposerSlashingsCount  uint64  `db:"proposerslashingscount"`
	AttesterSlashingsCount  uint64  `db:"attesterslashingscount"`
	AttestationsCount       uint64  `db:"attestationscount"`
	DepositsCount           uint64  `db:"depositscount"`
	WithdrawalCount         uint64  `db:"withdrawalcount"`
	VoluntaryExitsCount     uint64  `db:"voluntaryexitscount"`
	ValidatorsCount         uint64  `db:"validatorscount"`
	AverageValidatorBalance uint64  `db:"averagevalidatorbalance"`
	Finalized               bool    `db:"finalized"`
	EligibleEther           uint64  `db:"eligibleether"`
	GlobalParticipationRate float64 `db:"globalparticipationrate"`
	VotedEther              uint64  `db:"votedether"`
}

// BlockVote stores a vote for a given block
type BlockVote struct {
	Validator      uint64 `db:"validator"`
	IncludedIn     uint64 `db:"included_in"`
	CommitteeIndex uint64 `db:"committee_index"`
}

// BlockPageTransaction is a struct to hold execution transactions on the block page
type BlockPageTransaction struct {
	BlockSlot    uint64 `db:"block_slot"`
	BlockIndex   uint64 `db:"block_index"`
	TxHash       []byte `db:"txhash"`
	AccountNonce uint64 `db:"nonce"`
	// big endian
	Price       []byte `db:"gas_price"`
	PricePretty string
	GasLimit    uint64 `db:"gas_limit"`
	Sender      []byte `db:"sender"`
	Recipient   []byte `db:"recipient"`
	// big endian
	Amount       []byte `db:"amount"`
	AmountPretty string
	Payload      []byte `db:"payload"`

	// TODO: transaction type

	MaxPriorityFeePerGas uint64 `db:"max_priority_fee_per_gas"`
	MaxFeePerGas         uint64 `db:"max_fee_per_gas"`
}

// BlockPageAttestation is a struct to hold attestations on the block page
type BlockPageAttestation struct {
	BlockSlot       uint64        `db:"block_slot"`
	BlockIndex      uint64        `db:"block_index"`
	AggregationBits []byte        `db:"aggregationbits"`
	Validators      pq.Int64Array `db:"validators"`
	Signature       []byte        `db:"signature"`
	Slot            uint64        `db:"slot"`
	CommitteeIndex  uint64        `db:"committeeindex"`
	BeaconBlockRoot []byte        `db:"beaconblockroot"`
	SourceEpoch     uint64        `db:"source_epoch"`
	SourceRoot      []byte        `db:"source_root"`
	TargetEpoch     uint64        `db:"target_epoch"`
	TargetRoot      []byte        `db:"target_root"`
}

// BlockPageDeposit is a struct to hold data for deposits on the block page
type BlockPageDeposit struct {
	PublicKey             []byte `db:"publickey"`
	WithdrawalCredentials []byte `db:"withdrawalcredentials"`
	Amount                uint64 `db:"amount"`
	Signature             []byte `db:"signature"`
}

// BlockPageVoluntaryExits is a struct to hold data for voluntary exits on the block page
type BlockPageVoluntaryExits struct {
	ValidatorIndex uint64 `db:"validatorindex"`
	Signature      []byte `db:"signature"`
}

// BlockPageAttesterSlashing is a struct to hold data for attester slashings on the block page
type BlockPageAttesterSlashing struct {
	BlockSlot                   uint64        `db:"block_slot"`
	BlockIndex                  uint64        `db:"block_index"`
	Attestation1Indices         pq.Int64Array `db:"attestation1_indices"`
	Attestation1Signature       []byte        `db:"attestation1_signature"`
	Attestation1Slot            uint64        `db:"attestation1_slot"`
	Attestation1Index           uint64        `db:"attestation1_index"`
	Attestation1BeaconBlockRoot []byte        `db:"attestation1_beaconblockroot"`
	Attestation1SourceEpoch     uint64        `db:"attestation1_source_epoch"`
	Attestation1SourceRoot      []byte        `db:"attestation1_source_root"`
	Attestation1TargetEpoch     uint64        `db:"attestation1_target_epoch"`
	Attestation1TargetRoot      []byte        `db:"attestation1_target_root"`
	Attestation2Indices         pq.Int64Array `db:"attestation2_indices"`
	Attestation2Signature       []byte        `db:"attestation2_signature"`
	Attestation2Slot            uint64        `db:"attestation2_slot"`
	Attestation2Index           uint64        `db:"attestation2_index"`
	Attestation2BeaconBlockRoot []byte        `db:"attestation2_beaconblockroot"`
	Attestation2SourceEpoch     uint64        `db:"attestation2_source_epoch"`
	Attestation2SourceRoot      []byte        `db:"attestation2_source_root"`
	Attestation2TargetEpoch     uint64        `db:"attestation2_target_epoch"`
	Attestation2TargetRoot      []byte        `db:"attestation2_target_root"`
	SlashedValidators           []int64
}

// BlockPageProposerSlashing is a struct to hold data for proposer slashings on the block page
type BlockPageProposerSlashing struct {
	BlockSlot         uint64 `db:"block_slot"`
	BlockIndex        uint64 `db:"block_index"`
	BlockRoot         []byte `db:"block_root" json:"block_root"`
	ProposerIndex     uint64 `db:"proposerindex"`
	Header1Slot       uint64 `db:"header1_slot"`
	Header1ParentRoot []byte `db:"header1_parentroot"`
	Header1StateRoot  []byte `db:"header1_stateroot"`
	Header1BodyRoot   []byte `db:"header1_bodyroot"`
	Header1Signature  []byte `db:"header1_signature"`
	Header2Slot       uint64 `db:"header2_slot"`
	Header2ParentRoot []byte `db:"header2_parentroot"`
	Header2StateRoot  []byte `db:"header2_stateroot"`
	Header2BodyRoot   []byte `db:"header2_bodyroot"`
	Header2Signature  []byte `db:"header2_signature"`
}

// BlockPageData is a struct block data used in the block page
type BlockPageData struct {
	Epoch                  uint64  `db:"epoch"`
	EpochFinalized         bool    `db:"epoch_finalized"`
	EpochParticipationRate float64 `db:"epoch_participation_rate"`
	Slot                   uint64  `db:"slot"`
	Ts                     time.Time
	NextSlot               uint64
	PreviousSlot           uint64
	Proposer               uint64  `db:"proposer"`
	Status                 uint64  `db:"status"`
	BlockRoot              []byte  `db:"blockroot"`
	ParentRoot             []byte  `db:"parentroot"`
	StateRoot              []byte  `db:"stateroot"`
	Signature              []byte  `db:"signature"`
	RandaoReveal           []byte  `db:"randaoreveal"`
	Graffiti               []byte  `db:"graffiti"`
	ProposerName           string  `db:"name"`
	Eth1dataDepositroot    []byte  `db:"eth1data_depositroot"`
	Eth1dataDepositcount   uint64  `db:"eth1data_depositcount"`
	Eth1dataBlockhash      []byte  `db:"eth1data_blockhash"`
	SyncAggregateBits      []byte  `db:"syncaggregate_bits"`
	SyncAggregateSignature []byte  `db:"syncaggregate_signature"`
	SyncAggParticipation   float64 `db:"syncaggregate_participation"`
	ProposerSlashingsCount uint64  `db:"proposerslashingscount"`
	AttesterSlashingsCount uint64  `db:"attesterslashingscount"`
	AttestationsCount      uint64  `db:"attestationscount"`
	DepositsCount          uint64  `db:"depositscount"`
	WithdrawalCount        uint64  `db:"withdrawalcount"`
	BLSChangeCount         uint64  `db:"bls_change_count"`
	VoluntaryExitscount    uint64  `db:"voluntaryexitscount"`
	SlashingsCount         uint64
	VotesCount             uint64
	VotingValidatorsCount  uint64
	Mainnet                bool

	ExecParentHash        []byte        `db:"exec_parent_hash"`
	ExecFeeRecipient      []byte        `db:"exec_fee_recipient"`
	ExecStateRoot         []byte        `db:"exec_state_root"`
	ExecReceiptsRoot      []byte        `db:"exec_receipts_root"`
	ExecLogsBloom         []byte        `db:"exec_logs_bloom"`
	ExecRandom            []byte        `db:"exec_random"`
	ExecBlockNumber       sql.NullInt64 `db:"exec_block_number"`
	ExecGasLimit          sql.NullInt64 `db:"exec_gas_limit"`
	ExecGasUsed           sql.NullInt64 `db:"exec_gas_used"`
	ExecTimestamp         sql.NullInt64 `db:"exec_timestamp"`
	ExecTime              time.Time
	ExecExtraData         []byte        `db:"exec_extra_data"`
	ExecBaseFeePerGas     sql.NullInt64 `db:"exec_base_fee_per_gas"`
	ExecBlockHash         []byte        `db:"exec_block_hash"`
	ExecTransactionsCount uint64        `db:"exec_transactions_count"`

	Transactions []*BlockPageTransaction

	Withdrawals []*Withdrawals

	ExecutionData *Eth1BlockPageData

	Attestations      []*BlockPageAttestation // Attestations included in this block
	VoluntaryExits    []*BlockPageVoluntaryExits
	Votes             []*BlockVote // Attestations that voted for that block
	AttesterSlashings []*BlockPageAttesterSlashing
	ProposerSlashings []*BlockPageProposerSlashing
	SyncCommittee     []uint64 // TODO: Setting it to contain the validator index

	Tags       TagMetadataSlice `db:"tags"`
	IsValidMev bool             `db:"is_valid_mev"`
}

func (u *BlockPageData) MarshalJSON() ([]byte, error) {
	type Alias BlockPageData
	return json.Marshal(&struct {
		BlockRoot string
		Ts        int64
		*Alias
	}{
		BlockRoot: fmt.Sprintf("%x", u.BlockRoot),
		Ts:        u.Ts.Unix(),
		Alias:     (*Alias)(u),
	})
}

type Eth1BlockPageData struct {
	Number                uint64
	PreviousBlock         uint64
	NextBlock             uint64
	TxCount               uint64
	WithdrawalCount       uint64
	UncleCount            uint64
	Hash                  string
	ParentHash            string
	MinerAddress          string
	MinerFormatted        template.HTML
	Reward                *big.Int
	MevReward             *big.Int
	MevBribe              *big.Int
	IsValidMev            bool
	MevRecipientFormatted template.HTML
	TxFees                *big.Int
	GasUsage              template.HTML
	GasLimit              uint64
	LowestGasPrice        *big.Int
	Ts                    time.Time
	Difficulty            *big.Int
	BaseFeePerGas         *big.Int
	BurnedFees            *big.Int
	Extra                 string
	Txs                   []Eth1BlockPageTransaction
	Uncles                []Eth1BlockPageData
	State                 string
}

type Eth1BlockPageTransaction struct {
	Hash          string
	HashFormatted template.HTML
	From          string
	FromFormatted template.HTML
	To            string
	ToFormatted   template.HTML
	Value         *big.Int
	Fee           *big.Int
	GasPrice      *big.Int
	Method        string
}

// IndexPageDataBlocks is a struct to hold detail data for the main web page
type BlocksPageDataBlocks struct {
	TotalCount           uint64        `db:"total_count"`
	Epoch                uint64        `json:"epoch"`
	Slot                 uint64        `json:"slot"`
	Ts                   time.Time     `json:"ts"`
	Proposer             uint64        `db:"proposer" json:"proposer"`
	ProposerFormatted    template.HTML `json:"proposer_formatted"`
	BlockRoot            []byte        `db:"blockroot" json:"block_root"`
	BlockRootFormatted   string        `json:"block_root_formatted"`
	ParentRoot           []byte        `db:"parentroot" json:"parent_root"`
	Attestations         uint64        `db:"attestationscount" json:"attestations"`
	Deposits             uint64        `db:"depositscount" json:"deposits"`
	Withdrawals          uint64        `db:"withdrawalcount" json:"withdrawals"`
	Exits                uint64        `db:"voluntaryexitscount" json:"exits"`
	Proposerslashings    uint64        `db:"proposerslashingscount" json:"proposerslashings"`
	Attesterslashings    uint64        `db:"attesterslashingscount" json:"attesterslashings"`
	SyncAggParticipation float64       `db:"syncaggregate_participation" json:"sync_aggregate_participation"`
	Status               uint64        `db:"status" json:"status"`
	StatusFormatted      template.HTML `json:"status_formatted"`
	Votes                uint64        `db:"votes" json:"votes"`
	Graffiti             []byte        `db:"graffiti"`
	ProposerName         string        `db:"name"`
}

type ClElInt64 struct {
	El    int64
	Cl    int64
	Total int64
}

type ClElFloat64 struct {
	El    float64
	Cl    float64
	Total float64
}

type AddValidatorWatchlistModal struct {
	CsrfField       template.HTML
	ValidatorIndex  uint64
	ValidatorPubkey string
	Events          []EventNameCheckbox
}

type EventNameCheckbox struct {
	EventLabel string
	EventName
	Active  bool
	Warning template.HTML
	Info    template.HTML
}

// ValidatorPageData is a struct to hold data for the validators page
type ValidatorPageData struct {
	Epoch                                    uint64 `db:"epoch"`
	ValidatorIndex                           uint64 `db:"validatorindex"`
	PublicKey                                []byte `db:"pubkey"`
	WithdrawableEpoch                        uint64 `db:"withdrawableepoch"`
	WithdrawCredentials                      []byte `db:"withdrawalcredentials"`
	CurrentBalance                           uint64 `db:"balance"`
	BalanceActivation                        uint64 `db:"balanceactivation"`
	EffectiveBalance                         uint64 `db:"effectivebalance"`
	Slashed                                  bool   `db:"slashed"`
	SlashedBy                                uint64
	SlashedAt                                uint64
	SlashedFor                               string
	ActivationEligibilityEpoch               uint64         `db:"activationeligibilityepoch"`
	ActivationEpoch                          uint64         `db:"activationepoch"`
	ExitEpoch                                uint64         `db:"exitepoch"`
	Index                                    uint64         `db:"index"`
	LastAttestationSlot                      *uint64        `db:"lastattestationslot"`
	Name                                     string         `db:"name"`
	Pool                                     string         `db:"pool"`
	Tags                                     pq.StringArray `db:"tags"`
	WithdrawableTs                           time.Time
	ActivationEligibilityTs                  time.Time
	ActivationTs                             time.Time
	ExitTs                                   time.Time
	Status                                   string `db:"status"`
	BlocksCount                              uint64
	ScheduledBlocksCount                     uint64
	MissedBlocksCount                        uint64
	OrphanedBlocksCount                      uint64
	ProposedBlocksCount                      uint64
	UnmissedBlocksPercentage                 float64 // missed/(executed+orphaned+scheduled)
	AttestationsCount                        uint64
	ExecutedAttestationsCount                uint64
	MissedAttestationsCount                  uint64
	OrphanedAttestationsCount                uint64
	UnmissedAttestationsPercentage           float64 // missed/(executed+orphaned)
	StatusProposedCount                      uint64
	StatusMissedCount                        uint64
	DepositsCount                            uint64
	WithdrawalCount                          uint64
	SlashingsCount                           uint64
	PendingCount                             uint64
	SyncCount                                uint64
	ScheduledSyncCount                       uint64
	ParticipatedSyncCount                    uint64
	MissedSyncCount                          uint64
	OrphanedSyncCount                        uint64
	UnmissedSyncPercentage                   float64       // missed/(participated+orphaned)
	IncomeToday                              ClElInt64     `json:"incomeToday"`
	Income1d                                 ClElInt64     `json:"income1d"`
	Income7d                                 ClElInt64     `json:"income7d"`
	Income31d                                ClElInt64     `json:"income31d"`
	IncomeTotal                              ClElInt64     `json:"incomeTotal"`
	IncomeTotalFormatted                     template.HTML `json:"incomeTotalFormatted"`
	Apr7d                                    ClElFloat64   `json:"apr7d"`
	Apr31d                                   ClElFloat64   `json:"apr31d"`
	Apr365d                                  ClElFloat64   `json:"apr365d"`
	ProposalLuck                             float64
	SyncLuck                                 float64
	ProposalEstimate                         *time.Time
	SyncEstimate                             *time.Time
	AvgSlotInterval                          *time.Duration
	AvgSyncInterval                          *time.Duration
	Rank7d                                   int64 `db:"rank7d"`
	RankCount                                int64 `db:"rank_count"`
	RankPercentage                           float64
	Proposals                                [][]uint64
	IncomeHistoryChartData                   []*ChartDataPoint
	ExecutionIncomeHistoryData               []*ChartDataPoint
	Deposits                                 *ValidatorDeposits
	Eth1DepositAddress                       []byte
	FlashMessage                             string
	Watchlist                                []*TaggedValidators
	SubscriptionFlash                        []interface{}
	User                                     *User
	AverageAttestationInclusionDistance      float64
	AttestationInclusionEffectiveness        float64
	CsrfField                                template.HTML
	NetworkStats                             *IndexPageData
	ChurnRate                                uint64
	QueuePosition                            uint64
	EstimatedActivationTs                    time.Time
	EstimatedActivationEpoch                 uint64
	InclusionDelay                           int64
	CurrentAttestationStreak                 uint64
	LongestAttestationStreak                 uint64
	IsRocketpool                             bool
	Rocketpool                               *RocketpoolValidatorPageData
	ShowMultipleWithdrawalCredentialsWarning bool
	CappellaHasHappened                      bool
	BLSChange                                *BLSChange
	IsWithdrawableAddress                    bool
	EstimatedNextWithdrawal                  template.HTML
	AddValidatorWatchlistModal               *AddValidatorWatchlistModal
	NextWithdrawalRow                        [][]interface{}
}

type RocketpoolValidatorPageData struct {
	NodeAddress          *[]byte    `db:"node_address"`
	MinipoolAddress      *[]byte    `db:"minipool_address"`
	MinipoolNodeFee      *float64   `db:"minipool_node_fee"`
	MinipoolDepositType  *string    `db:"minipool_deposit_type"`
	MinipoolStatus       *string    `db:"minipool_status"`
	MinipoolStatusTime   *time.Time `db:"minipool_status_time"`
	NodeTimezoneLocation *string    `db:"node_timezone_location"`
	NodeRPLStake         *string    `db:"node_rpl_stake"`
	NodeMinRPLStake      *string    `db:"node_min_rpl_stake"`
	NodeMaxRPLStake      *string    `db:"node_max_rpl_stake"`
	CumulativeRPL        *string    `db:"rpl_cumulative_rewards"`
	SmoothingClaimed     *string    `db:"claimed_smoothing_pool"`
	SmoothingUnclaimed   *string    `db:"unclaimed_smoothing_pool"`
	UnclaimedRPL         *string    `db:"unclaimed_rpl_rewards"`
	SmoothingPoolOptIn   bool       `db:"smoothing_pool_opted_in"`
	PenaltyCount         int        `db:"penalty_count"`
	RocketscanUrl        string     `db:"-"`
	NodeDepositBalance   *string    `db:"node_deposit_balance"`
	NodeRefundBalance    *string    `db:"node_refund_balance"`
	UserDepositBalance   *string    `db:"user_deposit_balance"`
	IsVacant             bool       `db:"is_vacant"`
	Version              *string    `db:"version"`
	NodeDepositCredit    *string    `db:"deposit_credit"`
	EffectiveRPLStake    *string    `db:"effective_rpl_stake"`
}

type ValidatorEarnings struct {
	Income1d                ClElInt64     `json:"income1d"`
	Income7d                ClElInt64     `json:"income7d"`
	Income31d               ClElInt64     `json:"income31d"`
	IncomeTotal             ClElInt64     `json:"incomeTotal"`
	Apr7d                   ClElFloat64   `json:"apr"`
	Apr31d                  ClElFloat64   `json:"apr31d"`
	Apr365d                 ClElFloat64   `json:"apr365d"`
	TotalDeposits           int64         `json:"totalDeposits"`
	TotalWithdrawals        uint64        `json:"totalWithdrawals"`
	EarningsInPeriodBalance int64         `json:"earningsInPeriodBalance"`
	EarningsInPeriod        int64         `json:"earningsInPeriod"`
	EpochStart              int64         `json:"epochStart"`
	EpochEnd                int64         `json:"epochEnd"`
	LastDayFormatted        template.HTML `json:"lastDayFormatted"`
	LastWeekFormatted       template.HTML `json:"lastWeekFormatted"`
	LastMonthFormatted      template.HTML `json:"lastMonthFormatted"`
	TotalFormatted          template.HTML `json:"totalFormatted"`
	TotalChangeFormatted    template.HTML `json:"totalChangeFormatted"`
	TotalBalance            template.HTML `json:"totalBalance"`
}

type ValidatorHistory struct {
	Epoch             uint64                `db:"epoch" json:"epoch,omitempty"`
	BalanceChange     sql.NullInt64         `db:"balancechange" json:"balance_change,omitempty"`
	AttesterSlot      sql.NullInt64         `db:"attestatation_attesterslot" json:"attester_slot,omitempty"`
	InclusionSlot     sql.NullInt64         `db:"attestation_inclusionslot" json:"inclusion_slot,omitempty"`
	AttestationStatus uint64                `db:"attestation_status" json:"attestation_status,omitempty"`
	ProposalStatus    sql.NullInt64         `db:"proposal_status" json:"proposal_status,omitempty"`
	ProposalSlot      sql.NullInt64         `db:"proposal_slot" json:"proposal_slot,omitempty"`
	IncomeDetails     *ValidatorEpochIncome `db:"-" json:"income_details,omitempty"`
	WithdrawalStatus  sql.NullInt64         `db:"withdrawal_status" json:"withdrawal_status,omitempty"`
	WithdrawalSlot    sql.NullInt64         `db:"withdrawal_slot" json:"withdrawal_slot,omitempty"`
}

// ValidatorAttestationSlashing is a struct to hold data of an attestation-slashing
type ValidatorAttestationSlashing struct {
	Epoch                  uint64        `db:"epoch" json:"epoch,omitempty"`
	Slot                   uint64        `db:"slot" json:"slot,omitempty"`
	Proposer               uint64        `db:"proposer" json:"proposer,omitempty"`
	Attestestation1Indices pq.Int64Array `db:"attestation1_indices" json:"attestation1_indices,omitempty"`
	Attestestation2Indices pq.Int64Array `db:"attestation2_indices" json:"attestation2_indices,omitempty"`
}

type ValidatorProposerSlashing struct {
	Epoch         uint64 `db:"epoch" json:"epoch,omitempty"`
	Slot          uint64 `db:"slot" json:"slot,omitempty"`
	Proposer      uint64 `db:"proposer" json:"proposer,omitempty"`
	ProposerIndex uint64 `db:"proposerindex" json:"proposer_index,omitempty"`
}

type MyCryptoSignature struct {
	Address string `json:"address"`
	Msg     string `json:"msg"`
	Sig     string `json:"sig"`
	Version string `json:"version"`
}

type ValidatorStatsTablePageData struct {
	ValidatorIndex uint64
	Rows           []*ValidatorStatsTableRow
	Currency       string
}

type ValidatorStatsTableRow struct {
	ValidatorIndex         uint64
	Day                    int64         `db:"day"`
	StartBalance           sql.NullInt64 `db:"start_balance"`
	EndBalance             sql.NullInt64 `db:"end_balance"`
	Income                 int64         `db:"cl_rewards_gwei"`
	IncomeExchangeRate     float64       `db:"-"`
	IncomeExchangeCurrency string        `db:"-"`
	IncomeExchanged        float64       `db:"-"`
	MinBalance             sql.NullInt64 `db:"min_balance"`
	MaxBalance             sql.NullInt64 `db:"max_balance"`
	StartEffectiveBalance  sql.NullInt64 `db:"start_effective_balance"`
	EndEffectiveBalance    sql.NullInt64 `db:"end_effective_balance"`
	MinEffectiveBalance    sql.NullInt64 `db:"min_effective_balance"`
	MaxEffectiveBalance    sql.NullInt64 `db:"max_effective_balance"`
	MissedAttestations     sql.NullInt64 `db:"missed_attestations"`
	OrphanedAttestations   sql.NullInt64 `db:"orphaned_attestations"`
	ProposedBlocks         sql.NullInt64 `db:"proposed_blocks"`
	MissedBlocks           sql.NullInt64 `db:"missed_blocks"`
	OrphanedBlocks         sql.NullInt64 `db:"orphaned_blocks"`
	AttesterSlashings      sql.NullInt64 `db:"attester_slashings"`
	ProposerSlashings      sql.NullInt64 `db:"proposer_slashings"`
	Deposits               sql.NullInt64 `db:"deposits"`
	DepositsAmount         sql.NullInt64 `db:"deposits_amount"`
	ParticipatedSync       sql.NullInt64 `db:"participated_sync"`
	MissedSync             sql.NullInt64 `db:"missed_sync"`
	OrphanedSync           sql.NullInt64 `db:"orphaned_sync"`
}

// ValidatorsPageData is a struct to hold data about the validators page
type ValidatorsPageData struct {
	TotalCount           uint64
	DepositedCount       uint64
	PendingCount         uint64
	ActiveCount          uint64
	ActiveOnlineCount    uint64
	ActiveOfflineCount   uint64
	SlashingCount        uint64
	SlashingOnlineCount  uint64
	SlashingOfflineCount uint64
	Slashed              uint64
	ExitingCount         uint64
	ExitingOnlineCount   uint64
	ExitingOfflineCount  uint64
	ExitedCount          uint64
	VoluntaryExitsCount  uint64
	UnknownCount         uint64
	Validators           []*ValidatorsPageDataValidators
	CappellaHasHappened  bool
}

// ValidatorsPageDataValidators is a struct to hold data about validators for the validators page
type ValidatorsPageDataValidators struct {
	TotalCount                 uint64 `db:"total_count"`
	Epoch                      uint64 `db:"epoch"`
	PublicKey                  []byte `db:"pubkey"`
	ValidatorIndex             uint64 `db:"validatorindex"`
	WithdrawableEpoch          uint64 `db:"withdrawableepoch"`
	CurrentBalance             uint64 `db:"balance"`
	EffectiveBalance           uint64 `db:"effectivebalance"`
	Slashed                    bool   `db:"slashed"`
	ActivationEligibilityEpoch uint64 `db:"activationeligibilityepoch"`
	ActivationEpoch            uint64 `db:"activationepoch"`
	ExitEpoch                  uint64 `db:"exitepoch"`
	LastAttestationSlot        *int64 `db:"lastattestationslot"`
	Name                       string `db:"name"`
	State                      string `db:"state"`
	MissedProposals            uint64 `db:"missedproposals"`
	ExecutedProposals          uint64 `db:"executedproposals"`
	MissedAttestations         uint64 `db:"missedattestations"`
	ExecutedAttestations       uint64 `db:"executedattestations"`
	Performance7d              int64  `db:"performance7d"`
}

type ValidatorSlashing struct {
	Epoch                  uint64        `db:"epoch" json:"epoch,omitempty"`
	Slot                   uint64        `db:"slot" json:"slot,omitempty"`
	Proposer               uint64        `db:"proposer" json:"proposer,omitempty"`
	SlashedValidator       *uint64       `db:"slashedvalidator" json:"slashed_validator,omitempty"`
	Attestestation1Indices pq.Int64Array `db:"attestation1_indices" json:"attestation1_indices,omitempty"`
	Attestestation2Indices pq.Int64Array `db:"attestation2_indices" json:"attestation2_indices,omitempty"`
	Type                   string        `db:"type" json:"type"`
}

type WithdrawalsPageData struct {
	Stats           *Stats
	WithdrawalChart *ChartsPageDataChart
	Withdrawals     *DataTableResponse
	BlsChanges      *DataTableResponse
}

type WithdrawalStats struct {
	WithdrawalsCount             uint64
	WithdrawalsTotal             uint64
	BLSChangeCount               uint64
	ValidatorsWithBLSCredentials uint64
}

type ChangeWithdrawalCredentialsPageData struct {
	FlashMessage string
	CsrfField    template.HTML
	RecaptchaKey string
}

type DepositsPageData struct {
	*Stats
	DepositContract string
	DepositChart    *ChartsPageDataChart
}

type EthOneDepositLeaderBoardPageData struct {
	DepositContract string
}

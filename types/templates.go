package types

import (
	"database/sql"
	"fmt"
	"html/template"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
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

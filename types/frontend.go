package types

import (
	"html/template"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	ValidatorBalanceDecreasedEventName               EventName = "validator_balance_decreased"
	ValidatorMissedProposalEventName                 EventName = "validator_proposal_missed"
	ValidatorExecutedProposalEventName               EventName = "validator_proposal_submitted"
	ValidatorMissedAttestationEventName              EventName = "validator_attestation_missed"
	ValidatorGotSlashedEventName                     EventName = "validator_got_slashed"
	ValidatorDidSlashEventName                       EventName = "validator_did_slash"
	ValidatorIsOfflineEventName                      EventName = "validator_is_offline"
	ValidatorReceivedWithdrawalEventName             EventName = "validator_withdrawal"
	ValidatorReceivedDepositEventName                EventName = "validator_received_deposit"
	NetworkSlashingEventName                         EventName = "network_slashing"
	NetworkValidatorActivationQueueFullEventName     EventName = "network_validator_activation_queue_full"
	NetworkValidatorActivationQueueNotFullEventName  EventName = "network_validator_activation_queue_not_full"
	NetworkValidatorExitQueueFullEventName           EventName = "network_validator_exit_queue_full"
	NetworkValidatorExitQueueNotFullEventName        EventName = "network_validator_exit_queue_not_full"
	NetworkLivenessIncreasedEventName                EventName = "network_liveness_increased"
	EthClientUpdateEventName                         EventName = "eth_client_update"
	MonitoringMachineOfflineEventName                EventName = "monitoring_machine_offline"
	MonitoringMachineDiskAlmostFullEventName         EventName = "monitoring_hdd_almostfull"
	MonitoringMachineCpuLoadEventName                EventName = "monitoring_cpu_load"
	MonitoringMachineMemoryUsageEventName            EventName = "monitoring_memory_usage"
	MonitoringMachineSwitchedToETH2FallbackEventName EventName = "monitoring_fallback_eth2inuse"
	MonitoringMachineSwitchedToETH1FallbackEventName EventName = "monitoring_fallback_eth1inuse"
	TaxReportEventName                               EventName = "user_tax_report"
	RocketpoolCommissionThresholdEventName           EventName = "rocketpool_commision_threshold"
	RocketpoolNewClaimRoundStartedEventName          EventName = "rocketpool_new_claimround"
	RocketpoolColleteralMinReached                   EventName = "rocketpool_colleteral_min"
	RocketpoolColleteralMaxReached                   EventName = "rocketpool_colleteral_max"
	SyncCommitteeSoon                                EventName = "validator_synccommittee_soon"
)

type MachineMetricSystemUser struct {
	UserID                    uint64
	Machine                   string
	CurrentData               *MachineMetricSystem
	CurrentDataInsertTs       int64
	FiveMinuteOldData         *MachineMetricSystem
	FiveMinuteOldDataInsertTs int64
}

type Eth1AddressSearchItem struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	Token   string `json:"token"`
}

type NotificationChannel string

type EventName string

type Tag string

const (
	ValidatorTagsWatchlist Tag = "watchlist"
)

var NotificationChannelLabels map[NotificationChannel]template.HTML = map[NotificationChannel]template.HTML{
	EmailNotificationChannel:          "Email Notification",
	PushNotificationChannel:           "Push Notification",
	WebhookNotificationChannel:        `Webhook Notification (<a href="/user/webhooks">configure</a>)`,
	WebhookDiscordNotificationChannel: "Discord Notification",
}

const (
	EmailNotificationChannel          NotificationChannel = "email"
	PushNotificationChannel           NotificationChannel = "push"
	WebhookNotificationChannel        NotificationChannel = "webhook"
	WebhookDiscordNotificationChannel NotificationChannel = "webhook_discord"
)

var NotificationChannels = []NotificationChannel{
	EmailNotificationChannel,
	PushNotificationChannel,
	WebhookNotificationChannel,
	WebhookDiscordNotificationChannel,
}

type RawMempoolResponse struct {
	Pending map[string]map[int]*RawMempoolTransaction `json:"pending"`
	Queued  map[string]map[int]*RawMempoolTransaction `json:"queued"`
	BaseFee map[string]map[int]*RawMempoolTransaction `json:"baseFee"`

	TxsByHash map[common.Hash]*RawMempoolTransaction
}

func (mempool RawMempoolResponse) FindTxByHash(txHashString string) *RawMempoolTransaction {
	return mempool.TxsByHash[common.HexToHash(txHashString)]
}

type RawMempoolTransaction struct {
	Hash             common.Hash     `json:"hash"`
	From             *common.Address `json:"from"`
	To               *common.Address `json:"to"`
	Value            *hexutil.Big    `json:"value"`
	Gas              *hexutil.Big    `json:"gas"`
	GasFeeCap        *hexutil.Big    `json:"maxFeePerGas,omitempty"`
	GasTipCap        *hexutil.Big    `json:"maxPriorityFeePerGas,omitempty"`
	GasPrice         *hexutil.Big    `json:"gasPrice"`
	Nonce            *hexutil.Big    `json:"nonce"`
	Input            *string         `json:"input"`
	TransactionIndex *hexutil.Big    `json:"transactionIndex"`
}

type GasNowPageData struct {
	Code int `json:"code"`
	Data struct {
		Rapid     *big.Int `json:"rapid"`
		Fast      *big.Int `json:"fast"`
		Standard  *big.Int `json:"standard"`
		Slow      *big.Int `json:"slow"`
		Timestamp int64    `json:"timestamp"`
		Price     float64  `json:"price,omitempty"`
		PriceUSD  float64  `json:"priceUSD"`
		Currency  string   `json:"currency,omitempty"`
	} `json:"data"`
}

type TaggedValidators struct {
	UserID             uint64 `db:"user_id"`
	Tag                string `db:"tag"`
	ValidatorPublickey []byte `db:"validator_publickey"`
	Validator          *Validator
	Events             []EventName `db:"events"`
}

type EventNameDesc struct {
	Desc    string
	Event   EventName
	Info    template.HTML
	Warning template.HTML
}

// this is the source of truth for the validator events that are supported by the user/notification page
var AddWatchlistEvents = []EventNameDesc{
	{
		Desc:  "Validator is Offline",
		Event: ValidatorIsOfflineEventName,
		Info:  template.HTML(`<i data-toggle="tooltip" data-html="true" title="<div class='text-left'>Will trigger a notifcation:<br><ul><li>Once you have been offline for 3 epochs</li><li>Every 32 Epochs (~3 hours) during your downtime</li><li>Once you are back online again</li></ul></div>" class="fas fa-question-circle"></i>`),
	},
	{
		Desc:  "Proposals missed",
		Event: ValidatorMissedProposalEventName,
	},
	{
		Desc:  "Proposals submitted",
		Event: ValidatorExecutedProposalEventName,
	},
	{
		Desc:  "Validator got slashed",
		Event: ValidatorGotSlashedEventName,
	},
	{
		Desc:  "Sync committee",
		Event: SyncCommitteeSoon,
	},
	{
		Desc:    "Attestations missed",
		Event:   ValidatorMissedAttestationEventName,
		Warning: template.HTML(`<i data-toggle="tooltip" title="Will trigger every epoch (6.4 minutes) during downtime" class="fas fa-exclamation-circle text-warning"></i>`),
	},
	{
		Desc:  "Withdrawal processed",
		Event: ValidatorReceivedWithdrawalEventName,
		Info:  template.HTML(`<i data-toggle="tooltip" data-html="true" title="<div class='text-left'>Will trigger a notifcation when:<br><ul><li>A partial withdrawal is processed</li><li>Your validator exits and its full balance is withdrawn</li></ul> <div>Requires that your validator has 0x01 credentials</div></div>" class="fas fa-question-circle"></i>`),
	},
}

package types

import (
	"html/template"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

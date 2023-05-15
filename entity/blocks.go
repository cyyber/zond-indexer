package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

type BlockData struct {
	Hash         []byte
	ParentHash   []byte
	UncleHash    []byte
	Coinbase     []byte
	Root         []byte
	TxHash       []byte
	ReceiptHash  []byte
	Difficulty   []byte
	Number       uint64
	GasLimit     uint64
	GasUsed      uint64
	Time         primitive.DateTime
	Extra        []byte
	MixDigest    []byte
	Bloom        []byte
	BaseFee      []byte
	Uncles       []*BlockData
	Transactions []*Eth1Transaction
	Withdrawals  []*Eth1Withdrawal
}

type Eth1Transaction struct {
	Type                 uint32
	Nonce                uint64
	GasPrice             []byte
	MaxPriorityFeePerGas []byte
	MaxFeePerGas         []byte
	Gas                  uint64
	Value                []byte
	Data                 []byte
	To                   []byte
	From                 []byte
	ChainId              []byte
	AccessList           []*AccessList
	Hash                 []byte
	// Receipt fields
	ContractAddress    []byte
	CommulativeGasUsed uint64
	GasUsed            uint64
	LogsBloom          []byte
	Status             uint64
	ErrorMsg           string
	Logs               []*Eth1Log
	// Internal transactions
	Itx []*Eth1InternalTransaction
}

type AccessList struct {
	Address     []byte
	StorageKeys [][]byte
}

type Eth1Log struct {
	Address []byte
	Data    []byte
	Removed bool
	Topics  [][]byte
}

type Eth1InternalTransaction struct {
	Type     string
	From     []byte
	To       []byte
	Value    []byte
	ErrorMsg string
	Path     string
}

type Eth1Withdrawal struct {
	Index          uint64
	ValidatorIndex uint64
	Address        []byte
	Amount         []byte
}

package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

type BlockIndex struct {
	ChainId                  string
	Type                     string
	Hash                     []byte
	ParentHash               []byte
	UncleHash                []byte
	Coinbase                 []byte
	Difficulty               []byte
	Number                   uint64
	GasLimit                 uint64
	GasUsed                  uint64
	Time                     primitive.Timestamp
	BaseFee                  []byte
	UncleCount               uint64
	TransactionCount         uint64
	Mev                      []byte
	LowestGasPrice           []byte
	HighestGasPrice          []byte
	TxReward                 []byte
	UncleReward              []byte
	InternalTransactionCount uint64
}

type TransactionIndex struct {
	ChainId            string
	Hash               []byte
	Type               string
	TransactionType    string
	BlockNumber        uint64
	Time               primitive.Timestamp
	MethodId           []byte
	From               []byte
	To                 []byte
	Value              []byte
	TxFee              []byte
	GasPrice           []byte
	IsContractCreation bool
	InvokesContract    bool
	ErrorMsg           string
}

type InternalTransactionIndex struct {
	ChainId         string
	ParentHash      []byte
	Type            string
	BlockNumber     uint64
	TransactionType string
	Time            primitive.Timestamp
	From            []byte
	To              []byte
	Value           []byte
}

type ERC20Index struct {
	ChainId      string
	ParentHash   []byte
	Type         string
	BlockNumber  uint64
	TokenAddress []byte
	Time         primitive.Timestamp
	From         []byte
	To           []byte
	Value        []byte
}

type ERC721Index struct {
	ChainId      string
	ParentHash   []byte
	Type         string
	BlockNumber  uint64
	TokenAddress []byte
	Time         primitive.Timestamp
	From         []byte
	To           []byte
	TokenId      []byte
}

type ERC1155Index struct {
	ChainId      string
	ParentHash   []byte
	Type         string
	BlockNumber  uint64
	TokenAddress []byte
	Time         primitive.Timestamp
	From         []byte
	To           []byte
	TokenId      []byte
	Value        []byte
	Operator     []byte
}

type UncleBlocksIndex struct {
	ChainId     string
	BlockNumber uint64
	Type        string
	Number      uint64
	GasLimit    uint64
	GasUsed     uint64
	BaseFee     []byte
	Difficulty  []byte
	Time        primitive.Timestamp
	Reward      []byte
}

type WithdrawalIndex struct {
	ChainId        string
	BlockNumber    uint64
	Type           string
	Index          uint64
	ValidatorIndex uint64
	Address        []byte
	Amount         []byte
	Time           primitive.Timestamp
}

type Indexes struct {
	Type  string
	Key   string
	Value string
}

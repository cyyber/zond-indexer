package entity

import (
	"math/big"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccountMetadataFamily struct {
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ChainId string
	Type    string
	Name    string
	Balance big.Int
	Token   string
	Address string
}

type ContractMetadataFamily struct {
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ChainId string
	Name    string
	Abi     []byte
	Address string
}

type ERC20MetadataFamily struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ChainId      string
	Type         string
	Address      string
	Name         string
	Logo         []byte
	Logoformat   string
	Price        []byte
	Description  string
	Decimals     []byte
	TotalSupply  []byte
	Symbol       string
	OfficialSite string
}

type ERC721MetadataFamily struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string
	Logo        []byte
	Logoformat  string
	Price       []byte
	Description string
	Decimals    uint64
	TotalSupply uint64
	Symbol      string
}

type ERC1155MetadataFamily struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string
	Logo        []byte
	Logoformat  string
	Price       []byte
	Description string
	Decimals    uint64
	TotalSupply uint64
	Symbol      string
}

type Series struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ChainID  string
	Time     primitive.Timestamp
	Type     string
	Slow     []byte
	Standard []byte
	Fast     []byte
	Rapid    []byte
}

type BalanceUpdates struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ChainId  string
	Type     string
	Key      string
	Token    []byte
	Address  []byte
	Balance  []byte
	Metadata *ERC20MetadataFamily
}

type BlockMetadataUpdates struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	BlockNumber uint64
	BlockHash   string
	ChainId     string
	Type        string
	Keys        string
}

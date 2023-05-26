package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

type AccountMetadataFamily struct {
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Type    string
	Name    string
	Balance uint64
	Token   string
	Address string
}

type ContractMetadataFamily struct {
	ID   primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name string
	Abi  []byte
}

type ERC20MetadataFamily struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
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
	ChainId     uint64
	Type        string
	Keys        string
}

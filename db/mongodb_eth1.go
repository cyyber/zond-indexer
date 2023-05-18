package db

import (
	"context"
	"time"

	"github.com/Prajjawalk/zond-indexer/entity"
	"github.com/Prajjawalk/zond-indexer/types"
	"github.com/Prajjawalk/zond-indexer/utils"
	"go.mongodb.org/mongo-driver/bson"
)

type IndexFilter string

const (
	FILTER_TIME           IndexFilter = "TIME"
	FILTER_TO             IndexFilter = "TO"
	FILTER_FROM           IndexFilter = "FROM"
	FILTER_TOKEN_RECEIVED IndexFilter = "TOKEN_RECEIVED"
	FILTER_TOKEN_SENT     IndexFilter = "TOKEN_SENT"
	FILTER_METHOD         IndexFilter = "METHOD"
	FILTER_CONTRACT       IndexFilter = "CONTRACT"
	FILTER_ERROR          IndexFilter = "ERROR"
)

const (
	DATA                           = "data"
	METADATA_UPDATES               = "metadata_updates"
	METADATA                       = "metadata"
	BLOCKS                         = "blocks"
	DATA_COLUMN                    = "d"
	INDEX_COLUMN                   = "i"
	DEFAULT_FAMILY_BLOCKS          = "default"
	METADATA_UPDATES_FAMILY_BLOCKS = "blocks"
	ACCOUNT_METADATA_FAMILY        = "a"
	CONTRACT_METADATA_FAMILY       = "c"
	ERC20_METADATA_FAMILY          = "erc20"
	ERC721_METADATA_FAMILY         = "erc721"
	ERC1155_METADATA_FAMILY        = "erc1155"
	writeRowLimit                  = 10000
	MAX_INT                        = 9223372036854775807
	MIN_INT                        = -9223372036854775808
)

const (
	ACCOUNT_COLUMN_NAME = "NAME"
	ACCOUNT_IS_CONTRACT = "ISCONTRACT"

	CONTRACT_NAME = "CONTRACTNAME"
	CONTRACT_ABI  = "ABI"

	ERC20_COLUMN_DECIMALS    = "DECIMALS"
	ERC20_COLUMN_TOTALSUPPLY = "TOTALSUPPLY"
	ERC20_COLUMN_SYMBOL      = "SYMBOL"

	ERC20_COLUMN_PRICE = "PRICE"

	ERC20_COLUMN_NAME           = "NAME"
	ERC20_COLUMN_DESCRIPTION    = "DESCRIPTION"
	ERC20_COLUMN_LOGO           = "LOGO"
	ERC20_COLUMN_LOGO_FORMAT    = "LOGOFORMAT"
	ERC20_COLUMN_LINK           = "LINK"
	ERC20_COLUMN_OGIMAGE        = "OGIMAGE"
	ERC20_COLUMN_OGIMAGE_FORMAT = "OGIMAGEFORMAT"
)

var ZERO_ADDRESS []byte = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

var (
	ERC20TOPIC   []byte
	ERC721TOPIC  []byte
	ERC1155Topic []byte
)

func (mongodb *Mongo) GetDataTable() interface{} {
	return mongodb.Db.Collection(DATA)
}

func (mongodb *Mongo) GetMetadataUpdatesTable() interface{} {
	return mongodb.Db.Collection(METADATA_UPDATES)
}

func (mongodb *Mongo) GetMetadatTable() interface{} {
	return mongodb.Db.Collection(METADATA)
}

func (mongodb *Mongo) SaveBlock(block *types.Eth1Block) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	blockInput := &entity.BlockData{}
	blockInput.Eth1Block = *block
	blockInput.ChainId = mongodb.ChainId
	doc, err := utils.ToDoc(block)
	if err != nil {
		return err
	}

	_, err = mongodb.Db.Collection(BLOCKS).InsertOne(ctx, doc)
	if err != nil {
		return err
	}
	return nil
}

func (mongodb *Mongo) GetBlockFromBlocksTable(number uint64) (*types.Eth1Block, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "number", Value: number}}
	cursor, err := mongodb.Db.Collection(BLOCKS).Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var results []*entity.BlockData
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return &results[0].Eth1Block, nil
}

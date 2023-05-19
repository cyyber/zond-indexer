package db

import (
	"context"
	"fmt"
	"time"

	"github.com/Prajjawalk/zond-indexer/entity"
	"github.com/Prajjawalk/zond-indexer/types"
	"github.com/Prajjawalk/zond-indexer/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (mongodb *Mongo) CheckForGapsInBlocksTable(lookback int) (gapFound bool, start int, end int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}}
	cursor, err := mongodb.Db.Collection(BLOCKS).Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "createdat", Value: -1}}).SetLimit(int64(lookback)))
	if err != nil {
		return false, 0, 0, err
	}
	var results []*entity.BlockData

	if err = cursor.All(ctx, &results); err != nil {
		logger.Errorf("error while parsing blocks: %v", err)
	}

	previous := uint64(0)
	for _, result := range results {
		c := result.Eth1Block.Number
		if previous != 0 && previous != c+1 {
			gapFound = true
			start = int(c)
			end = int(previous)
			logger.Fatalf("found gap between block %v and block %v in blocks table", previous, c)
			break
		}
		previous = c
	}

	return gapFound, start, end, err
}

func (mongodb *Mongo) GetLastBlockInBlocksTable() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}}
	cursor, err := mongodb.Db.Collection(BLOCKS).Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "createdat", Value: -1}}).SetLimit(int64(1)))
	if err != nil {
		return 0, err
	}

	var results []*entity.BlockData

	if err = cursor.All(ctx, &results); err != nil {
		logger.Errorf("error while parsing blocks: %v", err)
	}

	return int(results[0].Eth1Block.Number), nil
}

func (mongodb *Mongo) CheckForGapsInDataTable(lookback int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "type", Value: "blockindex"}}
	var results []*entity.BlockIndex
	cursor, err := mongodb.Db.Collection(DATA).Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "createdat", Value: -1}}).SetLimit(int64(lookback)))
	if err != nil {
		return err
	}
	if err = cursor.All(ctx, &results); err != nil {
		logger.Errorf("error while parsing block data: %v", err)
	}

	previous := uint64(0)
	for _, result := range results {
		c := result.Number
		if previous != 0 && previous != c+1 {
			logger.Fatalf("found gap between block %v and block %v in blocks table", previous, c)
			break
		}
		previous = c
	}
	return nil
}

func (mongodb *Mongo) GetLastBlockInDataTable() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "type", Value: "blockindex"}}
	var results []*entity.BlockIndex
	cursor, err := mongodb.Db.Collection(DATA).Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "createdat", Value: -1}}).SetLimit(int64(1)))
	if err != nil {
		return 0, err
	}
	if err = cursor.All(ctx, &results); err != nil {
		logger.Errorf("error while parsing block data: %v", err)
	}

	return int(results[0].Number), nil
}

func (mongodb *Mongo) GetMostRecentBlockFromDataTable() (*types.Eth1BlockIndexed, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "type", Value: "blockindex"}}
	var results []*entity.BlockIndex
	cursor, err := mongodb.Db.Collection(DATA).Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "createdat", Value: -1}}).SetLimit(int64(1)))
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &results); err != nil {
		logger.Errorf("error while parsing block data: %v", err)
	}

	return &types.Eth1BlockIndexed{
		Hash:                     results[0].Hash,
		ParentHash:               results[0].ParentHash,
		UncleHash:                results[0].UncleHash,
		Coinbase:                 results[0].Coinbase,
		Difficulty:               results[0].Difficulty,
		Number:                   results[0].Number,
		GasLimit:                 results[0].GasLimit,
		GasUsed:                  results[0].GasUsed,
		Time:                     timestamppb.New(time.Unix(int64(results[0].Time.T), 0)),
		BaseFee:                  results[0].BaseFee,
		UncleCount:               results[0].UncleCount,
		TransactionCount:         results[0].TransactionCount,
		Mev:                      results[0].Mev,
		LowestGasPrice:           results[0].LowestGasPrice,
		HighestGasPrice:          results[0].HighestGasPrice,
		TxReward:                 results[0].TxReward,
		UncleReward:              results[0].UncleReward,
		InternalTransactionCount: results[0].InternalTransactionCount,
	}, nil
}

func (mongodb *Mongo) GetFullBlockDescending(start, limit uint64) ([]*types.Eth1Block, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*60))
	defer cancel()

	if start < 1 || limit < 1 || limit > start {
		return nil, fmt.Errorf("invalid block range provided (start: %v, limit: %v)", start, limit)
	}

	blocks := make([]*types.Eth1Block, 0, limit)
	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "number", Value: bson.D{{Key: "lte", Value: start}}}}
	cursor, err := mongodb.Db.Collection(BLOCKS).Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "number", Value: -1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &blocks); err != nil {
		logger.Errorf("error while parsing block data: %v", err)
	}

	return blocks, nil
}

// GetFullBlockDescending gets blocks starting at block start
func (mongodb *Mongo) GetFullBlocksDescending(stream chan<- *types.Eth1Block, high, low uint64) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*180))
	defer cancel()

	if high < 1 || low < 1 || high < low {
		return fmt.Errorf("invalid block range provided (start: %v, limit: %v)", high, low)
	}
	limit := high - low
	blocks := make([]*types.Eth1Block, 0, limit)
	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "number", Value: bson.D{{Key: "lte", Value: high}}}}
	cursor, err := mongodb.Db.Collection(BLOCKS).Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "number", Value: -1}}).SetLimit(int64(limit)))
	if err != nil {
		return err
	}
	if err = cursor.All(ctx, &blocks); err != nil {
		logger.Errorf("error while parsing block data: %v", err)
	}

	for _, block := range blocks {
		stream <- block
	}
	return nil
}

func (mongodb *Mongo) GetBlocksIndexedMultiple(blockNumbers []uint64, limit uint64) ([]*types.Eth1BlockIndexed, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	blocks := make([]*types.Eth1BlockIndexed, 0, 100)

	var results []*entity.BlockIndex
	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "type", Value: "blockindex"}, {Key: "number", Value: bson.D{{Key: "number", Value: bson.D{{Key: "$in", Value: blockNumbers}}}}}}
	cursor, err := mongodb.Db.Collection(DATA).Find(ctx, filter, options.Find().SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &results); err != nil {
		logger.Errorf("error while parsing block data: %v", err)
	}

	for i, result := range results {
		blocks[i] = &types.Eth1BlockIndexed{
			Hash:                     result.Hash,
			ParentHash:               result.ParentHash,
			UncleHash:                result.UncleHash,
			Coinbase:                 result.Coinbase,
			Difficulty:               result.Difficulty,
			Number:                   result.Number,
			GasLimit:                 result.GasLimit,
			GasUsed:                  result.GasUsed,
			Time:                     timestamppb.New(time.Unix(int64(result.Time.T), 0)),
			BaseFee:                  result.BaseFee,
			UncleCount:               result.UncleCount,
			TransactionCount:         result.TransactionCount,
			Mev:                      result.Mev,
			LowestGasPrice:           result.LowestGasPrice,
			HighestGasPrice:          result.HighestGasPrice,
			TxReward:                 result.TxReward,
			UncleReward:              result.UncleReward,
			InternalTransactionCount: result.InternalTransactionCount,
		}
	}
	return blocks, nil
}

func (mongodb *Mongo) GetBlocksDescending(start, limit uint64) ([]*types.Eth1BlockIndexed, error) {
	if start < 1 || limit < 1 || limit > start {
		return nil, fmt.Errorf("invalid block range provided (start: %v, limit: %v)", start, limit)
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	var results []*entity.BlockIndex
	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "type", Value: "blockindex"}, {Key: "number", Value: bson.D{{Key: "lte", Value: start}}}}
	cursor, err := mongodb.Db.Collection(DATA).Find(ctx, filter, options.Find().SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &results); err != nil {
		logger.Errorf("error while parsing block data: %v", err)
	}

	blocks := make([]*types.Eth1BlockIndexed, 0, 100)
	for i, result := range results {
		blocks[i] = &types.Eth1BlockIndexed{
			Hash:                     result.Hash,
			ParentHash:               result.ParentHash,
			UncleHash:                result.UncleHash,
			Coinbase:                 result.Coinbase,
			Difficulty:               result.Difficulty,
			Number:                   result.Number,
			GasLimit:                 result.GasLimit,
			GasUsed:                  result.GasUsed,
			Time:                     timestamppb.New(time.Unix(int64(result.Time.T), 0)),
			BaseFee:                  result.BaseFee,
			UncleCount:               result.UncleCount,
			TransactionCount:         result.TransactionCount,
			Mev:                      result.Mev,
			LowestGasPrice:           result.LowestGasPrice,
			HighestGasPrice:          result.HighestGasPrice,
			TxReward:                 result.TxReward,
			UncleReward:              result.UncleReward,
			InternalTransactionCount: result.InternalTransactionCount,
		}
	}
	return blocks, nil
}

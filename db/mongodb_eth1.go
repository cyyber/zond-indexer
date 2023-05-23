package db

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/Prajjawalk/zond-indexer/entity"
	"github.com/Prajjawalk/zond-indexer/erc1155"
	"github.com/Prajjawalk/zond-indexer/erc20"
	"github.com/Prajjawalk/zond-indexer/erc721"
	"github.com/Prajjawalk/zond-indexer/types"
	"github.com/Prajjawalk/zond-indexer/utils"
	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	eth_types "github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

func (mongodb *Mongo) TransformBlock(block *types.Eth1Block, cache *freecache.Cache) (interface{}, interface{}, error) {
	var bulkData []mongo.WriteModel
	var bulkMetadataUpdates []mongo.WriteModel

	idx := entity.BlockIndex{
		ChainId:    mongodb.ChainId,
		Type:       "blockindex",
		Hash:       block.GetHash(),
		ParentHash: block.GetParentHash(),
		UncleHash:  block.GetUncleHash(),
		Coinbase:   block.GetCoinbase(),
		Difficulty: block.GetDifficulty(),
		Number:     block.GetNumber(),
		GasLimit:   block.GetGasLimit(),
		GasUsed:    block.GetGasUsed(),
		Time:       primitive.Timestamp{T: uint32(block.GetTime().AsTime().Unix()), I: 0},
		BaseFee:    block.GetBaseFee(),
		// Duration:               uint64(block.GetTime().AsTime().Unix() - previous.GetTime().AsTime().Unix()),
		UncleCount:       uint64(len(block.GetUncles())),
		TransactionCount: uint64(len(block.GetTransactions())),
		// BaseFeeChange:          new(big.Int).Sub(new(big.Int).SetBytes(block.GetBaseFee()), new(big.Int).SetBytes(previous.GetBaseFee())).Bytes(),
		// BlockUtilizationChange: new(big.Int).Sub(new(big.Int).Div(big.NewInt(int64(block.GetGasUsed())), big.NewInt(int64(block.GetGasLimit()))), new(big.Int).Div(big.NewInt(int64(previous.GetGasUsed())), big.NewInt(int64(previous.GetGasLimit())))).Bytes(),
	}

	uncleReward := big.NewInt(0)
	r := new(big.Int)

	for _, uncle := range block.Uncles {

		if len(block.Difficulty) == 0 { // no uncle rewards in PoS
			continue
		}

		r.Add(big.NewInt(int64(uncle.GetNumber())), big.NewInt(8))
		r.Sub(r, big.NewInt(int64(block.GetNumber())))
		r.Mul(r, utils.Eth1BlockReward(block.GetNumber(), block.Difficulty))
		r.Div(r, big.NewInt(8))

		r.Div(utils.Eth1BlockReward(block.GetNumber(), block.Difficulty), big.NewInt(32))
		uncleReward.Add(uncleReward, r)
	}

	idx.UncleReward = uncleReward.Bytes()

	var maxGasPrice *big.Int
	var minGasPrice *big.Int
	txReward := big.NewInt(0)

	for _, t := range block.GetTransactions() {
		price := new(big.Int).SetBytes(t.GasPrice)

		if minGasPrice == nil {
			minGasPrice = price
		}
		if maxGasPrice == nil {
			maxGasPrice = price
		}

		if price.Cmp(maxGasPrice) > 0 {
			maxGasPrice = price
		}

		if price.Cmp(minGasPrice) < 0 {
			minGasPrice = price
		}

		txFee := new(big.Int).Mul(new(big.Int).SetBytes(t.GasPrice), big.NewInt(int64(t.GasUsed)))

		if len(block.BaseFee) > 0 {
			effectiveGasPrice := math.BigMin(new(big.Int).Add(new(big.Int).SetBytes(t.MaxPriorityFeePerGas), new(big.Int).SetBytes(block.BaseFee)), new(big.Int).SetBytes(t.MaxFeePerGas))
			proposerGasPricePart := new(big.Int).Sub(effectiveGasPrice, new(big.Int).SetBytes(block.BaseFee))

			if proposerGasPricePart.Cmp(big.NewInt(0)) >= 0 {
				txFee = new(big.Int).Mul(proposerGasPricePart, big.NewInt(int64(t.GasUsed)))
			} else {
				logger.Errorf("error minerGasPricePart is below 0 for tx %v: %v", t.Hash, proposerGasPricePart)
				txFee = big.NewInt(0)
			}

		}

		txReward.Add(txReward, txFee)

		for _, itx := range t.Itx {
			if itx.Path == "[]" || bytes.Equal(itx.Value, []byte{0x0}) { // skip top level call & empty calls
				continue
			}
			idx.InternalTransactionCount++
		}
	}

	idx.TxReward = txReward.Bytes()

	// logger.Infof("tx reward for block %v is %v", block.Number, txReward.String())

	if maxGasPrice != nil {
		idx.LowestGasPrice = minGasPrice.Bytes()

	}
	if minGasPrice != nil {
		idx.HighestGasPrice = maxGasPrice.Bytes()
	}

	idx.Mev = CalculateMevFromBlock(block).Bytes()

	// Mark Coinbase for balance update
	mongodb.markBalanceUpdate(idx.Coinbase, []byte{0x0}, &bulkMetadataUpdates, cache)

	doc, err := utils.ToDoc(idx)
	if err != nil {
		return nil, nil, err
	}
	insertBlock := mongo.NewInsertOneModel().SetDocument(doc)
	bulkData = append(bulkData, insertBlock)

	indexes := []string{
		// Index blocks by the miners address
		fmt.Sprintf("%s:I:B:%x:TIME:%s", mongodb.ChainId, block.GetCoinbase(), block.Time),
	}

	blockIdentifier := fmt.Sprintf("%s:B:%09d", mongodb.ChainId, block.GetNumber())

	for _, idx := range indexes {
		mut := &entity.Indexes{
			Type:  "index",
			Key:   idx,
			Value: blockIdentifier,
		}
		doc, err := utils.ToDoc(mut)
		if err != nil {
			return nil, nil, err
		}
		insertBlock := mongo.NewInsertOneModel().SetDocument(doc)
		bulkData = append(bulkData, insertBlock)
	}

	return bulkData, bulkMetadataUpdates, nil
}

func CalculateMevFromBlock(block *types.Eth1Block) *big.Int {
	mevReward := big.NewInt(0)

	for _, tx := range block.GetTransactions() {
		for _, itx := range tx.GetItx() {
			//log.Printf("%v - %v", common.HexToAddress(itx.To), common.HexToAddress(block.Miner))
			if common.BytesToAddress(itx.To) == common.BytesToAddress(block.GetCoinbase()) {
				mevReward = new(big.Int).Add(mevReward, new(big.Int).SetBytes(itx.GetValue()))
			}
		}

	}
	return mevReward
}

func CalculateTxFeesFromBlock(block *types.Eth1Block) *big.Int {
	txFees := new(big.Int)
	for _, tx := range block.Transactions {
		txFees.Add(txFees, CalculateTxFeeFromTransaction(tx, new(big.Int).SetBytes(block.BaseFee)))
	}
	return txFees
}

func CalculateTxFeeFromTransaction(tx *types.Eth1Transaction, blockBaseFee *big.Int) *big.Int {
	// calculate tx fee depending on tx type
	txFee := new(big.Int).SetUint64(tx.GasUsed)
	if tx.Type == uint32(2) {
		// multiply gasused with min(baseFee + maxpriorityfee, maxfee)
		if normalGasPrice, maxGasPrice := new(big.Int).Add(blockBaseFee, new(big.Int).SetBytes(tx.MaxPriorityFeePerGas)), new(big.Int).SetBytes(tx.MaxFeePerGas); normalGasPrice.Cmp(maxGasPrice) <= 0 {
			txFee.Mul(txFee, normalGasPrice)
		} else {
			txFee.Mul(txFee, maxGasPrice)
		}
	} else {
		txFee.Mul(txFee, new(big.Int).SetBytes(tx.GasPrice))
	}
	return txFee
}

func (mongodb *Mongo) TransformTx(blk *types.Eth1Block, cache *freecache.Cache) (interface{}, interface{}, error) {
	var bulkData []mongo.WriteModel
	var bulkMetadataUpdates []mongo.WriteModel

	for i, tx := range blk.Transactions {
		if i > 9999 {
			return nil, nil, fmt.Errorf("unexpected number of transactions in block expected at most 9999 but got: %v, tx: %x", i, tx.GetHash())
		}

		to := tx.GetTo()
		isContract := false
		if !bytes.Equal(tx.GetContractAddress(), ZERO_ADDRESS) {
			to = tx.GetContractAddress()
			isContract = true
		}
		// logger.Infof("sending to: %x", to)
		invokesContract := false
		if len(tx.GetItx()) > 0 || tx.GetGasUsed() > 21000 || tx.GetErrorMsg() != "" {
			invokesContract = true
		}

		method := make([]byte, 0)
		if len(tx.GetData()) > 3 {
			method = tx.GetData()[:4]
		}
		fee := new(big.Int).Mul(new(big.Int).SetBytes(tx.GetGasPrice()), big.NewInt(int64(tx.GetGasUsed()))).Bytes()

		indexedTx := &entity.TransactionIndex{
			ChainId:            mongodb.ChainId,
			Type:               "transactionindex",
			Hash:               tx.GetHash(),
			BlockNumber:        blk.GetNumber(),
			Time:               primitive.Timestamp{T: uint32(blk.GetTime().AsTime().Unix()), I: 0},
			MethodId:           method,
			From:               tx.GetFrom(),
			To:                 to,
			Value:              tx.GetValue(),
			TxFee:              fee,
			GasPrice:           tx.GetGasPrice(),
			IsContractCreation: isContract,
			InvokesContract:    invokesContract,
			ErrorMsg:           tx.GetErrorMsg(),
		}

		// Mark Sender and Recipient for balance update
		mongodb.markBalanceUpdate(indexedTx.From, []byte{0x0}, bulkMetadataUpdates, cache)
		mongodb.markBalanceUpdate(indexedTx.To, []byte{0x0}, bulkMetadataUpdates, cache)

		if len(indexedTx.Hash) != 32 {
			logger.Fatalf("retrieved hash of length %v for a tx in block %v", len(indexedTx.Hash), blk.GetNumber())
		}

		doc, err := utils.ToDoc(indexedTx)
		if err != nil {
			return nil, nil, err
		}
		insertBlock := mongo.NewInsertOneModel().SetDocument(doc)
		bulkData = append(bulkData, insertBlock)
		// indexes := []mongo.IndexModel{{Keys: bson.D{{Key: "to", Value: -1}, {Key: "time", Value: -1}, {Key: "block", Value: -1}, {Key: "method", Value: -1}, {Key: "from", Value: -1}, {Key: "iscontractcreation", Value: -1}, {Key: "errormsg", Value: -1}}}}
		indexes := []string{
			fmt.Sprintf("%s:I:TX:%x:TO:%x:%s:%019d", mongodb.ChainId, tx.GetFrom(), to, blk.GetTime(), i),
			fmt.Sprintf("%s:I:TX:%x:TIME:%s:%019d", mongodb.ChainId, tx.GetFrom(), blk.GetTime(), i),
			fmt.Sprintf("%s:I:TX:%x:BLOCK:%09d:%d", mongodb.ChainId, tx.GetFrom(), blk.GetNumber(), i),
			fmt.Sprintf("%s:I:TX:%x:METHOD:%x:%019d:%d", mongodb.ChainId, tx.GetFrom(), method, blk.GetTime(), i),
			fmt.Sprintf("%s:I:TX:%x:FROM:%x:%019d:%d", mongodb.ChainId, to, tx.GetFrom(), blk.GetTime(), i),
			fmt.Sprintf("%s:I:TX:%x:TIME:%019d:%d", mongodb.ChainId, to, blk.GetTime(), i),
			fmt.Sprintf("%s:I:TX:%x:BLOCK:%09d:%d", mongodb.ChainId, to, blk.GetNumber(), i),
			fmt.Sprintf("%s:I:TX:%x:METHOD:%x:%019d:%d", mongodb.ChainId, to, method, blk.GetTime(), i),
		}

		if indexedTx.ErrorMsg != "" {
			indexes = append(indexes, fmt.Sprintf("%s:I:TX:%x:ERROR:%019d:%d", mongodb.ChainId, tx.GetFrom(), blk.GetTime(), i))
			indexes = append(indexes, fmt.Sprintf("%s:I:TX:%x:ERROR:%019d:%d", mongodb.ChainId, to, blk.GetTime(), i))
		}

		if indexedTx.IsContractCreation {
			indexes = append(indexes, fmt.Sprintf("%s:I:TX:%x:CONTRACT:%019d:%d", mongodb.ChainId, tx.GetFrom(), blk.GetTime(), i))
			indexes = append(indexes, fmt.Sprintf("%s:I:TX:%x:CONTRACT:%019d:%d", mongodb.ChainId, to, blk.GetTime(), i))
		}

		txIdentifier := fmt.Sprintf("%s:TX:%x", mongodb.ChainId, tx.GetHash())

		for _, idx := range indexes {
			mut := &entity.Indexes{
				Type:  "index",
				Key:   idx,
				Value: txIdentifier,
			}
			doc, err := utils.ToDoc(mut)
			if err != nil {
				return nil, nil, err
			}
			insertBlock := mongo.NewInsertOneModel().SetDocument(doc)
			bulkData = append(bulkData, insertBlock)
		}
	}

	return bulkData, bulkMetadataUpdates, nil
}

func (mongodb *Mongo) TransformItx(blk *types.Eth1Block, cache *freecache.Cache) (interface{}, interface{}, error) {
	var bulkData []mongo.WriteModel
	var bulkMetadataUpdates []mongo.WriteModel

	for i, tx := range blk.GetTransactions() {
		if i > 9999 {
			return nil, nil, fmt.Errorf("unexpected number of transactions in block expected at most 9999 but got: %v, tx: %x", i, tx.GetHash())
		}
		for j, idx := range tx.GetItx() {
			if j > 999999 {
				return nil, nil, fmt.Errorf("unexpected number of internal transactions in block expected at most 999999 but got: %v, tx: %x", j, tx.GetHash())
			}

			if idx.Path == "[]" || bytes.Equal(idx.Value, []byte{0x0}) { // skip top level call & empty calls
				continue
			}
			indexedItx := &entity.InternalTransactionIndex{
				ChainId:     mongodb.ChainId,
				ParentHash:  tx.GetHash(),
				BlockNumber: blk.GetNumber(),
				Time:        primitive.Timestamp{T: uint32(blk.GetTime().AsTime().Unix()), I: 0},
				Type:        idx.GetType(),
				From:        idx.GetFrom(),
				To:          idx.GetTo(),
				Value:       idx.GetValue(),
			}
			mongodb.markBalanceUpdate(indexedItx.To, []byte{0x0}, bulkMetadataUpdates, cache)
			mongodb.markBalanceUpdate(indexedItx.From, []byte{0x0}, bulkMetadataUpdates, cache)

			doc, err := utils.ToDoc(indexedItx)
			if err != nil {
				return nil, nil, err
			}
			insertBlock := mongo.NewInsertOneModel().SetDocument(doc)
			bulkData = append(bulkData, insertBlock)

			indexes := []string{
				fmt.Sprintf("%s:I:ITX:%x:TO:%x:%019d:%d:%d", mongodb.ChainId, idx.GetFrom(), idx.GetTo(), blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ITX:%x:FROM:%x:%019d:%d:%d", mongodb.ChainId, idx.GetTo(), idx.GetFrom(), blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ITX:%x:TIME:%019d:%d:%d", mongodb.ChainId, idx.GetFrom(), blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ITX:%x:TIME:%019d:%d:%d", mongodb.ChainId, idx.GetTo(), blk.GetTime(), i, j),
			}

			itxIdentifier := fmt.Sprintf("%s:ITX:%x:%d", mongodb.ChainId, tx.GetHash(), j)
			for _, idx := range indexes {
				mut := &entity.Indexes{
					Type:  "index",
					Key:   idx,
					Value: itxIdentifier,
				}
				doc, err := utils.ToDoc(mut)
				if err != nil {
					return nil, nil, err
				}
				insertBlock := mongo.NewInsertOneModel().SetDocument(doc)
				bulkData = append(bulkData, insertBlock)
			}

		}
	}

	return bulkData, bulkMetadataUpdates, nil
}

func (mongodb *Mongo) TransformERC20(blk *types.Eth1Block, cache *freecache.Cache) (interface{}, interface{}, error) {
	var bulkData []mongo.WriteModel
	var bulkMetadataUpdates []mongo.WriteModel

	filterer, err := erc20.NewErc20Filterer(common.Address{}, nil)
	if err != nil {
		log.Printf("error creating filterer: %v", err)
	}

	for i, tx := range blk.GetTransactions() {
		if i > 9999 {
			return nil, nil, fmt.Errorf("unexpected number of transactions in block expected at most 9999 but got: %v, tx: %x", i, tx.GetHash())
		}

		for j, log := range tx.GetLogs() {
			if j > 99999 {
				return nil, nil, fmt.Errorf("unexpected number of logs in block expected at most 99999 but got: %v tx: %x", j, tx.GetHash())
			}

			if len(log.GetTopics()) != 3 || !bytes.Equal(log.GetTopics()[0], erc20.TransferTopic) {
				continue
			}

			topics := make([]common.Hash, 0, len(log.GetTopics()))

			for _, lTopic := range log.GetTopics() {
				topics = append(topics, common.BytesToHash(lTopic))
			}

			ethLog := eth_types.Log{
				Address:     common.BytesToAddress(log.GetAddress()),
				Data:        log.Data,
				Topics:      topics,
				BlockNumber: blk.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(i),
				BlockHash:   common.BytesToHash(blk.GetHash()),
				Index:       uint(j),
				Removed:     log.GetRemoved(),
			}
			transfer, _ := filterer.ParseTransfer(ethLog)
			if transfer == nil {
				continue
			}

			value := []byte{}
			if transfer != nil && transfer.Value != nil {
				value = transfer.Value.Bytes()
			}
			indexedLog := &entity.ERC20Index{
				ChainId:      mongodb.ChainId,
				Type:         "erc20index",
				ParentHash:   tx.GetHash(),
				BlockNumber:  blk.GetNumber(),
				Time:         primitive.Timestamp{T: uint32(blk.GetTime().AsTime().Unix()), I: 0},
				TokenAddress: log.Address,
				From:         transfer.From.Bytes(),
				To:           transfer.To.Bytes(),
				Value:        value,
			}
			mongodb.markBalanceUpdate(indexedLog.From, indexedLog.TokenAddress, bulkMetadataUpdates, cache)
			mongodb.markBalanceUpdate(indexedLog.To, indexedLog.TokenAddress, bulkMetadataUpdates, cache)
			doc, err := utils.ToDoc(indexedLog)
			if err != nil {
				return nil, nil, err
			}
			insertBlock := mongo.NewInsertOneModel().SetDocument(doc)
			bulkData = append(bulkData, insertBlock)

			indexes := []string{
				fmt.Sprintf("%s:I:ERC20:%x:TIME:%019d:%d:%d", mongodb.ChainId, indexedLog.From, blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ERC20:%x:TIME:%019d:%d:%d", mongodb.ChainId, indexedLog.To, blk.GetTime(), i, j),

				fmt.Sprintf("%s:I:ERC20:%x:ALL:TIME:%019d:%d:%d", mongodb.ChainId, indexedLog.TokenAddress, blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ERC20:%x:%x:TIME:%019d:%d:%d", mongodb.ChainId, indexedLog.TokenAddress, indexedLog.From, blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ERC20:%x:%x:TIME:%019d:%d:%d", mongodb.ChainId, indexedLog.TokenAddress, indexedLog.To, blk.GetTime(), i, j),

				fmt.Sprintf("%s:I:ERC20:%x:TO:%x:%019d:%d:%d", mongodb.ChainId, indexedLog.From, indexedLog.To, blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ERC20:%x:FROM:%x:%019d:%d:%d", mongodb.ChainId, indexedLog.To, indexedLog.From, blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ERC20:%x:TOKEN_SENT:%x:%019d:%d:%d", mongodb.ChainId, indexedLog.From, indexedLog.TokenAddress, blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ERC20:%x:TOKEN_RECEIVED:%x:%019d:%d:%d", mongodb.ChainId, indexedLog.To, indexedLog.TokenAddress, blk.GetTime(), i, j),
			}

			erc20Identifier := fmt.Sprintf("%s:ERC20:%x:%d", mongodb.ChainId, tx.GetHash(), j)

			for _, idx := range indexes {
				mut := &entity.Indexes{
					Type:  "index",
					Key:   idx,
					Value: erc20Identifier,
				}
				doc, err := utils.ToDoc(mut)
				if err != nil {
					return nil, nil, err
				}
				insertBlock := mongo.NewInsertOneModel().SetDocument(doc)
				bulkData = append(bulkData, insertBlock)
			}
		}
	}

	return bulkData, bulkMetadataUpdates, nil
}

func (mongodb *Mongo) TransformERC721(blk *types.Eth1Block, cache *freecache.Cache) (interface{}, interface{}, error) {
	var bulkData []mongo.WriteModel
	var bulkMetadataUpdates []mongo.WriteModel

	filterer, err := erc721.NewErc721Filterer(common.Address{}, nil)
	if err != nil {
		log.Printf("error creating filterer: %v", err)
	}

	for i, tx := range blk.GetTransactions() {
		if i > 9999 {
			return nil, nil, fmt.Errorf("unexpected number of transactions in block expected at most 9999 but got: %v, tx: %x", i, tx.GetHash())
		}
		for j, log := range tx.GetLogs() {
			if j > 99999 {
				return nil, nil, fmt.Errorf("unexpected number of logs in block expected at most 99999 but got: %v tx: %x", j, tx.GetHash())
			}
			if len(log.GetTopics()) != 4 || !bytes.Equal(log.GetTopics()[0], erc721.TransferTopic) {
				continue
			}

			topics := make([]common.Hash, 0, len(log.GetTopics()))

			for _, lTopic := range log.GetTopics() {
				topics = append(topics, common.BytesToHash(lTopic))
			}

			ethLog := eth_types.Log{
				Address:     common.BytesToAddress(log.GetAddress()),
				Data:        log.Data,
				Topics:      topics,
				BlockNumber: blk.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(i),
				BlockHash:   common.BytesToHash(blk.GetHash()),
				Index:       uint(j),
				Removed:     log.GetRemoved(),
			}

			transfer, _ := filterer.ParseTransfer(ethLog)
			if transfer == nil {
				continue
			}

			tokenId := new(big.Int)
			if transfer != nil && transfer.TokenId != nil {
				tokenId = transfer.TokenId
			}

			indexedLog := &entity.ERC721Index{
				ChainId:      mongodb.ChainId,
				Type:         "erc721index",
				ParentHash:   tx.GetHash(),
				BlockNumber:  blk.GetNumber(),
				Time:         primitive.Timestamp{T: uint32(blk.GetTime().AsTime().Unix()), I: 0},
				TokenAddress: log.Address,
				From:         transfer.From.Bytes(),
				To:           transfer.To.Bytes(),
				TokenId:      tokenId.Bytes(),
			}
			doc, err := utils.ToDoc(indexedLog)
			if err != nil {
				return nil, nil, err
			}
			insertBlock := mongo.NewInsertOneModel().SetDocument(doc)
			bulkData = append(bulkData, insertBlock)

			indexes := []string{
				fmt.Sprintf("%s:I:ERC721:%x:TIME:%019d:%d:%d", mongodb.ChainId, indexedLog.From, blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ERC721:%x:TIME:%019d:%d:%d", mongodb.ChainId, indexedLog.To, blk.GetTime(), i, j),

				fmt.Sprintf("%s:I:ERC721:%x:ALL:TIME:%019d:%d:%d", mongodb.ChainId, indexedLog.TokenAddress, blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ERC721:%x:%x:TIME:%019d:%d:%d", mongodb.ChainId, indexedLog.TokenAddress, indexedLog.From, blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ERC721:%x:%x:TIME:%019d:%d:%d", mongodb.ChainId, indexedLog.TokenAddress, indexedLog.To, blk.GetTime(), i, j),

				fmt.Sprintf("%s:I:ERC721:%x:TO:%x:%019d:%d:%d", mongodb.ChainId, indexedLog.From, indexedLog.To, blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ERC721:%x:FROM:%x:%019d:%d:%d", mongodb.ChainId, indexedLog.To, indexedLog.From, blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ERC721:%x:TOKEN_SENT:%x:%019d:%d:%d", mongodb.ChainId, indexedLog.From, indexedLog.TokenAddress, blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ERC721:%x:TOKEN_RECEIVED:%x:%019d:%d:%d", mongodb.ChainId, indexedLog.To, indexedLog.TokenAddress, blk.GetTime(), i, j),
			}

			erc721Identifier := fmt.Sprintf("%s:ERC721:%x:%d", mongodb.ChainId, tx.GetHash(), j)

			for _, idx := range indexes {
				mut := &entity.Indexes{
					Type:  "index",
					Key:   idx,
					Value: erc721Identifier,
				}
				doc, err := utils.ToDoc(mut)
				if err != nil {
					return nil, nil, err
				}
				insertBlock := mongo.NewInsertOneModel().SetDocument(doc)
				bulkData = append(bulkData, insertBlock)
			}
		}
	}

	return bulkData, bulkMetadataUpdates, nil
}

func (mongodb *Mongo) TransformERC1155(blk *types.Eth1Block, cache *freecache.Cache) (interface{}, interface{}, error) {
	var bulkData []mongo.WriteModel
	var bulkMetadataUpdates []mongo.WriteModel

	filterer, err := erc1155.NewErc1155Filterer(common.Address{}, nil)
	if err != nil {
		log.Printf("error creating filterer: %v", err)
	}

	for i, tx := range blk.GetTransactions() {
		if i > 9999 {
			return nil, nil, fmt.Errorf("unexpected number of transactions in block expected at most 9999 but got: %v, tx: %x", i, tx.GetHash())
		}
		for j, log := range tx.GetLogs() {
			if j > 99999 {
				return nil, nil, fmt.Errorf("unexpected number of logs in block expected at most 99999 but got: %v tx: %x", j, tx.GetHash())
			}
			// no events emitted continue
			if len(log.GetTopics()) != 4 || (!bytes.Equal(log.GetTopics()[0], erc1155.TransferBulkTopic) && !bytes.Equal(log.GetTopics()[0], erc1155.TransferSingleTopic)) {
				continue
			}

			topics := make([]common.Hash, 0, len(log.GetTopics()))

			for _, lTopic := range log.GetTopics() {
				topics = append(topics, common.BytesToHash(lTopic))
			}

			ethLog := eth_types.Log{
				Address:     common.BytesToAddress(log.GetAddress()),
				Data:        log.Data,
				Topics:      topics,
				BlockNumber: blk.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(i),
				BlockHash:   common.BytesToHash(blk.GetHash()),
				Index:       uint(j),
				Removed:     log.GetRemoved(),
			}

			indexedLog := &entity.ERC1155Index{}
			indexedLog.ChainId = mongodb.ChainId
			indexedLog.Type = "erc1155index"
			transferBatch, _ := filterer.ParseTransferBatch(ethLog)
			transferSingle, _ := filterer.ParseTransferSingle(ethLog)
			if transferBatch == nil && transferSingle == nil {
				continue
			}

			if transferBatch != nil {
				ids := make([][]byte, 0, len(transferBatch.Ids))
				for _, id := range transferBatch.Ids {
					ids = append(ids, id.Bytes())
				}

				values := make([][]byte, 0, len(transferBatch.Values))
				for _, val := range transferBatch.Values {
					values = append(values, val.Bytes())
				}

				if len(ids) != len(values) {
					logrus.Errorf("error parsing erc1155 batch transfer logs. Expected len(ids): %v len(values): %v to be the same", len(ids), len(values))
					continue
				}
				for ti := range ids {
					indexedLog.BlockNumber = blk.GetNumber()
					indexedLog.Time = primitive.Timestamp{T: uint32(blk.GetTime().AsTime().Unix()), I: 0}
					indexedLog.ParentHash = tx.GetHash()
					indexedLog.From = transferBatch.From.Bytes()
					indexedLog.To = transferBatch.To.Bytes()
					indexedLog.Operator = transferBatch.Operator.Bytes()
					indexedLog.TokenId = ids[ti]
					indexedLog.Value = values[ti]
					indexedLog.TokenAddress = log.GetAddress()
				}
			} else if transferSingle != nil {
				indexedLog.BlockNumber = blk.GetNumber()
				indexedLog.Time = primitive.Timestamp{T: uint32(blk.GetTime().AsTime().Unix()), I: 0}
				indexedLog.ParentHash = tx.GetHash()
				indexedLog.From = transferSingle.From.Bytes()
				indexedLog.To = transferSingle.To.Bytes()
				indexedLog.Operator = transferSingle.Operator.Bytes()
				indexedLog.TokenId = transferSingle.Id.Bytes()
				indexedLog.Value = transferSingle.Value.Bytes()
				indexedLog.TokenAddress = log.GetAddress()
			}

			doc, err := utils.ToDoc(indexedLog)
			if err != nil {
				return nil, nil, err
			}
			insertBlock := mongo.NewInsertOneModel().SetDocument(doc)
			bulkData = append(bulkData, insertBlock)

			indexes := []string{
				fmt.Sprintf("%s:I:ERC1155:%x:TIME:%019d:%d:%d", mongodb.ChainId, indexedLog.From, blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ERC1155:%x:TIME:%019d:%d:%d", mongodb.ChainId, indexedLog.To, blk.GetTime(), i, j),

				fmt.Sprintf("%s:I:ERC1155:%x:ALL:TIME:%019d:%d:%d", mongodb.ChainId, indexedLog.TokenAddress, blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ERC1155:%x:%x:TIME:%019d:%d:%d", mongodb.ChainId, indexedLog.TokenAddress, indexedLog.From, blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ERC1155:%x:%x:TIME:%019d:%d:%d", mongodb.ChainId, indexedLog.TokenAddress, indexedLog.To, blk.GetTime(), i, j),

				fmt.Sprintf("%s:I:ERC1155:%x:TO:%x:%019d:%d:%d", mongodb.ChainId, indexedLog.From, indexedLog.To, blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ERC1155:%x:FROM:%x:%019d:%d:%d", mongodb.ChainId, indexedLog.To, indexedLog.From, blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ERC1155:%x:TOKEN_SENT:%x:%019d:%d:%d", mongodb.ChainId, indexedLog.From, indexedLog.TokenAddress, blk.GetTime(), i, j),
				fmt.Sprintf("%s:I:ERC1155:%x:TOKEN_RECEIVED:%x:%019d:%d:%d", mongodb.ChainId, indexedLog.To, indexedLog.TokenAddress, blk.GetTime(), i, j),
			}

			erc1155Identifier := fmt.Sprintf("%s:ERC1155:%x:%d", mongodb.ChainId, tx.GetHash(), j)

			for _, idx := range indexes {
				mut := &entity.Indexes{
					Type:  "index",
					Key:   idx,
					Value: erc1155Identifier,
				}
				doc, err := utils.ToDoc(mut)
				if err != nil {
					return nil, nil, err
				}
				insertBlock := mongo.NewInsertOneModel().SetDocument(doc)
				bulkData = append(bulkData, insertBlock)
			}
		}
	}

	return bulkData, bulkMetadataUpdates, nil
}

func (mongodb *Mongo) TransformUncle(block *types.Eth1Block, cache *freecache.Cache) (interface{}, interface{}, error) {
	var bulkData []mongo.WriteModel
	var bulkMetadataUpdates []mongo.WriteModel

	for i, uncle := range block.Uncles {
		if i > 99 {
			return nil, nil, fmt.Errorf("unexpected number of uncles in block expected at most 99 but got: %v", i)
		}

		r := new(big.Int)

		if len(block.Difficulty) > 0 {
			r.Add(big.NewInt(int64(uncle.GetNumber())), big.NewInt(8))
			r.Sub(r, big.NewInt(int64(block.GetNumber())))
			r.Mul(r, utils.Eth1BlockReward(block.GetNumber(), block.Difficulty))
			r.Div(r, big.NewInt(8))

			r.Div(utils.Eth1BlockReward(block.GetNumber(), block.Difficulty), big.NewInt(32))
		}

		uncleIndexed := entity.UncleBlocksIndex{
			Number:      uncle.GetNumber(),
			BlockNumber: block.GetNumber(),
			GasLimit:    uncle.GetGasLimit(),
			GasUsed:     uncle.GetGasUsed(),
			BaseFee:     uncle.GetBaseFee(),
			Difficulty:  uncle.GetDifficulty(),
			Time:        primitive.Timestamp{T: uint32(block.GetTime().AsTime().Unix()), I: 0},
			Reward:      r.Bytes(),
		}

		mongodb.markBalanceUpdate(uncle.Coinbase, []byte{0x0}, bulkMetadataUpdates, cache)

		doc, err := utils.ToDoc(uncleIndexed)
		if err != nil {
			return nil, nil, err
		}
		insertBlock := mongo.NewInsertOneModel().SetDocument(doc)
		bulkData = append(bulkData, insertBlock)

		indexes := []string{
			// Index uncle by the miners address
			fmt.Sprintf("%s:I:U:%x:TIME:%019d:%d", mongodb.ChainId, uncle.GetCoinbase(), block.Time, i),
		}

		uncleBlockIdentifier := fmt.Sprintf("%s:U:%09d:%d", mongodb.ChainId, block.GetNumber(), i)
		for _, idx := range indexes {
			mut := &entity.Indexes{
				Type:  "index",
				Key:   idx,
				Value: uncleBlockIdentifier,
			}
			doc, err := utils.ToDoc(mut)
			if err != nil {
				return nil, nil, err
			}
			insertBlock := mongo.NewInsertOneModel().SetDocument(doc)
			bulkData = append(bulkData, insertBlock)
		}
	}

	return bulkData, bulkMetadataUpdates, nil
}

func (mongodb *Mongo) TransformWithdrawals(block *types.Eth1Block, cache *freecache.Cache) (interface{}, interface{}, error) {
	var bulkData []mongo.WriteModel
	var bulkMetadataUpdates []mongo.WriteModel

	if len(block.Withdrawals) > int(utils.Config.Chain.Config.MaxWithdrawalsPerPayload) {
		return nil, nil, fmt.Errorf("unexpected number of withdrawals in block expected at most %v but got: %v", utils.Config.Chain.Config.MaxWithdrawalsPerPayload, len(block.Withdrawals))
	}

	for _, withdrawal := range block.Withdrawals {
		withdrawalIndexed := entity.WithdrawalIndex{
			BlockNumber:    block.Number,
			Index:          withdrawal.Index,
			ValidatorIndex: withdrawal.ValidatorIndex,
			Address:        withdrawal.Address,
			Amount:         withdrawal.Amount,
			Time:           primitive.Timestamp{T: uint32(block.Time.AsTime().Unix()), I: 0},
		}

		mongodb.markBalanceUpdate(withdrawal.Address, []byte{0x0}, bulkMetadataUpdates, cache)

		doc, err := utils.ToDoc(withdrawalIndexed)
		if err != nil {
			return nil, nil, err
		}
		insertBlock := mongo.NewInsertOneModel().SetDocument(doc)
		bulkData = append(bulkData, insertBlock)

		indexes := []string{
			// Index withdrawal by address
			fmt.Sprintf("%s:I:W:%x:TIME:%019d:%d", mongodb.ChainId, withdrawal.Address, block.Time, int(withdrawal.Index)),
		}

		withdrawalIndexIdentifier := fmt.Sprintf("%s:W:%09d:%d", mongodb.ChainId, block.GetNumber(), int(withdrawal.Index))
		for _, idx := range indexes {
			mut := &entity.Indexes{
				Type:  "index",
				Key:   idx,
				Value: withdrawalIndexIdentifier,
			}
			doc, err := utils.ToDoc(mut)
			if err != nil {
				return nil, nil, err
			}
			insertBlock := mongo.NewInsertOneModel().SetDocument(doc)
			bulkData = append(bulkData, insertBlock)
		}
	}

	return bulkData, bulkMetadataUpdates, nil
}

func (mongodb *Mongo) markBalanceUpdate(address []byte, token []byte, mutations interface{}, cache *freecache.Cache) {
	balanceUpdateCacheKey := []byte(fmt.Sprintf("%s:B:%x:%x", mongodb.ChainId, address, token)) // format is B: for balance update as chainid:prefix:address (token id will be encoded as column name)
	if _, err := cache.Get(balanceUpdateCacheKey); err != nil {
		update := entity.BalanceUpdates{
			ChainId: mongodb.ChainId,
			Token:   token,
			Address: address,
		}
		doc, _ := utils.ToDoc(update)
		insertBlock := mongo.NewInsertOneModel().SetDocument(doc)
		mutations = append([]interface{}{mutations}, insertBlock)
		cache.Set(balanceUpdateCacheKey, []byte{0x1}, int((time.Hour * 48).Seconds()))
	}
}

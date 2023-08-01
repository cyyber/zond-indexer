package handlers

import (
	"fmt"
	"math/big"

	"github.com/Prajjawalk/zond-indexer/db"
	"github.com/Prajjawalk/zond-indexer/rpc"
	"github.com/Prajjawalk/zond-indexer/services"
	"github.com/Prajjawalk/zond-indexer/types"
	"github.com/Prajjawalk/zond-indexer/utils"
)

func GetExecutionBlockPageData(number uint64, limit int) (*types.Eth1BlockPageData, error) {
	block, err := db.MongodbClient.GetBlockFromBlocksTable(number)
	if diffToHead := int64(services.LatestEth1BlockNumber()) - int64(number); err != nil && diffToHead < 0 && diffToHead >= -5 {
		block, _, err = rpc.CurrentErigonClient.GetBlock(int64(number))
	}
	if err != nil {
		return nil, err
	}

	// retrieve address names from bigtable
	names := make(map[string]string)
	names[string(block.Coinbase)] = ""
	for _, tx := range block.Transactions {
		names[string(tx.From)] = ""
		names[string(tx.To)] = ""
	}
	for _, uncle := range block.Uncles {
		names[string(uncle.Coinbase)] = ""
	}
	names, _, err = db.MongodbClient.GetAddressesNamesArMetadata(&names, nil)
	if err != nil {
		return nil, err
	}

	// calculate total block reward and set lowest gas price
	txs := []types.Eth1BlockPageTransaction{}
	txFees := new(big.Int)
	lowestGasPrice := big.NewInt(1 << 62)
	for _, tx := range block.Transactions {
		// sum txFees
		txFee := db.CalculateTxFeeFromTransaction(tx, new(big.Int).SetBytes(block.BaseFee))
		txFees.Add(txFees, txFee)

		effectiveGasPrice := big.NewInt(0)
		if gasUsed := new(big.Int).SetUint64(tx.GasUsed); gasUsed.Cmp(big.NewInt(0)) != 0 {
			// calculate effective gas price
			effectiveGasPrice = new(big.Int).Div(txFee, gasUsed)
			if effectiveGasPrice.Cmp(lowestGasPrice) < 0 {
				lowestGasPrice = effectiveGasPrice
			}
		}

		// set tx to if tx is contract creation
		if tx.To == nil && len(tx.Itx) >= 1 {
			tx.To = tx.Itx[0].To
			names[string(tx.To)] = "Contract Creation"
		}

		method := make([]byte, 0)
		if len(tx.GetData()) > 3 && (len(tx.GetItx()) > 0 || tx.GetGasUsed() > 21000 || tx.GetErrorMsg() != "") {
			method = tx.GetData()[:4]
		}
		txs = append(txs, types.Eth1BlockPageTransaction{
			Hash:          fmt.Sprintf("%#x", tx.Hash),
			HashFormatted: utils.FormatAddressWithLimits(tx.Hash, "", false, "tx", 15, 18, true),
			From:          fmt.Sprintf("%#x", tx.From),
			FromFormatted: utils.FormatAddressWithLimits(tx.From, names[string(tx.From)], false, "address", 15, 20, true),
			To:            fmt.Sprintf("%#x", tx.To),
			ToFormatted:   utils.FormatAddressWithLimits(tx.To, names[string(tx.To)], names[string(tx.To)] == "Contract Creation" || len(method) > 0, "address", 15, 20, true),
			Value:         new(big.Int).SetBytes(tx.Value),
			Fee:           txFee,
			GasPrice:      effectiveGasPrice,
			Method:        fmt.Sprintf("%#x", method),
		})
	}

	blockReward := utils.Eth1BlockReward(block.Number, block.Difficulty)

	uncleInclusionRewards := new(big.Int)
	uncleInclusionRewards.Div(blockReward, big.NewInt(32)).Mul(uncleInclusionRewards, big.NewInt(int64(len(block.Uncles))))
	uncles := []types.Eth1BlockPageData{}
	for _, uncle := range block.Uncles {
		reward := big.NewInt(int64(uncle.Number - block.Number + 8))
		reward.Mul(reward, blockReward).Div(reward, big.NewInt(8))
		uncles = append(uncles, types.Eth1BlockPageData{
			Number:       uncle.Number,
			MinerAddress: fmt.Sprintf("%#x", uncle.Coinbase),
			//MinerFormatted: utils.FormatAddress(uncle.Coinbase, nil, names[string(uncle.Coinbase)], false, false, false),
			MinerFormatted: utils.FormatAddressWithLimits(uncle.Coinbase, names[string(uncle.Coinbase)], false, "address", 42, 42, true),
			Reward:         reward,
			Extra:          string(uncle.Extra),
		})
	}

	if limit > 0 {
		if len(txs) > limit {
			txs = txs[:limit]
		} else {
			txs = txs[:0]
		}
	}

	burnedEth := new(big.Int).Mul(new(big.Int).SetBytes(block.BaseFee), big.NewInt(int64(block.GasUsed)))
	blockReward.Add(blockReward, txFees).Add(blockReward, uncleInclusionRewards).Sub(blockReward, burnedEth)
	nextBlock := number + 1
	if nextBlock > services.LatestEth1BlockNumber() {
		nextBlock = 0
	}
	eth1BlockPageData := types.Eth1BlockPageData{
		Number:        number,
		PreviousBlock: number - 1,
		NextBlock:     nextBlock,
		TxCount:       uint64(len(block.Transactions)),
		UncleCount:    uint64(len(block.Uncles)),
		Hash:          fmt.Sprintf("%#x", block.Hash),
		ParentHash:    fmt.Sprintf("%#x", block.ParentHash),
		MinerAddress:  fmt.Sprintf("%#x", block.Coinbase),
		//MinerFormatted: utils.FormatAddress(block.Coinbase, nil, names[string(block.Coinbase)], false, false, false),
		MinerFormatted: utils.FormatAddressWithLimits(block.Coinbase, names[string(block.Coinbase)], false, "address", 42, 42, true),
		Reward:         blockReward,
		MevReward:      db.CalculateMevFromBlock(block),
		TxFees:         txFees,
		GasUsage:       utils.FormatBlockUsage(block.GasUsed, block.GasLimit),
		GasLimit:       block.GasLimit,
		LowestGasPrice: lowestGasPrice,
		Ts:             block.GetTime().AsTime(),
		Difficulty:     new(big.Int).SetBytes(block.Difficulty),
		BaseFeePerGas:  new(big.Int).SetBytes(block.BaseFee),
		BurnedFees:     burnedEth,
		Extra:          fmt.Sprintf("%#x", block.Extra),
		Txs:            txs,
		Uncles:         uncles,
	}

	var relaysData struct {
		MevRecipient []byte          `db:"proposer_fee_recipient"`
		MevBribe     types.WeiString `db:"value"`
	}
	// try to get mev rewards from relays_blocks table
	err = db.ReaderDb.Get(&relaysData, `SELECT proposer_fee_recipient, value FROM relays_blocks WHERE relays_blocks.exec_block_hash = $1 limit 1`, block.Hash)
	if err == nil {
		eth1BlockPageData.MevBribe = relaysData.MevBribe.BigInt()
		eth1BlockPageData.MevRecipientFormatted = utils.FormatAddressWithLimits(relaysData.MevRecipient, names[string(relaysData.MevRecipient)], false, "address", 42, 42, true)
	}
	return &eth1BlockPageData, nil
}
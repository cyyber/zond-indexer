package handlers

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Prajjawalk/zond-indexer/db"
	"github.com/Prajjawalk/zond-indexer/services"
	"github.com/Prajjawalk/zond-indexer/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"golang.org/x/exp/maps"
)

// ApiEth1Deposit godoc
// @Summary Get an eth1 deposit by its eth1 transaction hash
// @Tags Execution
// @Produce  json
// @Param  txhash path string true "Eth1 transaction hash"
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/eth1deposit/{txhash} [get]
func ApiEth1Deposit(c *gin.Context) {
	w := c.Writer
	r := c.Request
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	eth1TxHash, err := hex.DecodeString(strings.Replace(vars["txhash"], "0x", "", -1))
	if err != nil {
		sendErrorResponse(w, r.URL.String(), "invalid eth1 tx hash provided")
		return
	}

	rows, err := db.ReaderDb.Query("SELECT amount, block_number, block_ts, from_address, merkletree_index, publickey, removed, signature, tx_hash, tx_index, tx_input, valid_signature, withdrawal_credentials FROM eth1_deposits WHERE tx_hash = $1", eth1TxHash)
	if err != nil {
		sendErrorResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}
	defer rows.Close()

	returnQueryResults(rows, w, r)
}

func getValidatorExecutionPerformance(queryIndices []uint64) ([]types.ExecutionPerformanceResponse, error) {
	latestEpoch := services.LatestEpoch()
	last30dTimestamp := time.Now().Add(-31 * 24 * time.Hour)
	last7dTimestamp := time.Now().Add(-7 * 24 * time.Hour)
	last1dTimestamp := time.Now().Add(-1 * 24 * time.Hour)

	monthRange := latestEpoch - 7200
	if latestEpoch < 7200 {
		monthRange = 0
	}

	var execBlocks []types.ExecBlockProposer
	err := db.ReaderDb.Select(&execBlocks,
		`SELECT 
			exec_block_number, 
			proposer 
			FROM blocks 
		WHERE proposer = ANY($1) 
		AND exec_block_number IS NOT NULL 
		AND exec_block_number > 0 
		AND epoch > $2`,
		pq.Array(queryIndices),
		monthRange, // 32d range
	)
	if err != nil {
		return nil, fmt.Errorf("error cannot get proposed blocks from db with indicies: %+v and epoch: %v, err: %w", queryIndices, latestEpoch, err)
	}

	blockList, blockToProposerMap := getBlockNumbersAndMapProposer(execBlocks)

	blocks, err := db.MongodbClient.GetBlocksIndexedMultiple(blockList, 10000)
	if err != nil {
		return nil, fmt.Errorf("error cannot get blocks from bigtable using GetBlocksIndexedMultiple: %w", err)
	}

	resultPerProposer := make(map[uint64]types.ExecutionPerformanceResponse)

	relaysData, err := db.GetRelayDataForIndexedBlocks(blocks)
	if err != nil {
		// logger.WithError(err).Errorf("can not get relays data")
		return nil, fmt.Errorf("error can not get relays data: %w", err)
	}

	for _, block := range blocks {
		proposer := blockToProposerMap[block.Number].Proposer
		result, ok := resultPerProposer[proposer]
		if !ok {
			result = types.ExecutionPerformanceResponse{
				Performance1d:  big.NewInt(0),
				Performance7d:  big.NewInt(0),
				Performance31d: big.NewInt(0),
				ValidatorIndex: proposer,
			}
		}

		txFees := big.NewInt(0).SetBytes(block.TxReward)
		mev := big.NewInt(0).SetBytes(block.Mev)
		income := big.NewInt(0).Add(txFees, mev)

		var mevBribe *big.Int = big.NewInt(0)
		relayData, ok := relaysData[common.BytesToHash(block.Hash)]
		if ok {
			mevBribe = relayData.MevBribe.BigInt()
		}

		var producerReward *big.Int
		if mevBribe.Int64() == 0 {
			producerReward = income
		} else {
			producerReward = mevBribe
		}

		if block.Time.AsTime().After(last30dTimestamp) {
			result.Performance31d = result.Performance31d.Add(result.Performance31d, producerReward)
		}
		if block.Time.AsTime().After(last7dTimestamp) {
			result.Performance7d = result.Performance7d.Add(result.Performance7d, producerReward)
		}
		if block.Time.AsTime().After(last1dTimestamp) {
			result.Performance1d = result.Performance1d.Add(result.Performance1d, producerReward)
		}

		resultPerProposer[proposer] = result
	}

	return maps.Values(resultPerProposer), nil
}

func getBlockNumbersAndMapProposer(data []types.ExecBlockProposer) ([]uint64, map[uint64]types.ExecBlockProposer) {
	blockList := []uint64{}
	blockToProposerMap := make(map[uint64]types.ExecBlockProposer)
	for _, execBlock := range data {
		blockList = append(blockList, execBlock.ExecBlock)
		blockToProposerMap[execBlock.ExecBlock] = execBlock
	}
	return blockList, blockToProposerMap
}

func findExecBlockNumbersByProposerIndex(indices []uint64, offset, limit uint64) ([]uint64, map[uint64]types.ExecBlockProposer, error) {
	var blockListSub []types.ExecBlockProposer
	err := db.ReaderDb.Select(&blockListSub,
		`SELECT 
			exec_block_number,
			proposer,
			slot,
			epoch  
		FROM blocks 
		WHERE proposer = ANY($1)
		AND exec_block_number IS NOT NULL AND exec_block_number > 0 
		ORDER BY exec_block_number DESC
		OFFSET $2 LIMIT $3`,
		pq.Array(indices),
		offset,
		limit,
	)
	if err != nil {
		return nil, nil, err
	}
	blockList, blockProposerMap := getBlockNumbersAndMapProposer(blockListSub)
	return blockList, blockProposerMap, nil
}

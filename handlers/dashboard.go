package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Prajjawalk/zond-indexer/db"
	"github.com/Prajjawalk/zond-indexer/types"
	"github.com/Prajjawalk/zond-indexer/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"golang.org/x/sync/errgroup"
)

var ErrTooManyValidators = errors.New("too many validators")

func parseValidatorsFromQueryString(str string, validatorLimit int) ([]uint64, error) {
	if str == "" {
		return []uint64{}, nil
	}

	strSplit := strings.Split(str, ",")
	strSplitLen := len(strSplit)

	// we only support up to [validatorLimit] validators
	if strSplitLen > validatorLimit {
		return []uint64{}, ErrTooManyValidators
	}

	validators := make([]uint64, strSplitLen)
	keys := make(map[uint64]bool, strSplitLen)

	for i, vStr := range strSplit {
		v, err := strconv.ParseUint(vStr, 10, 64)
		if err != nil {
			return []uint64{}, err
		}
		// make sure keys are uniq
		if exists := keys[v]; exists {
			continue
		}
		keys[v] = true
		validators[i] = v
	}

	return validators, nil
}

func getExecutionChartData(indices []uint64, currency string) ([]*types.ChartDataPoint, error) {
	var limit uint64 = 300
	blockList, consMap, err := findExecBlockNumbersByProposerIndex(indices, 0, limit)
	if err != nil {
		return nil, err
	}

	blocks, err := db.MongodbClient.GetBlocksIndexedMultiple(blockList, limit)
	if err != nil {
		return nil, err
	}
	relaysData, err := db.GetRelayDataForIndexedBlocks(blocks)
	if err != nil {
		return nil, err
	}

	var chartData = make([]*types.ChartDataPoint, len(blocks))
	epochsPerDay := utils.EpochsPerDay()

	for i := len(blocks) - 1; i >= 0; i-- {
		consData := consMap[blocks[i].Number]
		day := int64(consData.Epoch / epochsPerDay)
		color := "#90ed7d"
		totalReward, _ := utils.WeiToEther(utils.Eth1TotalReward(blocks[i])).Float64()
		relayData, ok := relaysData[common.BytesToHash(blocks[i].Hash)]
		if ok {
			totalReward, _ = utils.WeiToEther(relayData.MevBribe.BigInt()).Float64()
		}

		//balanceTs := blocks[i].GetTime().AsTime().Unix()

		chartData[len(blocks)-1-i] = &types.ChartDataPoint{
			X:     float64(utils.DayToTime(day).Unix() * 1000), //float64(balanceTs * 1000),
			Y:     utils.ExchangeRateForCurrency(currency) * totalReward,
			Color: color,
		}
	}
	return chartData, nil
}

func DashboardDataProposals(c *gin.Context) {
	w := c.Writer
	r := c.Request
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	validatorLimit := getUserPremium(r).MaxValidators
	filterArr, err := parseValidatorsFromQueryString(q.Get("validators"), validatorLimit)
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}
	filter := pq.Array(filterArr)

	proposals := []struct {
		Slot   uint64
		Status uint64
	}{}

	err = db.ReaderDb.Select(&proposals, `
		SELECT slot, status
		FROM blocks
		WHERE proposer = ANY($1)
		ORDER BY slot`, filter)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error retrieving block-proposals")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	proposalsResult := make([][]uint64, len(proposals))
	for i, b := range proposals {
		proposalsResult[i] = []uint64{
			uint64(utils.SlotToTime(b.Slot).Unix()),
			b.Status,
		}
	}

	err = json.NewEncoder(w).Encode(proposalsResult)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error enconding json response")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

// DashboardDataBalance retrieves the income history of a set of validators
func DashboardDataBalance(c *gin.Context) {
	w := c.Writer
	r := c.Request
	currency := GetCurrency(r)

	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	validatorLimit := getUserPremium(r).MaxValidators
	queryValidators, err := parseValidatorsFromQueryString(q.Get("validators"), validatorLimit)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error parsing validators from query string")
		http.Error(w, "Invalid query", 400)
		return
	}
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}
	if len(queryValidators) < 1 {
		http.Error(w, "Invalid query", 400)
		return
	}

	incomeHistoryChartData, _, err := db.GetValidatorIncomeHistoryChart(queryValidators, currency)
	if err != nil {
		logger.Errorf("failed to genereate income history chart data for dashboard view: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(incomeHistoryChartData)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error enconding json response")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

// Dashboard Chart that combines balance data and
func DashboardDataBalanceCombined(c *gin.Context) {
	w := c.Writer
	r := c.Request
	currency := GetCurrency(r)

	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	validatorLimit := getUserPremium(r).MaxValidators

	queryValidators, err := parseValidatorsFromQueryString(q.Get("validators"), validatorLimit)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error parsing validators from query string")
		http.Error(w, "Invalid query", 400)
		return
	}
	if len(queryValidators) < 1 {
		http.Error(w, "Invalid query", 400)
		return
	}

	g, _ := errgroup.WithContext(context.Background())
	var incomeHistoryChartData []*types.ChartDataPoint
	var executionChartData []*types.ChartDataPoint
	g.Go(func() error {
		incomeHistoryChartData, _, err = db.GetValidatorIncomeHistoryChart(queryValidators, currency)
		return err
	})

	g.Go(func() error {
		executionChartData, err = getExecutionChartData(queryValidators, currency)
		return err
	})

	err = g.Wait()
	if err != nil {
		logger.Errorf("combined balance chart %v", err)
		sendErrorResponse(w, r.URL.String(), err.Error())
		return
	}

	var response struct {
		ConsensusChartData []*types.ChartDataPoint `json:"consensusChartData"`
		ExecutionChartData []*types.ChartDataPoint `json:"executionChartData"`
	}
	response.ConsensusChartData = incomeHistoryChartData
	response.ExecutionChartData = executionChartData

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error enconding json response")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

package db

import (
	"github.com/Prajjawalk/zond-indexer/types"
	"github.com/Prajjawalk/zond-indexer/utils"
	"github.com/lib/pq"
)

func GetValidatorIncomeHistoryChart(validator_indices []uint64, currency string) ([]*types.ChartDataPoint, int64, error) {
	incomeHistory, currentDayIncome, err := GetValidatorIncomeHistory(validator_indices, 0, 0)
	if err != nil {
		return nil, 0, err
	}
	var clRewardsSeries = make([]*types.ChartDataPoint, len(incomeHistory))

	for i := 0; i < len(incomeHistory); i++ {
		color := "#7cb5ec"
		if incomeHistory[i].ClRewards < 0 {
			color = "#f7a35c"
		}
		balanceTs := utils.DayToTime(incomeHistory[i].Day)
		clRewardsSeries[i] = &types.ChartDataPoint{X: float64(balanceTs.Unix() * 1000), Y: utils.ExchangeRateForCurrency(currency) * (float64(incomeHistory[i].ClRewards) / 1e9), Color: color}
	}
	return clRewardsSeries, currentDayIncome, err
}

func GetValidatorIncomeHistory(validator_indices []uint64, lowerBoundDay uint64, upperBoundDay uint64) ([]types.ValidatorIncomeHistory, int64, error) {
	if upperBoundDay == 0 {
		upperBoundDay = 65536
	}
	queryValidatorsArr := pq.Array(validator_indices)

	var result []types.ValidatorIncomeHistory
	err := ReaderDb.Select(&result, `
		SELECT 
			day, 
			SUM(COALESCE(cl_rewards_gwei, 0)) AS cl_rewards_gwei,
			SUM(COALESCE(end_balance, 0)) AS end_balance
		FROM validator_stats 
		WHERE validatorindex = ANY($1) AND day BETWEEN $2 AND $3 
		GROUP BY day 
		ORDER BY day
	;`, queryValidatorsArr, lowerBoundDay, upperBoundDay)
	if err != nil {
		return nil, 0, err
	}

	// retrieve rewards for epochs not yet in stats
	currentDayIncome := int64(0)
	if upperBoundDay == 65536 {
		lastDay := int64(0)
		if len(result) > 0 {
			lastDay = result[len(result)-1].Day
		} else {
			err = ReaderDb.Get(&lastDay, "SELECT COALESCE(MAX(day), 0) FROM validator_stats_status")
			if err != nil {
				return nil, 0, err
			}
		}

		currentDay := uint64(lastDay + 1)
		startEpoch := currentDay * utils.EpochsPerDay()
		endEpoch := startEpoch + utils.EpochsPerDay() - 1
		income, err := MongodbClient.GetValidatorIncomeDetailsHistory(validator_indices, startEpoch, endEpoch)

		if err != nil {
			return nil, 0, err
		}

		for _, ids := range income {
			for _, id := range ids {
				currentDayIncome += id.TotalClRewards()
			}
		}

		result = append(result, types.ValidatorIncomeHistory{
			Day:       int64(currentDay),
			ClRewards: currentDayIncome,
		})
	}

	return result, currentDayIncome, err
}

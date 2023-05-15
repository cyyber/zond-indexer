package interfaces

import "github.com/Prajjawalk/zond-indexer/entity"

type IncomeDetails interface {
	SaveValidatorIncomeDetails(epoch uint64, rewards map[uint64]*entity.IncomeDetailsColumnFamily) error
	GetEpochIncomeHistoryDescending(startEpoch uint64, endEpoch uint64) (*entity.IncomeDetailsColumnFamily, error)
	GetEpochIncomeHistory(epoch uint64) (*entity.IncomeDetailsColumnFamily, error)
	GetValidatorIncomeDetailsHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]map[uint64]*entity.IncomeDetailsColumnFamily, error)
	GetAggregatedValidatorIncomeDetailsHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]*entity.IncomeDetailsColumnFamily, error)
}

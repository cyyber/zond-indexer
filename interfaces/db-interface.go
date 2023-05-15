package interfaces

import (
	"math/big"
	"time"

	"github.com/Prajjawalk/zond-indexer/entity"
	"github.com/Prajjawalk/zond-indexer/types"
	"github.com/coocood/freecache"
)

type Database interface {
	Close()
	GetClient() interface{}
	SaveMachineMetric(process string, userID uint64, machine string, data *entity.MachineMetrics) error
	GetMachineMetricsMachineNames(userID uint64) ([]string, error)
	GetMachineMetricsMachineCount(userID uint64) (uint64, error)
	GetMachineMetricsNode(userID uint64, limit, offset int) ([]*types.MachineMetricNode, error)
	GetMachineMetricsValidator(userID uint64, limit, offset int) ([]*types.MachineMetricValidator, error)
	GetMachineMetricsSystem(userID uint64, limit, offset int) ([]*types.MachineMetricSystem, error)
	GetMachineMetricsForNotifications(rowKeys interface{}) (map[uint64]map[string]*types.MachineMetricSystemUser, error)
	SaveValidatorBalances(epoch uint64, validators []*types.Validator) error
	SaveAttestationAssignments(epoch uint64, assignments map[string]uint64) error
	SaveProposalAssignments(epoch uint64, assignments map[uint64]uint64) error
	SaveSyncCommitteesAssignments(startSlot, endSlot uint64, validators []uint64) error
	SaveAttestations(blocks map[uint64]map[string]*types.Block) error
	SaveProposals(blocks map[uint64]map[string]*types.Block) error
	SaveSyncComitteeDuties(blocks map[uint64]map[string]*types.Block) error
	GetValidatorBalanceHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorBalance, error)
	GetValidatorAttestationHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorAttestation, error)
	GetValidatorSyncDutiesHistoryOrdered(validatorIndex uint64, startEpoch uint64, endEpoch uint64, reverseOrdering bool) ([]*types.ValidatorSyncParticipation, error)
	GetValidatorSyncDutiesHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorSyncParticipation, error)
	GetValidatorMissedAttestationsCount(validators []uint64, firstEpoch uint64, lastEpoch uint64) (map[uint64]*types.ValidatorMissedAttestationsStatistic, error)
	GetValidatorSyncDutiesStatistics(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]*types.ValidatorSyncDutiesStatistic, error)
	GetValidatorEffectiveness(validators []uint64, epoch uint64) ([]*types.ValidatorEffectiveness, error)
	GetValidatorBalanceStatistics(startEpoch, endEpoch uint64) (map[uint64]*types.ValidatorBalanceStatistic, error)
	GetValidatorProposalHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorProposal, error)
	SaveValidatorIncomeDetails(epoch uint64, rewards map[uint64]*entity.IncomeDetailsColumnFamily) error
	GetEpochIncomeHistoryDescending(startEpoch uint64, endEpoch uint64) (*entity.IncomeDetailsColumnFamily, error)
	GetEpochIncomeHistory(epoch uint64) (*entity.IncomeDetailsColumnFamily, error)
	GetValidatorIncomeDetailsHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]map[uint64]*entity.IncomeDetailsColumnFamily, error)
	GetAggregatedValidatorIncomeDetailsHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]*entity.IncomeDetailsColumnFamily, error)
	DeleteEpoch(epoch uint64) error

	GetDataTable() interface{}
	GetMetadataUpdatesTable() interface{}
	GetMetadatTable() interface{}
	SaveBlock(block *entity.BlockData) error
	SaveBlocks(block *entity.BlockData) error
	GetBlockFromBlocksTable(number uint64) (*entity.BlockData, error)
	CheckForGapsInBlocksTable(lookback int) (gapFound bool, start int, end int, err error)
	GetLastBlockInBlocksTable() (int, error)
	CheckForGapsInDataTable(lookback int) error
	GetLastBlockInDataTable() (int, error)
	GetMostRecentBlockFromDataTable() (*entity.BlockIndex, error)
	GetFullBlockDescending(start, limit uint64) ([]*entity.BlockData, error)
	GetFullBlocksDescending(stream chan<- *entity.BlockData, high, low uint64) error
	GetBlocksIndexedMultiple(blockNumbers []uint64, limit uint64) ([]*entity.BlockIndex, error)
	GetBlocksDescending(start, limit uint64) ([]*entity.BlockIndex, error)
	TransformBlock(block *entity.BlockData, cache *freecache.Cache) error
	TransformTx(blk *entity.BlockData, cache *freecache.Cache) error
	TransformItx(blk *entity.BlockData, cache *freecache.Cache) error
	TransformERC20(blk *entity.BlockData, cache *freecache.Cache) error
	TransformERC721(blk *entity.BlockData, cache *freecache.Cache) error
	TransformERC1155(blk *entity.BlockData, cache *freecache.Cache) error
	TransformUncle(block *entity.BlockData, cache *freecache.Cache) error
	TransformWithdrawals(block *entity.BlockData, cache *freecache.Cache) error
	GetEth1TxForAddress(prefix string, limit int64) ([]*entity.BlockData, string, error)
	GetAddressesNamesArMetadata(names *map[string]string, inputMetadata *map[string]*entity.ERC20MetadataFamily) (map[string]string, map[string]*entity.ERC20MetadataFamily, error)
	GetIndexedEth1Transaction(txHash []byte) (*entity.TransactionIndex, error)
	GetAddressTransactionsTableData(address []byte, search string, pageToken string) (*types.DataTableResponse, error)
	GetEth1BlocksForAddress(prefix string, limit int64) ([]*entity.BlockIndex, string, error)
	GetAddressBlocksMinedTableData(address string, search string, pageToken string) (*types.DataTableResponse, error)
	GetEth1UnclesForAddress(prefix string, limit int64) ([]*entity.UncleBlocksIndex, string, error)
	GetAddressUnclesMinedTableData(address string, search string, pageToken string) (*types.DataTableResponse, error)
	GetEth1ItxForAddress(prefix string, limit int64) ([]*entity.InternalTransactionIndex, string, error)
	GetAddressInternalTableData(address []byte, search string, pageToken string) (*types.DataTableResponse, error)
	GetInternalTransfersForTransaction(transaction []byte, from []byte) ([]types.Transfer, error)
	GetArbitraryTokenTransfersForTransaction(transaction []byte) ([]*types.Transfer, error)
	GetEth1ERC20ForAddress(prefix string, limit int64) ([]*entity.ERC20Index, string, error)
	GetAddressErc20TableData(address []byte, search string, pageToken string) (*types.DataTableResponse, error)
	GetEth1ERC721ForAddress(prefix string, limit int64) ([]*entity.ERC721Index, string, error)
	GetAddressErc721TableData(address string, search string, pageToken string) (*types.DataTableResponse, error)
	GetEth1ERC1155ForAddress(prefix string, limit int64) ([]*entity.ERC1155Index, string, error)
	GetAddressErc1155TableData(address string, search string, pageToken string) (*types.DataTableResponse, error)
	GetMetadataUpdates(prefix string, startToken string, limit int) ([]string, []*types.AddressBalance, error)
	GetMetadata(startToken string, limit int) ([]string, []*types.AddressBalance, error)
	GetMetadataForAddress(address []byte) (*types.AddressMetadata, error)
	GetBalanceForAddress(address []byte, token []byte) (*types.AddressBalance, error)
	GetERC20MetadataForAddress(address []byte) (*types.ERC20Metadata, error)
	SaveERC20Metadata(address []byte, metadata *types.ERC20Metadata) error
	GetAddressName(address []byte) (string, error)
	GetAddressNames(addresses map[string]string) error
	SaveAddressName(address []byte, name string) error
	GetContractMetadata(address []byte) (*entity.ContractMetadataFamily, error)
	SaveContractMetadata(address []byte, metadata *entity.ContractMetadataFamily) error
	SaveBalances(balances []*types.AddressBalance, deleteKeys []string) error
	SaveERC20TokenPrices(prices []*types.ERC20TokenPrice) error
	SaveBlockKeys(blockNumber uint64, blockHash []byte, keys string) error
	GetBlockKeys(blockNumber uint64, blockHash []byte) ([]string, error)
	DeleteBlock(blockNumber uint64, blockHash []byte) error
	GetEth1TxForToken(prefix string, limit int64) ([]*entity.ERC20Index, string, error)
	GetTokenTransactionsTableData(token []byte, address []byte, pageToken string) (*types.DataTableResponse, error)
	SearchForAddress(addressPrefix []byte, limit int) ([]*types.AddressSearchItem, error)
	markBalanceUpdate(address []byte, token []byte, cache *freecache.Cache)
	SaveGasNowHistory(slow, standard, rapid, fast *big.Int) error
	GetGasNowHistory(ts, pastTs time.Time) ([]types.GasNowHistory, error)
}

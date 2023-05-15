package interfaces

import (
	"math/big"

	"github.com/Prajjawalk/zond-indexer/entity"
	"github.com/coocood/freecache"
)

type BlockInterface interface {
	SaveBlock(block *entity.BlockData) error
	GetBlockFromBlocksTable(number uint64) (*entity.BlockData, error)
	GetFullBlockDescending(start, limit uint64) ([]*entity.BlockData, error)
	GetFullBlocksDescending(stream chan<- *entity.BlockData, high, low uint64) error
	CalculateMevFromBlock(block *entity.BlockData) *big.Int
	CalculateTxFeesFromBlock(block *entity.BlockData) *big.Int
	TransformBlock(block *entity.BlockData, cache *freecache.Cache) error
	TransformTx(blk *entity.BlockData, cache *freecache.Cache) error
	TransformItx(blk *entity.BlockData, cache *freecache.Cache) error
	TransformERC20(blk *entity.BlockData, cache *freecache.Cache) error
	TransformERC721(blk *entity.BlockData, cache *freecache.Cache) error
	TransformERC1155(blk *entity.BlockData, cache *freecache.Cache) error
	TransformUncle(block *entity.BlockData, cache *freecache.Cache) error
	TransformWithdrawals(block *entity.BlockData, cache *freecache.Cache) error
}

type Eth1Transaction interface {
	CalculateTxFeeFromTransaction(tx *entity.Eth1Transaction, blockBaseFee *big.Int) *big.Int
}

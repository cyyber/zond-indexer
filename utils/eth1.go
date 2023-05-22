package utils

import "math/big"

func Eth1BlockReward(blockNumber uint64, difficulty []byte) *big.Int {

	if len(difficulty) == 0 { // no block rewards for PoS blocks
		return big.NewInt(0)
	}

	if blockNumber < 4370000 {
		return big.NewInt(5e+18)
	} else if blockNumber < 7280000 {
		return big.NewInt(3e+18)
	} else {
		return big.NewInt(2e+18)
	}
}

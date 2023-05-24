package utils

import "github.com/ethereum/go-ethereum/common"

func FixAddressCasing(add string) string {
	return common.HexToAddress(add).Hex()
}

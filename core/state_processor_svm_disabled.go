// +build !sputnikvm

package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

type FeatureDisabledError struct{}

func (e *FeatureDisabledError) Error() string {
	return "'sputnikvm' feature is disabled, please rebuild 'geth' with 'sputnikvm' tag, or use a built-in evm"
}

func ApplySputnikTransaction(config *params.ChainConfig, bc ChainContext, author *common.Address, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, error) {
	return nil, 0, new(FeatureDisabledError)
}

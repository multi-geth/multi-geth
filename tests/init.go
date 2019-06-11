// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"

	xchainparity "github.com/etclabscore/eth-x-chainspec/parity"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/params"
)

var chainspecsDir = filepath.Join(".", "chainspecs")

func mustReadChainspec(specFilename string) *params.ChainConfig {
	chainspecFile := filepath.Join(chainspecsDir, specFilename)
	b, err := ioutil.ReadFile(chainspecFile)
	if err != nil {
		return nil
	}
	return mustReadChainspecBytes(b)
}

func mustReadChainspecBytes(input []byte) *params.ChainConfig {
	pc := xchainparity.Config{}
	err := json.Unmarshal(input, &pc)
	if err != nil {
		panic("unmarshal chainspec")
	}

	mggen := core.ParityConfigToMultiGethGenesis(&pc)
	if mggen == nil {
		panic("nil genesis converted")
	}
	return mggen.Config

}

var (
	FrontierChainspec             = mustReadChainspec("frontier_test.json")
	HomesteadChainspec            = mustReadChainspec("homestead_test.json")
	EIP150Chainspec               = mustReadChainspec("eip150_test.json")
	EIP158Chainspec               = mustReadChainspec("eip161_test.json")
	ByzantiumChainspec            = mustReadChainspec("byzantium_test.json")
	ConstantinopleChainspec       = mustReadChainspec("constantinople_test.json")
	ConstantinopleFixChainspec    = mustReadChainspec("st_peters_test.json")
	EIP158ToByzantiumAt5Chainspec = mustReadChainspec("transition_test.json")
)

// Forks table defines supported forks and their chain config.
var Forks = map[string]*params.ChainConfig{
	// https://github.com/paritytech/parity-ethereum/blob/1871275ecdf02431bf67d09a1b25be8ff8916e3a/ethcore/src/client/evm_test_client.rs#L98
	// https://github.com/paritytech/parity-ethereum/blob/0199acbece836c49e07410796c40c185e9051451/ethcore/src/ethereum/mod.rs#L129
	"Frontier":             FrontierChainspec,
	"Homestead":            HomesteadChainspec,
	"EIP150":               EIP150Chainspec,
	"EIP158":               EIP158Chainspec,
	"Byzantium":            ByzantiumChainspec,
	"Constantinople":       ConstantinopleChainspec,
	"ConstantinopleFix":    ConstantinopleFixChainspec,
	"EIP158ToByzantiumAt5": EIP158ToByzantiumAt5Chainspec,

	// "Frontier": {
	// 	ChainID: big.NewInt(1),
	// },
	// "Homestead": {
	// 	ChainID:        big.NewInt(1),
	// 	HomesteadBlock: big.NewInt(0),
	// },
	// "EIP150": {
	// 	ChainID:        big.NewInt(1),
	// 	HomesteadBlock: big.NewInt(0),
	// 	EIP150Block:    big.NewInt(0),
	// },
	// "EIP158": {
	// 	ChainID:        big.NewInt(1),
	// 	HomesteadBlock: big.NewInt(0),
	// 	EIP150Block:    big.NewInt(0),
	// 	EIP155Block:    big.NewInt(0),
	// 	EIP158Block:    big.NewInt(0),
	// },
	// "Byzantium": {
	// 	ChainID:        big.NewInt(1),
	// 	HomesteadBlock: big.NewInt(0),
	// 	EIP150Block:    big.NewInt(0),
	// 	EIP155Block:    big.NewInt(0),
	// 	EIP158Block:    big.NewInt(0),
	// 	DAOForkBlock:   big.NewInt(0),
	// 	ByzantiumBlock: big.NewInt(0),
	// },
	// "Constantinople": {
	// 	ChainID:             big.NewInt(1),
	// 	HomesteadBlock:      big.NewInt(0),
	// 	EIP150Block:         big.NewInt(0),
	// 	EIP155Block:         big.NewInt(0),
	// 	EIP158Block:         big.NewInt(0),
	// 	DAOForkBlock:        big.NewInt(0),
	// 	ByzantiumBlock:      big.NewInt(0),
	// 	ConstantinopleBlock: big.NewInt(0),
	// 	PetersburgBlock:     big.NewInt(10000000),
	// },
	// "ConstantinopleFix": {
	// 	ChainID:             big.NewInt(1),
	// 	HomesteadBlock:      big.NewInt(0),
	// 	EIP150Block:         big.NewInt(0),
	// 	EIP155Block:         big.NewInt(0),
	// 	EIP158Block:         big.NewInt(0),
	// 	DAOForkBlock:        big.NewInt(0),
	// 	ByzantiumBlock:      big.NewInt(0),
	// 	ConstantinopleBlock: big.NewInt(0),
	// 	PetersburgBlock:     big.NewInt(0),
	// },
	"FrontierToHomesteadAt5": {
		ChainID:        big.NewInt(1),
		HomesteadBlock: big.NewInt(5),
	},
	"HomesteadToEIP150At5": {
		ChainID:        big.NewInt(1),
		HomesteadBlock: big.NewInt(0),
		EIP150Block:    big.NewInt(5),
	},
	"HomesteadToDaoAt5": {
		ChainID:        big.NewInt(1),
		HomesteadBlock: big.NewInt(0),
		DAOForkBlock:   big.NewInt(5),
		DAOForkSupport: true,
	},
	// "EIP158ToByzantiumAt5": {
	// 	ChainID:        big.NewInt(1),
	// 	HomesteadBlock: big.NewInt(0),
	// 	EIP150Block:    big.NewInt(0),
	// 	EIP155Block:    big.NewInt(0),
	// 	EIP158Block:    big.NewInt(0),
	// 	ByzantiumBlock: big.NewInt(5),
	// },
	"ByzantiumToConstantinopleAt5": {
		ChainID:        big.NewInt(1),
		HomesteadBlock: big.NewInt(0),
		EIP150Block:    big.NewInt(0),
		EIP155Block:    big.NewInt(0),
		EIP158Block:    big.NewInt(0),
		ByzantiumBlock: big.NewInt(0),
		BlockRewardSchedule: params.BlockRewardScheduleT{
			new(big.Int).SetUint64(uint64(0)): new(big.Int).SetUint64(uint64(0x29a2241af62c0000)),
			new(big.Int).SetUint64(uint64(5)): new(big.Int).SetUint64(uint64(0x1bc16d674ec80000)),
		},
		DifficultyBombDelays: params.DifficultyBombDelaysT{
			new(big.Int).SetUint64(uint64(0)): new(big.Int).SetUint64(uint64(0x2dc6c0)),
			new(big.Int).SetUint64(uint64(5)): new(big.Int).SetUint64(uint64(0x1e8480)),
		},
	},
}

// UnsupportedForkError is returned when a test requests a fork that isn't implemented.
type UnsupportedForkError struct {
	Name string
}

func (e UnsupportedForkError) Error() string {
	return fmt.Sprintf("unsupported fork %q", e.Name)
}

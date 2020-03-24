// Copyright 2019 The multi-geth Authors
// This file is part of the multi-geth library.
//
// The multi-geth library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The multi-geth library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the multi-geth library. If not, see <http://www.gnu.org/licenses/>.
package ethash

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

func accumulateECIP1017Rewards(config *params.ChainConfig, state *state.StateDB, header *types.Header, uncles []*types.Header) {
	blockReward := FrontierBlockReward

	// Ensure value 'era' is configured.
	eraLen := config.ECIP1017EraBlock
	era := getBlockEra(header.Number, eraLen)
	wr := getBlockWinnerRewardByEra(era, blockReward)                    // wr "winner reward". 5, 4, 3.2, 2.56, ...
	wurs := getBlockWinnerRewardForUnclesByEra(era, uncles, blockReward) // wurs "winner uncle rewards"
	wr.Add(wr, wurs)
	state.AddBalance(header.Coinbase, wr) // $$

	// Reward uncle miners.
	for _, uncle := range uncles {
		ur := getBlockUncleRewardByEra(era, header, uncle, blockReward)
		state.AddBalance(uncle.Coinbase, ur) // $$
	}
}

func ecip1010Explosion(config *params.ChainConfig, next *big.Int, exPeriodRef *big.Int) {
	// https://github.com/ethereumproject/ECIPs/blob/master/ECIPs/ECIP-1010.md

	explosionBlock := new(big.Int).Add(config.ECIP1010PauseBlock, config.ECIP1010Length)
	if next.Cmp(explosionBlock) < 0 {
		exPeriodRef.Set(config.ECIP1010PauseBlock)
	} else {
		exPeriodRef.Sub(exPeriodRef, config.ECIP1010Length)
	}
}

// GetBlockEra gets which "Era" a given block is within, given an era length (ecip-1017 has era=5,000,000 blocks)
// Returns a zero-index era number, so "Era 1": 0, "Era 2": 1, "Era 3": 2 ...
func getBlockEra(blockNum, eraLength *big.Int) *big.Int {
	// If genesis block or impossible negative-numbered block, return zero-val.
	if blockNum.Sign() < 1 {
		return new(big.Int)
	}

	remainder := big.NewInt(0).Mod(big.NewInt(0).Sub(blockNum, big.NewInt(1)), eraLength)
	base := big.NewInt(0).Sub(blockNum, remainder)

	d := big.NewInt(0).Div(base, eraLength)
	dremainder := big.NewInt(0).Mod(d, big.NewInt(1))

	return new(big.Int).Sub(d, dremainder)
}

// As of "Era 2" (zero-index era 1), uncle miners and winners are rewarded equally for each included block.
// So they share this function.
func getEraUncleBlockReward(era *big.Int, blockReward *big.Int) *big.Int {
	return new(big.Int).Div(getBlockWinnerRewardByEra(era, blockReward), big32)
}

// GetBlockUncleRewardByEra gets called _for each uncle miner_ associated with a winner block's uncles.
func getBlockUncleRewardByEra(era *big.Int, header, uncle *types.Header, blockReward *big.Int) *big.Int {
	// Era 1 (index 0):
	//   An extra reward to the winning miner for including uncles as part of the block, in the form of an extra 1/32 (0.15625ETC) per uncle included, up to a maximum of two (2) uncles.
	if era.Cmp(big.NewInt(0)) == 0 {
		r := new(big.Int)
		r.Add(uncle.Number, big8) // 2,534,998 + 8              = 2,535,006
		r.Sub(r, header.Number)   // 2,535,006 - 2,534,999        = 7
		r.Mul(r, blockReward)     // 7 * 5e+18               = 35e+18
		r.Div(r, big8)            // 35e+18 / 8                            = 7/8 * 5e+18

		return r
	}
	return getEraUncleBlockReward(era, blockReward)
}

// GetBlockWinnerRewardForUnclesByEra gets called _per winner_, and accumulates rewards for each included uncle.
// Assumes uncles have been validated and limited (@ func (v *BlockValidator) VerifyUncles).
func getBlockWinnerRewardForUnclesByEra(era *big.Int, uncles []*types.Header, blockReward *big.Int) *big.Int {
	r := big.NewInt(0)

	for range uncles {
		r.Add(r, getEraUncleBlockReward(era, blockReward)) // can reuse this, since 1/32 for winner's uncles remain unchanged from "Era 1"
	}
	return r
}

// GetRewardByEra gets a block reward at disinflation rate.
// Constants MaxBlockReward, DisinflationRateQuotient, and DisinflationRateDivisor assumed.
func getBlockWinnerRewardByEra(era *big.Int, blockReward *big.Int) *big.Int {
	if era.Cmp(big.NewInt(0)) == 0 {
		return new(big.Int).Set(blockReward)
	}

	// MaxBlockReward _r_ * (4/5)**era == MaxBlockReward * (4**era) / (5**era)
	// since (q/d)**n == q**n / d**n
	// qed
	var q, d, r *big.Int = new(big.Int), new(big.Int), new(big.Int)

	q.Exp(params.DisinflationRateQuotient, era, nil)
	d.Exp(params.DisinflationRateDivisor, era, nil)

	r.Mul(blockReward, q)
	r.Div(r, d)

	return r
}

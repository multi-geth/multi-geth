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

func accumulateEcip1017FRewards(config *params.ChainConfig, state *state.StateDB, header *types.Header, uncles []*types.Header) {
	blockReward := FrontierBlockReward

	// Ensure value 'era' is configured.
	eraLen := config.ECIP1017EraRounds
	era := GetBlockEra(header.Number, eraLen)
	wr := GetBlockWinnerRewardByEra(era, blockReward)                    // wr "winner reward". 5, 4, 3.2, 2.56, ...
	wurs := GetBlockWinnerRewardForUnclesByEra(era, uncles, blockReward) // wurs "winner uncle rewards"
	wr.Add(wr, wurs)
	state.AddBalance(header.Coinbase, wr) // $$

	// Reward uncle miners.
	for _, uncle := range uncles {
		ur := GetBlockUncleRewardByEra(era, header, uncle, blockReward)
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
func blockEra(blockNum, eraLength *big.Int) *big.Int {
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

func calcDifficultyClassic(config *params.ChainConfig, time uint64, parent *types.Header) *big.Int {
	next := new(big.Int).Add(parent.Number, big1)
	out := new(big.Int)

	// ADJUSTMENT algorithms
	if config.IsByan(next) {
		// https://github.com/ethereum/EIPs/issues/100
		// algorithm:
		// diff = (parent_diff +
		//         (parent_diff / 2048 * max((2 if len(parent.uncles) else 1) - ((timestamp - parent.timestamp) // 9), -99))
		//        ) + 2^(periodCount - 2)
		out.Div(parent_time_delta(time, parent), big9)

		if parent.UncleHash == types.EmptyUncleHash {
			out.Sub(big1, out)
		} else {
			out.Sub(big2, out)
		}
		out.Set(math.BigMax(out, bigMinus99))
		out.Mul(parent_diff_over_dbd(parent), out)
		out.Add(out, parent.Difficulty)

	} else if config.IsEIP2F(next) {
		// https://github.com/ethereum/EIPs/blob/master/EIPS/eip-2.md
		// algorithm:
		// diff = (parent_diff +
		//         (parent_diff / 2048 * max(1 - (block_timestamp - parent_timestamp) // 10, -99))
		//        )
		out.Div(parent_time_delta(time, parent), big10)
		out.Sub(big1, out)
		out.Set(math.BigMax(out, bigMinus99))
		out.Mul(parent_diff_over_dbd(parent), out)
		out.Add(out, parent.Difficulty)

	} else {
		// FRONTIER
		// algorithm:
		// diff =
		//   if parent_block_time_delta < params.DurationLimit
		//      parent_diff + (parent_diff // 2048)
		//   else
		//      parent_diff - (parent_diff // 2048)
		out.Set(parent.Difficulty)
		if parent_time_delta(time, parent).Cmp(params.DurationLimit) < 0 {
			out.Add(out, parent_diff_over_dbd(parent))
		} else {
			out.Sub(out, parent_diff_over_dbd(parent))
		}
	}

	// after adjustment and before bomb
	out.Set(math.BigMax(out, params.MinimumDifficulty))

	// EXPLOSION delays

	// exPeriodRef the explosion clause's reference point
	exPeriodRef := new(big.Int).Add(parent.Number, big1)

	if config.IsBombDisposal(next) {
		return out

	} else if config.IsEIP1234F(next) {
		// calcDifficultyEIP1234 is the difficulty adjustment algorithm for Constantinople.
		// The calculation uses the Byzantium rules, but with bomb offset 5M.
		// Specification EIP-1234: https://eips.ethereum.org/EIPS/eip-1234
		// Note, the calculations below looks at the parent number, which is 1 below
		// the block number. Thus we remove one from the delay given

		// calculate a fake block number for the ice-age delay
		// Specification: https://eips.ethereum.org/EIPS/eip-1234
		fakeBlockNumber := new(big.Int)
		if parent.Number.Cmp(big.NewInt(4999999)) >= 0 {
			fakeBlockNumber = fakeBlockNumber.Sub(parent.Number, big.NewInt(4999999))
		}
		exPeriodRef.Set(fakeBlockNumber)

	} else if config.IsEIP649F(next) {
		// The calculation uses the Byzantium rules, with bomb offset of 3M.
		// Specification EIP-649: https://eips.ethereum.org/EIPS/eip-649
		// Related meta-ish EIP-669: https://github.com/ethereum/EIPs/pull/669
		// Note, the calculations below looks at the parent number, which is 1 below
		// the block number. Thus we remove one from the delay given

		fakeBlockNumber := new(big.Int)
		if parent.Number.Cmp(big.NewInt(2999999)) >= 0 {
			fakeBlockNumber = fakeBlockNumber.Sub(parent.Number, big.NewInt(2999999))
		}
		exPeriodRef.Set(fakeBlockNumber)

	} else if config.IsECIP1010(next) {
		ecip1010Explosion(config, next, exPeriodRef)
	}

	// EXPLOSION

	// the 'periodRef' (from above) represents the many ways of hackishly modifying the reference number
	// (ie the 'currentBlock') in order to lie to the function about what time it really is
	//
	//   2^(( periodRef // EDP) - 2)
	//
	x := new(big.Int)
	x.Div(exPeriodRef, params.ExpDiffPeriod) // (periodRef // EDP)
	if x.Cmp(big1) > 0 {                     // if result large enough (not in algo explicitly)
		x.Sub(x, big2)      // - 2
		x.Exp(big2, x, nil) // 2^
	} else {
		x.SetUint64(0)
	}
	out.Add(out, x)
	return out
}

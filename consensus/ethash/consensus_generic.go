// Copyright 2017 The go-ethereum Authors
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

package ethash

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

// parent_time_delta is a convenience fn for CalcDifficulty
func parentTimeDelta(t uint64, p *types.Header) *big.Int {
	return new(big.Int).Sub(new(big.Int).SetUint64(t), new(big.Int).SetUint64(p.Time))
}

// parent_diff_over_dbd is a  convenience fn for CalcDifficulty
func parentDiffOverDbd(p *types.Header) *big.Int {
	return new(big.Int).Div(p.Difficulty, params.DifficultyBoundDivisor)
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func calcDifficultyGeneric(config *params.ChainConfig, time uint64, parent *types.Header) *big.Int {
	next := new(big.Int).Add(parent.Number, big1)
	out := new(big.Int)

	// ADJUSTMENT algorithms
	if config.IsByzantium(next) {
		// https://github.com/ethereum/EIPs/issues/100
		// algorithm:
		// diff = (parent_diff +
		//         (parent_diff / 2048 * max((2 if len(parent.uncles) else 1) - ((timestamp - parent.timestamp) // 9), -99))
		//        ) + 2^(periodCount - 2)
		out.Div(parentTimeDelta(time, parent), big9)

		if parent.UncleHash == types.EmptyUncleHash {
			out.Sub(big1, out)
		} else {
			out.Sub(big2, out)
		}
		out.Set(math.BigMax(out, bigMinus99))
		out.Mul(parentDiffOverDbd(parent), out)
		out.Add(out, parent.Difficulty)

	} else if config.IsHomestead(next) {
		// https://github.com/ethereum/EIPs/blob/master/EIPS/eip-2.md
		// algorithm:
		// diff = (parent_diff +
		//         (parent_diff / 2048 * max(1 - (block_timestamp - parent_timestamp) // 10, -99))
		//        )
		out.Div(parentTimeDelta(time, parent), big10)
		out.Sub(big1, out)
		out.Set(math.BigMax(out, bigMinus99))
		out.Mul(parentDiffOverDbd(parent), out)
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
		if parentTimeDelta(time, parent).Cmp(params.DurationLimit) < 0 {
			out.Add(out, parentDiffOverDbd(parent))
		} else {
			out.Sub(out, parentDiffOverDbd(parent))
		}
	}

	// after adjustment and before bomb
	out.Set(math.BigMax(out, params.MinimumDifficulty))

	// EXPLOSION delays

	// exPeriodRef the explosion clause's reference point
	exPeriodRef := new(big.Int).Add(parent.Number, big1)

	if config.IsBombDisposal(next) {
		return out
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

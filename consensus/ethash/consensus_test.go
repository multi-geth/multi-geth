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
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

type diffTest struct {
	ParentTimestamp    uint64
	ParentDifficulty   *big.Int
	CurrentTimestamp   uint64
	CurrentBlocknumber *big.Int
	CurrentDifficulty  *big.Int
}

func (d *diffTest) UnmarshalJSON(b []byte) (err error) {
	var ext struct {
		ParentTimestamp    string
		ParentDifficulty   string
		CurrentTimestamp   string
		CurrentBlocknumber string
		CurrentDifficulty  string
	}
	if err := json.Unmarshal(b, &ext); err != nil {
		return err
	}

	d.ParentTimestamp = math.MustParseUint64(ext.ParentTimestamp)
	d.ParentDifficulty = math.MustParseBig256(ext.ParentDifficulty)
	d.CurrentTimestamp = math.MustParseUint64(ext.CurrentTimestamp)
	d.CurrentBlocknumber = math.MustParseBig256(ext.CurrentBlocknumber)
	d.CurrentDifficulty = math.MustParseBig256(ext.CurrentDifficulty)

	return nil
}

func TestCalcDifficulty(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "tests", "testdata", "BasicTests", "difficulty.json"))
	if err != nil {
		t.Skip(err)
	}
	defer file.Close()

	tests := make(map[string]diffTest)
	err = json.NewDecoder(file).Decode(&tests)
	if err != nil {
		t.Fatal(err)
	}

	config := &params.ChainConfig{HomesteadBlock: big.NewInt(1150000)}

	for name, test := range tests {
		number := new(big.Int).Sub(test.CurrentBlocknumber, big.NewInt(1))
		diff := CalcDifficulty(config, test.CurrentTimestamp, &types.Header{
			Number:     number,
			Time:       test.ParentTimestamp,
			Difficulty: test.ParentDifficulty,
		})
		if diff.Cmp(test.CurrentDifficulty) != 0 {
			t.Error(name, "failed. Expected", test.CurrentDifficulty, "and calculated", diff)
		}
	}
}

func TestCalcDifficultyDifficultyDelayConfigVSForkFeatureConfig(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "tests", "testdata", "BasicTests", "difficulty.json"))
	if err != nil {
		t.Skip(err)
	}
	defer file.Close()

	tests := make(map[string]diffTest)
	err = json.NewDecoder(file).Decode(&tests)
	if err != nil {
		t.Fatal(err)
	}

	mainA := &params.ChainConfig{}
	mainB := &params.ChainConfig{}

	*mainA = *params.MainnetChainConfig
	*mainB = *params.MainnetChainConfig

	testN := big.NewInt(10000000)
	mainA.DifficultyBombDelays = nil
	if !mainA.IsEIP1234F(testN) || !mainA.IsEIP649F(testN) {
		t.Fatal("test requires reference config to use fork features to compare to difficulty delay map")
	}
	if len(mainA.DifficultyBombDelays) != 0 {
		t.Fatal("test requires reference config to use empty difficulty delay map")
	}

	mainB.ByzantiumBlock = nil
	mainB.ConstantinopleBlock = nil
	mainB.EIP100FBlock = mainA.ByzantiumBlock // Needs for EIP100 adjustment
	if mainB.IsEIP1234F(testN) || mainB.IsEIP649F(testN) {
		t.Fatal("test requires compared config to use no fork features configure difficulty delay (map only)")
	}
	if len(mainB.DifficultyBombDelays) == 0 {
		t.Fatal("test requires compared config to use existing difficulty delay map")
	}

	for name, test := range tests {
		number := new(big.Int).Sub(test.CurrentBlocknumber, big.NewInt(1))
		diffA := CalcDifficulty(mainA, test.CurrentTimestamp, &types.Header{
			Number:     number,
			Time:       test.ParentTimestamp,
			Difficulty: test.ParentDifficulty,
		})
		diffB := CalcDifficulty(mainB, test.CurrentTimestamp, &types.Header{
			Number:     number,
			Time:       test.ParentTimestamp,
			Difficulty: test.ParentDifficulty,
		})

		if diffA.Cmp(diffB) != 0 {
			t.Errorf("%s: want: %v, got: %v", name, diffA, diffB)
		}
	}

	for name, test := range tests {
		diffA := CalcDifficulty(mainA, test.CurrentTimestamp, &types.Header{
			Number:     testN,
			Time:       test.ParentTimestamp,
			Difficulty: test.ParentDifficulty,
		})
		diffB := CalcDifficulty(mainB, test.CurrentTimestamp, &types.Header{
			Number:     testN,
			Time:       test.ParentTimestamp,
			Difficulty: test.ParentDifficulty,
		})

		if diffA.Cmp(diffB) != 0 {
			t.Errorf("%s: want: %v w/ %v, got: %v w/ %v", name, diffA, mainA, diffB, mainB.DifficultyBombDelays)
		}
	}

}

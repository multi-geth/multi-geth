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
package params

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

var (
	// Genesis hashes to enforce below configs on.
	EllaismGenesisHash = common.HexToHash("0x4d7df65052bb21264d6ad2d6fe2d5578a36be12f71bf8d0559b0c15c4dc539b5")

	// EllaismChainConfig is the chain parameters to run a node on the Ellaism main network.
	EllaismChainConfig = &ChainConfig{
		ChainID:             big.NewInt(64),
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        nil,
		DAOForkSupport:      false,
		EIP150Block:         big.NewInt(0),
		EIP150Hash:          common.HexToHash("0x4d7df65052bb21264d6ad2d6fe2d5578a36be12f71bf8d0559b0c15c4dc539b5"),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         nil,
		ByzantiumBlock:      big.NewInt(2000000),
		DisposalBlock:       big.NewInt(0),
		ConstantinopleBlock: nil,
		PetersburgBlock:     nil,
		IstanbulBlock:       nil,
		ECIP1017EraBlock:    big.NewInt(10000000),
		EIP160Block:         big.NewInt(0),
		EIP161DisableBlock:  big.NewInt(2000000),
		Ethash:              new(EthashConfig),
	}
)

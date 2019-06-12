package core

import (
	"fmt"
	"math/big"
	"strings"

	xchain "github.com/etclabscore/eth-x-chainspec"
	xchainparity "github.com/etclabscore/eth-x-chainspec/parity"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

func bigMax(a, b *big.Int) *big.Int {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}

	if a.Cmp(b) > 0 {
		return a
	}
	return b
}

// ParityConfigFromMultiGethGenesis creates an xchain parity config from core.Genesis value.
func ParityConfigFromMultiGethGenesis(name string, c *xchainparity.Config, mgg *Genesis) error {
	if c == nil {
		c = &xchainparity.Config{}
	}
	if c.Genesis == nil {
		c.Genesis = &xchainparity.ConfigGenesis{}
	}
	if c.Params == nil {
		c.Params = &xchainparity.ConfigParams{}
	}

	c.Name = name
	if mgg.Config.NetworkID != 0 {
		c.Params.NetworkID = xchain.FromUint64(mgg.Config.NetworkID)
	}

	c.Genesis.Seal.Ethereum.Nonce = xchain.BlockNonce(types.EncodeNonce(mgg.Nonce))
	c.Genesis.Seal.Ethereum.MixHash = mgg.Mixhash
	c.Genesis.Timestamp = xchain.FromUint64(mgg.Timestamp)
	c.Genesis.GasLimit = xchain.FromUint64(mgg.GasLimit)
	c.Genesis.GasUsed = xchain.FromUint64(mgg.GasUsed)
	c.Genesis.Difficulty = xchain.FromUint64(mgg.Difficulty.Uint64())
	c.Genesis.Author = &mgg.Coinbase
	c.Genesis.ParentHash = &mgg.ParentHash
	c.Genesis.ExtraData = mgg.ExtraData

	ParityConfigWithPrecompiledContractsFromMultiGeth(c, mgg)

	for a, v := range mgg.Alloc {
		n := xchain.ConfigAccountNonce(v.Nonce)
		// If the account belongs to a precompiled contract then don't overwrite the
		// existing val.
		if pre, ok := c.Accounts[a.Hex()]; ok {
			pre.Nonce = &n
			pre.Balance = v.Balance.String()
			pre.Code = v.Code
			pre.Storage = v.Storage
			c.Accounts[a.Hex()] = pre
		} else {
			pv := xchainparity.ConfigAccountValue{
				Nonce:   &n,
				Balance: v.Balance.String(),
				Code:    v.Code,
				Storage: v.Storage,
			}
			c.Accounts[a.Hex()] = pv
		}
	}

	zero := xchain.Uint64(0)
	c.Params.AccountStartNonce = &zero

	c.Params.ChainID = xchain.FromUint64(mgg.Config.ChainID.Uint64())
	setParityDAOConfigFromMultiGeth(c.Params, mgg.Config)
	if mgg.Config.EIP150Block != nil {
		c.Params.EIP150Transition = xchain.FromUint64(mgg.Config.EIP150Block.Uint64())
	}
	if mgg.Config.EIP155Block != nil {
		c.Params.EIP155Transition = xchain.FromUint64(mgg.Config.EIP155Block.Uint64())
	}
	if mgg.Config.EIP160FBlock != nil || mgg.Config.EIP158Block != nil {
		b := new(big.Int).Set(bigMax(mgg.Config.EIP158Block, mgg.Config.EIP160FBlock))
		c.Params.EIP160Transition = xchain.FromUint64(b.Uint64())
	}
	if mgg.Config.EIP161FBlock != nil || mgg.Config.EIP158Block != nil {
		b := new(big.Int).Set(bigMax(mgg.Config.EIP158Block, mgg.Config.EIP161FBlock))
		c.Params.EIP161abcTransition = xchain.FromUint64(b.Uint64())
		c.Params.EIP161dTransition = xchain.FromUint64(b.Uint64())
	}
	if mgg.Config.EIP170FBlock != nil || mgg.Config.EIP158Block != nil {
		c.Params.MaxCodeSize = xchain.FromUint64(uint64(params.MaxCodeSize))
		b := new(big.Int).Set(bigMax(mgg.Config.EIP158Block, mgg.Config.EIP170FBlock))
		c.Params.MaxCodeSizeTransition = xchain.FromUint64(b.Uint64())
	}
	if mgg.Config.EIP140FBlock != nil || mgg.Config.ByzantiumBlock != nil {
		b := new(big.Int).Set(bigMax(mgg.Config.ByzantiumBlock, mgg.Config.EIP140FBlock))
		c.Params.EIP140Transition = xchain.FromUint64(b.Uint64())
	}
	if mgg.Config.EIP211FBlock != nil || mgg.Config.ByzantiumBlock != nil {
		b := new(big.Int).Set(bigMax(mgg.Config.ByzantiumBlock, mgg.Config.EIP211FBlock))
		c.Params.EIP211Transition = xchain.FromUint64(b.Uint64())
	}
	if mgg.Config.EIP214FBlock != nil || mgg.Config.ByzantiumBlock != nil {
		b := new(big.Int).Set(bigMax(mgg.Config.ByzantiumBlock, mgg.Config.EIP214FBlock))
		c.Params.EIP214Transition = xchain.FromUint64(b.Uint64())
	}
	if mgg.Config.EIP658FBlock != nil || mgg.Config.ByzantiumBlock != nil {
		b := new(big.Int).Set(bigMax(mgg.Config.ByzantiumBlock, mgg.Config.EIP658FBlock))
		c.Params.EIP658Transition = xchain.FromUint64(b.Uint64())
	}
	if mgg.Config.EIP145FBlock != nil || mgg.Config.ConstantinopleBlock != nil {
		b := new(big.Int).Set(bigMax(mgg.Config.ConstantinopleBlock, mgg.Config.EIP145FBlock))
		c.Params.EIP145Transition = xchain.FromUint64(b.Uint64())
	}
	if mgg.Config.EIP1014FBlock != nil || mgg.Config.ConstantinopleBlock != nil {
		b := new(big.Int).Set(bigMax(mgg.Config.ConstantinopleBlock, mgg.Config.EIP1014FBlock))
		c.Params.EIP1014Transition = xchain.FromUint64(b.Uint64())
	}
	if mgg.Config.EIP1052FBlock != nil || mgg.Config.ConstantinopleBlock != nil {
		b := new(big.Int).Set(bigMax(mgg.Config.ConstantinopleBlock, mgg.Config.EIP1052FBlock))
		c.Params.EIP1052Transition = xchain.FromUint64(b.Uint64())
	}
	if mgg.Config.EIP1283FBlock != nil || mgg.Config.ConstantinopleBlock != nil {
		b := new(big.Int).Set(bigMax(mgg.Config.ConstantinopleBlock, mgg.Config.EIP1283FBlock))
		c.Params.EIP1283Transition = xchain.FromUint64(b.Uint64())
	}
	if mgg.Config.PetersburgBlock != nil {
		c.Params.EIP1283DisableTransition = xchain.FromUint64(mgg.Config.PetersburgBlock.Uint64())
	}
	if mgg.Config.EWASMBlock != nil {
		c.Params.WASMActivationTransition = xchain.FromUint64(mgg.Config.EWASMBlock.Uint64())
	}

	if mgg.Config.Ethash != nil {
		if c.EngineOpt.ParityConfigEngineEthash == nil {
			c.EngineOpt.ParityConfigEngineEthash = &xchainparity.ConfigEngineEthash{}
		}
		if mgg.Config.HomesteadBlock != nil {
			c.EngineOpt.ParityConfigEngineEthash.Params.HomesteadTransition = xchain.FromUint64(mgg.Config.HomesteadBlock.Uint64())
		}
		if mgg.Config.EIP100FBlock != nil || mgg.Config.ByzantiumBlock != nil {
			b := new(big.Int).Set(bigMax(mgg.Config.EIP100FBlock, mgg.Config.ByzantiumBlock))
			c.EngineOpt.ParityConfigEngineEthash.Params.EIP100BTransition = xchain.FromUint64(b.Uint64())
		}
		if mgg.Config.DisposalBlock != nil {
			c.EngineOpt.ParityConfigEngineEthash.Params.BombDefuseTransition = xchain.FromUint64(mgg.Config.DisposalBlock.Uint64())
		}
		if mgg.Config.ECIP1010PauseBlock != nil {
			c.EngineOpt.ParityConfigEngineEthash.Params.Ecip1010PauseTransition = xchain.FromUint64(mgg.Config.ECIP1010PauseBlock.Uint64())
			// assume if pause is set, so is continue
			if mgg.Config.ECIP1010Length != nil {
				c.EngineOpt.ParityConfigEngineEthash.Params.Ecip1010ContinueTransition = xchain.FromUint64(new(big.Int).Add(mgg.Config.ECIP1010PauseBlock, mgg.Config.ECIP1010Length).Uint64())
			}
		}
		if mgg.Config.ECIP1017EraRounds != nil {
			c.EngineOpt.ParityConfigEngineEthash.Params.Ecip1017EraRounds = xchain.FromUint64(mgg.Config.ECIP1017EraRounds.Uint64())
		}

		for k, v := range mgg.Config.DifficultyBombDelays {
			c.EngineOpt.ParityConfigEngineEthash.Params.DifficultyBombDelays[xchain.Uint64(k.Uint64())] = xchain.FromUint64(v.Uint64())
		}
		for k, v := range mgg.Config.BlockRewardSchedule {
			b := hexutil.Big(*v)
			c.EngineOpt.ParityConfigEngineEthash.Params.BlockReward[xchain.Uint64(k.Uint64())] = &b
		}

	} else if mgg.Config.Clique != nil {
		if c.EngineOpt.ParityConfigEngineClique == nil {
			c.EngineOpt.ParityConfigEngineClique = &xchainparity.ConfigEngineClique{}
		}
		c.EngineOpt.ParityConfigEngineClique.Params.Period = mgg.Config.Clique.Period
		c.EngineOpt.ParityConfigEngineClique.Params.Epoch = mgg.Config.Clique.Epoch
	}

	return nil
}

func ParityConfigBuiltinContracts(c *xchainparity.Config) (builtins []xchainparity.ConfigAccountValueBuiltin) {
	for _, v := range c.Accounts {
		if v.Builtin != nil {
			b := v.Builtin
			builtins = append(builtins, *b)
		}
	}
	return
}

// ToMultiGethGenesis converts a Parity chainspec to the corresponding MultiGeth datastructure.
// Note that the return value 'core.Genesis' includes the respective 'params.ChainConfig' values.
func ParityConfigToMultiGethGenesis(c *xchainparity.Config) *Genesis {
	mgc := &params.ChainConfig{}
	if pars := c.Params; pars != nil {
		if err := checkUnsupportedValsMust(pars); err != nil {
			panic(err)
		}

		mgc.NetworkID = pars.NetworkID.Uint64()
		mgc.ChainID = pars.ChainID.Big()

		// Defaults according to Parity documentation https://wiki.parity.io/Chain-specification.html
		if mgc.ChainID == nil && pars.NetworkID != nil {
			mgc.ChainID = pars.NetworkID.Big()
		}

		// DAO
		setMultiGethDAOConfigsFromParity(mgc, pars)

		// Tangerine
		mgc.EIP150Block = pars.EIP150Transition.Big()
		// mgc.EIP150Hash // optional@mg

		// Spurious
		mgc.EIP155Block = pars.EIP155Transition.Big()
		mgc.EIP160FBlock = pars.EIP160Transition.Big()
		mgc.EIP161FBlock = pars.EIP161abcTransition.Big() // and/or d
		mgc.EIP170FBlock = pars.MaxCodeSizeTransition.Big()
		if mgc.EIP170FBlock != nil && uint64(*pars.MaxCodeSize) != uint64(24576) {
			panic(fmt.Sprintf("%v != %v - unsupported configuration value", *pars.MaxCodeSize, 24576))
		}

		// Byzantium
		// 100
		mgc.EIP140FBlock = pars.EIP140Transition.Big()
		// 198
		mgc.EIP211FBlock = pars.EIP211Transition.Big() // FIXME this might actually be for EIP210. :-$
		// 212
		// 213
		mgc.EIP214FBlock = pars.EIP214Transition.Big()
		// 649 - metro diff bomb, block reward
		mgc.EIP658FBlock = pars.EIP658Transition.Big()

		parityBuiltins := ParityConfigBuiltinContracts(c)
		for _, pc := range parityBuiltins {
			if pc.ActivateAt != nil {
				switch *pc.Name {
				case "modexp":
					mgc.EIP198FBlock = new(big.Int).Set(pc.ActivateAt.Big())
				case "alt_bn128_pairing":
					mgc.EIP212FBlock = new(big.Int).Set(pc.ActivateAt.Big())
				case "alt_bn128_add", "alt_bn128_mul":
					mgc.EIP213FBlock = new(big.Int).Set(pc.ActivateAt.Big())
				case "ripemd160", "ecrecover", "sha256", "identity":
				default:
					panic("unsupported builtin contract: " + *pc.Name)
				}
			}
		}

		// Constantinople
		mgc.EIP145FBlock = pars.EIP145Transition.Big()
		mgc.EIP1014FBlock = pars.EIP1014Transition.Big()
		mgc.EIP1052FBlock = pars.EIP1052Transition.Big()
		mgc.EIP1283FBlock = pars.EIP1283Transition.Big()
		mgc.PetersburgBlock = pars.EIP1283DisableTransition.Big()

		mgc.EWASMBlock = pars.WASMActivationTransition.Big()
	}

	if ethc := c.EngineOpt.ParityConfigEngineEthash; ethc != nil {

		pars := ethc.Params

		mgc.Ethash = &params.EthashConfig{}

		mgc.HomesteadBlock = pars.HomesteadTransition.Big()
		mgc.EIP100FBlock = pars.EIP100BTransition.Big()
		mgc.DisposalBlock = pars.BombDefuseTransition.Big()
		mgc.ECIP1010PauseBlock = pars.Ecip1010PauseTransition.Big()
		mgc.ECIP1010Length = func() *big.Int {
			if pars.Ecip1010ContinueTransition != nil {
				return new(big.Int).Sub(pars.Ecip1010ContinueTransition.Big(), pars.Ecip1010PauseTransition.Big())
			} else if pars.Ecip1010PauseTransition == nil && pars.Ecip1010ContinueTransition == nil {
				return nil
			}
			return big.NewInt(0)
		}()
		mgc.ECIP1017EraRounds = pars.Ecip1017EraRounds.Big()

		mgc.DifficultyBombDelays = params.DifficultyBombDelaysT{}
		for k, v := range pars.DifficultyBombDelays {
			mgc.DifficultyBombDelays[new(big.Int).SetUint64(k.Uint64())] = new(big.Int).SetUint64(v.Uint64())
		}
		mgc.BlockRewardSchedule = params.BlockRewardScheduleT{}
		for k, v := range pars.BlockReward {
			mgc.BlockRewardSchedule[new(big.Int).SetUint64(k.Uint64())] = new(big.Int).Set(v.ToInt())
		}

	} else if ethc := c.EngineOpt.ParityConfigEngineClique; ethc != nil {

		pars := ethc.Params

		mgc.Clique = &params.CliqueConfig{
			Period: pars.Period,
			Epoch:  pars.Epoch,
		}

	} else {
		return nil
	}
	mgg := &Genesis{
		Config: mgc,
	}
	if c.Genesis != nil {
		seal := c.Genesis.Seal.Ethereum

		mgg.Nonce = seal.Nonce.Uint64()
		mgg.Mixhash = seal.MixHash
		mgg.Timestamp = c.Genesis.Timestamp.Uint64()
		mgg.GasLimit = c.Genesis.GasLimit.Uint64()
		mgg.GasUsed = c.Genesis.GasUsed.Uint64()
		mgg.Difficulty = c.Genesis.Difficulty.Big()
		mgg.Coinbase = *c.Genesis.Author
		mgg.ParentHash = *c.Genesis.ParentHash
		mgg.ExtraData = c.Genesis.ExtraData
	}
	if c.Accounts != nil {
		mgg.Alloc = GenesisAlloc{}

	accountsloop:
		for k, v := range c.Accounts {
			bal, ok := xchain.ParseBig256(v.Balance)
			if !ok {
				panic("error setting genesis account balance")
			}
			var nonce uint64
			if v.Nonce != nil {
				nonce = uint64(*v.Nonce)
			}

			addr := common.HexToAddress(strings.ToLower(k))
			for _, i := range []byte{1, 2, 3, 4, 5, 6, 7, 8} {
				if common.BytesToAddress([]byte{i}) == common.HexToAddress(strings.ToLower(k)) {
					if bal.Sign() < 1 {
						continue accountsloop
					}

				}
			}

			// if _, ok := vm.PrecompiledContractsForConfig(params.AllEthashProtocolChanges, big.NewInt(0))[addr]; ok && bal.Sign() < 1 {
			// 	continue accountsloop
			// }

			mgg.Alloc[addr] = GenesisAccount{
				Nonce:   nonce,
				Balance: bal,
				Code:    v.Code,
				Storage: v.Storage,
			}
		}
	}
	return mgg
}

func checkUnsupportedValsMust(pars *xchainparity.ConfigParams) error {
	// FIXME
	if pars.EIP161abcTransition.Uint64() != pars.EIP161dTransition.Uint64() {
		panic("not supported")
	}
	// TODO...
	// unsupportedValuesMust := map[interface{}]interface{}{
	// 	pars.AccountStartNonce:                       uint64(0),
	// 	pars.MaximumExtraDataSize:                    uint64(32),
	// 	pars.MinGasLimit:                             uint64(5000),
	// 	pars.SubProtocolName:                         "",
	// 	pars.ValidateChainIDTransition:               nil,
	// 	pars.ValidateChainReceiptsTransition:         nil,
	// 	pars.DustProtectionTransition:                nil,
	// 	pars.NonceCapIncrement:                       nil,
	// 	pars.RemoveDustContracts:                     false,
	// 	pars.EIP210Transition:                        nil,
	// 	pars.EIP210ContractAddress:                   nil,
	// 	pars.EIP210ContractCode:                      nil,
	// 	pars.ApplyReward:                             false,
	// 	pars.TransactionPermissionContract:           nil,
	// 	pars.TransactionPermissionContractTransition: nil,
	// 	pars.KIP4Transition:                          nil,
	// 	pars.KIP6Transition:                          nil,
	// }
	// i := -1
	// for k, v := range unsupportedValuesMust {
	// 	i++
	// 	if v == nil && k == nil {
	// 		continue
	// 	}
	// 	if v != nil && !reflect.DeepEqual(k, v) {
	// 		panic(fmt.Sprintf("%d: %v != %v - unsupported configuration value", i, k, v))
	// 	}
	// }
	return nil
}

// NOTE this should NEVER be needed. The chains with DAO settings are already canonical and have existing chainspecs.
// There is no need to replicate this information.
func setParityDAOConfigFromMultiGeth(pars *xchainparity.ConfigParams, mgc *params.ChainConfig) {
	// noop
}

func setMultiGethDAOConfigsFromParity(mgc *params.ChainConfig, pars *xchainparity.ConfigParams) {
	if pars.ForkCanonHash != nil {
		if (*pars.ForkCanonHash == common.HexToHash("0x4985f5ca3d2afbec36529aa96f74de3cc10a2a4a6c44f2157a57d2c6059a11bb")) ||
			(*pars.ForkCanonHash == common.HexToHash("0x3e12d5c0f8d63fbc5831cc7f7273bd824fa4d0a9a4102d65d99a7ea5604abc00")) {

			mgc.DAOForkBlock = new(big.Int).SetUint64(pars.ForkBlock.Uint64())
			mgc.DAOForkSupport = true
		}
		if *pars.ForkCanonHash == common.HexToHash("0x94365e3a8c0b35089c1d1195081fe7489b528a84b22199c916180db8b28ade7f") {
			mgc.DAOForkBlock = new(big.Int).SetUint64(pars.ForkBlock.Uint64())
		}
	}
}

func ParityConfigWithPrecompiledContractsFromMultiGeth(c *xchainparity.Config, mgg *Genesis) {
	c.Accounts = make(xchainparity.ConfigAccounts, 0)

	ecrecover := "ecrecover"
	c.Accounts[common.BytesToAddress([]byte{1}).Hex()] = xchainparity.ConfigAccountValue{
		Builtin: &xchainparity.ConfigAccountValueBuiltin{
			Name: &ecrecover,
			PricingOpt: xchainparity.ConfigAccountValueBuiltinPricing{
				ConfigAccountValueBuiltinPricingLinear: &xchainparity.ConfigAccountValueBuiltinPricingLinear{
					Base: 3000,
					Word: 0,
				},
			},
		},
	}

	sha256 := "sha256"
	c.Accounts[common.BytesToAddress([]byte{2}).Hex()] = xchainparity.ConfigAccountValue{
		Builtin: &xchainparity.ConfigAccountValueBuiltin{
			Name: &sha256,
			PricingOpt: xchainparity.ConfigAccountValueBuiltinPricing{
				ConfigAccountValueBuiltinPricingLinear: &xchainparity.ConfigAccountValueBuiltinPricingLinear{
					Base: 60,
					Word: 12,
				},
			},
		},
	}

	ripemd160 := "ripemd160"
	c.Accounts[common.BytesToAddress([]byte{3}).Hex()] = xchainparity.ConfigAccountValue{
		Builtin: &xchainparity.ConfigAccountValueBuiltin{
			Name: &ripemd160,
			PricingOpt: xchainparity.ConfigAccountValueBuiltinPricing{
				ConfigAccountValueBuiltinPricingLinear: &xchainparity.ConfigAccountValueBuiltinPricingLinear{
					Base: 600,
					Word: 120,
				},
			},
		},
	}

	identity := "identity"
	c.Accounts[common.BytesToAddress([]byte{4}).Hex()] = xchainparity.ConfigAccountValue{
		Builtin: &xchainparity.ConfigAccountValueBuiltin{
			Name: &identity,
			PricingOpt: xchainparity.ConfigAccountValueBuiltinPricing{
				ConfigAccountValueBuiltinPricingLinear: &xchainparity.ConfigAccountValueBuiltinPricingLinear{
					Base: 15,
					Word: 3,
				},
			},
		},
	}

	if mgg.Config.EIP198FBlock != nil || mgg.Config.ByzantiumBlock != nil {
		b := new(big.Int).Set(bigMax(mgg.Config.EIP198FBlock, mgg.Config.ByzantiumBlock))
		modexp := "modexp"
		c.Accounts[common.BytesToAddress([]byte{5}).Hex()] = xchainparity.ConfigAccountValue{
			Builtin: &xchainparity.ConfigAccountValueBuiltin{
				Name:       &modexp,
				ActivateAt: xchain.FromUint64(b.Uint64()),
				PricingOpt: xchainparity.ConfigAccountValueBuiltinPricing{
					ConfigAccountValueBuiltinPricingModexp: &xchainparity.ConfigAccountValueBuiltinPricingModexp{
						Divisor: 20,
					},
				},
			},
		}

	}

	if mgg.Config.EIP212FBlock != nil || mgg.Config.ByzantiumBlock != nil {
		b := new(big.Int).Set(bigMax(mgg.Config.EIP212FBlock, mgg.Config.ByzantiumBlock))
		alt_bn128_pairing := "alt_bn128_pairing"
		c.Accounts[common.BytesToAddress([]byte{8}).Hex()] = xchainparity.ConfigAccountValue{
			Builtin: &xchainparity.ConfigAccountValueBuiltin{
				Name:       &alt_bn128_pairing,
				ActivateAt: xchain.FromUint64(b.Uint64()),
				PricingOpt: xchainparity.ConfigAccountValueBuiltinPricing{
					ConfigAccountValueBuiltinPricingAltBN128Pairing: &xchainparity.ConfigAccountValueBuiltinPricingAltBN128Pairing{
						Base: 100000,
						Pair: 80000,
					},
				},
			},
		}

	}

	if mgg.Config.EIP213FBlock != nil || mgg.Config.ByzantiumBlock != nil {
		b := new(big.Int).Set(bigMax(mgg.Config.EIP213FBlock, mgg.Config.ByzantiumBlock))
		alt_bn128_add := "alt_bn128_add"
		c.Accounts[common.BytesToAddress([]byte{6}).Hex()] = xchainparity.ConfigAccountValue{
			Builtin: &xchainparity.ConfigAccountValueBuiltin{
				Name:       &alt_bn128_add,
				ActivateAt: xchain.FromUint64(b.Uint64()),
				PricingOpt: xchainparity.ConfigAccountValueBuiltinPricing{
					ConfigAccountValueBuiltinPricingLinear: &xchainparity.ConfigAccountValueBuiltinPricingLinear{
						Base: 500,
						Word: 0,
					},
				},
			},
		}

		alt_bn128_mul := "alt_bn128_mul"
		c.Accounts[common.BytesToAddress([]byte{7}).Hex()] = xchainparity.ConfigAccountValue{
			Builtin: &xchainparity.ConfigAccountValueBuiltin{
				Name:       &alt_bn128_mul,
				ActivateAt: xchain.FromUint64(b.Uint64()),
				PricingOpt: xchainparity.ConfigAccountValueBuiltinPricing{
					ConfigAccountValueBuiltinPricingLinear: &xchainparity.ConfigAccountValueBuiltinPricingLinear{
						Base: 40000,
						Word: 0,
					},
				},
			},
		}
	}
}

package parity

import (
	"fmt"
	"math/big"
	"strings"

	xchain "github.com/etclabscore/eth-x-chainspec"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

func (c *Config) BuiltinContracts() (builtins []ConfigAccountValueBuiltin) {
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
func (c *Config) ToMultiGethGenesis() *core.Genesis {
	mgc := &params.ChainConfig{}
	if pars := c.Params; pars != nil {
		if err := checkUnsupportedValsMust(pars); err != nil {
			panic(err)
		}

		mgc.ChainID = pars.ChainID.Big()

		// Defaults according to Parity documentation https://wiki.parity.io/Chain-specification.html
		if mgc.ChainID == nil && pars.NetworkID != nil {
			mgc.ChainID = pars.NetworkID.Big()
		}

		// DAO
		setDAOConfigs(mgc, pars)

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

		parityBuiltins := c.BuiltinContracts()
		for _, pc := range parityBuiltins {
			if pc.ActivateAt != nil {
				switch *pc.Name {
				case "modexp":
					mgc.EIP198FBlock = new(big.Int).Set(pc.ActivateAt.Big())
				case "alt_bn128_pairing":
					mgc.EIP212FBlock = new(big.Int).Set(pc.ActivateAt.Big())
				case "alt_bn128_add", "alt_bn128_mul":
					mgc.EIP213FBlock = new(big.Int).Set(pc.ActivateAt.Big())
				default:
					// panic("unsupported builtin contract: " + *pc.Name)
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
	mgg := &core.Genesis{
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
		mgg.Alloc = core.GenesisAlloc{}

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
			if _, ok := vm.PrecompiledContractsForConfig(params.AllEthashProtocolChanges, big.NewInt(0))[addr]; ok && bal.Sign() < 1 {
				continue accountsloop
			}

			mgg.Alloc[addr] = core.GenesisAccount{
				Nonce:   nonce,
				Balance: bal,
				Code:    v.Code,
				Storage: v.Storage,
			}
		}
	}
	return mgg
}

func checkUnsupportedValsMust(pars *ConfigParams) error {
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

func setDAOConfigs(mgc *params.ChainConfig, pars *ConfigParams) {
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

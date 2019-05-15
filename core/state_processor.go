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

package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereumproject/evm-ffi/go/sputnikvm"
	"math/big"
)

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	config *params.ChainConfig // Chain configuration options
	bc     *BlockChain         // Canonical block chain
	engine consensus.Engine    // Consensus engine used for block rewards
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, bc *BlockChain, engine consensus.Engine) *StateProcessor {
	return &StateProcessor{
		config: config,
		bc:     bc,
		engine: engine,
	}
}

// Process processes the state changes according to the Ethereum rules by running
// the transaction messages using the statedb and applying any rewards to both
// the processor (coinbase) and any included uncles.
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (p *StateProcessor) Process(block *types.Block, statedb *state.StateDB, cfg vm.Config) (types.Receipts, []*types.Log, uint64, error) {
	var (
		receipts types.Receipts
		usedGas  = new(uint64)
		header   = block.Header()
		allLogs  []*types.Log
		gp       = new(GasPool).AddGas(block.GasLimit())
	)
	// Mutate the block and state according to any hard-fork specs
	if p.config.DAOForkSupport && p.config.DAOForkBlock != nil && p.config.DAOForkBlock.Cmp(block.Number()) == 0 {
		misc.ApplyDAOHardFork(statedb)
	}
	// Iterate over and process the individual transactions
	for i, tx := range block.Transactions() {
		statedb.Prepare(tx.Hash(), block.Hash(), i)
		receipt, _, err := ApplyTransaction(p.config, p.bc, nil, gp, statedb, header, tx, usedGas, cfg)
		if err != nil {
			return nil, nil, 0, err
		}
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, receipt.Logs...)
	}
	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	p.engine.Finalize(p.bc, header, statedb, block.Transactions(), block.Uncles())

	return receipts, allLogs, *usedGas, nil
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(config *params.ChainConfig, bc ChainContext, author *common.Address, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, error) {
	if cfg.EVMInterpreter == "svm" {
		return applySputnikTransaction(config, bc, author, gp, statedb, header, tx, usedGas, cfg)
	}
	return applyTransaction(config, bc, author, gp, statedb, header, tx, usedGas, cfg)
}

// applyTransaction is the standard transaction application function, using the built in go evm.
func applyTransaction(config *params.ChainConfig, bc ChainContext, author *common.Address, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, error) {
	msg, err := tx.AsMessage(types.MakeSigner(config, header.Number))
	if err != nil {
		return nil, 0, err
	}
	// Create a new context to be used in the EVM environment
	context := NewEVMContext(msg, header, bc, author)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmenv := vm.NewEVM(context, statedb, config, cfg)
	// Apply the transaction to the current state (included in the env)
	_, gas, failed, err := ApplyMessage(vmenv, msg, gp)
	if err != nil {
		return nil, 0, err
	}
	// Update the state with pending changes
	var root []byte
	if config.IsEIP658F(header.Number) {
		statedb.Finalise(config.IsEIP161F(header.Number))
	} else {
		root = statedb.IntermediateRoot(config.IsEIP161F(header.Number)).Bytes()
	}
	*usedGas += gas

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing whether the root touch-delete accounts.
	receipt := types.NewReceipt(root, failed, *usedGas)
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = gas
	// if the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(vmenv.Context.Origin, tx.Nonce())
	}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = statedb.GetLogs(tx.Hash())
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
	receipt.BlockHash = statedb.BlockHash()
	receipt.BlockNumber = header.Number
	receipt.TransactionIndex = uint(statedb.TxIndex())

	return receipt, gas, err
}

func precheckSputnikVMTransaction(config *params.ChainConfig, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64) error {
	// Convert transaction to message
	msg, err := tx.AsMessage(types.MakeSigner(config, header.Number))
	if err != nil {
		return err
	}

	// Check nonce
	if msg.CheckNonce() {
		nonce := statedb.GetNonce(msg.From())
		if nonce < msg.Nonce() {
			return ErrNonceTooHigh
		} else if nonce > msg.Nonce() {
			return ErrNonceTooLow
		}
	}

	// Check if there's enough balance for gas
	mgval := new(big.Int).Mul(new(big.Int).SetUint64(msg.Gas()), tx.GasPrice())
	if statedb.GetBalance(msg.From()).Cmp(mgval) < 0 {
		return errInsufficientBalanceForGas
	}

	// Check if needed gas is not greater then GasLimit
	if *usedGas+msg.Gas() > header.GasLimit {
		return ErrGasLimitReached
	}

	// No errors, pre-check finished
	return nil
}

func applySputnikTransaction(config *params.ChainConfig, bc ChainContext, author *common.Address, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, error) {
	// Pre-check is needed as SputnikVM-FFI relies on Valid Transactions to be provided.
	err := precheckSputnikVMTransaction(config, statedb, header, tx, usedGas)
	if err != nil {
		return nil, 0, err
	}

	asSputnikAddress := func(a common.Address) [20]byte {
		var addr [20]byte
		addressBytes := a.Bytes()
		for i := 0; i < 20; i++ {
			addr[i] = addressBytes[i]
		}
		return addr
	}

	asSputnikHash := func(h common.Hash) [32]byte {
		var hash [32]byte
		hashBytes := h.Bytes()
		for i := 0; i < 32; i++ {
			hash[i] = hashBytes[i]
		}
		return hash
	}

	asEthAddress := func(a [20]byte) common.Address {
		return common.BytesToAddress(a[:])
	}

	msg, err := tx.AsMessage(types.MakeSigner(config, header.Number))
	if err != nil {
		return nil, 0, err
	}
	var addr []byte
	if tx.To() != nil {
		addr = tx.To().Bytes()
	}
	vmtx := sputnikvm.Transaction{
		Caller:   asSputnikAddress(msg.From()),
		GasPrice: tx.GasPrice(),
		GasLimit: new(big.Int).SetUint64(tx.Gas()),
		Address:  addr,
		Value:    tx.Value(),
		Input:    tx.Data(),
		Nonce:    new(big.Int).SetUint64(tx.Nonce()),
	}
	vmheader := sputnikvm.HeaderParams{
		Beneficiary: asSputnikAddress(header.Coinbase),
		Timestamp:   header.Time,
		Number:      header.Number,
		Difficulty:  header.Difficulty,
		GasLimit:    new(big.Int).SetUint64(header.GasLimit),
	}
	currentNumber := header.Number

	// Get SputnikVM's corresponding chain config.
	// TODO: handle chains that are not networkid=1 (ETH main), eg testnets, custom chains with custom state staring nonces
	patch := makeSputnikVMPatch(config, header)
	vm := sputnikvm.NewDynamic(patch, &vmtx, &vmheader)

OUTER:
	for {
		ret := vm.Fire()
		switch ret.Typ() {
		case sputnikvm.RequireNone:
			break OUTER
		case sputnikvm.RequireAccount:
			address := ret.Address()
			ethAddress := asEthAddress(address)
			if statedb.Exist(ethAddress) {
				vm.CommitAccount(address, new(big.Int).SetUint64(statedb.GetNonce(ethAddress)),
					statedb.GetBalance(ethAddress), statedb.GetCode(ethAddress))
				break
			}
			vm.CommitNonexist(address)
		case sputnikvm.RequireAccountCode:
			address := ret.Address()
			ethAddress := asEthAddress(address)
			if statedb.Exist(ethAddress) {
				vm.CommitAccountCode(address, statedb.GetCode(ethAddress))
				break
			}
			vm.CommitNonexist(address)
		case sputnikvm.RequireAccountStorage:
			address := ret.Address()
			ethAddress := asEthAddress(address)
			key := common.BigToHash(ret.StorageKey())
			if statedb.Exist(ethAddress) {
				value := statedb.GetState(ethAddress, key).Big()
				sKey := ret.StorageKey()
				vm.CommitAccountStorage(address, sKey, value)
				break
			}
			vm.CommitNonexist(address)
		case sputnikvm.RequireBlockhash:
			number := ret.BlockNumber()
			hash := asSputnikHash(GetHashFn(header, bc)(number.Uint64()))
			vm.CommitBlockhash(number, hash)
		}
	}

	// VM execution is finished at this point. We apply changes to the statedb.
	for _, account := range vm.AccountChanges() {
		switch account.Typ() {
		case sputnikvm.AccountChangeIncreaseBalance:
			ethAddress := asEthAddress(account.Address())
			amount := account.ChangedAmount()
			statedb.AddBalance(ethAddress, amount)
		case sputnikvm.AccountChangeDecreaseBalance:
			ethAddress := asEthAddress(account.Address())
			amount := account.ChangedAmount()
			balance := new(big.Int).Sub(statedb.GetBalance(ethAddress), amount)
			statedb.SetBalance(ethAddress, balance)
		case sputnikvm.AccountChangeRemoved:
			ethAddress := asEthAddress(account.Address())
			statedb.Suicide(ethAddress)
		case sputnikvm.AccountChangeFull:
			ethAddress := asEthAddress(account.Address())
			code := account.Code()
			nonce := account.Nonce()
			balance := account.Balance()
			statedb.SetBalance(ethAddress, balance)
			statedb.SetNonce(ethAddress, nonce.Uint64())
			statedb.SetCode(ethAddress, code)
			for _, item := range account.ChangedStorage() {
				statedb.SetState(ethAddress, common.BigToHash(item.Key), common.BigToHash(item.Value))
			}
		case sputnikvm.AccountChangeCreate:
			ethAddress := asEthAddress(account.Address())
			code := account.Code()
			nonce := account.Nonce()
			balance := account.Balance()
			statedb.SetBalance(ethAddress, balance)
			statedb.SetNonce(ethAddress, nonce.Uint64())
			statedb.SetCode(ethAddress, code)
			for _, item := range account.Storage() {
				statedb.SetState(ethAddress, common.BigToHash(item.Key), common.BigToHash(item.Value))
			}
		default:
			panic("unreachable")
		}
	}
	for _, log := range vm.Logs() {
		var topics []common.Hash
		for _, t := range log.Topics {
			topics = append(topics, common.BytesToHash(t[:]))
		}
		// statelog := evm.NewLog(log.Address, log.Topics, log.Data, header.Number.Uint64())
		statedb.AddLog(&types.Log{
			Address:     asEthAddress(log.Address),
			Topics:      topics,
			Data:        log.Data,
			BlockNumber: currentNumber.Uint64(),
		})
	}

	// Update the state with pending changes
	var root []byte
	if config.IsEIP658F(header.Number) {
		statedb.Finalise(true)
	} else {
		root = statedb.IntermediateRoot(config.IsEIP161F(header.Number)).Bytes()
	}
	gas := vm.UsedGas().Uint64()
	*usedGas += gas

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing whether the root touch-delete accounts.
	receipt := types.NewReceipt(root, vm.Failed(), *usedGas)
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = gas

	// if the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(msg.From(), tx.Nonce())
	}

	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = statedb.GetLogs(tx.Hash())
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	// Free the patch and destroy machine
	vm.Free()
	patch.Free()

	return receipt, gas, err
}

func makeSputnikVMPatch(config *params.ChainConfig, header *types.Header) sputnikvm.DynamicPatch {
	gasTable := config.GasTable(header.Number)

	// Zero == unlimited
	codeDepositLimit := 0
	if config.IsEIP170F(header.Number) {
		codeDepositLimit = params.MaxCodeSize
	}

	// Helper to convert uint64 tp big.Int
	toBigInt := func(x uint64) *big.Int {
		return new(big.Int).SetUint64(x)
	}

	rules := config.Rules(header.Number)

	// Calculate the upfront CREATE cost (it's lower on Frontier)
	createGasCost := params.CreateGas
	if !rules.IsEIP2F {
		createGasCost = 0
	}

	// Build list of enabled precompile
	enabledPrecompileds := [][20]byte{
		common.BytesToAddress([]byte{1}),
		common.BytesToAddress([]byte{2}),
		common.BytesToAddress([]byte{3}),
		common.BytesToAddress([]byte{4}),
	}

	if rules.IsEIP198F {
		enabledPrecompileds = append(enabledPrecompileds,
			common.BytesToAddress([]byte{5}))
	}

	if rules.IsEIP213F {
		enabledPrecompileds = append(enabledPrecompileds,
			common.BytesToAddress([]byte{6}))
		enabledPrecompileds = append(enabledPrecompileds,
			common.BytesToAddress([]byte{7}))
	}

	if rules.IsEIP212F {
		enabledPrecompileds = append(enabledPrecompileds,
			common.BytesToAddress([]byte{8}))
	}

	patchBuilder := sputnikvm.DynamicPatchBuilder{
		CodeDepositLimit:            uint(codeDepositLimit),
		CallStackLimit:              uint(params.CallCreateDepth),
		GasExtcode:                  toBigInt(gasTable.ExtcodeCopy),
		GasBalance:                  toBigInt(gasTable.Balance),
		GasSload:                    toBigInt(gasTable.SLoad),
		GasSuicide:                  toBigInt(gasTable.Suicide),
		GasSuicideNewAccount:        toBigInt(gasTable.CreateBySuicide),
		GasCall:                     toBigInt(gasTable.Calls),
		GasExpbyte:                  toBigInt(gasTable.ExpByte),
		GasTransactionCreate:        toBigInt(createGasCost),
		ForceCodeDeposit:            !rules.IsEIP2F,
		HasDelegateCall:             rules.IsEIP7F,
		HasStaticCall:               rules.IsEIP214F,
		HasRevert:                   rules.IsEIP140F,
		HasReturnData:               rules.IsEIP211F,
		HasBitwiseShift:             rules.IsEIP145F,
		HasCreate2:                  rules.IsEIP1014F,
		HasExtCodeHash:              rules.IsEIP1052F,
		HasReducedSstoreGasMetering: rules.IsEIP1283F,
		ErrOnCallWithMoreGas:        !rules.IsEIP150,
		CallCreateL64AfterGas:       rules.IsEIP150,
		MemoryLimit:                 ^uint(0), // Reversed 0 is max unsigned integer value for uint
		EnabledContracts:            enabledPrecompileds,
	}

	var initialNonce uint64
	var initialCreateNonce uint64
	if rules.IsEIP161F {
		initialCreateNonce = 1
	}
	accountPatch := sputnikvm.DynamicAccountPatch{
		InitialNonce:          toBigInt(initialNonce),
		InitialCreateNonce:    toBigInt(initialCreateNonce),
		EmptyConsideredExists: !rules.IsEIP161F,
		AllowPartialChange:    true,
	}

	dynamicPatch := sputnikvm.NewDynamicPatch(&patchBuilder, &accountPatch)

	return dynamicPatch
}

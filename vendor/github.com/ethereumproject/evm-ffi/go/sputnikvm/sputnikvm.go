// This is a go binding for SputnikVM, an Ethereum Virtual Machine.
//
// Copyright (c) ETC Dev Team 2017. Licensed under Apache-2.0.
// Authored by Wei Tang <hi@that.world>
package sputnikvm

// #include "../../c/sputnikvm.h"
// #include <stdlib.h>
//
// sputnikvm_address sputnikvm_require_value_read_account(sputnikvm_require_value v) {
//   return v.account;
// }
//
// sputnikvm_require_value_account_storage sputnikvm_require_value_read_account_storage(sputnikvm_require_value v) {
//   return v.account_storage;
// }
//
// sputnikvm_u256 sputnikvm_require_value_read_blockhash(sputnikvm_require_value v) {
//   return v.blockhash;
// }
//
// sputnikvm_account_change_value_balance sputnikvm_account_change_value_read_balance(sputnikvm_account_change_value v) {
//   return v.balance;
// }
//
// sputnikvm_account_change_value_all sputnikvm_account_change_value_read_all(sputnikvm_account_change_value v) {
//   return v.all;
// }
//
// sputnikvm_address sputnikvm_account_change_value_read_removed(sputnikvm_account_change_value v) {
//   return v.removed;
// }
import "C"

import (
	"math/big"
	"unsafe"
)

type AccountChangeType int

const (
	AccountChangeIncreaseBalance = iota
	AccountChangeDecreaseBalance
	AccountChangeFull
	AccountChangeCreate
	AccountChangeRemoved
)

type AccountChangeStorageItem struct {
	Key   *big.Int
	Value *big.Int
}

type AccountChange struct {
	info    C.sputnikvm_account_change
	storage []AccountChangeStorageItem
	code    []byte
}

func (change *AccountChange) Typ() AccountChangeType {
	switch change.info.typ {
	case C.account_change_increase_balance:
		return AccountChangeIncreaseBalance
	case C.account_change_decrease_balance:
		return AccountChangeDecreaseBalance
	case C.account_change_full:
		return AccountChangeFull
	case C.account_change_create:
		return AccountChangeCreate
	case C.account_change_removed:
		return AccountChangeRemoved
	default:
		panic("unreachable")
	}
}

func (change *AccountChange) Address() [20]byte {
	switch change.Typ() {
	case AccountChangeIncreaseBalance, AccountChangeDecreaseBalance:
		balance := C.sputnikvm_account_change_value_read_balance(change.info.value)
		return FromCAddress(balance.address)
	case AccountChangeFull, AccountChangeCreate:
		all := C.sputnikvm_account_change_value_read_all(change.info.value)
		return FromCAddress(all.address)
	case AccountChangeRemoved:
		removed := C.sputnikvm_account_change_value_read_removed(change.info.value)
		return FromCAddress(removed)
	default:
		panic("unreachable")
	}
}

func (change *AccountChange) ChangedAmount() *big.Int {
	switch change.Typ() {
	case AccountChangeIncreaseBalance, AccountChangeDecreaseBalance:
		balance := C.sputnikvm_account_change_value_read_balance(change.info.value)
		return FromCU256(balance.amount)
	default:
		panic("Incorrect usage")
	}
}

func (change *AccountChange) Nonce() *big.Int {
	switch change.Typ() {
	case AccountChangeFull, AccountChangeCreate:
		all := C.sputnikvm_account_change_value_read_all(change.info.value)
		return FromCU256(all.nonce)
	default:
		panic("incorrect usage")
	}
}

func (change *AccountChange) Balance() *big.Int {
	switch change.Typ() {
	case AccountChangeFull, AccountChangeCreate:
		all := C.sputnikvm_account_change_value_read_all(change.info.value)
		return FromCU256(all.balance)
	default:
		panic("incorrect usage")
	}
}

func (change *AccountChange) Code() []byte {
	switch change.Typ() {
	case AccountChangeFull, AccountChangeCreate:
		return change.code
	default:
		panic("incorrect usage")
	}
}

func (change *AccountChange) Storage() []AccountChangeStorageItem {
	switch change.Typ() {
	case AccountChangeCreate:
		return change.storage
	default:
		panic("incorrect usage")
	}
}

func (change *AccountChange) ChangedStorage() []AccountChangeStorageItem {
	switch change.Typ() {
	case AccountChangeFull:
		return change.storage
	default:
		panic("incorrect usage")
	}
}

type RequireType int

const (
	RequireNone = iota
	RequireAccount
	RequireAccountCode
	RequireAccountStorage
	RequireBlockhash
)

type Require struct {
	c C.sputnikvm_require
}

func (require *Require) Typ() RequireType {
	switch require.c.typ {
	case C.require_none:
		return RequireNone
	case C.require_account:
		return RequireAccount
	case C.require_account_code:
		return RequireAccountCode
	case C.require_account_storage:
		return RequireAccountStorage
	case C.require_blockhash:
		return RequireBlockhash
	default:
		panic("unreachable")
	}
}

func (require *Require) Address() [20]byte {
	switch require.Typ() {
	case RequireAccount, RequireAccountCode:
		return FromCAddress(C.sputnikvm_require_value_read_account(require.c.value))
	case RequireAccountStorage:
		return FromCAddress(C.sputnikvm_require_value_read_account_storage(require.c.value).address)
	default:
		panic("incorrect usage")
	}
}

func (require *Require) StorageKey() *big.Int {
	switch require.Typ() {
	case RequireAccountStorage:
		storage := C.sputnikvm_require_value_read_account_storage(require.c.value)
		return FromCU256(storage.key)
	default:
		panic("incorrect usage")
	}
}

func (require *Require) BlockNumber() *big.Int {
	switch require.Typ() {
	case RequireBlockhash:
		number := C.sputnikvm_require_value_read_blockhash(require.c.value)
		return FromCU256(number)
	default:
		panic("incorrect usage")
	}
}

type DynamicPatchBuilder struct {
	CodeDepositLimit uint
	CallStackLimit uint
	GasExtcode *big.Int
	GasBalance *big.Int
	GasSload *big.Int
	GasSuicide *big.Int
	GasSuicideNewAccount *big.Int
	GasCall *big.Int
	GasExpbyte *big.Int
	GasTransactionCreate *big.Int
	ForceCodeDeposit bool
	HasDelegateCall bool
	HasStaticCall bool
	HasRevert bool
	HasReturnData bool
	HasBitwiseShift bool
	HasCreate2 bool
	HasExtCodeHash bool
	HasReducedSstoreGasMetering bool
	ErrOnCallWithMoreGas bool
	CallCreateL64AfterGas bool
	MemoryLimit uint
	EnabledContracts [][20]byte
}

func ToCDynamicPatchBuilder(v *DynamicPatchBuilder) C.sputnikvm_dynamic_patch_builder {
	cEnabledContracts := make([]C.sputnikvm_address, len(v.EnabledContracts))
	for i := 0; i < len(v.EnabledContracts); i++ {
		cEnabledContracts[i] = ToCAddress(v.EnabledContracts[i])
	}

	cbuilder := new(C.sputnikvm_dynamic_patch_builder)
	cbuilder.code_deposit_limit = C.ulong(v.CodeDepositLimit)
	cbuilder.callstack_limit = C.ulong(v.CallStackLimit)
	cbuilder.gas_extcode = ToCGas(v.GasExtcode)
	cbuilder.gas_balance = ToCGas(v.GasBalance)
	cbuilder.gas_sload = ToCGas(v.GasSload)
	cbuilder.gas_suicide = ToCGas(v.GasSuicide)
	cbuilder.gas_suicide_new_account = ToCGas(v.GasSuicideNewAccount)
	cbuilder.gas_call = ToCGas(v.GasCall)
	cbuilder.gas_expbyte = ToCGas(v.GasExpbyte)
	cbuilder.gas_transaction_create = ToCGas(v.GasTransactionCreate)
	cbuilder.force_code_deposit = C.bool(v.ForceCodeDeposit)
	cbuilder.has_delegate_call = C.bool(v.HasDelegateCall)
	cbuilder.has_static_call = C.bool(v.HasStaticCall)
	cbuilder.has_revert = C.bool(v.HasRevert)
	cbuilder.has_return_data = C.bool(v.HasReturnData)
	cbuilder.has_bitwise_shift = C.bool(v.HasBitwiseShift)
	cbuilder.has_create2 = C.bool(v.HasCreate2)
	cbuilder.has_extcodehash = C.bool(v.HasExtCodeHash)
	cbuilder.has_reduced_sstore_gas_metering = C.bool(v.HasReducedSstoreGasMetering)
	cbuilder.err_on_call_with_more_gas = C.bool(v.ErrOnCallWithMoreGas)
	cbuilder.call_create_l64_after_gas = C.bool(v.CallCreateL64AfterGas)
	cbuilder.memory_limit = C.ulong(v.MemoryLimit)
	cbuilder.enabled_contracts = &cEnabledContracts[0]
	cbuilder.enabled_contracts_length = C.ulong(len(cEnabledContracts))

	return *cbuilder
}

type DynamicAccountPatch struct {
	InitialNonce *big.Int
	InitialCreateNonce *big.Int
	EmptyConsideredExists bool
	AllowPartialChange bool
}

func ToCDynamicAccountPatch(v *DynamicAccountPatch) C.sputnikvm_dynamic_account_patch {
	cpatch := new(C.sputnikvm_dynamic_account_patch)
	cpatch.initial_nonce = ToCU256(v.InitialNonce)
	cpatch.initial_create_nonce = ToCU256(v.InitialCreateNonce)
	cpatch.empty_considered_exists = C.bool(v.EmptyConsideredExists)
	cpatch.allow_partial_change = C.bool(v.AllowPartialChange)
	return *cpatch
}

type DynamicPatch struct {
	ptr C.sputnikvm_dynamic_patch
}

func NewDynamicPatch(builder *DynamicPatchBuilder, accountPatch *DynamicAccountPatch) DynamicPatch {
	cbuilder := ToCDynamicPatchBuilder(builder)
	cpatch := ToCDynamicAccountPatch(accountPatch)
	return DynamicPatch{ C.dynamic_patch_new(cbuilder, cpatch)}
}

func (p *DynamicPatch) Free() {
	C.dynamic_patch_free(p.ptr)
}

type Log struct {
	Address [20]byte
	Topics  [][32]byte
	Data    []byte
}

type VM struct {
	c *C.sputnikvm_vm_t
}

type Transaction struct {
	Caller   [20]byte
	GasPrice *big.Int
	GasLimit *big.Int
	Address  []byte // If it is nil, then we take it as a Create transaction.
	Value    *big.Int
	Input    []byte
	Nonce    *big.Int
}

type HeaderParams struct {
	Beneficiary [20]byte
	Timestamp   uint64
	Number      *big.Int
	Difficulty  *big.Int
	GasLimit    *big.Int
}

func PrintCU256(v C.sputnikvm_u256) {
	C.print_u256(v)
}

func ToCU256(v *big.Int) C.sputnikvm_u256 {
	bytes := v.Bytes()
	cu256 := new(C.sputnikvm_u256)
	for i := 0; i < 32; i++ {
		if i < (32 - len(bytes)) {
			continue
		}
		cu256.data[i] = C.uchar(bytes[i-(32-len(bytes))])
	}
	return *cu256
}

func FromCU256(v C.sputnikvm_u256) *big.Int {
	bytes := new([32]byte)
	for i := 0; i < 32; i++ {
		bytes[i] = byte(v.data[i])
	}
	i := new(big.Int)
	i.SetBytes(bytes[0:32])
	return i
}

func ToCGas(v *big.Int) C.sputnikvm_gas {
	bytes := v.Bytes()
	cgas := new(C.sputnikvm_gas)
	for i := 0; i < 32; i++ {
		if i < (32 - len(bytes)) {
			continue
		}
		cgas.data[i] = C.uchar(bytes[i-(32-len(bytes))])
	}
	return *cgas
}

func FromCGas(v C.sputnikvm_gas) *big.Int {
	bytes := new([32]byte)
	for i := 0; i < 32; i++ {
		bytes[i] = byte(v.data[i])
	}
	i := new(big.Int)
	i.SetBytes(bytes[0:32])
	return i
}

// NOTE: We can use non-nil len and capped bytes because these functions are never
// intended to handle the case when a transaction is passed a nil 'To' value
// indicating a contract creation transaction. Using bytes with explicit lengths
// helps reader and developer differentiate Address vs. Hash values, as well
// as being generally more descriptive code.
func ToCAddress(v [20]byte) C.sputnikvm_address {
	caddress := new(C.sputnikvm_address)
	for i := 0; i < 20; i++ {
		caddress.data[i] = C.uchar(v[i])
	}
	return *caddress
}

func FromCAddress(v C.sputnikvm_address) [20]byte {
	var b [20]byte
	for i := 0; i < 20; i++ {
		b[i] = byte(v.data[i])
	}
	return b
}

func ToCH256(v [32]byte) C.sputnikvm_h256 {
	chash := new(C.sputnikvm_h256)
	for i := 0; i < 32; i++ {
		chash.data[i] = C.uchar(v[i])
	}
	return *chash
}

func FromCH256(v C.sputnikvm_h256) [32]byte {
	var b [32]byte
	for i := 0; i < 32; i++ {
		b[i] = byte(v.data[i])
	}
	return b
}

func toCTransaction(transaction *Transaction) (*C.sputnikvm_transaction, unsafe.Pointer) {
	// Malloc input length memory and must be freed manually.

	ctransaction := new(C.sputnikvm_transaction)
	cinput := C.malloc(C.size_t(len(transaction.Input)))
	for i := 0; i < len(transaction.Input); i++ {
		i_cinput := unsafe.Pointer(uintptr(cinput) + uintptr(i))
		*(*C.uchar)(i_cinput) = C.uchar(transaction.Input[i])
	}
	ctransaction.caller = ToCAddress(transaction.Caller)
	ctransaction.gas_price = ToCGas(transaction.GasPrice)
	ctransaction.gas_limit = ToCGas(transaction.GasLimit)
	if len(transaction.Address) == 0 {
		ctransaction.action = C.sputnikvm_action(C.CREATE_ACTION)
	} else {
		ctransaction.action = C.sputnikvm_action(C.CALL_ACTION)
		var baddr [20]byte
		for i := 0; i < 20; i++ {
			baddr[i] = transaction.Address[i]
		}
		ctransaction.action_address = ToCAddress(baddr)
	}
	ctransaction.value = ToCU256(transaction.Value)
	ctransaction.input = (*C.uchar)(cinput)
	ctransaction.input_len = C.uint(len(transaction.Input))
	ctransaction.nonce = ToCU256(transaction.Nonce)

	return ctransaction, cinput
}

func ToCHeaderParams(header *HeaderParams) *C.sputnikvm_header_params {
	cheader := new(C.sputnikvm_header_params)
	cheader.beneficiary = ToCAddress(header.Beneficiary)
	cheader.timestamp = C.ulonglong(header.Timestamp)
	cheader.number = ToCU256(header.Number)
	cheader.difficulty = ToCU256(header.Difficulty)
	cheader.gas_limit = ToCGas(header.GasLimit)

	return cheader
}

func NewDynamic(patch DynamicPatch, transaction *Transaction, header *HeaderParams) *VM {
	ctransaction, cinput := toCTransaction(transaction)
	cheader := ToCHeaderParams(header)

	cvm := C.sputnikvm_new_dynamic(patch.ptr, *ctransaction, *cheader)
	C.free(cinput)

	vm := new(VM)
	vm.c = cvm

	return vm
}

func (vm *VM) Fire() Require {
	ret := C.sputnikvm_fire(vm.c)
	return Require{
		c: ret,
	}
}

func (vm *VM) Free() {
	C.sputnikvm_free(vm.c)
}

func (vm *VM) CommitAccount(address [20]byte, nonce *big.Int, balance *big.Int, code []byte) {
	caddress := ToCAddress(address)
	cnonce := ToCU256(nonce)
	cbalance := ToCU256(balance)
	ccode := C.malloc(C.size_t(len(code)))
	for i := 0; i < len(code); i++ {
		i_ccode := unsafe.Pointer(uintptr(ccode) + uintptr(i))
		*(*C.uchar)(i_ccode) = C.uchar(code[i])
	}

	C.sputnikvm_commit_account(vm.c, caddress, cnonce, cbalance, (*C.uchar)(ccode), C.uint(len(code)))
	C.free(ccode)
}

func (vm *VM) CommitAccountCode(address [20]byte, code []byte) {
	caddress := ToCAddress(address)
	ccode := C.malloc(C.size_t(len(code)))
	for i := 0; i < len(code); i++ {
		i_ccode := unsafe.Pointer(uintptr(ccode) + uintptr(i))
		*(*C.uchar)(i_ccode) = C.uchar(code[i])
	}

	C.sputnikvm_commit_account_code(vm.c, caddress, (*C.uchar)(ccode), C.uint(len(code)))
	C.free(ccode)
}

func (vm *VM) CommitAccountStorage(address [20]byte, key *big.Int, value *big.Int) {
	caddress := ToCAddress(address)
	ckey := ToCU256(key)
	cvalue := ToCU256(value)

	C.sputnikvm_commit_account_storage(vm.c, caddress, ckey, cvalue)
}

func (vm *VM) CommitNonexist(address [20]byte) {
	caddress := ToCAddress(address)
	C.sputnikvm_commit_nonexist(vm.c, caddress)
}

func (vm *VM) CommitBlockhash(number *big.Int, hash [32]byte) {
	cnumber := ToCU256(number)
	chash := ToCH256(hash)
	C.sputnikvm_commit_blockhash(vm.c, cnumber, chash)
}

func (vm *VM) UsedGas() *big.Int {
	cgas := C.sputnikvm_used_gas(vm.c)
	return FromCGas(cgas)
}

func (vm *VM) Logs() []Log {
	logs := make([]Log, 0)
	l := uint(C.sputnikvm_logs_len(vm.c))
	clogs := C.malloc(C.size_t(C.sizeof_sputnikvm_log * l))
	C.sputnikvm_logs_copy_info(vm.c, (*C.sputnikvm_log)(clogs), C.uint(l))
	for i := 0; i < int(l); i++ {
		i_clog := unsafe.Pointer(uintptr(clogs) + (uintptr(i) * uintptr(C.sizeof_sputnikvm_log)))
		clog := (*C.sputnikvm_log)(i_clog)
		address := FromCAddress(clog.address)
		topics := make([][32]byte, 0)
		for j := 0; j < int(uint(clog.topic_len)); j++ {
			topics = append(topics, FromCH256(C.sputnikvm_logs_topic(vm.c, C.uint(i), C.uint(j))))
		}
		cdata := C.malloc(C.size_t(clog.data_len))
		C.sputnikvm_logs_copy_data(vm.c, C.uint(i), (*C.uchar)(cdata), C.uint(clog.data_len))
		data := make([]byte, int(uint(clog.data_len)))
		for j := 0; j < int(uint(clog.data_len)); j++ {
			j_cdata := unsafe.Pointer(uintptr(cdata) + uintptr(j))
			data[j] = byte(*(*C.uchar)(j_cdata))
		}
		logs = append(logs, Log{
			Address: address,
			Topics:  topics,
			Data:    data,
		})
		C.free(cdata)
	}
	C.free(clogs)
	return logs
}

func (vm *VM) AccountChanges() []AccountChange {
	changes := make([]AccountChange, 0)
	l := uint(C.sputnikvm_account_changes_len(vm.c))
	cchanges := C.malloc(C.size_t(C.sizeof_sputnikvm_account_change * l))
	C.sputnikvm_account_changes_copy_info(vm.c, (*C.sputnikvm_account_change)(cchanges), C.uint(l))
	for i := 0; i < int(l); i++ {
		i_cchange := unsafe.Pointer(uintptr(cchanges) + (uintptr(i) * uintptr(C.sizeof_sputnikvm_account_change)))
		cchange := (*C.sputnikvm_account_change)(i_cchange)
		change := AccountChange{
			info:    *cchange,
			storage: make([]AccountChangeStorageItem, 0),
			code:    make([]byte, 0),
		}
		switch change.Typ() {
		case AccountChangeIncreaseBalance, AccountChangeDecreaseBalance, AccountChangeRemoved:
			changes = append(changes, change)
		case AccountChangeCreate, AccountChangeFull:
			all := C.sputnikvm_account_change_value_read_all(change.info.value)
			address := all.address
			storage_len := all.storage_len
			code_len := all.code_len

			cstorage := C.malloc(C.size_t(C.sizeof_sputnikvm_account_change_storage * uint(storage_len)))
			C.sputnikvm_account_changes_copy_storage(vm.c, address, (*C.sputnikvm_account_change_storage)(cstorage), storage_len)
			storage := make([]AccountChangeStorageItem, 0)
			for j := 0; j < int(uint(storage_len)); j++ {
				j_cstorage := unsafe.Pointer(uintptr(cstorage) + (uintptr(j) * uintptr(C.sizeof_sputnikvm_account_change_storage)))
				citem := (*C.sputnikvm_account_change_storage)(j_cstorage)
				storage = append(storage, AccountChangeStorageItem{
					Key:   FromCU256(citem.key),
					Value: FromCU256(citem.value),
				})
			}
			C.free(cstorage)

			ccode := C.malloc(C.size_t(uint(code_len)))
			C.sputnikvm_account_changes_copy_code(vm.c, address, (*C.uchar)(ccode), code_len)
			code := make([]byte, int(uint(code_len)))
			for j := 0; j < int(uint(code_len)); j++ {
				j_ccode := unsafe.Pointer(uintptr(ccode) + uintptr(j))
				code[j] = byte(*(*C.uchar)(j_ccode))
			}
			C.free(ccode)

			change.storage = storage
			change.code = code
			changes = append(changes, change)
		default:
			panic("unreachable")
		}
	}
	C.free(cchanges)
	return changes
}

func (vm *VM) Failed() bool {
	return uint(C.sputnikvm_status_failed(vm.c)) == 1
}

/* LEGACY ETC SPECIFIC API */

func NewFrontier(transaction *Transaction, header *HeaderParams) *VM {
	ctransaction, cinput := toCTransaction(transaction)
	cheader := ToCHeaderParams(header)

	cvm := C.sputnikvm_new_frontier(*ctransaction, *cheader)
	C.free(cinput)

	vm := new(VM)
	vm.c = cvm

	return vm
}

func NewHomestead(transaction *Transaction, header *HeaderParams) *VM {
	ctransaction, cinput := toCTransaction(transaction)
	cheader := ToCHeaderParams(header)

	cvm := C.sputnikvm_new_homestead(*ctransaction, *cheader)
	C.free(cinput)

	vm := new(VM)
	vm.c = cvm

	return vm
}

func NewEIP150(transaction *Transaction, header *HeaderParams) *VM {
	ctransaction, cinput := toCTransaction(transaction)
	cheader := ToCHeaderParams(header)

	cvm := C.sputnikvm_new_eip150(*ctransaction, *cheader)
	C.free(cinput)

	vm := new(VM)
	vm.c = cvm

	return vm
}

func NewEIP160(transaction *Transaction, header *HeaderParams) *VM {
	ctransaction, cinput := toCTransaction(transaction)
	cheader := ToCHeaderParams(header)

	cvm := C.sputnikvm_new_eip160(*ctransaction, *cheader)
	C.free(cinput)

	vm := new(VM)
	vm.c = cvm

	return vm
}

func NewMordenFrontier(transaction *Transaction, header *HeaderParams) *VM {
	ctransaction, cinput := toCTransaction(transaction)
	cheader := ToCHeaderParams(header)

	cvm := C.sputnikvm_new_morden_frontier(*ctransaction, *cheader)
	C.free(cinput)

	vm := new(VM)
	vm.c = cvm

	return vm
}

func NewMordenHomestead(transaction *Transaction, header *HeaderParams) *VM {
	ctransaction, cinput := toCTransaction(transaction)
	cheader := ToCHeaderParams(header)

	cvm := C.sputnikvm_new_morden_homestead(*ctransaction, *cheader)
	C.free(cinput)

	vm := new(VM)
	vm.c = cvm

	return vm
}

func NewMordenEIP150(transaction *Transaction, header *HeaderParams) *VM {
	ctransaction, cinput := toCTransaction(transaction)
	cheader := ToCHeaderParams(header)

	cvm := C.sputnikvm_new_morden_eip150(*ctransaction, *cheader)
	C.free(cinput)

	vm := new(VM)
	vm.c = cvm

	return vm
}

func NewMordenEIP160(transaction *Transaction, header *HeaderParams) *VM {
	ctransaction, cinput := toCTransaction(transaction)
	cheader := ToCHeaderParams(header)

	cvm := C.sputnikvm_new_morden_eip160(*ctransaction, *cheader)
	C.free(cinput)

	vm := new(VM)
	vm.c = cvm

	return vm
}

func NewCustomFrontier(transaction *Transaction, header *HeaderParams) *VM {
	ctransaction, cinput := toCTransaction(transaction)
	cheader := ToCHeaderParams(header)

	cvm := C.sputnikvm_new_custom_frontier(*ctransaction, *cheader)
	C.free(cinput)

	vm := new(VM)
	vm.c = cvm

	return vm
}

func NewCustomHomestead(transaction *Transaction, header *HeaderParams) *VM {
	ctransaction, cinput := toCTransaction(transaction)
	cheader := ToCHeaderParams(header)

	cvm := C.sputnikvm_new_custom_homestead(*ctransaction, *cheader)
	C.free(cinput)

	vm := new(VM)
	vm.c = cvm

	return vm
}

func NewCustomEIP150(transaction *Transaction, header *HeaderParams) *VM {
	ctransaction, cinput := toCTransaction(transaction)
	cheader := ToCHeaderParams(header)

	cvm := C.sputnikvm_new_custom_eip150(*ctransaction, *cheader)
	C.free(cinput)

	vm := new(VM)
	vm.c = cvm

	return vm
}

func NewCustomEIP160(transaction *Transaction, header *HeaderParams) *VM {
	ctransaction, cinput := toCTransaction(transaction)
	cheader := ToCHeaderParams(header)

	cvm := C.sputnikvm_new_custom_eip160(*ctransaction, *cheader)
	C.free(cinput)

	vm := new(VM)
	vm.c = cvm

	return vm
}

func SetCustomInitialNonce(nonce *big.Int) {
	cnonce := ToCU256(nonce)
	C.sputnikvm_set_custom_initial_nonce(cnonce)
}
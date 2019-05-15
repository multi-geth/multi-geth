#[cfg(not(feature = "std"))] use core::ffi::c_void;
#[cfg(feature = "std")] use std::ffi::c_void;
#[cfg(not(feature = "std"))] use core::slice;
#[cfg(feature = "std")] use std::slice;

use evm::{DynamicPatch, DynamicAccountPatch};
use evm_network::PRECOMPILEDS;
use smallvec::SmallVec;

use crate::common::{c_gas, c_u256};
use crate::c_address;

#[repr(C)]
pub struct dynamic_patch_builder {
    /// Maximum contract size. 0 for unlimited.
    pub code_deposit_limit: usize,
    /// Limit of the call stack.
    pub callstack_limit: usize,
    /// Gas paid for extcode.
    pub gas_extcode: c_gas,
    /// Gas paid for BALANCE opcode.
    pub gas_balance: c_gas,
    /// Gas paid for SLOAD opcode.
    pub gas_sload: c_gas,
    /// Gas paid for SUICIDE opcode.
    pub gas_suicide: c_gas,
    /// Gas paid for SUICIDE opcode when it hits a new account.
    pub gas_suicide_new_account: c_gas,
    /// Gas paid for CALL opcode.
    pub gas_call: c_gas,
    /// Gas paid for EXP opcode for every byte.
    pub gas_expbyte: c_gas,
    /// Gas paid for a contract creation transaction.
    pub gas_transaction_create: c_gas,
    /// Whether to force code deposit even if it does not have enough
    /// gas.
    pub force_code_deposit: bool,
    /// Whether the EVM has DELEGATECALL opcode.
    pub has_delegate_call: bool,
    /// Whether the EVM has STATICCALL opcode.
    pub has_static_call: bool,
    /// Whether the EVM has REVERT opcode.
    pub has_revert: bool,
    /// Whether the EVM has RETURNDATASIZE and RETURNDATACOPY opcode.
    pub has_return_data: bool,
    /// Whether the EVM has SHL, SHR and SAR
    pub has_bitwise_shift: bool,
    /// Whether the EVM has CREATE2
    pub has_create2: bool,
    /// Whether the EVM has EXTCODEHASH
    pub has_extcodehash: bool,
    /// Whether EVM should implement the EIP1283 gas metering scheme for SSTORE opcode
    pub has_reduced_sstore_gas_metering: bool,
    /// Whether to throw out of gas error when
    /// CALL/CALLCODE/DELEGATECALL requires more than maximum amount
    /// of gas.
    pub err_on_call_with_more_gas: bool,
    /// If true, only consume at maximum l64(after_gas) when
    /// CALL/CALLCODE/DELEGATECALL.
    pub call_create_l64_after_gas: bool,
    /// Maximum size of the memory, in bytes.
    /// NOTE: **NOT** runtime-configurable by block number
    pub memory_limit: usize,
    /// Enabled precompiled contracts array
    pub enabled_contracts: *const c_address,
    pub enabled_contracts_length: usize,
}

#[repr(C)]
pub struct dynamic_account_patch {
    initial_nonce: c_u256,
    initial_create_nonce: c_u256,
    empty_considered_exists: bool,
    allow_partial_change: bool,
}

impl From<dynamic_account_patch> for DynamicAccountPatch {
    fn from(p: dynamic_account_patch) -> Self {
        Self {
            initial_nonce: p.initial_nonce.into(),
            initial_create_nonce: p.initial_create_nonce.into(),
            empty_considered_exists: p.empty_considered_exists,
            allow_partial_change: p.allow_partial_change,
        }
    }
}

pub type dynamic_patch_box = c_void;

#[no_mangle]
extern "C" fn dynamic_patch_new(builder: dynamic_patch_builder, account_patch: dynamic_account_patch) -> *mut dynamic_patch_box {
    let mut enabled_contracts = SmallVec::new();
    let c_enabled_contracts = unsafe {  slice::from_raw_parts(builder.enabled_contracts, builder.enabled_contracts_length) };
    for c_address in c_enabled_contracts {
        let address = (*c_address).into();
        enabled_contracts.push(address);
    };

    let patch = DynamicPatch {
        account_patch: DynamicAccountPatch::from(account_patch),
        code_deposit_limit: if builder.code_deposit_limit == 0 { None } else { Some(builder.code_deposit_limit) },
        callstack_limit: builder.callstack_limit,
        gas_extcode: builder.gas_extcode.into(),
        gas_balance: builder.gas_balance.into(),
        gas_sload: builder.gas_sload.into(),
        gas_suicide: builder.gas_suicide.into(),
        gas_suicide_new_account: builder.gas_suicide_new_account.into(),
        gas_call: builder.gas_call.into(),
        gas_expbyte: builder.gas_expbyte.into(),
        gas_transaction_create: builder.gas_transaction_create.into(),
        force_code_deposit: builder.force_code_deposit,
        has_delegate_call: builder.has_delegate_call,
        has_static_call: builder.has_static_call,
        has_revert: builder.has_revert,
        has_return_data: builder.has_return_data,
        has_create2: builder.has_create2,
        has_bitwise_shift: builder.has_bitwise_shift,
        has_extcodehash: builder.has_extcodehash,
        has_reduced_sstore_gas_metering: builder.has_reduced_sstore_gas_metering,
        err_on_call_with_more_gas: builder.err_on_call_with_more_gas,
        call_create_l64_after_gas: builder.call_create_l64_after_gas,
        memory_limit: builder.memory_limit,
        enabled_precompileds: enabled_contracts,
        precompileds: &PRECOMPILEDS,
    };

    Box::into_raw(Box::new(patch)) as *mut dynamic_patch_box
}

#[no_mangle]
pub extern "C" fn dynamic_patch_free(patch: *mut dynamic_patch_box) {
    if patch.is_null() { return }
    // It's safe to erase type of AccountPatch as it's a size-less generic parameter
    unsafe { Box::from_raw(patch as *mut DynamicPatch); }
}
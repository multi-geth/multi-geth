#![cfg(feature = "legacy")]

use crate::common::c_u256;
use crate::{
    c_transaction,
    c_header_params,
    sputnikvm_new,
};

use bigint::U256;

use evm::VM;
use evm_network_classic::{MainnetFrontierPatch, MainnetHomesteadPatch, MainnetEIP150Patch, MainnetEIP160Patch};
use evm::AccountPatch;
use evm_network_classic::{FrontierPatch, HomesteadPatch, EIP150Patch, EIP160Patch};

use lazy_static::lazy_static;

#[derive(Copy, Clone, Default)]
pub struct MordenAccountPatch;
impl AccountPatch for MordenAccountPatch {
    fn initial_nonce(&self) -> U256 { U256::from(1048576) }
    fn initial_create_nonce(&self) -> U256 { self.initial_nonce() }
    fn empty_considered_exists(&self) -> bool { true }
}

pub type MordenFrontierPatch = FrontierPatch<MordenAccountPatch>;
pub type MordenHomesteadPatch = HomesteadPatch<MordenAccountPatch>;
pub type MordenEIP150Patch = EIP150Patch<MordenAccountPatch>;
pub type MordenEIP160Patch = EIP160Patch<MordenAccountPatch>;

static mut CUSTOM_INITIAL_NONCE: Option<U256> = None;

#[derive(Copy, Clone, Default)]
pub struct CustomAccountPatch;
impl AccountPatch for CustomAccountPatch {
    fn initial_nonce(&self) -> U256 { U256::from(unsafe { CUSTOM_INITIAL_NONCE.unwrap() }) }
    fn initial_create_nonce(&self) -> U256 { self.initial_nonce() }
    fn empty_considered_exists(&self) -> bool { true }
}

pub type CustomFrontierPatch = FrontierPatch<CustomAccountPatch>;
pub type CustomHomesteadPatch = HomesteadPatch<CustomAccountPatch>;
pub type CustomEIP150Patch = EIP150Patch<CustomAccountPatch>;
pub type CustomEIP160Patch = EIP160Patch<CustomAccountPatch>;

#[no_mangle]
#[deprecated(since = "0.11.0", note = "Ethereum Classic specific FFI interface is deprecated, use the network-agnostic API instead.")]
pub unsafe extern "C" fn sputnikvm_set_custom_initial_nonce(v: c_u256) {
    let v: U256 = v.into();
    CUSTOM_INITIAL_NONCE = Some(v)
}

lazy_static! {
    static ref MAINNET_FRONTIER_PATCH: MainnetFrontierPatch = FrontierPatch::default();
    static ref MAINNET_HOMESTEAD_PATCH: MainnetHomesteadPatch = HomesteadPatch::default();
    static ref MAINNET_EIP150_PATCH: MainnetEIP150Patch = EIP150Patch::default();
    static ref MAINNET_EIP160_PATCH: MainnetEIP160Patch = EIP160Patch::default();
    
    static ref MORDEN_FRONTIER_PATCH: MordenFrontierPatch = FrontierPatch::default();
    static ref MORDEN_HOMESTEAD_PATCH: MordenHomesteadPatch = HomesteadPatch::default();
    static ref MORDEN_EIP150_PATCH: MordenEIP150Patch = EIP150Patch::default();
    static ref MORDEN_EIP160_PATCH: MordenEIP160Patch = EIP160Patch::default();
    
    static ref CUSTOM_FRONTIER_PATCH: CustomFrontierPatch = FrontierPatch::default();
    static ref CUSTOM_HOMESTEAD_PATCH: CustomHomesteadPatch = HomesteadPatch::default();
    static ref CUSTOM_EIP150_PATCH: CustomEIP150Patch = EIP150Patch::default();
    static ref CUSTOM_EIP160_PATCH: CustomEIP160Patch = EIP160Patch::default();
}

#[no_mangle]
#[deprecated(since = "0.11.0", note = "Ethereum Classic specific FFI interface is deprecated, use the network-agnostic API instead.")]
pub extern "C" fn sputnikvm_new_frontier(
    transaction: c_transaction, header: c_header_params
) -> *mut Box<VM> {
    sputnikvm_new(&*MAINNET_FRONTIER_PATCH, transaction, header)
}

#[no_mangle]
#[deprecated(since = "0.11.0", note = "Ethereum Classic specific FFI interface is deprecated, use the network-agnostic API instead.")]
pub extern "C" fn sputnikvm_new_homestead(
    transaction: c_transaction, header: c_header_params
) -> *mut Box<VM> {
    sputnikvm_new(&*MAINNET_HOMESTEAD_PATCH, transaction, header)
}

#[no_mangle]
#[deprecated(since = "0.11.0", note = "Ethereum Classic specific FFI interface is deprecated, use the network-agnostic API instead.")]
pub extern "C" fn sputnikvm_new_eip150(
    transaction: c_transaction, header: c_header_params
) -> *mut Box<VM> {
    sputnikvm_new(&*MAINNET_EIP150_PATCH, transaction, header)
}

#[no_mangle]
#[deprecated(since = "0.11.0", note = "Ethereum Classic specific FFI interface is deprecated, use the network-agnostic API instead.")]
pub extern "C" fn sputnikvm_new_eip160(
    transaction: c_transaction, header: c_header_params
) -> *mut Box<VM> {
    sputnikvm_new(&*MAINNET_EIP160_PATCH, transaction, header)
}

#[no_mangle]
#[deprecated(since = "0.11.0", note = "Ethereum Classic specific FFI interface is deprecated, use the network-agnostic API instead.")]
pub extern "C" fn sputnikvm_new_morden_frontier(
    transaction: c_transaction, header: c_header_params
) -> *mut Box<VM> {
    sputnikvm_new(&*MORDEN_FRONTIER_PATCH, transaction, header)
}

#[no_mangle]
#[deprecated(since = "0.11.0", note = "Ethereum Classic specific FFI interface is deprecated, use the network-agnostic API instead.")]
pub extern "C" fn sputnikvm_new_morden_homestead(
    transaction: c_transaction, header: c_header_params
) -> *mut Box<VM> {
    sputnikvm_new(&*MORDEN_HOMESTEAD_PATCH, transaction, header)
}

#[no_mangle]
#[deprecated(since = "0.11.0", note = "Ethereum Classic specific FFI interface is deprecated, use the network-agnostic API instead.")]
pub extern "C" fn sputnikvm_new_morden_eip150(
    transaction: c_transaction, header: c_header_params
) -> *mut Box<VM> {
    sputnikvm_new(&*MORDEN_EIP160_PATCH, transaction, header)
}

#[no_mangle]
#[deprecated(since = "0.11.0", note = "Ethereum Classic specific FFI interface is deprecated, use the network-agnostic API instead.")]
pub extern "C" fn sputnikvm_new_morden_eip160(
    transaction: c_transaction, header: c_header_params
) -> *mut Box<VM> {
    sputnikvm_new(&*MORDEN_EIP160_PATCH, transaction, header)
}

#[no_mangle]
#[deprecated(since = "0.11.0", note = "Ethereum Classic specific FFI interface is deprecated, use the network-agnostic API instead.")]
pub extern "C" fn sputnikvm_new_custom_frontier(
    transaction: c_transaction, header: c_header_params
) -> *mut Box<VM> {
    sputnikvm_new(&*CUSTOM_FRONTIER_PATCH, transaction, header)
}

#[no_mangle]
#[deprecated(since = "0.11.0", note = "Ethereum Classic specific FFI interface is deprecated, use the network-agnostic API instead.")]
pub extern "C" fn sputnikvm_new_custom_homestead(
    transaction: c_transaction, header: c_header_params
) -> *mut Box<VM> {
    sputnikvm_new(&*CUSTOM_HOMESTEAD_PATCH, transaction, header)
}

#[no_mangle]
#[deprecated(since = "0.11.0", note = "Ethereum Classic specific FFI interface is deprecated, use the network-agnostic API instead.")]
pub extern "C" fn sputnikvm_new_custom_eip150(
    transaction: c_transaction, header: c_header_params
) -> *mut Box<VM> {
    sputnikvm_new(&*CUSTOM_EIP150_PATCH, transaction, header)
}

#[no_mangle]
#[deprecated(since = "0.11.0", note = "Ethereum Classic specific FFI interface is deprecated, use the network-agnostic API instead.")]
pub extern "C" fn sputnikvm_new_custom_eip160(
    transaction: c_transaction, header: c_header_params
) -> *mut Box<VM> {
    sputnikvm_new(&*CUSTOM_EIP160_PATCH, transaction, header)
}

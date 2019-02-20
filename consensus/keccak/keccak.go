// Copyright 2019 The go-ethereum Authors
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

// Package ethash implements the ethash proof-of-work consensus engine.
package keccak

import (
	"errors"
	"math/big"
	"sync"
	"time"
	"unsafe"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/rpc"
)

var ErrInvalidDumpMagic = errors.New("invalid dump magic")

var (
	// two256 is a big integer representing 2^256
	two256 = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))

	// sharedEthash is a full instance that can be shared between multiple users.
)

// isLittleEndian returns whether the local system is running in little or big
// endian byte order.
func isLittleEndian() bool {
	n := uint32(0x01020304)
	return *(*byte)(unsafe.Pointer(&n)) == 0x04
}

// Mode defines the type and amount of PoW verification an ethash engine makes.
type Mode uint

const (
	ModeNormal Mode = iota
	ModeShared
	ModeTest
	ModeFake
	ModeFullFake
)

// sealTask wraps a seal block with relative result channel for remote sealer thread.
type sealTask struct {
	block   *types.Block
	results chan<- *types.Block
}

// mineResult wraps the pow solution parameters for the specified block.
type mineResult struct {
	nonce types.BlockNonce
	hash  common.Hash

	errc chan error
}

// hashrate wraps the hash rate submitted by the remote sealer.
type hashrate struct {
	id   common.Hash
	ping time.Time
	rate uint64

	done chan struct{}
}

// sealWork wraps a seal work package for remote sealer.
type sealWork struct {
	errc chan error
	res  chan [4]string
}

// Ethash is a consensus engine based on proof-of-work implementing the ethash
// algorithm.
type Keccak struct {
	// Mining related fields
	threads  int           // Number of threads to mine on if mining
	update   chan struct{} // Notification channel to update mining parameters
	hashrate metrics.Meter // Meter tracking the average hashrate

	// Remote sealer related fields
	workCh       chan *sealTask   // Notification channel to push new work and relative result channel to remote sealer
	fetchWorkCh  chan *sealWork   // Channel used for remote sealer to fetch mining work
	submitWorkCh chan *mineResult // Channel used for remote sealer to submit their mining result
	fetchRateCh  chan chan uint64 // Channel used to gather submitted hash rate for local or remote sealer.
	submitRateCh chan *hashrate   // Channel used for remote sealer to submit their mining hashrate

	// The fields below are hooks for testing
	shared    *Keccak       // Shared PoW verifier to avoid cache regeneration
	fakeFail  uint64        // Block number which fails PoW check even in fake mode
	fakeDelay time.Duration // Time delay to sleep for before returning from verify

	lock      sync.Mutex      // Ensures thread safety for the in-memory caches and mining fields
	closeOnce sync.Once       // Ensures exit channel will not be closed twice.
	exitCh    chan chan error // Notification channel to exiting backend threads
}

// New creates a full sized ethash PoW scheme and starts a background thread for
// remote mining, also optionally notifying a batch of remote services of new work
// packages.
func New(notify []string, noverify bool) *Keccak {
	keccak := &Keccak{
		update:       make(chan struct{}),
		hashrate:     metrics.NewMeterForced(),
		workCh:       make(chan *sealTask),
		fetchWorkCh:  make(chan *sealWork),
		submitWorkCh: make(chan *mineResult),
		fetchRateCh:  make(chan chan uint64),
		submitRateCh: make(chan *hashrate),
		exitCh:       make(chan chan error),
	}
	go keccak.remote(notify, noverify)
	return keccak
}

// NewTester creates a small sized ethash PoW scheme useful only for testing
// purposes.
func NewTester(notify []string, noverify bool) *Keccak {
	keccak := &Keccak{
		update:       make(chan struct{}),
		hashrate:     metrics.NewMeterForced(),
		workCh:       make(chan *sealTask),
		fetchWorkCh:  make(chan *sealWork),
		submitWorkCh: make(chan *mineResult),
		fetchRateCh:  make(chan chan uint64),
		submitRateCh: make(chan *hashrate),
		exitCh:       make(chan chan error),
	}
	go keccak.remote(notify, noverify)
	return keccak
}

// Close closes the exit channel to notify all backend threads exiting.
func (keccak *Keccak) Close() error {
	var err error
	keccak.closeOnce.Do(func() {
		// Short circuit if the exit channel is not allocated.
		if keccak.exitCh == nil {
			return
		}
		errc := make(chan error)
		keccak.exitCh <- errc
		err = <-errc
		close(keccak.exitCh)
	})
	return err
}

// cache tries to retrieve a verification cache for the specified block number
// by first checking against a list of in-memory caches, then against caches
// stored on disk, and finally generating one if none can be found.
// func (keccak *Keccak) cache(block uint64) *cache {
// 	epoch := block / epochLength
// 	currentI, futureI := ethash.caches.get(epoch)
// 	current := currentI.(*cache)

// 	// Wait for generation finish.
// 	current.generate(ethash.config.CacheDir, ethash.config.CachesOnDisk, ethash.config.PowMode == ModeTest)

// 	// If we need a new future cache, now's a good time to regenerate it.
// 	if futureI != nil {
// 		future := futureI.(*cache)
// 		go future.generate(ethash.config.CacheDir, ethash.config.CachesOnDisk, ethash.config.PowMode == ModeTest)
// 	}
// 	return current
// }

// dataset tries to retrieve a mining dataset for the specified block number
// by first checking against a list of in-memory datasets, then against DAGs
// stored on disk, and finally generating one if none can be found.
//
// If async is specified, not only the future but the current DAG is also
// generates on a background thread.
// func (ethash *Ethash) dataset(block uint64, async bool) *dataset {
// 	// Retrieve the requested ethash dataset
// 	epoch := block / epochLength
// 	currentI, futureI := ethash.datasets.get(epoch)
// 	current := currentI.(*dataset)

// 	// If async is specified, generate everything in a background thread
// 	if async && !current.generated() {
// 		go func() {
// 			current.generate(ethash.config.DatasetDir, ethash.config.DatasetsOnDisk, ethash.config.PowMode == ModeTest)

// 			if futureI != nil {
// 				future := futureI.(*dataset)
// 				future.generate(ethash.config.DatasetDir, ethash.config.DatasetsOnDisk, ethash.config.PowMode == ModeTest)
// 			}
// 		}()
// 	} else {
// 		// Either blocking generation was requested, or already done
// 		current.generate(ethash.config.DatasetDir, ethash.config.DatasetsOnDisk, ethash.config.PowMode == ModeTest)

// 		if futureI != nil {
// 			future := futureI.(*dataset)
// 			go future.generate(ethash.config.DatasetDir, ethash.config.DatasetsOnDisk, ethash.config.PowMode == ModeTest)
// 		}
// 	}
// 	return current
// }

// Threads returns the number of mining threads currently enabled. This doesn't
// necessarily mean that mining is running!
func (keccak *Keccak) Threads() int {
	keccak.lock.Lock()
	defer keccak.lock.Unlock()

	return keccak.threads
}

// SetThreads updates the number of mining threads currently enabled. Calling
// this method does not start mining, only sets the thread count. If zero is
// specified, the miner will use all cores of the machine. Setting a thread
// count below zero is allowed and will cause the miner to idle, without any
// work being done.
func (keccak *Keccak) SetThreads(threads int) {
	keccak.lock.Lock()
	defer keccak.lock.Unlock()

	// If we're running a shared PoW, set the thread count on that instead
	if keccak.shared != nil {
		keccak.shared.SetThreads(threads)
		return
	}
	// Update the threads and ping any running seal to pull in any changes
	keccak.threads = threads
	select {
	case keccak.update <- struct{}{}:
	default:
	}
}

// Hashrate implements PoW, returning the measured rate of the search invocations
// per second over the last minute.
// Note the returned hashrate includes local hashrate, but also includes the total
// hashrate of all remote miner.
func (keccak *Keccak) Hashrate() float64 {
	// Short circuit if we are run the ethash in normal/test mode.
	var res = make(chan uint64, 1)

	select {
	case keccak.fetchRateCh <- res:
	case <-keccak.exitCh:
		// Return local hashrate only if ethash is stopped.
		return keccak.hashrate.Rate1()
	}

	// Gather total submitted hash rate of remote sealers.
	return keccak.hashrate.Rate1() + float64(<-res)
}

// APIs implements consensus.Engine, returning the user facing RPC APIs.
func (keccak *Keccak) APIs(chain consensus.ChainReader) []rpc.API {
	// In order to ensure backward compatibility, we exposes ethash RPC APIs
	// to both eth and ethash namespaces.
	return []rpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   &API{keccak},
			Public:    true,
		},
		{
			Namespace: "keccak",
			Version:   "1.0",
			Service:   &API{keccak},
			Public:    true,
		},
	}
}

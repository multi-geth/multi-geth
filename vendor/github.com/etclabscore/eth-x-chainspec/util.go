// Package xchain @utils.go contains utilites for data structure manipulation.
// At the time of writing these are pertinent only to Parity data values, and
// might move to that package at some point.

package xchain

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Uint64 uint64

func (u *Uint64) UnmarshalJSON(input []byte) error {
	if input[0] == '"' {
		uq, err := strconv.Unquote(strings.ToLower(string(input)))
		if err != nil {
			return err
		}
		input = []byte(uq)
	}
	ui, err := strconv.ParseUint(string(input), 0, 64)
	if err != nil {
		return err
	}
	*u = Uint64(ui)
	return nil

	return nil
}

func (u *Uint64) MarshalJSON() ([]byte, error) {
	x := hexutil.Uint64(uint64(*u))
	return json.Marshal(x)
}

// MarshalText implements encoding.TextMarshaler.
func (u Uint64) MarshalText() ([]byte, error) {
	return hexutil.Uint64(u).MarshalText()
}

func (u *Uint64) Big() *big.Int {
	if u == nil {
		return nil
	}
	return new(big.Int).SetUint64(uint64(*u))
}

func (u *Uint64) Uint64() uint64 {
	if u == nil {
		return 0
	}
	return uint64(*u)
}

func FromUint64(i uint64) *Uint64 {
	u := Uint64(i)
	return &u
}

type BlockReward map[Uint64]*hexutil.Big

func (br *BlockReward) UnmarshalJSON(input []byte) error {
	if input[0] != '{' {
		sinput, err := strconv.Unquote(string(input))
		if err != nil {
			return err
		}
		bb, err := hexutil.DecodeBig(sinput)
		if err != nil {
			return err
		}
		*br = BlockReward{Uint64(0): (*hexutil.Big)(bb)}
		return nil
	}

	type BlockRewardMap map[string]string
	m := BlockRewardMap{}
	if err := json.Unmarshal(input, &m); err != nil {
		return err
	}
	var bbr = BlockReward{}
	for k, v := range m {
		var u Uint64
		err := u.UnmarshalJSON([]byte(k))
		if err != nil {
			return err
		}
		var hb = new(hexutil.Big)
		err = hb.UnmarshalJSON([]byte(strconv.Quote(v)))
		if err != nil {
			return err
		}
		bbr[u] = hb
	}
	*br = bbr
	return nil
}

type BTreeMap map[Uint64]*Uint64

func (btm *BTreeMap) UnmarshalJSON(input []byte) error {
	type IntermediateMap map[string]interface{}
	// m := make(map[string]interface{})
	m := IntermediateMap{}
	err := json.Unmarshal(input, &m)
	if err != nil {
		return err
	}
	var bbtm = BTreeMap{}
	for k, v := range m {
		var ku Uint64
		err := ku.UnmarshalJSON([]byte(k))
		if err != nil {
			return err
		}
		var vu Uint64
		vv, ok := v.(float64)
		if ok {
			vu = Uint64(vv)
		} else {
			vs, ok := v.(string)
			if ok {
				err = vu.UnmarshalJSON([]byte(vs))
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("could not assert btree map type")
			}
		}
		bbtm[ku] = &vu
	}
	*btm = bbtm
	return nil
}

// A BlockNonce is a 64-bit hash which proves (combined with the
// mix-hash) that a sufficient amount of computation has been carried
// out on a block.
type BlockNonce [8]byte

// EncodeNonce converts the given integer to a block nonce.
func EncodeNonce(i uint64) BlockNonce {
	var n BlockNonce
	binary.BigEndian.PutUint64(n[:], i)
	return n
}

// Uint64 returns the integer value of a block nonce.
func (n BlockNonce) Uint64() uint64 {
	return binary.BigEndian.Uint64(n[:])
}

// MarshalText encodes n as a hex string with 0x prefix.
func (n BlockNonce) MarshalText() ([]byte, error) {
	return hexutil.Bytes(n[:]).MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (n *BlockNonce) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("BlockNonce", input, n[:])
}

// ParseBig256 parses s as a 256 bit integer in decimal or hexadecimal syntax.
// Leading zeros are accepted. The empty string parses as zero.
func ParseBig256(s string) (*big.Int, bool) {
	if s == "" {
		return new(big.Int), true
	}
	var bigint *big.Int
	var ok bool
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		bigint, ok = new(big.Int).SetString(s[2:], 16)
	} else {
		bigint, ok = new(big.Int).SetString(s, 10)
	}
	if ok && bigint.BitLen() > 256 {
		bigint, ok = nil, false
	}
	return bigint, ok
}

type ConfigAccountNonce uint64

func (n *ConfigAccountNonce) UnmarshalJSON(input []byte) error {
	if input[0] == '"' {
		uq, err := strconv.Unquote(string(input))
		if err != nil {
			return err
		}
		input = []byte(uq)
	}
	if strings.Contains(string(input), "x") {
		uu, err := hexutil.DecodeUint64(string(input))
		if err != nil {
			return err
		}
		*n = ConfigAccountNonce(uu)
		return nil
	}
	u, err := strconv.ParseUint(string(input), 10, 64)
	if err != nil {
		return err
	}
	*n = ConfigAccountNonce(u)
	return nil
}

func (n ConfigAccountNonce) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%d"`, uint64(n))), nil
}

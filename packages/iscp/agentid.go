// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscp

import (
	"encoding/hex"
	"strings"

	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/hive.go/serializer"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/mr-tron/base58"
	"golang.org/x/xerrors"
)

// AgentID represents address on the ledger with optional hname
// If address is and alias address and hname is != 0 the agent id is interpreted as
// ad contract id
type AgentID struct {
	a iotago.Address
	h Hname
}

var NilAgentID AgentID

func init() {
	NilAgentID = AgentID{
		a: nil,
		h: 0,
	}
}

// NewAgentID makes new AgentID
func NewAgentID(addr iotago.Address, hname Hname) *AgentID {
	return &AgentID{
		a: addr,
		h: hname,
	}
}

// TODO move somewhere else
func AddressFromMatshalUtil(mu *marshalutil.MarshalUtil) (iotago.Address, error) {
	typeByte, err := mu.ReadByte()
	if err != nil {
		return nil, err
	}
	addr, err := iotago.AddressSelector(uint32(typeByte))
	if err != nil {
		return nil, err
	}
	length, err := addr.Deserialize(mu.Bytes(), serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil, err
	}
	mu.ReadSeek(mu.ReadOffset() + length)
	return addr, nil
}

func AgentIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (*AgentID, error) {
	var err error
	ret := &AgentID{}
	if ret.a, err = AddressFromMatshalUtil(mu); err != nil {
		return nil, err
	}
	if ret.h, err = HnameFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	return ret, nil
}

func AgentIDFromBytes(data []byte) (*AgentID, error) {
	return AgentIDFromMarshalUtil(marshalutil.New(data))
}

func NewAgentIDFromBase58EncodedString(s string) (*AgentID, error) {
	data, err := base58.Decode(s)
	if err != nil {
		return nil, err
	}
	return AgentIDFromBytes(data)
}

// NewAgentIDFromString parses the human-readable string representation
func NewAgentIDFromString(s string) (*AgentID, error) {
	if len(s) < 2 {
		return nil, xerrors.New("NewAgentIDFromString: invalid length")
	}
	if s[:2] != "A/" {
		return nil, xerrors.New("NewAgentIDFromString: wrong prefix")
	}
	parts := strings.Split(s[2:], "::")
	if len(parts) != 2 {
		return nil, xerrors.New("NewAgentIDFromString: wrong format")
	}
	addrBytes, err := hex.DecodeString(parts[0])
	if err != nil {
		return nil, xerrors.Errorf("NewAgentIDFromString: %v", err)
	}
	addr, err := AddressFromMatshalUtil(marshalutil.New(addrBytes))
	if err != nil {
		return nil, xerrors.Errorf("NewAgentIDFromString: %v", err)
	}
	hname, err := HnameFromString(parts[1])
	if err != nil {
		return nil, xerrors.Errorf("NewAgentIDFromString: %v", err)
	}
	return NewAgentID(addr, hname), nil
}

// NewRandomAgentID creates random AgentID
func NewRandomAgentID() *AgentID {
	raddr := RandomChainID()
	return NewAgentID(raddr.AsAddress(), Hn("testName"))
}

func (a *AgentID) Address() iotago.Address {
	return a.a
}

func (a *AgentID) Hname() Hname {
	return a.h
}

func (a *AgentID) Bytes() []byte {
	if a.a == nil {
		panic("AgentID.Bytes: address == nil")
	}
	addressBytes, err := a.a.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil
	}
	mu := marshalutil.New()
	mu.WriteBytes(addressBytes)
	mu.Write(a.h)
	return mu.Bytes()
}

func (a *AgentID) Equals(a1 *AgentID) bool {
	if !a.a.Equal(a1.a) {
		return false
	}
	if a.h != a1.h {
		return false
	}
	return true
}

// String human readable string
func (a *AgentID) String() string {
	return "A/" + a.a.String() + "::" + a.h.String()
}

func (a *AgentID) Base58() string {
	return base58.Encode(a.Bytes())
}

func (a *AgentID) IsNil() bool {
	return a.Equals(&NilAgentID)
}

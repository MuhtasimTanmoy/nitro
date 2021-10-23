//
// Copyright 2021, Offchain Labs, Inc. All rights reserved.
//

package precompiles

import (
	"errors"
	"math/big"
)

type ArbSys struct {
	Address           addr
	L2ToL1Transaction func(*stateDB, addr, addr, huge, huge, huge, huge, huge, huge, huge, []byte)
}

func (con *ArbSys) ArbBlockNumber(caller addr, st *stateDB) (huge, error) {
	return nil, errors.New("unimplemented")
}

func (con *ArbSys) ArbBlockNumberGasCost() uint64 {
	return 0
}

func (con *ArbSys) ArbChainID(caller addr, st *stateDB) (huge, error) {
	return big.NewInt(412345), nil
}

func (con *ArbSys) ArbChainIDGasCost() uint64 {
	return 0
}

func (con *ArbSys) ArbOSVersion(caller addr) (huge, error) {
	return big.NewInt(1000), nil
}

func (con *ArbSys) ArbOSVersionGasCost() uint64 {
	return 0
}

func (con *ArbSys) GetStorageAt(caller addr, st *stateDB, address addr, index huge) (huge, error) {
	return nil, errors.New("unimplemented")
}

func (con *ArbSys) GetStorageAtGasCost(address addr, index huge) uint64 {
	return 0
}

func (con *ArbSys) GetStorageGasAvailable(caller addr, st *stateDB) (huge, error) {
	return nil, errors.New("unimplemented")
}

func (con *ArbSys) GetStorageGasAvailableGasCost() uint64 {
	return 0
}

func (con *ArbSys) GetTransactionCount(caller addr, st *stateDB, account addr) (huge, error) {
	return nil, errors.New("unimplemented")
}

func (con *ArbSys) GetTransactionCountGasCost(account addr) uint64 {
	return 0
}

func (con *ArbSys) IsTopLevelCall(caller addr, st *stateDB) (bool, error) {
	return false, errors.New("unimplemented")
}

func (con *ArbSys) IsTopLevelCallGasCost() uint64 {
	return 0
}

func (con *ArbSys) MapL1SenderContractAddressToL2Alias(
	caller addr,
	sender addr,
	dest addr,
) (addr, error) {
	return addr{}, errors.New("unimplemented")
}

func (con *ArbSys) MapL1SenderContractAddressToL2AliasGasCost(sender addr, dest addr) uint64 {
	return 0
}

func (con *ArbSys) MyCallersAddressWithoutAliasing(caller addr, st *stateDB) (addr, error) {
	return addr{}, errors.New("unimplemented")
}

func (con *ArbSys) MyCallersAddressWithoutAliasingGasCost() uint64 {
	return 0
}

func (con *ArbSys) SendTxToL1(
	caller addr,
	st *stateDB,
	value huge,
	destination addr,
	calldataForL1 []byte,
) (huge, error) {
	return nil, errors.New("unimplemented")
}

func (con *ArbSys) SendTxToL1GasCost(destination addr, calldataForL1 []byte) uint64 {
	return 0
}

func (con *ArbSys) WasMyCallersAddressAliased(caller addr, st *stateDB) (bool, error) {
	return false, errors.New("unimplemented")
}

func (con *ArbSys) WasMyCallersAddressAliasedGasCost() uint64 {
	return 0
}

func (con *ArbSys) WithdrawEth(caller addr, st *stateDB, value huge, destination addr) (huge, error) {
	return nil, errors.New("unimplemented")
}

func (con *ArbSys) WithdrawEthGasCost(destination addr) uint64 {
	return 0
}

package avsregistry

import (
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	blsapkreg "github.com/Layr-Labs/eigensdk-go/contracts/bindings/BLSApkRegistry"
	regcoord "github.com/Layr-Labs/eigensdk-go/contracts/bindings/RegistryCoordinator"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/types"
)

type AvsRegistrySubscriber interface {
	SubscribeToNewPubkeyRegistrations() (chan *blsapkreg.ContractBLSApkRegistryNewPubkeyRegistration, event.Subscription, error)
	SubscribeToOperatorSocketUpdates() (chan *regcoord.ContractZrRegistryCoordinatorOperatorSocketUpdate, event.Subscription, error)
}

type AvsRegistryChainSubscriber struct {
	logger         logging.Logger
	regCoord       *regcoord.ContractZrRegistryCoordinator
	blsApkRegistry blsapkreg.ContractBLSApkRegistryFilters
}

// forces EthSubscriber to implement the chainio.Subscriber interface
var _ AvsRegistrySubscriber = (*AvsRegistryChainSubscriber)(nil)

func NewAvsRegistryChainSubscriber(
	logger logging.Logger,
	regCoord *regcoord.ContractZrRegistryCoordinator,
	blsApkRegistry blsapkreg.ContractBLSApkRegistryFilters,
) (*AvsRegistryChainSubscriber, error) {
	return &AvsRegistryChainSubscriber{
		logger:         logger,
		regCoord:       regCoord,
		blsApkRegistry: blsApkRegistry,
	}, nil
}

func BuildAvsRegistryChainSubscriber(
	regCoordAddr common.Address,
	ethWsClient eth.Client,
	logger logging.Logger,
) (*AvsRegistryChainSubscriber, error) {
	regCoord, err := regcoord.NewContractZrRegistryCoordinator(regCoordAddr, ethWsClient)
	if err != nil {
		return nil, types.WrapError(errors.New("Failed to create RegistryCoordinator contract"), err)
	}
	blsApkRegAddr, err := regCoord.BlsApkRegistry(&bind.CallOpts{})
	if err != nil {
		return nil, types.WrapError(errors.New("Failed to get BLSApkRegistry address from RegistryCoordinator"), err)
	}
	blsApkReg, err := blsapkreg.NewContractBLSApkRegistry(blsApkRegAddr, ethWsClient)
	if err != nil {
		return nil, types.WrapError(errors.New("Failed to create BLSApkRegistry contract"), err)
	}
	return NewAvsRegistryChainSubscriber(logger, regCoord, blsApkReg)
}

func (s *AvsRegistryChainSubscriber) SubscribeToNewPubkeyRegistrations() (chan *blsapkreg.ContractBLSApkRegistryNewPubkeyRegistration, event.Subscription, error) {
	newPubkeyRegistrationChan := make(chan *blsapkreg.ContractBLSApkRegistryNewPubkeyRegistration)
	sub, err := s.blsApkRegistry.WatchNewPubkeyRegistration(
		&bind.WatchOpts{}, newPubkeyRegistrationChan, nil,
	)
	if err != nil {
		return nil, nil, types.WrapError(errors.New("Failed to subscribe to NewPubkeyRegistration events"), err)
	}
	return newPubkeyRegistrationChan, sub, nil
}

func (s *AvsRegistryChainSubscriber) SubscribeToOperatorSocketUpdates() (chan *regcoord.ContractZrRegistryCoordinatorOperatorSocketUpdate, event.Subscription, error) {
	operatorSocketUpdateChan := make(chan *regcoord.ContractZrRegistryCoordinatorOperatorSocketUpdate)
	sub, err := s.regCoord.WatchOperatorSocketUpdate(
		&bind.WatchOpts{}, operatorSocketUpdateChan, nil,
	)
	if err != nil {
		return nil, nil, types.WrapError(errors.New("Failed to subscribe to OperatorSocketUpdate events"), err)
	}
	return operatorSocketUpdateChan, sub, nil
}

package contract

import (
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// Echo the argument back as the response
func (s *SmartContract) Echo(ctx contractapi.TransactionContextInterface, arg string) (string, error) {
	return arg, nil
}

// Put a value for a given ledger key and return the value
func (s *SmartContract) Put(ctx contractapi.TransactionContextInterface, key string, value string) (string, error) {
	if err := ctx.GetStub().PutState(key, []byte(value)); err != nil {
		return "", fmt.Errorf("failed to put state to ledger: %w", err)
	}

	return value, nil
}

// Get the value for a given ledger key
func (s *SmartContract) Get(ctx contractapi.TransactionContextInterface, key string) (string, error) {
	value, err := ctx.GetStub().GetState(key)
	if err != nil {
		return "", fmt.Errorf("failed to get state from ledger: %w", err)
	}

	return string(value), nil
}

/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type serverConfig struct {
	CCID    string
	Address string
}

// SmartContract provides functions for managing an asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
type PointTransaction struct {
	Owner     string `json:"Owner"`
	Value     string `json:"value"`
	Merchant  string `json:"merchant"`
	Status    string `json:"status"` // [pending, archived, birthday, campaign, adjustment, bouns, lifecard]
	CreatedAt string `json:"created_at"`

	Order    string `json:"order"`
	Stockist string `json:"stockist"`
	Campaign string `json:"campaign"`
	Giftee   string `json:"giftee"`
	Gifter   string `json:"gifter"`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Key    string `json:"Key"`
	Record *PointTransaction
}

// InitLedger adds a base set of point transations to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	transations := []PointTransaction{
		PointTransaction{Owner: "Xiaoming", Value: "300", Merchant: "zh-CN", Status: "Birthday", CreatedAt: "20210929"},
	}

	for i, transation := range transations {
		transationAsBytes, _ := json.Marshal(transation)
		err := ctx.GetStub().PutState("Transation"+strconv.Itoa(i), transationAsBytes)

		if err != nil {
			return fmt.Errorf("Failed to put to world state. %s", err.Error())
		}
	}

	return nil
}

func (s *SmartContract) CreateBirthdayTransation(ctx contractapi.TransactionContextInterface, transationId string, owner string, value string, merchant string, createdAt string) error {
	PointTransaction := PointTransaction{
		Owner:     owner,
		Value:     value,
		Merchant:  merchant,
		CreatedAt: createdAt,
		Status: "Birthday",
	}

	transationAsBytes, _ := json.Marshal(PointTransaction)

	return ctx.GetStub().PutState(transationId, transationAsBytes)
}

func (s *SmartContract) CreateOrderTransaction(ctx contractapi.TransactionContextInterface, transationId string, owner string, value string, merchant string, createdAt string, order string, stockist string) error {
	PointTransaction := PointTransaction{
		Owner:     owner,
		Value:     value,
		Merchant:  merchant,
		CreatedAt: createdAt,
		Order:     order,
		Stockist:  stockist,
		Status: "Order",
	}

	transationAsBytes, _ := json.Marshal(PointTransaction)

	return ctx.GetStub().PutState(transationId, transationAsBytes)
}

func (s *SmartContract) CreateCampaignTransation(ctx contractapi.TransactionContextInterface, transationId string, owner string, value string, merchant string, createdAt string, campaign string) error {
	PointTransaction := PointTransaction{
		Owner:     owner,
		Value:     value,
		Merchant:  merchant,
		CreatedAt: createdAt,
		Campaign:  campaign,
		Status: "Campaign",
	}

	transationAsBytes, _ := json.Marshal(PointTransaction)

	return ctx.GetStub().PutState(transationId, transationAsBytes)
}

func (s *SmartContract) QueryTransation(ctx contractapi.TransactionContextInterface, id string) (*PointTransaction, error) {
	transactionJson, err := ctx.GetStub().GetState(id)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if transactionJson == nil {
		return nil, fmt.Errorf("%s does not exist in world state", id)
	}

	var transaction PointTransaction
	err = json.Unmarshal(transactionJson, &transaction)
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

func (s *SmartContract) QueryAllTransactions(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []QueryResult{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		pointTransaction := new(PointTransaction)
		err = json.Unmarshal(queryResponse.Value, pointTransaction)
		if err != nil {
			return nil, err
		}

		queryResult := QueryResult{Key: queryResponse.Key, Record: pointTransaction}
		results = append(results, queryResult)
	}

	return results, nil
}

func main() {
	// See chaincode.env.example
	config := serverConfig{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
	}

	chaincode, err := contractapi.NewChaincode(&SmartContract{})

	if err != nil {
		log.Panicf("error create points-transfer chaincode: %s", err)
	}

	server := &shim.ChaincodeServer{
		CCID:    config.CCID,
		Address: config.Address,
		CC:      chaincode,
		TLSProps: getTLSProperties(),
	}

	if err := server.Start(); err != nil {
		log.Panicf("error starting points-transfer chaincode: %s", err)
	}
}

func getTLSProperties() shim.TLSProperties {
	// Check if chaincode is TLS enabled
	tlsDisabledStr := getEnvOrDefault("CHAINCODE_TLS_DISABLED", "true")
	key := getEnvOrDefault("CHAINCODE_TLS_KEY", "")
	cert := getEnvOrDefault("CHAINCODE_TLS_CERT", "")
	clientCACert := getEnvOrDefault("CHAINCODE_CLIENT_CA_CERT", "")

	// convert tlsDisabledStr to boolean
	tlsDisabled := getBoolOrDefault(tlsDisabledStr, false)
	var keyBytes, certBytes, clientCACertBytes []byte
	var err error

	if !tlsDisabled {
		keyBytes, err = ioutil.ReadFile(key)
		if err != nil {
			log.Panicf("error while reading the crypto file: %s", err)
		}
		certBytes, err = ioutil.ReadFile(cert)
		if err != nil {
			log.Panicf("error while reading the crypto file: %s", err)
		}
	}
	// Did not request for the peer cert verification
	if clientCACert != "" {
		clientCACertBytes, err = ioutil.ReadFile(clientCACert)
		if err != nil {
			log.Panicf("error while reading the crypto file: %s", err)
		}
	}

	return shim.TLSProperties{
		Disabled: tlsDisabled,
		Key: keyBytes,
		Cert: certBytes,
		ClientCACerts: clientCACertBytes,
	}
}

func getEnvOrDefault(env, defaultVal string) string {
	value, ok := os.LookupEnv(env)
	if !ok {
		value = defaultVal
	}
	return value
}

// Note that the method returns default value if the string
// cannot be parsed!
func getBoolOrDefault(value string, defaultVal bool) bool {
	parsed, err := strconv.ParseBool(value)
	if err!= nil {
		return defaultVal
	}
	return parsed
}

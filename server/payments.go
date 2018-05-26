package server

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/RTradeLtd/Temporal/bindings/payments"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// NewPaymentsContract is used to generate a new payment contract handler
func (sm *ServerManager) NewPaymentsContract(address common.Address) error {
	contract, err := payments.NewPayments(address, sm.Client)
	if err != nil {
		return err
	}
	sm.PaymentsContract = contract
	return nil
}

// RegisterPaymentForUploader is used to register a payment for the given uploader
func (sm *ServerManager) RegisterPaymentForUploader(uploaderAddress string, contentHash string, retentionPeriodInMonths *big.Int, chargeAmountInWei *big.Int, method uint8) (*types.Transaction, error) {
	if method > 1 || method < 0 {
		return nil, errors.New("invalid payment method. 0 = RTC, 1 = ETH")
	}
	// since the contract function defines a fixed length byte array, we will need to convert a byte slice before calling the function
	var b [32]byte
	// convert hash to byte slice
	data := []byte(contentHash)
	// runs the byte slice through keccak256
	hashedCIDByte := crypto.Keccak256(data)
	// convert to hash
	hashedCID := common.BytesToHash(hashedCIDByte)
	// convert byte slice to byte array
	copy(b[:], hashedCID.Bytes()[:32])
	sm.Auth.GasPrice = big.NewInt(int64(22000000000))
	// call the register payments function, indicating that we are expecting a user to upload the particular file
	// we hash the content identifier hash before submitting to the blockchain such that we preserve users privacy
	// but allow them to audit the contracts and data stores themselves, by hashing their plaintext content identifier hashes
	tx, err := sm.PaymentsContract.RegisterPayment(sm.Auth, common.HexToAddress(uploaderAddress), b, retentionPeriodInMonths, chargeAmountInWei, method)
	if err != nil {
		return nil, err
	}
	sm.RegisterWaitForAndProcessPaymentsReceivedEventForAddress(uploaderAddress, contentHash)

	return tx, nil
}

func (sm *ServerManager) RegisterWaitForAndProcessPaymentsReceivedEventForAddress(address string, cid string) {
	var processed bool
	var ch = make(chan *payments.PaymentsPaymentRegistered)
	sub, err := sm.PaymentsContract.WatchPaymentRegistered(nil, ch, []common.Address{common.HexToAddress(address)})
	if err != nil {
		log.Fatal(err)
	}
	queueManager, err := queue.Initialize(queue.PaymentRegisterQueue)
	if err != nil {
		log.Fatal(err)
	}
	for {
		if processed {
			break
		}
		select {
		case err := <-sub.Err():
			fmt.Println("Error parsing event ", err)
			log.Fatal(err)
		case evLog := <-ch:
			pr := queue.PaymentRegister{}
			uploader := evLog.Uploader
			hashedCID := evLog.HashedCID
			paymentID := evLog.PaymentID
			pr.UploaderAddress = uploader.String()
			pr.CID = cid
			pr.HashedCID = fmt.Sprintf("%s", hex.EncodeToString(hashedCID[:]))
			pr.PaymentID = fmt.Sprintf("0x%s", hex.EncodeToString(paymentID[:]))
			queueManager.PublishMessage(pr)
			processed = true
			break
		}
	}
}

func (sm *ServerManager) WaitForAndProcessPaymentsReceivedEvent() {
	// create the channel for which we will receive payments on
	var ch = make(chan *payments.PaymentsPaymentReceivedNoIndex)
	// create a subscription for th eevent passing in messages to teh chanenl we just established
	sub, err := sm.PaymentsContract.WatchPaymentReceivedNoIndex(nil, ch)
	if err != nil {
		log.Fatal(err)
	}

	queueManager, err := queue.Initialize(queue.PaymentReceivedQueue)
	if err != nil {
		log.Fatal(err)
	}
	// loop forever, waiting for and processing events
	for {
		select {
		case err := <-sub.Err():
			fmt.Println("Error parsing event ", err)
		case evLog := <-ch:
			uploader := evLog.Uploader
			paymentID := evLog.PaymentID
			pr := queue.PaymentReceived{}
			pr.UploaderAddress = uploader.String()
			pr.PaymentID = string(paymentID[:])
			queueManager.PublishMessage(pr)
		}
	}
}
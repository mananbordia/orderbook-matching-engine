package dex

import (
	"encoding/json"
	"errors"
	"fmt"

	. "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/sha3"
)

// OrderCancel is a group of params used for canceling an order previously
// sent to the matching engine. The OrderId and OrderHash must correspond to the
// same order. To be valid and be able to be processed by the matching engine,
// the OrderCancel must include a signature by the Maker of the order corresponding
// to the OrderHash.
type OrderCancel struct {
	OrderId   uint64     `json:"orderId"`
	PairID    Hash       `json:"pair"`
	OrderHash Hash       `json:"orderHash"`
	Hash      Hash       `json:"hash"`
	Signature *Signature `json:"signature"`
}

// NewOrderCancel returns a new empty OrderCancel object
func NewOrderCancel() *OrderCancel {
	return &OrderCancel{
		OrderId:   0,
		PairID:    Hash{},
		Hash:      Hash{},
		OrderHash: Hash{},
		Signature: &Signature{},
	}
}

// MarshalJSON returns the json encoded byte array representing the OrderCancel struct
func (oc *OrderCancel) MarshalJSON() ([]byte, error) {
	orderCancel := map[string]interface{}{
		"orderId":   oc.OrderId,
		"orderHash": oc.OrderHash,
		"pairID":    oc.PairID,
		"hash":      oc.Hash,
		"signature": map[string]interface{}{
			"V": oc.Signature.V,
			"R": oc.Signature.R,
			"S": oc.Signature.S,
		},
	}

	return json.Marshal(orderCancel)
}

func (oc *OrderCancel) String() string {
	return fmt.Sprintf("\nOrderCancel:\norderId: %d\nOrderHash: %x\nPairID: %x\nHash: %x\nSignature.V: %x\nSignature.R: %x\nSignature.S: %x\n\n",
		oc.OrderId, oc.OrderHash, oc.PairID, oc.Hash, oc.Signature.V, oc.Signature.R, oc.Signature.S)
}

// UnmarshalJSON creates an OrderCancel object from a json byte string
func (oc *OrderCancel) UnmarshalJSON(b []byte) error {
	parsed := map[string]interface{}{}

	err := json.Unmarshal(b, &parsed)
	if err != nil {
		return err
	}

	if parsed["orderId"] == nil {
		return errors.New("Order Id missing")
	}
	oc.OrderId = uint64(parsed["orderId"].(float64))
	// oc.OrderId, _ = strconv.ParseUint(parsed["orderId"].(string), 10, 64)

	if parsed["pairID"] == nil {
		return errors.New("Pair ID is missing")
	}
	oc.PairID = HexToHash(parsed["pairID"].(string))

	if parsed["orderHash"] == nil {
		return errors.New("Order Hash is missing")
	}
	oc.OrderHash = HexToHash(parsed["orderHash"].(string))

	if parsed["hash"] == nil {
		return errors.New("Hash is missing")
	}
	oc.Hash = HexToHash(parsed["hash"].(string))

	sig := parsed["signature"].(map[string]interface{})
	oc.Signature = &Signature{
		V: byte(sig["V"].(float64)),
		R: HexToHash(sig["R"].(string)),
		S: HexToHash(sig["S"].(string)),
	}

	return nil
}

// Decode takes a payload previously unmarshalled from a JSON byte string
// and decodes it into an OrderCancel object
func (oc *OrderCancel) Decode(orderCancel map[string]interface{}) error {
	if orderCancel["orderId"] == nil {
		return errors.New("Order Id missing")
	}
	oc.OrderId = uint64(orderCancel["orderId"].(float64))

	if orderCancel["pairID"] == nil {
		return errors.New("Pair ID is missing")
	}
	oc.PairID = HexToHash(orderCancel["pairID"].(string))

	if orderCancel["orderHash"] == nil {
		return errors.New("Order Hash is missing")
	}
	oc.OrderHash = HexToHash(orderCancel["orderHash"].(string))

	if orderCancel["hash"] == nil {
		return errors.New("Order Cancel Hash is missing")
	}
	oc.Hash = HexToHash(orderCancel["hash"].(string))

	sig := orderCancel["signature"].(map[string]interface{})
	oc.Signature = &Signature{
		V: byte(sig["V"].(float64)),
		R: HexToHash(sig["R"].(string)),
		S: HexToHash(sig["S"].(string)),
	}

	return nil
}

// VerifySignature returns a true value if the OrderCancel object signature
// corresponds to the Maker of the given order
func (oc *OrderCancel) VerifySignature(o *Order) (bool, error) {
	message := crypto.Keccak256(
		[]byte("\x19Ethereum Signed Message:\n32"),
		oc.Hash.Bytes(),
	)

	address, err := oc.Signature.Verify(BytesToHash(message))
	if err != nil {
		return false, err
	}

	if address != o.Maker {
		return false, errors.New("Recovered address is incorrect")
	}

	return true, nil
}

func (oc *OrderCancel) ComputeHash() Hash {
	sha := sha3.NewKeccak256()
	sha.Write(oc.OrderHash.Bytes())
	return BytesToHash(sha.Sum(nil))
}

// Sign first computes the order cancel hash, then signs and sets the signature
func (oc *OrderCancel) Sign(w *Wallet) error {
	h := oc.ComputeHash()
	sig, err := w.SignHash(h)
	if err != nil {
		return err
	}

	oc.Hash = h
	oc.Signature = sig
	return nil
}

package visor

import (
	"log"

	"encoding/json"
	"errors"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/coin"
)

// Encapsulates useful information from the coin.Blockchain
type BlockchainMetadata struct {
	// Most recent block's header
	Head ReadableBlockHeader `json:"head"`
	// Number of unspent outputs in the coin.Blockchain
	Unspents uint64 `json:"unspents"`
	// Number of known unconfirmed txns
	Unconfirmed uint64 `json:"unconfirmed"`
}

func NewBlockchainMetadata(v *Visor) BlockchainMetadata {
	head := v.Blockchain.Head().Head
	return BlockchainMetadata{
		Head:        NewReadableBlockHeader(&head),
		Unspents:    uint64(len(v.Blockchain.Unspent.Pool)),
		Unconfirmed: uint64(len(v.Unconfirmed.Txns)),
	}
}

// Wrapper around coin.Transaction, tagged with its status.  This allows us
// to include unconfirmed txns
type Transaction struct {
	Txn    coin.Transaction  `json:"txn"`
	Status TransactionStatus `json:"status"`
}

type TransactionStatus struct {
	// This txn is in the unconfirmed pool
	Unconfirmed bool `json:"unconfirmed"`
	// We can't find anything about this txn.  Be aware that the txn may be
	// in someone else's unconfirmed pool, and if valid, it may become a
	// confirmed txn in the future
	Unknown   bool `json:"unknown"`
	Confirmed bool `json:"confirmed"`
	// If confirmed, how many blocks deep in the chain it is. Will be at least
	// 1 if confirmed.
	Height uint64 `json:"height"`
}

func NewUnconfirmedTransactionStatus() TransactionStatus {
	return TransactionStatus{
		Unconfirmed: true,
		Unknown:     false,
		Confirmed:   false,
		Height:      0,
	}
}

func NewUnknownTransactionStatus() TransactionStatus {
	return TransactionStatus{
		Unconfirmed: false,
		Unknown:     true,
		Confirmed:   false,
		Height:      0,
	}
}

func NewConfirmedTransactionStatus(height uint64) TransactionStatus {
	if height == 0 {
		log.Panic("Invalid confirmed transaction height")
	}
	return TransactionStatus{
		Unconfirmed: false,
		Unknown:     false,
		Confirmed:   true,
		Height:      height,
	}
}

/*
type ReadableTransactionHeader struct {
	Hash string   `json:"hash"`
	Sigs []string `json:"sigs"`
}

func NewReadableTransactionHeader(t *coin.TransactionHeader) ReadableTransactionHeader {
	sigs := make([]string, len(t.Sigs))
	for i, _ := range t.Sigs {
		sigs[i] = t.Sigs[i].Hex()
	}
	return ReadableTransactionHeader{
		Hash: t.Hash.Hex(),
		Sigs: sigs,
	}
}
*/

type ReadableTransactionOutput struct {
	Address string `json:"dst"`
	Coins   uint64 `json:"coins"`
	Hours   uint64 `json:"hours"`
}

func NewReadableTransactionOutput(t *coin.TransactionOutput) ReadableTransactionOutput {
	return ReadableTransactionOutput{
		Address: t.Address.String(),
		Coins:   t.Coins,
		Hours:   t.Hours,
	}
}

/*
	Outputs
*/

/*
	Add a verbose version
*/
type ReadableOutput struct {
	Hash string `json:"hash"`

	Address string `json:"address"`
	Coins   uint64 `json:"coins"`
	Hours   uint64 `json:"hours"`
}

func NewReadableOutput(t coin.UxOut) ReadableOutput {
	return ReadableOutput{
		Hash:    t.Hash().Hex(),
		Address: t.Body.Address.String(),
		Coins:   t.Body.Coins,
		Hours:   t.Body.Hours,
	}
}

type ReadableTransaction struct {
	Length    uint32 `json:"length"`
	Type      uint8  `json:"type"`
	Hash      string `json:"hash"`
	InnerHash string `json:"inner_hash"`

	Sigs []string                    `json:"sigs"`
	In   []string                    `json:"inputs"`
	Out  []ReadableTransactionOutput `json:"outputs"`
}

func NewReadableTransaction(t *coin.Transaction) ReadableTransaction {

	sigs := make([]string, len(t.Sigs))
	for i, _ := range t.Sigs {
		sigs[i] = t.Sigs[i].Hex()
	}

	in := make([]string, len(t.In))
	for i, _ := range t.In {
		in[i] = t.In[i].Hex()
	}
	out := make([]ReadableTransactionOutput, len(t.Out))
	for i, _ := range t.Out {
		out[i] = NewReadableTransactionOutput(&t.Out[i])
	}
	return ReadableTransaction{
		Length:    t.Length,
		Type:      t.Type,
		Hash:      t.Hash().Hex(),
		InnerHash: t.InnerHash.Hex(),

		Sigs: sigs,
		In:   in,
		Out:  out,
	}
}

type ReadableBlockHeader struct {
	Version  uint32 `json:"version"`
	Time     uint64 `json:"timestamp"`
	BkSeq    uint64 `json:"seq"`
	Fee      uint64 `json:"fee"`
	PrevHash string `json:"prev_hash"`
	BodyHash string `json:"hash"`
}

func NewReadableBlockHeader(b *coin.BlockHeader) ReadableBlockHeader {
	return ReadableBlockHeader{
		Version:  b.Version,
		Time:     b.Time,
		BkSeq:    b.BkSeq,
		Fee:      b.Fee,
		PrevHash: b.PrevHash.Hex(),
		BodyHash: b.BodyHash.Hex(),
	}
}

type ReadableBlockBody struct {
	Transactions []ReadableTransaction `json:"txns"`
}

func NewReadableBlockBody(b *coin.BlockBody) ReadableBlockBody {
	txns := make([]ReadableTransaction, len(b.Transactions))
	for i, _ := range b.Transactions {
		txns[i] = NewReadableTransaction(&b.Transactions[i])
	}
	return ReadableBlockBody{
		Transactions: txns,
	}
}

type ReadableBlock struct {
	Head ReadableBlockHeader `json:"header"`
	Body ReadableBlockBody   `json:"body"`
}

func NewReadableBlock(b *coin.Block) ReadableBlock {
	return ReadableBlock{
		Head: NewReadableBlockHeader(&b.Head),
		Body: NewReadableBlockBody(&b.Body),
	}
}

/*
	Transactions to and from JSON
*/

type TransactionOutputJSON struct {
	Address string `json:"address"` // Address of receiver
	Coins   uint64 `json:"coins"`   // Number of coins
	Hours   uint64 `json:"hours"`   // Coin hours
}

func NewTransactionOutputJSON(ux coin.TransactionOutput) TransactionOutputJSON {
	var o TransactionOutputJSON
	o.Address = ux.Address.String()
	o.Coins = ux.Coins
	o.Hours = ux.Hours
	return o
}

func TransactionOutputFromJSON(in TransactionOutputJSON) (coin.TransactionOutput, error) {
	var tx coin.TransactionOutput

	addr, err := cipher.DecodeBase58Address(in.Address)
	if err != nil {
		return coin.TransactionOutput{}, errors.New("Adress decode fail")
	}
	tx.Address = addr
	tx.Coins = in.Coins
	tx.Hours = in.Hours
	return tx, nil
}

type TransactionJSON struct {
	Hash      string `json:"hash"`
	InnerHash string `json:"inner_hash"`

	Sigs []string                `json:"sigs"`
	In   []string                `json:"in"`
	Out  []TransactionOutputJSON `json:"out"`
}

func TransactionToJSON(tx coin.Transaction) string {

	var o TransactionJSON

	if err := tx.Verify(); err != nil {
		log.Panic("Transaction Invalid: Cannot serialize to JSON")
	}

	o.Sigs = make([]string, len(tx.Sigs))
	o.In = make([]string, len(tx.In))
	o.Out = make([]TransactionOutputJSON, len(tx.Out))

	for i, sig := range tx.Sigs {
		o.Sigs[i] = sig.Hex()
	}
	for i, x := range tx.In {
		o.In[i] = x.Hex() //hash to hex
	}
	for i, y := range tx.Out {
		o.Out[i] = NewTransactionOutputJSON(y)
	}

	b, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		log.Panic("Cannot serialize transaction as JSON")
	}
	return string(b)
}

func TransactionFromJSON(str string) (coin.Transaction, error) {

	var TxIn TransactionJSON
	err := json.Unmarshal([]byte(str), TxIn)

	if err != nil {
		errors.New("cannot deserialize")
	}

	var tx coin.Transaction

	tx.Sigs = make([]cipher.Sig, len(o.Sigs))
	tx.In = make([]cipher.SHA256, len(o.In))
	tx.Out = make([]cipher.TransactionOutput, len(o.Out))

	for i, sig := range txIn.Sigs {
		sig2, err := coin.SigFromHex(o.In[i])
		if err != nil {
			return coin.Transaction{}, errors.New("invalid signature")
		}
		tx.Sigs[i] = sig2
	}

	for i, in := range txIn.In {
		sig2, err := coin.SigFromHex(o.In[i])
		if err != nil {
			return coin.Transaction{}, errors.New("invalid signature")
		}
		tx.Sigs[i] = sig2
	}

	tx.Length = tx.Size()
	tx.Type = 0

	return coin.Transaction{}, nil
}

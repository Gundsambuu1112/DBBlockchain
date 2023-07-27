package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"strconv"
	"time"

	"github.com/tensor-programming/golang-blockchain/blockchain"
)

type Block struct {
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	Nonce        int
	Height       int
	Index        int64
	Timestamp    int64
	Data         blockchain.BlockData
}

type BlockData struct {
	Name  string
	Age   int
	Point int
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Serialize())
	}
	tree := NewMerkleTree(txHashes)

	return tree.RootNode.Data
}

func (b *Block) DeriveHash() {
	data := []byte(strconv.FormatInt(b.Index, 10) + strconv.FormatInt(b.Timestamp, 10) + string(b.Data.Name) + strconv.Itoa(b.Data.Age) + strconv.Itoa(b.Data.Point) + string(b.PrevHash) + strconv.Itoa(b.Nonce))
	hash := sha256.Sum256(data)
	b.Hash = hash[:]
}

func CreateBlock(name string, age, point int, txs []*Transaction, prevHash []byte) *Block {
	block := &Block{
		Hash:     nil,
		PrevHash: prevHash,
		Nonce:    0,
		Height:   0,
		Data: BlockData{
			Name:  name,
			Age:   age,
			Point: point,
		},
		Index:     0,
		Timestamp: time.Now().Unix(),
	}

	// Блокийн хэшийг үүсгэж, nonce-г тохируулна уу.
	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func Genesis(coinbase *Transaction) *Block {
	Name := "your_name" // Хүссэн нэрээр солино.
	age := 0            // Хүссэн насаар солино.
	point := 0          // Хүссэн OHOOooр солино.

	return CreateBlock(Name, age, point, []*Transaction{coinbase}, []byte{})
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)

	Handle(err)

	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)

	Handle(err)

	return &block
}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}

package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/dgraph-io/badger"
	"github.com/tensor-programming/golang-blockchain/cli"
)

type Block struct {
	Hash      []byte
	PrevHash  []byte
	Nonce     int
	Height    int
	Index     int64
	Timestamp int64
	Data      *Data
}

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
	Blocks   []*Block
}

type Data struct {
	Name  string
	Age   int
	Point int
}

func main() {
	// Replace "/path/to/your/database" with the actual path to your BadgerDB database
	dbPath := "/path/to/your/database"

	// BadgerDB мэдээллийн сангийн замыг зааж өгнө үү.
	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Шинэ Blockchain жишээг эхлүүлнэ үү.
	blockchain := &BlockChain{
		Blocks:   []*Block{},
		Database: db,
	}

	// Блокчейн өгөгдлүүдийг уншиж, блокчейн инстанцад нэмнэ.
	err = db.View(func(txn *badger.Txn) error {
		// Генезисийн блокоос эхлэх
		prevHash := []byte{}

		// Өгөгдлийн санд хадгалагдсан блокуудыг давт
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("block")
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek([]byte("block")); it.ValidForPrefix([]byte("block")); it.Next() {
			item := it.Item()
			blockData, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			// Блокийн өгөгдлийг Блокийн жишээ болгон салгах.
			block := DeserializeBlock(blockData)

			// Одоогийн блокийн өмнөх хэшийг тохируулна уу.
			block.PrevHash = prevHash

			// Блокийн хэшийг дахин тооцоол.
			block.DeriveHash()

			// Блокийг блокчэйнд хавсаргана уу.
			blockchain.Blocks = append(blockchain.Blocks, block)

			// Дараагийн блокийн өмнөх хэшийг шинэчилнэ үү.
			prevHash = block.Hash
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	// Блокчейн дэх блокуудын дата унших
	ReadDataFromBlockchain(blockchain)

	cmd := cli.CommandLine{}
	cmd.Run()
}

func ReadDataFromBlockchain(blockchain *BlockChain) {
	// Блокчейн дэх блокуудыг давт
	for _, block := range blockchain.Blocks {
		// Блок бүрт хадгалагдсан өгөгдөлд хандах
		data := block.Data
		// Шаардлагатай бол өгөгдлийг хэвлэх

		fmt.Printf("Name: %s, Age: %d, Point: %d\n", data.Name, data.Age, data.Point)
	}
}

func DeserializeBlock(data []byte) *Block {
	// Блокийн бүтцийн шинэ жишээг үүсгэ.
	block := &Block{}

	// Блокийн бүтцийн талбарт өгөгдлийг цувралаас ангижруул.
	err := json.Unmarshal(data, block)
	if err != nil {
		// Алдааг зохих ёсоор шийдвэрлэх (жишээ нь, алдаа буцаах эсвэл бүртгэх).
		return nil
	}

	return block
}

func (b *Block) DeriveHash() {
	data := []byte(strconv.FormatInt(b.Index, 10) + strconv.FormatInt(b.Timestamp, 10) + string(b.Data.Name) + strconv.Itoa(b.Data.Age) + strconv.Itoa(b.Data.Point) + string(b.PrevHash) + strconv.Itoa(b.Nonce))
	hash := sha256.Sum256(data)
	b.Hash = hash[:]
}

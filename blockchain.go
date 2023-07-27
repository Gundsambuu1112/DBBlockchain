package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "./tmp/blocks_%s"
	genesisData = "Genesis-ийн анхны гүйлгээ"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
	Data     []byte
	Blocks   []*Block
}

func DBexists(path string) bool {
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false
	}

	return true
}

func ReadDataFromBlockchain(blockchain *BlockChain) {
	// Блокчейн дэх блокуудыг давт
	for _, block := range blockchain.Blocks {
		// Блок бүрт хадгалагдсан өгөгдөлд хандах
		data := block.Data
		// Шаардлагатай бол өгөгдлийг боловсруулна

		fmt.Println(data)
	}
}

func ContinueBlockChain(nodeID string) *BlockChain {
	path := fmt.Sprintf(dbPath, nodeID)
	if !DBexists(path) {
		fmt.Println("Одоо байгаа блокчейн олдсонгүй, нэгийг үүсгээрэй!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions("")
	opts.Dir = path
	opts.ValueDir = path

	db, err := openDB(path, opts)
	Handle(err)

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(nil)

		return err
	})
	Handle(err)

	chain := BlockChain{
		LastHash: lastHash,
		Database: db,
		Data:     []byte("зарим өгөгдөл"),
	}

	return &chain
}

func InitBlockChain(address, nodeId string) *BlockChain {
	path := fmt.Sprintf(dbPath, nodeId)
	if DBexists(path) {
		fmt.Println("Блокчейн аль хэдийн бий болсон")
		runtime.Goexit()
	}

	var lastHash []byte
	opts := badger.DefaultOptions("")
	opts.Dir = path
	opts.ValueDir = path

	db, err := openDB(path, opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTx(address, genesisData)
		genesis := Genesis(cbtx)
		fmt.Println("Эхлэлийг бүтээсэн")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash

		return err
	})

	Handle(err)

	blockchain := BlockChain{
		LastHash: lastHash,
		Database: db,
		Data:     []byte("зарим өгөгдөл"),
	}
	return &blockchain
}

func (chain *BlockChain) AddBlock(block *Block) {
	err := chain.Database.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(block.Hash); err == nil {
			return nil
		}

		blockData := block.Serialize()
		err := txn.Set(block.Hash, blockData)
		Handle(err)

		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, _ := item.ValueCopy(nil)

		item, err = txn.Get(lastHash)
		Handle(err)
		lastBlockData, _ := item.ValueCopy(nil)

		lastBlock := Deserialize(lastBlockData)

		if block.Height > lastBlock.Height {
			err = txn.Set([]byte("lh"), block.Hash)
			Handle(err)
			chain.LastHash = block.Hash
		}

		return nil
	})
	Handle(err)
}

func (chain *BlockChain) GetBestHeight() int {
	var lastBlock Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, _ := item.ValueCopy(nil)

		item, err = txn.Get(lastHash)
		Handle(err)
		lastBlockData, _ := item.ValueCopy(nil)

		lastBlock = *Deserialize(lastBlockData)

		return nil
	})
	Handle(err)

	return lastBlock.Height
}

func (chain *BlockChain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(blockHash); err != nil {
			return errors.New("Блок олдсонгүй")
		} else {
			blockData, _ := item.ValueCopy(nil)

			block = *Deserialize(blockData)
		}
		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

func (chain *BlockChain) GetBlockHashes() [][]byte {
	var blocks [][]byte

	iter := chain.Iterator()

	for {
		block := iter.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return blocks
}

func (chain *BlockChain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte

	for _, tx := range transactions {
		if chain.VerifyTransaction(tx) != true {
			log.Panic("Хүчингүй гүйлгээ")
		}
	}

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(nil)

		item, err = txn.Get(lastHash)
		Handle(err)
		return err
	})
	Handle(err)

	Name := "your_name" // Replace with the desired address
	age := 0            // Replace with the desired age
	point := 0          // Replace with the desired point

	newBlock := CreateBlock(Name, age, point, transactions, lastHash)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err
	})
	Handle(err)

	return newBlock
}

func (chain *BlockChain) FindUTXO() map[string]TxOutputs {
	UTXO := make(map[string]TxOutputs)
	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return UTXO
}

func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Гүйлгээ байхгүй байна")
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf(`арилгах "LOCK": %s`, err)
	}
	retryOpts := originalOpts
	retryOpts.Truncate = true
	db, err := badger.Open(retryOpts)
	return db, err
}

func openDB(dir string, opts badger.Options) (*badger.DB, error) {
	if db, err := badger.Open(opts); err != nil {
		if strings.Contains(err.Error(), "LOCK") {
			if db, err := retry(dir, opts); err == nil {
				log.Println("өгөгдлийн сангийн түгжээг тайлсан, утгын бүртгэлийг таслав")
				return db, nil
			}
			log.Println("мэдээллийн сангийн түгжээг тайлж чадсангүй:", err)
		}
		return nil, err
	} else {
		return db, nil
	}
}

func LoadBlockchainFromDatabase() (*BlockChain, error) {
	// BadgerDB мэдээллийн сан руу очих замыг зааж өгнө үү.
	dbPath := "/path/to/your/database"

	// BadgerDB мэдээллийн санг нээх эсвэл холбогдох.
	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Шинэ Blockchain жишээг эхлүүлнэ үү.
	blockchain := &BlockChain{
		Blocks: []*Block{},
	}

	// Өгөгдлийн сангаас блокчейн өгөгдлийг татаж авах.
	err = db.View(func(txn *badger.Txn) error {
		// Генезисийн блокоос эхэл.
		prevHash := []byte{}

		// Өгөгдлийн санд хадгалагдсан блокуудыг давт.
		// блокчейн бүтцийг сэргээн засварлах.
		// өгөгдлийн сангаас блок өгөгдлийг татах замаар.
		// мөн блокчейн инстанцад нэмэх.
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
		return nil, err
	}

	return blockchain, nil
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

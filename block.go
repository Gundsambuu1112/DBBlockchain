package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

// Блокын бүтэц.
type Block struct {
	Index        int64  // Блокны индекс эсвэл өндөр
	Timestamp    int64  // Блок үүсгэх хугацааны тэмдэг
	Data         string // Блокод хадгалагдсан өгөгдөл
	PreviousHash string // Өмнөх блокийн хэш
	Hash         string // Одоогийн блокийн хэш
}

// Блокчейн бүтэц.
type Blockchain struct {
	Blocks []*Block // Блокчейн дэх блокуудын зүсмэлүүд
}

// CreateBlock нь өгөгдсөн өгөгдөл болон өмнөх хэштэй шинэ блок үүсгэдэг
func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{
		Index:        0, // Set the desired index value
		Timestamp:    time.Now().Unix(),
		Data:         data,
		PreviousHash: hex.EncodeToString(prevHash),
	}

	// Блок хэшийг тооцоол
	blockHash := calculateHash(block)
	block.Hash = blockHash

	return block
}

// accountHash нь SHA-256 алгоритмыг ашиглан блокийн хэшийг тооцдог
func calculateHash(block *Block) string {
	record := string(rune(block.Index)) + string(rune(block.Timestamp)) + block.Data + block.PreviousHash
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

func Genesis() {
	panic("unimplemented")
}

// CalculateHash нь блокийн хэшийг тооцдог.
func CalculateHash(block *Block) string {
	record := strconv.FormatInt(block.Index, 10) + strconv.FormatInt(block.Timestamp, 10) + block.Data + block.PreviousHash
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// GenerateBlock нь блокчэйнд шинэ блок үүсгэдэг.
func GenerateBlock(previousBlock *Block, data string) *Block {
	newBlock := &Block{
		Index:        previousBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		Data:         data,
		PreviousHash: previousBlock.Hash,
	}
	newBlock.Hash = CalculateHash(newBlock)
	return newBlock
}

// ValidateBlock нь түүний хэш болон өмнөх блокийн хэшийг шалгах замаар блокийн бүрэн бүтэн байдлыг шалгадаг.
func ValidateBlock(block, previousBlock *Block) bool {
	if block.PreviousHash != previousBlock.Hash {
		return false
	}
	hash := CalculateHash(block)
	if hash != block.Hash {
		return false
	}
	return true
}

func (block *Block) Verify() bool {
	hash := calculateHash(block)

	return hash == block.Hash
}

// ValidateChain нь блокчейн бүхэл бүтэн байдлыг шалгадаг.
func ValidateChain(chain []*Block) bool {
	for i := 1; i < len(chain); i++ {
		currentBlock := chain[i]
		previousBlock := chain[i-1]
		if !ValidateBlock(currentBlock, previousBlock) {
			return false
		}
	}
	return true
}

// AddBlock нь блокчэйнд шинэ блок нэмдэг.
func (bc *Blockchain) AddBlock(data string) {
	previousBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := GenerateBlock(previousBlock, data)
	bc.Blocks = append(bc.Blocks, newBlock)
}

// PrintChain нь блокчейн агуулгыг хэвлэдэг.
func (bc *Blockchain) PrintChain() {
	for _, block := range bc.Blocks {
		fmt.Printf("Index: %d\n", block.Index)
		fmt.Printf("Timestamp: %d\n", block.Timestamp)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Previous Hash: %s\n", block.PreviousHash)
		fmt.Printf("Hash: %s\n", block.Hash)
		fmt.Println("--------------------")
	}
}

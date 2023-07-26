package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
)

type Block struct {
	block_id     int64  // Блок_id эсвэл өндөр.
	Timestamp    int64  // Блок үүсгэх үеийн тэмдэг.
	Data         string // Блокод хадгалагдсан өгөгдөл.
	PreviousHash []byte // Өмнөх блокийн хэш.
	Hash         string // Одоогийн блокийн хэш.
	Index        int64  // Block index or height.
	Nonce        int64
}

// Блокчэйн бол блокчэйнийг төлөөлөх бүтэц юм.
type Blockchain struct {
	Blocks []*Block // Блокчэйнд блокуудыг хадгалахын тулд зүснэ үү.
	db     *sql.DB  // Өгөгдлийн сангийн холболт.
}

func main() {
	db, err := sql.Open("mysql", "root:8853d4E!@tcp(localhost:3306)/blockchain")
	if err != nil {
		fmt.Println("Өгөгдлийн санд холбогдож чадсангүй:", err)
		return
	}
	defer db.Close()

	fmt.Println("MySQL мэдээллийн санд холбогдсон")

	ReadBlockchainData(db)

	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	dataPtr := addCmd.String("data", "", "Өгөгдлийг блоклох (шаардлагатай)")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "add":
		addCmd.Parse(os.Args[2:])
		if *dataPtr == "" {
			log.Fatal("Блок өгөгдөл шаардлагатай. 'add -data <data>'-г ашиглах")
		}

		// Өгөгдсөн өгөгдөл болон хүндрэлийн түвшингээр createBlockAndAddToChain функцийг дуудна уу.
		createBlockAndAddToChain(*dataPtr, db)

	case "show":
		printBlockchain(db)
	case "help":
		printUsage()
	default:
		fmt.Println("Үл мэдэгдэх тушаал. Use 'help' боломжтой командуудыг харах.")
	}
}

func printUsage() {
	fmt.Println("Хэрэглээ:")
	fmt.Println("  add -data <data>: Заасан өгөгдөлтэй блокчэйнд шинэ блок нэмнэ .")
	fmt.Println("  show: Блокчэйнд хадгалагдсан өгөгдлийг харуул.")
	fmt.Println("  help: Энэ тусламжийн мессежийг харуул.")
}

func LoadBlockchainFromDatabase(db *sql.DB) (*Blockchain, error) {
	query := "SELECT block_id, timestamp, data, previous_hash, hash, nonce FROM blocks LIMIT 1"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	blocks := []*Block{}

	for rows.Next() {
		block := &Block{}
		var prevHashStr string

		err := rows.Scan(&block.block_id, &block.Timestamp, &block.Data, &prevHashStr, &block.Hash, &block.Index)
		if err != nil {
			return nil, err
		}

		// Өмнөх хэш мөрийг []byte руу хөрвүүлэх
		block.PreviousHash, err = hex.DecodeString(prevHashStr)
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, block)
	}

	blockchain := &Blockchain{
		Blocks: blocks,
	}

	return blockchain, nil
}

func NewBlock(data string, blockchain *Blockchain, difficulty int) *Block {
	// Өмнөх блокийн хэшийг аваарай.
	previousBlock := blockchain.Blocks[len(blockchain.Blocks)-1]
	previousHash := previousBlock.Hash

	// Шинэ блок үүсгэх.
	block := &Block{
		Index:        previousBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		Data:         data,
		PreviousHash: []byte(previousHash),
	}

	// proof-of-work ашиглан блокийн өвөрмөц хэшийг олоорой (хэшийг nonce-ээр тохируулна уу).
	for {
		hash := CalculateHash(block)
		if isValidHash(hash, difficulty) {
			block.Hash = hash
			break
		}
		block.Nonce++
	}

	// Блокчэйнд шинэ блок нэмнэ үү.
	blockchain.Blocks = append(blockchain.Blocks, block)

	return block
}

func (bc *Blockchain) AddBlockToDatabase(data string) error {
	// Өмнөх блокийн хэшийг аваарай.
	previousBlock := bc.Blocks[len(bc.Blocks)-1]
	previousHash := previousBlock.Hash

	// Шинэ блокийн хэшийг тооцоол.
	block := &Block{
		Index:        previousBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		Data:         data,
		PreviousHash: []byte(previousHash),
	}
	block.Hash = CalculateHash(block)

	// Мэдээллийн санд шинэ блок оруулах.
	query := "INSERT INTO blocks (timestamp, data, previous_hash, hash) VALUES (?, ?, ?, ?)"
	_, err := bc.db.Exec(query, block.Timestamp, block.Data, block.PreviousHash, block.Hash)
	if err != nil {
		return err
	}

	// Шинэ блокийг блокчэйнд нэмнэ үү.
	bc.Blocks = append(bc.Blocks, block)

	return nil
}

func isValidHash(hash string, difficulty int) bool {
	// Хэшийн эхний "хэцүү" тэмдэгтүүд тэг байгаа эсэхийг шалгана уу.
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}

func createBlockAndAddToChain(data string, db *sql.DB) {
	blockchain, err := LoadBlockchainFromDatabase(db)
	if err != nil {
		fmt.Println("Блокчейн ачаалж чадсангүй:", err)
		return
	}

	var latestBlock *Block
	if len(blockchain.Blocks) > 0 {
		latestBlock = blockchain.Blocks[len(blockchain.Blocks)-1]
	} else {
		latestBlock = &Block{
			block_id:     0,
			Timestamp:    time.Now().Unix(),
			Data:         "Genesis blockchain",
			PreviousHash: []byte{},
			Hash:         "",
			Index:        0,
			Nonce:        0,
		}
	}

	newBlock := &Block{
		block_id:     int64(len(blockchain.Blocks)),
		Timestamp:    time.Now().Unix(),
		Data:         data,
		PreviousHash: []byte(latestBlock.Hash),
	}

	// Шинэ блокийн хэшийг тооцоолж, ажлын нотлох баримтаар баталгаажуулна уу.
	hash := CalculateHash(newBlock)
	newBlock.Nonce = 0
	for !isValidHash(hash, 3) { // '3'-ыг хүссэн хүндрэлийн түвшинд шилжүүлээрэй.
		newBlock.Nonce++
		hash = CalculateHash(newBlock)
	}
	newBlock.Hash = hash

	// Шинэ блокийг блокчейн болон мэдээллийн санд нэмнэ үү.
	err = blockchain.AddBlock(newBlock)
	if err != nil {
		fmt.Println("Блокчэйнд блок нэмж чадсангүй:", err)
		return
	}

	fmt.Println("Блокыг амжилттай нэмсэн.")
}

func (bc *Blockchain) GetLatestBlock() *Block {
	if len(bc.Blocks) > 0 {
		return bc.Blocks[len(bc.Blocks)-1]
	}
	return nil
}

func (bc *Blockchain) AddBlock(block *Block) error {
	// Өгөгдсөн блокийн хэш нь блокчейн дэх сүүлийн блокийн хэштэй таарч байгаа эсэхийг шалгана уу.
	if len(bc.Blocks) > 0 {
		previousBlock := bc.Blocks[len(bc.Blocks)-1]
		previousHashBytes := []byte(previousBlock.Hash) // Convert previousBlock.Hash to []byte.
		if !bytes.Equal(previousHashBytes, block.PreviousHash) {
			return errors.New("өмнөх блок хэш буруу байна")
		}
	}

	// Шинэ блокийн хэшийг тооцоолж, индексийг тохируулна уу.
	block.Index = int64(len(bc.Blocks))
	block.Timestamp = time.Now().Unix()
	block.Hash = CalculateHash(block)

	// Блокийг мэдээллийн санд хадгал.
	stmt, err := bc.db.Prepare("INSERT INTO blocks (block_id, previous_hash, data, timestamp, hash) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(block.Index, hex.EncodeToString(block.PreviousHash), block.Data, block.Timestamp, block.Hash)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("өгөгдлийн санд блок оруулж чадсангүй")
	}

	// Шинэ блокыг гинжин хэлхээнд нэмнэ үү.
	bc.Blocks = append(bc.Blocks, block)

	return nil
}

// Length нь блокчейн уртыг буцаана.
func (bc *Blockchain) Length() int {
	return len(bc.Blocks)
}

// Блокчэйнд хадгалагдсан өгөгдлийг хэвлэх функц.
func printBlockchain(db *sql.DB) {
	blockchain, err := LoadBlockchainFromDatabase(db)
	if err != nil {
		fmt.Println("Блокчейн ачаалж чадсангүй:", err)
		return
	}

	fmt.Println("Блокчейн өгөгдөл:")
	for _, block := range blockchain.Blocks {
		fmt.Printf("Block_id: %d\n", block.block_id)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Previous Hash: %s\n", block.PreviousHash)
		fmt.Printf("Timestamp: %d\n", block.Timestamp)
		fmt.Printf("Index: %d\n", block.Index)
		fmt.Printf("Hash: %s\n", block.Hash)
		fmt.Println("----------")
	}
}

// GetBlocks нь блокчлоноос бүх блокуудыг татаж авдаг.
func (bc *Blockchain) GetBlocks() ([]Block, error) {
	query := "SELECT block_id, previous_hash, data, timestamp FROM blocks ORDER BY block_id ASC"
	rows, err := bc.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []Block
	for rows.Next() {
		var block Block
		err := rows.Scan(&block.block_id, &block.PreviousHash, &block.Data, &block.Timestamp, &block.Index, &block.Hash)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	if err := rows.Err(); err != nil {
		return nil, err // Алдааг буцаах замаар алдааны засварыг энд засна уу.
	}

	return blocks, nil
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

// GenerateBlock нь блокчэйнд шинэ блок үүсгэдэг.
func GenerateBlock(previousBlock *Block, data string) *Block {
	newBlock := &Block{
		Index:        previousBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		Data:         data,
		PreviousHash: []byte(previousBlock.Hash),
	}
	newBlock.Hash = CalculateHash(newBlock)
	return newBlock
}

func InsertDataIntoDatabase(db *sql.DB, block_id int, data string) error {
	// INSERT мэдэгдлийг бэлтгэ.
	stmt, err := db.Prepare("INSERT INTO blocks (block_id, data) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Өгөгдсөн өгөгдлөөр INSERT мэдэгдлийг гүйцэтгэнэ.
	_, err = stmt.Exec(block_id, data)
	if err != nil {
		return err
	}

	return nil
}

func NewBlockchain(db *sql.DB) *Blockchain {
	// Өгөгдлийн санд блок байгаа эсэхийг шалгана уу.
	rows, err := db.Query("SELECT block_id FROM blocks LIMIT 1")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	blocks := []*Block{}

	// Хэрэв мэдээллийн санд мөр байхгүй бол блокчэйнд генезис блок нэмнэ үү.
	if !rows.Next() {
		genesisBlock := &Block{
			block_id:     0,
			Timestamp:    time.Now().Unix(),
			Data:         "Genesis blockchain",
			PreviousHash: []byte{},
			Hash:         "",
			Index:        0,
			Nonce:        0,
		}
		genesisBlock.Hash = CalculateHash(genesisBlock)
		blocks = append(blocks, genesisBlock)
	}

	blockchain := &Blockchain{
		Blocks: blocks,
		db:     db,
	}

	return blockchain
}

func ReadBlockchainData(db *sql.DB) {
	// Өгөгдлийн сангаас блокчейн өгөгдлийг ачаална уу.
	blockchain, err := LoadBlockchainFromDatabase(db)
	if err != nil {
		fmt.Println("Блокчейн датаг ачаалж чадсангүй:", err)
		return
	}

	// Блокчейн өгөгдлийг боловсруулж, үйлдлүүдийг гүйцэтгэх.
	for _, block := range blockchain.Blocks {
		data := block.Data

		// Өгөгдөл боловсруулах үйлдлийг гүйцэтгэх.
		processedData := processData(data)

		// Боловсруулсан өгөгдөл дээр нэмэлт үйлдэл эсвэл хувиргалт хийх.

		// Жишээ: Боловсруулсан өгөгдөл дэх эгшиг ба гийгүүлэгчийн тоог тоол.
		vowelCount, consonantCount := countVowelsAndConsonants(processedData)

		// Жишээ: Боловсруулсан өгөгдлөөс цэг таслалыг арилгах.
		cleanedData := removePunctuation(processedData)

		// Жишээ: Боловсруулсан өгөгдлөөс тодорхой түлхүүр үгсийг олж, задлах.
		keywords := findKeywords(processedData)

		// Боловсруулсан өгөгдөл болон нэмэлт мэдээллийг хэвлэх.
		fmt.Println("Боловсруулсан өгөгдөл:", processedData)
		fmt.Println("Эгшиг тоо:", vowelCount)
		fmt.Println("Гийгүүлэгчийн тоо:", consonantCount)
		fmt.Println("Цэвэрлэсэн өгөгдөл:", cleanedData)
		fmt.Println("Түлхүүр үгс:", keywords)
	}
}

// processData нь таны өгөгдөл боловсруулах логикийн орлуулагч функц юм.
func processData(data string) string {
	numbers := strings.Split(data, ",")
	sum := 0
	for _, numStr := range numbers {
		num, err := strconv.Atoi(numStr)
		if err != nil {
			// Буруу тооны форматыг зохицуулах.
			continue
		}
		sum += num
	}

	processedData := strconv.Itoa(sum)

	return processedData
}

// countVowelsAndConsonants нь тухайн мөрөнд байгаа эгшиг ба гийгүүлэгчийн тоог тоолдог.
func countVowelsAndConsonants(data string) (int, int) {
	vowels := "aeiou"
	vowelCount := 0
	consonantCount := 0
	for _, char := range data {
		if unicode.IsLetter(char) {
			char = unicode.ToLower(char)
			if strings.ContainsRune(vowels, char) {
				vowelCount++
			} else {
				consonantCount++
			}
		}
	}
	return vowelCount, consonantCount
}

// RemovePunctuation нь өгөгдсөн мөрөөс цэг таслалыг арилгадаг.
func removePunctuation(data string) string {
	punctuations := `!"#$%&'()*+,-./:;<=>?@[\]^_` + "`" + `{|}~`
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune(punctuations, r) {
			return -1
		}
		return r
	}, data)
}

// findKeywords нь өгөгдсөн мөрөөс тодорхой түлхүүр үгсийг олж, задалдаг.
func findKeywords(data string) []string {
	var keywords []string
	words := strings.Fields(data)
	for _, word := range words {
		if strings.HasPrefix(word, "#") {
			// Түлхүүр үгнээс "#" тэмдгийг устгана уу.
			keyword := strings.TrimPrefix(word, "#")

			// Түлхүүр үгийг жижиг үсгээр хөрвүүлэх.
			keyword = strings.ToLower(keyword)

			// Түлхүүр үгнээс тусгай тэмдэгт, цэг таслалыг хас.
			keyword = cleanKeyword(keyword)

			// Түлхүүр үг хоосон биш эсвэл түлхүүр үгсийн хэсэгт байхгүй эсэхийг шалгана уу.
			if keyword != "" && !contains(keywords, keyword) {
				keywords = append(keywords, keyword)
			}
		}
	}

	return keywords
}

// Мөрөөс тусгай тэмдэгт болон цэг таслалыг арилгах туслах функц.
func cleanKeyword(keyword string) string {
	var cleanedKeyword strings.Builder
	for _, ch := range keyword {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			cleanedKeyword.WriteRune(ch)
		}
	}
	return cleanedKeyword.String()
}

// Зүсмэл нь тодорхой мөр агуулсан эсэхийг шалгах туслах функц.
func contains(slice []string, target string) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

// accountHash нь SHA-256 алгоритмыг ашиглан блокийн хэшийг тооцдог.
func CalculateHash(block *Block) string {
	record := strconv.FormatInt(block.Index, 10) + strconv.FormatInt(block.Timestamp, 10) + block.Data + hex.EncodeToString(block.PreviousHash)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// ValidateBlock нь түүний хэш болон өмнөх блокийн хэшийг шалгах замаар блокийн бүрэн бүтэн байдлыг шалгадаг.
func ValidateBlock(block, previousBlock *Block) bool {
	previousHashStr := hex.EncodeToString(block.PreviousHash) // Convert []byte to string.
	if previousHashStr != previousBlock.Hash {
		return false
	}
	hash := CalculateHash(block)
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

// Verify нь блокийн бүрэн бүтэн байдлыг шалгадаг бөгөөд хэрэв хүчинтэй бол үнэн, үгүй ​​бол худал гэж буцаана...
func (block *Block) Verify(previousBlock *Block) bool {
	// Блокийн индекс өмнөх блокийн индексээс их байгаа эсэхийг шалгана уу.
	if block.Index != previousBlock.Index+1 {
		return false
	}

	// Блокийн цагийн тэмдэг өмнөх блокийн цагийн тэмдэгээс их байгаа эсэхийг шалгана уу.
	if block.Timestamp <= previousBlock.Timestamp {
		return false
	}

	// Блокийн өмнөх хэш нь өмнөх блокийн хэштэй тохирч байгаа эсэхийг шалгаарай.
	if !bytes.Equal([]byte(previousBlock.Hash), block.PreviousHash) {
		return false
	}

	// Блокийн хэшийг тооцоолж, хадгалсан хэштэй тохирч байгаа эсэхийг шалгаарай.
	calculatedHash := CalculateHash(block)
	if calculatedHash == block.Hash {
		return false
	}

	return true
}

package blockchain

// Блокын бүтэц.
type Block struct {
	Index        int64  // Блокны индекс эсвэл өндөр.
	Timestamp    int64  // Блок үүсгэх хугацааны тэмдэг.
	Data         string // Блокод хадгалагдсан өгөгдөл.
	PreviousHash string // Өмнөх блокийн хэш.
	Hash         string // Одоогийн блокийн хэш.
	block_id     int
}

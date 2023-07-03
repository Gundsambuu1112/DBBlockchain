package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	// Өгөгдлийн сангийн холболтыг нээнэ үү.
	db, err := sql.Open("mysql", "root:8853d4E!@tcp(localhost:3306)/blockchain")
	if err != nil {
		fmt.Println("Өгөгдлийн санд холбогдож чадсангүй:", err)
		return
	}
	defer db.Close()

	// Холболтыг шалгана уу.
	err = db.Ping()
	if err != nil {
		fmt.Println("Өгөгдлийн санг ping хийж чадсангүй:", err)
		return
	}

	fmt.Println("MySQL мэдээллийн санд холбогдсон")

	//өгөгдлийн сангийн "блок" хүснэгтээс бүх мөрийг татахын тулд SQL мэдэгдлийг бэлтгэдэг.
	stmt, err := db.Prepare("SELECT * FROM blocks")
	if err != nil {
		fmt.Println("Мэдэгдэл бэлтгэхэд алдаа гарлаа:", err)
		return
	}
	defer stmt.Close()

	//өгөгдлийн сангаас үр дүнгийн багцыг авахын тулд бэлтгэсэн SQL мэдэгдлийг stmt.Query() гүйцэтгэнэ.
	rows, err := stmt.Query()
	if err != nil {
		fmt.Println("Асуултыг гүйцэтгэхэд алдаа гарлаа:", err)
		return
	}
	defer rows.Close()

	//дүнгийн багцын баганын төрлийг олж авна.
	columns, err := rows.ColumnTypes()
	if err != nil {
		fmt.Println("Баганын төрлийг авахад алдаа гарлаа:", err)
		return
	}

	//үр дүнгийн багц дахь багана бүрийн төрөл бүрийн шинж чанаруудад хандаж, тэдгээрийг төрөл, шинж чанарт нь үндэслэн өгөгдлийг зохих ёсоор удирдах боломж олгоно.
	for _, col := range columns {
		fmt.Println("Баганын нэр:", col.Name())
		fmt.Println("Баганын мэдээллийн сангийн төрлийн нэр:", col.DatabaseTypeName())

		nullable, _ := col.Nullable() // col.Nullable()-ийн үр дүнг тэг хувьсагчид онооно.
		fmt.Println("Багана хүчингүй болно:", nullable)

		precision, scale, ok := col.DecimalSize()
		if ok {
			fmt.Println("Аравтын бутархай баганын хэмжээ:", precision, scale)
		}

		length, ok := col.Length()
		if ok {
			fmt.Println("Баганын урт:", length)
		}

		fmt.Println("Баганын скан төрөл:", col.ScanType())
		fmt.Println("-----")
	}

	//гүйлгээг эхлүүлж, гүйлгээ эхлэхгүй бол алдааг зохицуулдаг. Энэ нь мөн гүйлгээ хийгдээгүй тохиолдолд буцаах дуудлагыг хойшлуулдаг.
	tx, err := db.Begin()
	if err != nil {
		fmt.Println("Гүйлгээг эхлүүлж чадсангүй:", err)
		return
	}
	defer tx.Rollback()

}

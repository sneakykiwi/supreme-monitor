package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
)

type Keyword struct {
	gorm.Model
	Value string
}

func GetKeywords(db *gorm.DB) []Keyword{
	var keywords []Keyword
	db.Find(&keywords)
	return keywords
}

func AddKeyword(db *gorm.DB, keyword string){
	kw := Keyword{Value: keyword}
	db.Create(&kw)
}

func RemoveKeyword(db *gorm.DB, keyword string){
	keywords := GetKeywords(db)
	for _, kw := range keywords{
		if kw.Value == keyword{
			fmt.Println("keyword deleted")
			db.Delete(&kw)
			return
		}
	}
}


type Product struct{
	gorm.Model
	SupremeID int
	Name string
	Image string
	Category string
	InStock bool
	LastTimeInStock int32 //timestamp needs to be bigger than max value
	Sizes string
	Color string
}


func Connect() *gorm.DB{
	databaseDSN := os.Getenv("DATABASE_DSN")
	db, err := gorm.Open(mysql.Open(databaseDSN), &gorm.Config{ Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil{
		panic(fmt.Sprintf("Failed to connect to database: %s", err))
	}
	_ = db.AutoMigrate(&Keyword{})
	_ = db.AutoMigrate(&Product{})

	return db
}
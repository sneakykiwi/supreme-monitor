package main

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	database "main/db"
	"main/helpers"
	"strings"
)



func main() {
	app := fiber.New()
	db := database.Connect()
	jobs := make(chan helpers.Job)
	go helpers.GetStock(db, jobs)
	go helpers.Monitor(db, jobs)
	app.Post("/add-keyword", func(ctx *fiber.Ctx) error {
		_ = addKeyword(ctx, db)
		return nil
	})
	app.Post("/remove-keyword", func(ctx *fiber.Ctx) error {
		_ = RemoveKeyword(ctx, db)
		return nil
	})
	app.Get("/view-keywords", func(ctx *fiber.Ctx) error {
		_ = ViewKeywords(ctx, db)
		return nil
	})

	app.Listen(":3000")
}

type KW struct{
	Value string
}

func addKeyword(c *fiber.Ctx, db *gorm.DB) error{
	body := c.Body()
	var kw *KW
	if err := json.Unmarshal(body, &kw); err != nil{
		fmt.Println(err)
	}

	keywords := database.GetKeywords(db)

	for _, keyword := range keywords{
		f := strings.TrimSpace(keyword.Value)
		if strings.ToLower(f) == kw.Value{
			err := c.JSON(fiber.Map{"state": "keyword_already_exists"})
			return err
		}
	}
	database.AddKeyword(db, kw.Value)
	err := c.JSON(fiber.Map{"state": "keyword_added"})
	return err
}


func RemoveKeyword(c *fiber.Ctx, db *gorm.DB) error{
	body := c.Body()
	var kw *KW
	if err := json.Unmarshal(body, &kw); err != nil{
			fmt.Println(err)
	}

	keywords := database.GetKeywords(db)

	for _, keyword := range keywords{
		f := strings.TrimSpace(keyword.Value)
		if strings.ToLower(f) == kw.Value{
			database.RemoveKeyword(db, kw.Value)
			err := c.JSON(fiber.Map{"state": "keyword_removed"})
			return err
		}
	}
	err := c.JSON(fiber.Map{"state": "keyword_not_found"})
	return err
}

func ViewKeywords(c *fiber.Ctx, db *gorm.DB) error{
	keywords := database.GetKeywords(db)
	err := c.JSON(fiber.Map{"keywords": keywords})
	return err
}
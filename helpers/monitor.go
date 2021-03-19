package helpers

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"io/ioutil"
	database "main/db"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SupremeItem struct {
	ID                int         `json:"id"`
	Name              string      `json:"name"`
	Currency          string      `json:"currency"`
	ImageURL          string      `json:"image_url"`
	Category 		  string
	Color 			  string
	Sizes             []struct {
		Name       string `json:"name"`
		ID         int    `json:"id"`
		StockLevel int    `json:"stock_level"`
		} `json:"sizes"`
}

type ItemResponse struct{
	Styles []SupremeItem `json:"styles"`
}

var products = make(chan SupremeItem)

func Monitor(db *gorm.DB, jobs <- chan Job) {
	go getProductData(products, jobs)

	for product := range products {
		var sizes strings.Builder
		sizes.Reset()
		var prod database.Product
		mu.Lock()
		db.Where("supreme_id = ?", product.ID).First(&prod)
		mu.Unlock()
		old_state := prod.InStock
		prod.InStock = false
		for _, size := range product.Sizes {
			if size.StockLevel > 0 {
				formattedForDb := fmt.Sprintf("%s, ", size.Name)
				sizes.WriteString(formattedForDb)
				prod.InStock = true
			}
		}
		if prod.SupremeID == 0 {
			dbProduct := database.Product{
				SupremeID:       product.ID,
				Name:            product.Name,
				Image:           product.ImageURL,
				Category:        product.Category,
				InStock:         prod.InStock,
				LastTimeInStock: int32(time.Now().Unix()),
				Sizes:           sizes.String(),
				Color:           product.Color,
			}
			mu.Lock()
			db.Create(&dbProduct)
			mu.Unlock()
		}else{
			prod.Color = product.Color
			prod.Sizes = sizes.String()
			if prod.InStock{
				prod.LastTimeInStock = int32(time.Now().Unix())
			}
			mu.Lock()
			db.Save(&prod)
			mu.Unlock()
		}
		mu.Lock()
		db.Where("supreme_id = ?", product.ID).First(&prod)
		mu.Unlock()
		if !old_state && prod.InStock{
			go sendWebhookMessage(prod)
			printInStock := fmt.Sprintf("[PRODUCT FOUND IN STOCK] [%s]", product.Name)
			fmt.Println(printInStock)
		}
	}
}
func getProductData(products chan <- SupremeItem, jobs <- chan Job){
	for job := range jobs{
		proxyUrl, err := url.Parse(getProxy())
		client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
		//client := &http.Client{}
		url := fmt.Sprintf("https://www.supremenewyork.com/shop/%d.json", job.id)
		r, err := client.Get(url)
		if err != nil {
			fmt.Println(err)
			continue
		}
		resp, _ := ioutil.ReadAll(r.Body)
		_ = r.Body.Close()
		var reply ItemResponse
		err = json.Unmarshal(resp, &reply)
		if err != nil {
			fmt.Println(err)
		}else {
			product := reply.Styles
			for _, p := range product {
				name := fmt.Sprintf("%s %s", job.itemName, p.Name)
				p.Color = p.Name
				p.Category = job.category
				p.Name = name
				products <- p
			}
		}
	}
}

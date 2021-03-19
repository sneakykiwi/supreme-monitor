package helpers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	database "main/db"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

type ResultProduct struct{
	Name          string `json:"name"`
	ID            int    `json:"id"`
	ImageURL      string `json:"image_url"`
	ImageURLHi    string `json:"image_url_hi"`
	Price         int    `json:"price"`
	SalePrice     int    `json:"sale_price"`
	NewItem       bool   `json:"new_item"`
	Position      int    `json:"position"`
	CategoryName  string `json:"category_name"`
	PriceEuro     int    `json:"price_euro"`
	SalePriceEuro int    `json:"sale_price_euro"`
}

type Reply struct {
	UniqueImageURLPrefixes []interface{} `json:"unique_image_url_prefixes"`
	ProductsAndCategories  map[string][]ResultProduct `json:"products_and_categories"`
	LastMobileAPIUpdate string `json:"last_mobile_api_update"`
	ReleaseDate         string `json:"release_date"`
	ReleaseWeek         string `json:"release_week"`
}


type Job struct{
	id       int
	category string
	price    int
	itemName string
}

var (
	mu sync.Mutex
)

func getProxy() string{
	mu.Lock()
	file, err := os.Open("proxies.txt")
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtlines []string
	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}
	_ = file.Close()

	if len(txtlines) == 0{
		panic("Please add proxies to proxies.txt")
	}

	index := rand.Intn(len(txtlines))
	mu.Unlock()
	return txtlines[index]
}

func StockRetriever(dataChannel chan <- Reply){
	for {
		proxyUrl, err := url.Parse(getProxy())
		client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
		//client := &http.Client{}
		r, err := client.Get("https://www.supremenewyork.com/mobile_stock.json")
		if err != nil {
			fmt.Println(err)
			continue
		}
		resp, _ := ioutil.ReadAll(r.Body)
		_ = r.Body.Close()
		//fmt.Println(r.Header)
		//os.Exit(1)
		var reply Reply
		err = json.Unmarshal(resp, &reply)
		if err != nil {
			fmt.Println("Stock Retriever ", err)
		}else {
			dataChannel <- reply
		}
	}
}



func GetStock(db *gorm.DB, jobs  chan <- Job){
	supremeData := make(chan Reply, 15)
	go StockRetriever(supremeData)
	for reply := range supremeData{
		keywords := database.GetKeywords(db)
		for _, category := range reply.ProductsAndCategories {
			for _, product := range category {
				if KeywordExistsInName(product.Name, keywords){
					jobs <- Job{id: product.ID, category: product.CategoryName, price: product.Price, itemName: product.Name}
				}
			}
		}

	}
}



func KeywordExistsInName(itemName string, keywords []database.Keyword) bool{
	for _, keyword := range keywords{
		if strings.Contains(strings.ToLower(itemName), strings.ToLower(keyword.Value)){
			return true
		}
	}
	return false
}

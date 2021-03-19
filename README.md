<h2>**Open Source Supreme Monitor Based on GoLang**</h2>

_A module built for personal use but ended up being worthy to have it open sourced._


<h3>**The module contains the following**</h3>
<h4>- Ability to **Add, Remove, View Keywords** through HTTP requests to the server of your choice</h4>
<h4>- Ability to save item status in any database supported by GORM</h4>
<h4>- Efficient use of resources</h4>

The project is built on GoLang as I felt it was fit for it. Some parts that are really worth noting:

- Stock Retriever
```go
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
```
This function is run as a `goroutine`, upon initialisation `StockRetriever()` is called as a goroutine and takes in `supremeData`
which is a channel that gives back structs of type `Reply` and has a buffer of 15. Below we access the items from the `supremeData` channel
and process them accordingly and sending them to the `jobs` channel as a `Job` structure for them to be handled individually.

- Retrieving Individual Item Data
```go
func getProductData(products chan <- SupremeItem, jobs <- chan Job){
	for job := range jobs{
		proxyUrl, err := url.Parse(getProxy())
		client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
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

```
This is the next part of the data, it takes in the structs from the `jobs` channel which are fed by the previous function, processes them
and then feeds them back to the `products` channel which basically confirms whether an item is newly instock or out of stock 
and decides to send or skip the webhook.
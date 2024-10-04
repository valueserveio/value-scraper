# value-scraper

Example POST request: 

```
curl -X POST http://localhost:8080/scrape \
     -H "Content-Type: application/json" \
     -d '{
           "customer_url": "http://example.com/customer",
           "product_url": "http://example.com/product"
         }'
```


With new changes run with ```go run .```
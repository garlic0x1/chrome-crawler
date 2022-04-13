# chrome-crawler
Drives chrome headless with chromedp https://github.com/chromedp/chromedp  
Queues links to a pool of chromium tabs and scrapes pages with goquery 

# Installation
Go install is broken for some reason  
Docker install:  
```
git clone https://github.com/garlic0x1/chrome-crawler
cd chrome-crawler
sudo docker build -t "garlic0x1/chrome-crawler" .
echo http://testphp.vulnweb.com/ | sudo docker run --rm -i garlic0x1/chrome-crawler
```

# Usage
Single URL:  
`echo https://example.com | chrome-crawler -s -u`  
Multiple URLs:  
`cat urls.txt | chrome-crawler -s -u`  

# Help
```
$ chrome-crawler -h
Usage of chrome-crawler:
  -d int
    	Depth to crawl. (default 2)
  -debug
    	Don't use headless. (slow but fun to watch)
  -proxy string
    	Proxy URL. Example: -proxy http://127.0.0.1:8080
  -r	Revisit URLs.
  -s	Show source.
  -t int
    	Number of chrome tabs to use concurrently. (default 8)
  -u	Show only unique URLs.
```

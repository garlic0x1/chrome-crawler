# chrome-crawler
Crawls URLs from stdin with headless chromium.  Use `-p` to enable form submission and find injection points for xss.  

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
Single URL, 4 deep:  
`echo https://example.com | chrome-crawler -u -s -d 4`  
Multiple URLs:  
`cat urls.txt | chrome-crawler -s -u`  
Submit Forms:  
`echo https://example.com | chrome-crawler -u -s -r -p`  
Wait For DOM to change:  
`echo https://example.com | chrome-crawler -u -s -r -p -w 2`  

# Help
```
$ chrome-crawler -h
Usage of chrome-crawler:
  -d int
    	Depth to crawl. (default 2)
  -debug
    	Don't use headless. (slow but fun to watch)
  -p	Find injection points.
  -proxy string
    	Proxy URL. Example: -proxy http://127.0.0.1:8080
  -r	Revisit URLs.
  -s	Show source.
  -t int
    	Number of chrome tabs to use concurrently. (default 10)
  -time int
    	Timeout per request. (default 10)
  -u	Show only unique URLs.
  -w int
    	Seconds to wait for DOM to load. (Use to find injections from AJAX reqs)
```

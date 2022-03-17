# chrome-crawler
Drives chrome headless with chromedp https://github.com/chromedp/chromedp  

# Installation:
With Go installed:  
`$ go install github.com/garlic0x1/chrome-crawler@main`  
  
# Usage:
```
$ ./chrome-crawler -h
Usage of ./chrome-crawler:
  -depth int
    	Depth to crawl (default 2)
  -r	Revisit URLs
  -tabs int
    	Number of chrome tabs to use concurrently (default 8)
  -u string
    	URL to crawl
  -uniq
    	Show only unique URLs
```

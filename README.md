# chrome-crawler
Drives chrome headless with chromedp https://github.com/chromedp/chromedp  

# Usage
Single URL:
`echo https://google.com | chrome-crawler -s -u`  
Multiple URLs:
`cat urls.txt | chrome-crawler -s -u`  

# Help
```
$ chrome-crawler -h
Usage of chrome-crawler:
  -d int
    	Depth to crawl. (default 2)
  -debug
    	Don't use headless mode.
  -r	Revisit URLs.
  -s	Show source.
  -t int
    	Number of chrome tabs to use concurrently. (default 8)
  -u	Show only unique URLs.
```

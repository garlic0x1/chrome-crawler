# chrome-crawler
Crawls URLs from stdin with headless chromium. Performs a passive crawl by default, but with `-p` flag it fills out forms and searches for reflected inputs  
Output structured data with `-json` or `-yaml`  

# Installation
Go install from remote source is broken for some reason  
From Source:  
```
git clone https://github.com/garlic0x1/chrome-crawler
cd chrome-crawler
go install .
```
Docker install:  
```
git clone https://github.com/garlic0x1/chrome-crawler
cd chrome-crawler
sudo docker build -t "garlic0x1/chrome-crawler" .
echo http://testphp.vulnweb.com/ | sudo docker run --rm -i garlic0x1/chrome-crawler
```

# Usage
Single URL, 4 deep, 20 tabs:  
```
echo https://example.com | chrome-crawler -u -s -d 4 -t 20
```  
Multiple URLs, disable headless and use proxy:  
```
cat urls.txt | chrome-crawler -u -s -debug-chrome -proxy http://localhost:8080
```  
Submit forms with header:  
```
echo https://example.com | chrome-crawler -u -s -r -p -head "Cookie: foo=bar;;Referer: http://example.com/"
```  
Wait for DOM to change:  
```
echo https://example.com | chrome-crawler -u -s -r -p -w 2
```  
Example toolchain:  
```
echo https://example.com | gau | chrome-crawler -u | url-miner -chrome -w wordlist.txt
```  
Example output on https://portswigger.net/web-security/cross-site-scripting/dom-based/lab-dom-xss-stored:
```
$ echo https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/ | chrome-crawler -s -u -d 3 -p -r -w 1
[href] https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/resources/labheader/css/academyLabHeader.css
[href] https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/resources/css/labsBlog.css
[href] https://portswigger.net/web-security/cross-site-scripting/dom-based/lab-dom-xss-stored
[href] https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/
...
[href] https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/post?postId=1
[form] https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/post/comment
[href] https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/post?postId=7
[script] https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/resources/labheader/js/labHeader.js
[href] https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/post?postId=3
[href] https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/post?postId=8
[href] https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/post?postId=6
[href] https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/post?postId=5
...
[reflect] https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/post/comment -> https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/post?postId=1
[reflect] https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/post/comment -> https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/post?postId=10
[script] https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/resources/js/loadCommentsWithVulnerableEscapeHtml.js
[href] https:Ezzx35jy
...
[href] https:=zzx35jy
[reflect] https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/post/comment -> https://acc71ffe1f561697c0e5040100190092.web-security-academy.net/post?postId=2
...
```

# Help
```
$ chrome-crawler -h
Usage of chrome-crawler:
  -d int
    	Depth to crawl. (default 2)
  -debug
    	Display error messages.
  -debug-chrome
    	Don't use headless. (slow but fun to watch)
  -head string
    	Custom headers separated by two semi-colons. Example: -h 'Cookie: foo=bar;;Referer: http://example.com/'
  -json
    	Output as JSON.
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
  -yaml
    	Output as YAML.
```

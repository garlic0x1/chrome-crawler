// Command eval is a chromedp example demonstrating how to evaluate javascript
// and retrieve the result.
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

func crawl(u string, ctx context.Context, queue chan string) {
	parsed, err := url.Parse(u)
	if err != nil {
		log.Println("failed to parse url", u, err)
	}
	host := parsed.Host
	// run task list
	var hrefs []*cdp.Node
	var forms []*cdp.Node
	err = chromedp.Run(ctx,
		chromedp.Navigate(u),
		chromedp.Nodes("a", &hrefs),
		chromedp.Nodes("form", &forms),
	)
	if err != nil {
		log.Fatal(err)
	}

	for _, href := range hrefs {
		if href.AttributeValue("href") != "" {
			link := "https://" + host + href.AttributeValue("href")
			log.Println("from crawl func", link)
			queue <- link
		}
	}

	for _, form := range forms {
		if form.AttributeValue("action") != "" {
			log.Println("form", form.AttributeValue("action"))
		}
	}
}

func main() {
	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	queue := make(chan string)

	go crawl("https://garlic0x1.com", ctx, queue)

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for link := range queue {
		fmt.Println(link)
		go crawl(link, ctx, queue)
	}
}

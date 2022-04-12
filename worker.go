package main

import (
	"context"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

func crawl(l item, ctx context.Context) {
	c1 := make(chan int, 1)

	go func() {
		var document string
		err := chromedp.Run(ctx,
			chromedp.Navigate(l.URL),
			chromedp.Evaluate("document.documentElement.innerHTML", &document),
		)
		if err != nil {
			c1 <- 1
			return
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(document))
		if err != nil {
			c1 <- 1
			return
		}

		doc.Find("*[href]").Each(func(index int, gitem *goquery.Selection) {
			href, _ := gitem.Attr("href")
			link := absoluteURL(l.URL, href)
			Results <- result{
				Source:  "href",
				Message: link,
			}
			if l.Level < Depth && inScope(link) && (Revisit || isUniqueURL(link)) {
				Queue <- item{
					URL:   link,
					Level: l.Level + 1,
				}
				Counter++
			}
		})

		doc.Find("script[src]").Each(func(index int, gitem *goquery.Selection) {
			src, _ := gitem.Attr("src")
			Results <- result{
				Source:  "script",
				Message: absoluteURL(l.URL, src),
			}
		})

		doc.Find("form").Each(func(index int, gitem *goquery.Selection) {
			action, _ := gitem.Attr("action")
			Results <- result{
				Source:  "form",
				Message: absoluteURL(l.URL, action),
			}
		})

		c1 <- 1
	}()

	// listen to timer and response, whichever happens first
	select {
	case _ = <-c1:
		Counter--
		return
	case <-time.After(time.Duration(timeout) * time.Second):
		Counter--
		return
	}
}

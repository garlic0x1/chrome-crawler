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
		// set cookies

		err := chromedp.Run(ctx,
			setCookies(),
		)

		var document string
		if l.Type == "href" {
			err := chromedp.Run(ctx,
				chromedp.Navigate(l.URL),
				chromedp.Evaluate(`document.getElementsByTagName('html')[0].innerHTML;`, &document),
			)
			if err != nil {
				c1 <- 1
				return
			}
		} else if l.Type == "form" {
			document = submitForm(l, ctx, c1)
		}

		oracle(document, l.URL)

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
					Type:  "href",
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
			method, _ := gitem.Attr("method")
			Results <- result{
				Source:  "form",
				Message: absoluteURL(l.URL, action),
			}

			f := item{
				Type:     "form",
				URL:      absoluteURL(l.URL, action),
				Location: l.URL,
				Level:    l.Level + 1,
				Method:   method,
			}

			gitem.Find("textarea").Each(func(index int, ginput *goquery.Selection) {
				name, _ := ginput.Attr("name")
				value, _ := ginput.Attr("value")
				flavor := "textarea"
				f.Inputs = append(f.Inputs, input{
					Type:  flavor,
					Name:  name,
					Value: value,
				})
			})
			gitem.Find("input").Each(func(index int, ginput *goquery.Selection) {
				name, _ := ginput.Attr("name")
				value, _ := ginput.Attr("value")
				flavor, _ := ginput.Attr("type")
				f.Inputs = append(f.Inputs, input{
					Type:  flavor,
					Name:  name,
					Value: value,
				})
			})
			if l.Level < Depth && inScope(f.URL) {
				Queue <- f
				Counter++
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

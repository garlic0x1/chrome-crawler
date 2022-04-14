package main

import (
	"context"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func crawl(l item, ctx context.Context) {
	c1 := make(chan int, 1)

	go func() {
		var document string
		if l.Type == "href" {
			// navigate to URL, and evaluate response
			document = getURL(l, ctx, c1)
		} else if l.Type == "form" {
			// submit form, and evaluate response
			document = submitForm(l, ctx, c1)
		}

		// search response for our injections
		oracle(document, l.URL)

		// create goquery object to process response
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(document))
		if err != nil {
			c1 <- 1
			return
		}

		// add all links to the queue
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

		// display all javascript files
		doc.Find("script[src]").Each(func(index int, gitem *goquery.Selection) {
			src, _ := gitem.Attr("src")
			Results <- result{
				Source:  "script",
				Message: absoluteURL(l.URL, src),
			}
		})

		// add forms to queue
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
			if l.Level < Depth && inScope(f.URL) && !Passive {
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

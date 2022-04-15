package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func timedCrawl(l item, ctx context.Context, timeout int) {
	c1 := make(chan int, 1)
	go func() {
		crawl(l, ctx)
		c1 <- 0
	}()
	select {
	case _ = <-c1:
		return
	case <-time.After(time.Duration(timeout) * time.Second):
		return
	}
}

func crawl(l item, ctx context.Context) {
	var document string
	if l.Type == "href" {
		// navigate to URL, and evaluate response
		document = getURL(l, ctx)
	} else if l.Type == "form" {
		// submit form, and evaluate response
		document = submitForm(l, ctx)
	}

	//log.Println(document)

	// search response for our injections
	oracle(document, l.URL)

	// create goquery object to process response
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(document))
	if err != nil {
		if Debug {
			log.Println(err)
		}
		return
	}

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
			pattern, _ := ginput.Attr("pattern")
			f.Inputs = append(f.Inputs, input{
				Type:    flavor,
				Name:    name,
				Value:   value,
				Pattern: pattern,
			})
		})
		if l.Level < Depth && inScope(f.URL) && !Passive {
			Queue <- f
			Counter++
		}
	})

	// add all links to the queue
	doc.Find("*[href]").Each(func(index int, gitem *goquery.Selection) {
		href, _ := gitem.Attr("href")
		href = strings.TrimSpace(href)
		link := absoluteURL(l.URL, href)
		Results <- result{
			Source:  "href",
			Message: link,
		}
		if l.Level < Depth && inScope(link) && (Revisit || isUniqueURL(link)) && filterImages(href) {
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
}

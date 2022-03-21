package main

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func crawl(l item, passctx context.Context, results chan string, queue chan item) {
	// open in a new tab
	ctx, cancel := chromedp.NewContext(passctx)
	defer cancel()

	c1 := make(chan int, 1)

	go func() {
		// run task list and store in slices
		var hrefs []string
		var formlist []item
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			if ev, ok := ev.(*page.EventJavascriptDialogOpening); ok {
				results <- "[reflected] " + ev.Message
			}
		})
		err := chromedp.Run(ctx,
			chromedp.Navigate(l.URL),
			chromedp.Evaluate(loadFile("getforms.js"), &formlist),
			chromedp.Evaluate(loadFile("getlinks.js"), &hrefs),
		)
		if err != nil {
			log.Println(err, l.URL)
			return
		}

		for _, href := range hrefs {
			ret := item{
				URL:   href,
				Level: l.Level + 1,
			}
			results <- "[href] " + ret.URL

			if REVISIT {
				if ret.Level < DEPTH && inScope(ret.URL) {
					// increment counter for every link found so we know to not stop yet
					COUNTER++
					// send back to queue to be further crawled
					queue <- ret
				}
			} else {
				if ret.Level < DEPTH && inScope(ret.URL) && isUniqueURL(ret.URL) {
					// increment counter for every link found so we know to not stop yet
					COUNTER++
					// send back to queue to be further crawled
					queue <- ret
				}
			}
		}

		//log.Println(formlist)
		for _, f := range formlist {
			if f.Reflected == "true" {
				results <- "[reflected] " + f.URL
			}

			results <- "[form] " + f.Method + " " + f.URL
		}
		c1 <- 1
	}()

	// listen to timer and response, whichever happens first
	select {
	case _ = <-c1:
		COUNTER--
		return
	case <-time.After(time.Duration(timeout) * time.Second):
		COUNTER--
		return
	}
}

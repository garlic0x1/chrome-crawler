package main

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

type forms struct {
	Forms  []item
	Hashes []string
}

func crawl(l item, passctx context.Context, results chan string, queue chan item) {
	// open in a new tab
	ctx, cancel := chromedp.NewContext(passctx)
	defer cancel()

	c1 := make(chan int, 1)

	go func() {
		// run task list and store in slices
		var hrefs []string
		var forms forms
		err := chromedp.Run(ctx,
			chromedp.Navigate(l.URL),
			chromedp.Evaluate(loadFile("getlinks.js"), &hrefs),
			chromedp.Evaluate(loadFile("getforms.js"), &forms),
		)
		if err != nil {
			log.Println(err, l.URL)
			return
		}
		// dont leave it open longer than we need

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

		for _, f := range forms.Forms {
			//	log.Println("gothere", f)

			ret := item{
				URL:    f.URL,
				Level:  l.Level + 1,
				Method: f.Method,
				Inputs: f.Inputs,
			}

			results <- "[form] " + ret.Method + " " + ret.URL
			//queue <- ret
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

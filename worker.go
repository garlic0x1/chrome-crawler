package main

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func crawl(l item, ctx context.Context) {
	c1 := make(chan int, 1)

	go func() {
		// run task list and store in slices
		var hrefs []string
		var formlist []item
		var scripts []string
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			if ev, ok := ev.(*page.EventJavascriptDialogOpening); ok {
				Results <- result{
					Source:  "reflector",
					Message: ev.Message,
				}
			}
		})
		err := chromedp.Run(ctx,
			chromedp.Navigate(l.URL),
			chromedp.Evaluate(loadFile("getforms.js"), &formlist),
			chromedp.Evaluate(loadFile("getlinks.js"), &hrefs),
			chromedp.Evaluate(loadFile("getscripts.js"), &scripts),
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
			Results <- result{
				Source:  "href",
				Message: ret.URL,
			}

			if Revisit {
				if ret.Level < Depth && inScope(ret.URL) {
					// increment counter for every link found so we know to not stop yet
					Counter++
					// send back to queue to be further crawled
					Queue <- ret
				}
			} else {
				if ret.Level < Depth && inScope(ret.URL) && isUniqueURL(ret.URL) {
					// increment counter for every link found so we know to not stop yet
					Counter++
					// send back to queue to be further crawled
					Queue <- ret
				}
			}
		}

		for _, script := range scripts {
			Results <- result{
				Source:  "script",
				Message: script,
			}
		}

		for _, f := range formlist {
			if f.Reflected == "true" {
				Results <- result{
					Source:  "reflector",
					Message: f.URL,
				}
			}

			Results <- result{
				Source:  "form",
				Message: f.URL,
			}

		}
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

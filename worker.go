package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/chromedp/chromedp"
)

func crawl(l item, passctx context.Context, results chan string, queue chan item) {
	// open in a new tab
	ctx, cancel := chromedp.NewContext(passctx)

	// run task list and store in slices
	var hrefs []string
	var forms []string
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
	cancel()

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

	for _, f := range forms {
		var umForm item

		//	log.Println("gothere", f)

		json.Unmarshal([]byte(f), &umForm)
		ret := item{
			URL:    umForm.URL,
			Level:  l.Level + 1,
			Method: umForm.Method,
			Inputs: umForm.Inputs,
		}

		results <- "[form] " + ret.Method + " " + ret.URL
		queue <- ret
	}

	// decrement counter for having looked at this link AFTER counting the child links
	COUNTER--
}

func submitForm(f item, passctx context.Context, results chan string, queue chan item) {
	// open in a new tab
	ctx, cancel := chromedp.NewContext(passctx)

	err := chromedp.Run(ctx,
		chromedp.Navigate(f.URL))
	if err != nil {
		log.Println(err, f.URL)
		return
	}

	for _, i := range f.Inputs {
		var q string
		sel := fmt.Sprintf(`//input[@name="%s"]`, i.Name)
		if i.Value == "" {
			q = randomString(8)
		} else {
			q = i.Value
		}
		err = chromedp.Run(ctx,
			chromedp.SendKeys(sel, q),
			chromedp.Submit(sel))
		if err != nil {
			log.Println(err, f.URL)
			return
		}
	}

	// run task list and store in slices
	var hrefs []string
	var forms []string
	err = chromedp.Run(ctx,
		chromedp.Evaluate(loadFile("getlinks.js"), &hrefs),
		chromedp.Evaluate(loadFile("getforms.js"), &forms),
	)
	if err != nil {
		log.Println(err, f.URL)
		return
	}
	// dont leave it open longer than we need
	cancel()

	for _, href := range hrefs {
		ret := item{
			URL:   href,
			Level: f.Level + 1,
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

	for _, f := range forms {
		results <- "[form] " + f
	}

	// decrement counter for having looked at this link AFTER counting the child links
	COUNTER--
}

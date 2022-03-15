// Command eval is a chromedp example demonstrating how to evaluate javascript
// and retrieve the result.
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

type link struct {
	URL   string
	Level int
}

type input struct {
	Type  string
	Name  string
	Value string
}

type form struct {
	Method string
	Action string
	Inputs []input
	Level  int
}

type injection struct {
	Hash         string
	FormLocation string
}

// Globals
var (
	DEPTH        = 2
	injectionMap []injection
	seededRand   *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func absoluteURL(protocol string, host string, u string) string {
	if len(u) > 8 {
		if u[:8] == "https://" || u[:7] == "http://" {
			return u
		}
	} else if u[0] == '/' {
		return protocol + "://" + host + u
	}
	return protocol + "://" + host + "/" + u

}

func crawl(l link, queue chan link) {
	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// parse link
	parsed, err := url.Parse(l.URL)
	if err != nil {
		log.Println("failed to parse url", l.URL, err)
	}
	protocol := parsed.Scheme
	host := parsed.Host

	// run task list
	var hrefs []*cdp.Node
	var forms []*cdp.Node
	err = chromedp.Run(ctx,
		chromedp.Navigate(l.URL),
		chromedp.Nodes("a", &hrefs),
		chromedp.Nodes("form", &forms),
	)
	if err != nil {
		log.Println(err, l.URL)
	}

	for _, href := range hrefs {
		if href.AttributeValue("href") != "" {
			log.Println("link", href.AttributeValue("href"))
			l := link{
				URL:   absoluteURL(protocol, host, href.AttributeValue("href")),
				Level: l.Level + 1,
			}

			if l.Level < DEPTH {
				queue <- l
			}
		}
	}

	for _, f := range forms {
		if f.AttributeValue("action") != "" {
			log.Println("form", f.AttributeValue("action"))
		}
	}
}

func main() {

	queue := make(chan link, 4)

	startlink := link{
		URL:   "https://garlic0x1.com",
		Level: 0,
	}

	go crawl(startlink, queue)

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for l := range queue {
		fmt.Fprintln(w, l.URL)
		go crawl(l, queue)
	}
}

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
	}
	if string(u[:1]) == "/" {

		return protocol + "://" + host + u
	}
	return protocol + "://" + host + "/" + u

	//log.Println("protocol:", protocol, "host:", host, "u:", u, "u[:1]:", u[:1])
}

func crawl(l link, timeout time.Duration, queue chan link) {
	// create context
	ctxbase, cancel := chromedp.NewContext(context.Background())

	ctx, cancel := context.WithTimeout(ctxbase, timeout*time.Second)
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
			//log.Println("link", href.AttributeValue("href"))
			ret := link{
				URL:   absoluteURL(protocol, host, href.AttributeValue("href")),
				Level: l.Level + 1,
			}

			if ret.Level < DEPTH {
				queue <- ret
			}
		}
	}

	for _, f := range forms {
		if f.AttributeValue("action") != "" {
			//log.Println("form", f.AttributeValue("action"))
		}
	}
}

func main() {

	queue := make(chan link, 4)

	startlink := link{
		URL:   "https://www.tiktok.com",
		Level: 0,
	}

	go crawl(startlink, time.Duration(10), queue)

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for l := range queue {
		fmt.Println(l.URL)
		go crawl(l, time.Duration(10), queue)
	}
}

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"sync"
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
	sm           sync.Map
	DEPTH        int
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

func crawl(l link, passctx context.Context, sem chan struct{}, timeout time.Duration) {
	ctx, cancel := chromedp.NewContext(passctx)
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
		//log.Println(err, l.URL)
	}

	var wg sync.WaitGroup
	for _, href := range hrefs {
		if href.AttributeValue("href") != "" {
			ret := link{
				URL:   absoluteURL(protocol, host, href.AttributeValue("href")),
				Level: l.Level + 1,
			}
			if isUnique(ret.URL) {
				fmt.Println("link", ret.URL)
			}

			if ret.Level < DEPTH {
				select {
				case sem <- struct{}{}:
					wg.Add(1)
					go func() {
						crawl(ret, ctx, sem, timeout)
						<-sem
						wg.Done()
					}()
				default:
					crawl(ret, ctx, sem, timeout)
				}
			}
		}
	}

	for _, f := range forms {
		if f.AttributeValue("action") != "" {
			//log.Println("form", f.AttributeValue("action"))
		}
	}
	wg.Wait()
}

func isUnique(url string) bool {
	_, present := sm.Load(url)
	if present {
		return false
	}
	sm.Store(url, true)
	return true
}

func main() {
	threads := flag.Int("tabs", 8, "Number of chrome tabs to use concurrently")
	depth := flag.Int("depth", 2, "Depth to crawl")
	//unique := flag.Bool("unique", false, "Show only unique urls")
	u := flag.String("url", "", "URL to crawl")
	flag.Parse()
	DEPTH = *depth + 1

	if *u == "" {
		fmt.Println("Please provide a url with -url")
		os.Exit(0)
	}

	timeout := time.Duration(10)
	//queue := make(chan link, 4)
	// set up concurrency limit
	sem := make(chan struct{}, *threads)

	// create context
	ctxbase, cancel := chromedp.NewContext(context.Background())

	ctx, cancel := context.WithTimeout(ctxbase, timeout*time.Second)
	defer cancel()

	startlink := link{
		URL:   *u,
		Level: 0,
	}

	crawl(startlink, ctx, sem, timeout)

}

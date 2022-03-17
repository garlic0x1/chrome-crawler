package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

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
	sm              sync.Map
	DEPTH           int
	SCOPE           string
	javascriptfuncs [2]string
	injectionMap    []injection
	seededRand      *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func crawl(l link, passctx context.Context, results chan string, sem chan struct{}) {
	var wg sync.WaitGroup

	// open in a new tab
	ctx, cancel := chromedp.NewContext(passctx)

	// run task list
	var hrefs []string
	var forms []string
	err := chromedp.Run(ctx,
		chromedp.Navigate(l.URL),
		chromedp.Evaluate(javascriptfuncs[0], &hrefs),
		chromedp.Evaluate(javascriptfuncs[1], &forms),
	)
	if err != nil {
		log.Println(err, l.URL)
		return
	}
	cancel()
	//fmt.Println(hrefs)

	for _, href := range hrefs {
		ret := link{
			URL:   href,
			Level: l.Level + 1,
		}
		results <- "[href] " + ret.URL

		if ret.Level < DEPTH && inScope(ret.URL) {
			select {
			case sem <- struct{}{}:
				wg.Add(1)
				go func() {
					//log.Println("crawling", ret.URL)
					crawl(ret, passctx, results, sem)
					<-sem
					wg.Done()
				}()
			default:
				crawl(ret, passctx, results, sem)
			}
		}
	}

	for _, f := range forms {
		//log.Println("form", f.AttributeValue("action"))
		results <- "[form] " + f
	}
	wg.Wait()
}

func inScope(u string) bool {
	return strings.Contains(u, SCOPE)
}
func isUnique(u string) bool {
	_, present := sm.Load(u)
	if present {
		return false
	}
	sm.Store(u, true)
	return true
}

func loadFiles() {
	// open files
	content, err := ioutil.ReadFile("getforms.js")
	if err != nil {
		log.Fatal(err)
	}
	javascriptfuncs[1] = string(content)
	content, err = ioutil.ReadFile("getlinks.js")
	if err != nil {
		log.Fatal(err)
	}
	javascriptfuncs[0] = string(content)
}

func main() {
	loadFiles()
	threads := flag.Int("tabs", 8, "Number of chrome tabs to use concurrently")
	//timeoutarg := flag.Int("timeout", 10, "Timeout in seconds")
	depth := flag.Int("depth", 2, "Depth to crawl")
	unique := flag.Bool("unique", false, "Show only unique urls")
	u := flag.String("url", "", "URL to crawl")
	flag.Parse()
	DEPTH = *depth
	// parse link
	parsed, err := url.Parse(*u)
	if err != nil {
		log.Println("failed to parse url", *u, err)
	}
	SCOPE = parsed.Host

	if *u == "" {
		fmt.Println("Please provide a url with -url")
		os.Exit(0)
	}

	//timeout := time.Duration(*timeoutarg)
	//queue := make(chan link, 4)
	// set up concurrency limit
	var wg sync.WaitGroup
	sem := make(chan struct{}, *threads)
	// results channel
	results := make(chan string)

	// create context
	ctx, cancel := chromedp.NewContext(context.Background())

	defer cancel()

	startlink := link{
		URL:   *u,
		Level: 0,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		crawl(startlink, ctx, results, sem)
		close(results)
	}()

	go func() {
		if *unique {
			for res := range results {
				if isUnique(res) {
					fmt.Println(res)
				}
			}
		}
		for res := range results {
			fmt.Println(res)
		}
	}()

	wg.Wait()
}

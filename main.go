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
	sm           sync.Map
	DEPTH        int
	SCOPE        string
	COUNTER      int
	injectionMap []injection
	seededRand   *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// spawns n workers listening to queue
func spawnWorkers(n int, passctx context.Context, results chan string, queue chan link) {
	for i := 0; i < n; i++ {
		go func() {
			// pops messages
			for message := range queue {
				crawl(message, passctx, results, queue)
			}
		}()
	}
}

func crawl(l link, passctx context.Context, results chan string, queue chan link) {
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
		ret := link{
			URL:   href,
			Level: l.Level + 1,
		}
		results <- "[href] " + ret.URL

		if ret.Level < DEPTH && inScope(ret.URL) {
			// increment counter for every link found so we know to not stop yet
			COUNTER++
			// send back to queue to be further crawled
			queue <- ret
		}
	}

	for _, f := range forms {
		results <- "[form] " + f
	}

	// decrement counter for having looked at this link AFTER counting the child links
	COUNTER--
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

// load the javascript functions
func loadFile(filename string) string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

func main() {
	threads := flag.Int("tabs", 8, "Number of chrome tabs to use concurrently")
	//timeoutarg := flag.Int("timeout", 10, "Timeout in seconds")
	depth := flag.Int("depth", 2, "Depth to crawl")
	unique := flag.Bool("uniq", false, "Show only unique urls")
	u := flag.String("u", "", "URL to crawl")
	flag.Parse()

	DEPTH = *depth

	// parse link to determine scope
	parsed, err := url.Parse(*u)
	if err != nil {
		log.Println("failed to parse url", *u, err)
	}
	SCOPE = parsed.Host

	if *u == "" {
		fmt.Println("Please provide a url with -u")
		os.Exit(0)
	}

	startlink := link{
		URL:   *u,
		Level: 0,
	}

	queue := make(chan link)
	results := make(chan string)
	COUNTER = 1

	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	go func() { queue <- startlink }()
	go func() { spawnWorkers(*threads, ctx, results, queue) }()

	// listen to results and output them
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
	for {
		if COUNTER < 1 {
			os.Exit(0)
		}
	}
}

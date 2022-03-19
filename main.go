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

	"github.com/chromedp/chromedp"
)

type input struct {
	Type  string
	Name  string
	Value string
}

type item struct {
	URL   string
	Level int

	// these values are set for forms only
	Method string
	Inputs []input
}

// Globals
var (
	sm         sync.Map
	visited    sync.Map
	REVISIT    bool
	DEPTH      int
	SCOPE      string
	COUNTER    int
	seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
)

// spawns n workers listening to queue
func spawnWorkers(n int, passctx context.Context, results chan string, queue chan item) {
	for i := 0; i < n; i++ {
		go func() {
			// pops messages
			for message := range queue {
				if len(message.Inputs) < 1 {
					crawl(message, passctx, results, queue)
				} else {
					submitForm(message, passctx, results, queue)
				}
			}
		}()
	}
}

func main() {
	threads := flag.Int("t", 8, "Number of chrome tabs to use concurrently")
	depth := flag.Int("d", 2, "Depth to crawl")
	unique := flag.Bool("uniq", false, "Show only unique URLs")
	revisit := flag.Bool("r", false, "Revisit URLs")
	u := flag.String("u", "", "URL to crawl")
	flag.Parse()

	DEPTH = *depth
	REVISIT = *revisit

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

	startlink := item{
		URL:   *u,
		Level: 0,
	}

	queue := make(chan item)
	results := make(chan string)
	COUNTER = 1

	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	go spawnWorkers(*threads, ctx, results, queue)

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

	queue <- startlink

	for {
		if COUNTER < 1 {
			os.Exit(0)
		}
	}
}

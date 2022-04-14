package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type input struct {
	Type       string
	Name       string
	Value      string
	Identifier string
}

type result struct {
	Source  string
	Message string
}

type item struct {
	Type  string
	URL   string
	Level int

	// these values are set for forms only
	Location string
	Method   string
	Inputs   []input
}

// Globals
var (
	Canary           = "http://zzx%djy"
	sm               sync.Map
	mu               = &sync.Mutex{}
	Injections       = make([]item, 0)
	Cookies          []*network.Cookie
	visited          sync.Map
	timeout          = 30
	InjectionCounter = 0
	Revisit          bool
	Depth            int
	Scope            []string
	Counter          int
	ShowSource       bool
	seededRand       *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	ChromeCtx context.Context
	Results   chan result
	Queue     chan item
)

func oracle(doc string, u string) {
	// check the response for injected stuff
	for i, inj := range Injections {
		for _, inp := range inj.Inputs {
			if inp.Identifier != "" && strings.Contains(doc, inp.Identifier) {
				Results <- result{
					Source:  "reflect",
					Message: Injections[i].URL + " -> " + u,
				}
			}
		}
	}
}

func writer(unique *bool) {
	for res := range Results {
		if !(*unique) || isUnique(res.Source+res.Message) {
			if ShowSource {
				fmt.Println("["+res.Source+"]", res.Message)
			} else {
				fmt.Println(res.Message)
			}
		}
	}
}

// goroutine to handle input
func reader() {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		u := s.Text()
		if Counter == -1 {
			Counter += 2
		} else {
			Counter++
		}
		// parse link to determine scope
		parsed, err := url.Parse(u)
		if err != nil {
			log.Println("failed to parse url", u, err)
		}
		Scope = append(Scope, parsed.Host)
		Queue <- item{
			Type:  "href",
			URL:   u,
			Level: 1,
		}
	}
}

// spawns n workers listening to queue
func spawnWorkers(n int) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			tab, cancel := chromedp.NewContext(ChromeCtx)
			defer cancel()
			// pop from queue
			for message := range Queue {
				crawl(message, tab)
			}
			wg.Done()
		}()
	}

	// wait for all jobs to be finished, then end program
	wg.Wait()
	close(Results)
}

func main() {
	threads := flag.Int("t", 8, "Number of chrome tabs to use concurrently.")
	depth := flag.Int("d", 2, "Depth to crawl.")
	unique := flag.Bool("u", false, "Show only unique URLs.")
	revisit := flag.Bool("r", false, "Revisit URLs.")
	showSource := flag.Bool("s", false, "Show source.")
	debug := flag.Bool("debug", false, "Don't use headless. (slow but fun to watch)")
	proxy := flag.String(("proxy"), "", "Proxy URL. Example: -proxy http://127.0.0.1:8080")

	flag.Parse()
	ShowSource = *showSource
	Depth = *depth
	Revisit = *revisit
	Counter = -1

	Queue = make(chan item)
	Results = make(chan result)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer(*proxy),
		// block all images
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.Flag("headless", !(*debug)))...)
	ChromeCtx = ctx
	defer cancel()

	go reader()
	// start workers with their own routines
	go spawnWorkers(*threads)
	go writer(unique)
	for true {
		//log.Println(Counter)
		if Counter == 0 {
			os.Exit(0)
		}
	}
}

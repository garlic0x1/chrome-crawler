package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"gopkg.in/yaml.v2"
)

type input struct {
	Type       string
	Name       string
	Value      string
	Identifier string
	Pattern    string
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

type result struct {
	Type      string
	URL       string
	Injection item
}

// Globals
var (
	Debug            bool
	Passive          bool
	Canary           = "zzx%djy"
	sm               sync.Map
	InjectionMap     sync.Map
	Headers          = make(map[string]interface{})
	Injections       = make([]item, 0)
	Cookies          []*network.Cookie
	visited          sync.Map
	InjectionCounter = 0
	Revisit          bool
	Depth            int
	Wait             int
	Scope            []string
	Counter          int
	ShowSource       bool
	seededRand       *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	ChromeCtx context.Context
	Results   chan result
	Queue     chan item
)

func writer(unique *bool, pjson *bool, pyaml *bool) {
	for res := range Results {
		if !(*unique) || isUnique(res.Type+res.URL+res.Injection.URL) {
			switch {
			case *pjson:
				b, err := json.Marshal(res)
				if err != nil {
					log.Println("Error:", err)
					continue
				}
				fmt.Println(string(b))
			case *pyaml:
				b, err := yaml.Marshal(res)
				if err != nil {
					log.Println("Error:", err)
					continue
				}
				fmt.Println(string(b))
			case !*pjson && !*pyaml:
				str := ""
				if res.Type == "reflect" {
					str += res.Injection.URL + " -> "
				}
				if ShowSource {
					fmt.Println("["+res.Type+"]", str+res.URL)
				} else {
					fmt.Println(str + res.URL)
				}
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
		if err != nil && Debug {
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
func spawnWorkers(n int, timeout int, done chan string) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			tab, cancel := chromedp.NewContext(ChromeCtx)
			defer cancel()
			// pop from queue
			for message := range Queue {
				timedCrawl(message, tab, timeout)
				//crawl(message, tab)
				Counter--
				if Counter == 0 {
					done <- "Counter == 0"
				}
			}
			wg.Done()
		}()
	}

	// wait for all jobs to be finished, then end program
	wg.Wait()
	close(Results)
}

func main() {
	unique := flag.Bool("u", false, "Show only unique URLs.")
	Json := flag.Bool("json", false, "Output as JSON.")
	Yaml := flag.Bool("yaml", false, "Output as YAML.")
	showSource := flag.Bool("s", false, "Show source.")
	threads := flag.Int("t", 10, "Number of chrome tabs to use concurrently.")
	depth := flag.Int("d", 2, "Depth to crawl.")
	revisit := flag.Bool("r", false, "Revisit URLs.")
	wait := flag.Int("w", 0, "Seconds to wait for DOM to load. (Use to find injections from AJAX reqs)")
	active := flag.Bool("p", false, "Find injection points.")
	cheaders := flag.String("head", "", "Custom headers separated by two semi-colons. Example: -h 'Cookie: foo=bar;;Referer: http://example.com/'")
	debugChrome := flag.Bool("debug-chrome", false, "Don't use headless. (slow but fun to watch)")
	debug := flag.Bool("debug", false, "Display error messages.")
	proxy := flag.String("proxy", "", "Proxy URL. Example: -proxy http://127.0.0.1:8080")
	timeout := flag.Int("time", 10, "Timeout per request.")

	flag.Parse()
	Wait = *wait
	Debug = *debug
	Passive = !(*active)
	ShowSource = *showSource
	Depth = *depth
	Revisit = *revisit
	Counter = -1

	err := parseHeaders(*cheaders)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing headers:", err)
		os.Exit(1)
	}

	Queue = make(chan item)
	Results = make(chan result)
	done := make(chan string, 1)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer(*proxy),
		// block all images
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.Flag("headless", !(*debugChrome)))...)
	ChromeCtx = ctx
	defer cancel()

	go reader()
	go spawnWorkers(*threads, *timeout, done)
	go writer(unique, Json, Yaml)

	_ = <-done
	return
}

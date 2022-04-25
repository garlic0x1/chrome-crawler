package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/lucasjones/reggen"
)

func oracle(doc string, u string) {
	// check the response for injected stuff
	for i := 0; i < InjectionCounter; i++ {
		if strings.Contains(doc, fmt.Sprintf(Canary, i)) {
			inj, ok := getInjection(fmt.Sprintf(Canary, i))
			if ok {
				Results <- result{
					Type:      "reflect",
					URL:       u,
					Injection: inj,
				}
			}
		}
	}
}

func filterImages(u string) bool {

	filters := []string{
		".jpg",
		".png",
		".jpeg",
		".ico",
		".pdf",
		".css",
	}

	parsed, err := url.Parse(u)
	if err != nil {
		if Debug {
			log.Println(err)
		}
		return false
	}
	for _, str := range filters {
		if strings.HasSuffix(parsed.Path, str) {
			return false
		}
	}
	return true
}

// parseHeaders does validation of headers input and saves it to a formatted map.
func parseHeaders(rawHeaders string) error {
	if rawHeaders != "" {
		if !strings.Contains(rawHeaders, ":") {
			return errors.New("headers flag not formatted properly (no colon to separate header and value)")
		}

		rawHeaders := strings.Split(rawHeaders, ";;")
		for _, header := range rawHeaders {
			var parts []string
			if strings.Contains(header, ": ") {
				parts = strings.SplitN(header, ": ", 2)
			} else if strings.Contains(header, ":") {
				parts = strings.SplitN(header, ":", 2)
			} else {
				continue
			}
			Headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return nil
}

func generateMatch(pattern string) string {
	// generate a single string
	str, err := reggen.Generate(pattern, 1)
	if err != nil {
		panic(err)
	}

	return str
}

func addInjection(canary string, f item) {
	InjectionMap.Store(canary, f)
}

func getInjection(canary string) (item, bool) {
	ret, ok := InjectionMap.Load(canary)
	return ret.(item), ok
}

func inScope(u string) bool {
	for _, host := range Scope {
		if strings.Contains(u, host) {
			return true
		}
	}
	return false
}

func isUnique(u string) bool {
	_, present := sm.Load(u)
	if present {
		return false
	}
	sm.Store(u, true)
	return true
}

func isUniqueURL(u string) bool {
	_, present := visited.Load(u)
	if present {
		return false
	}
	visited.Store(u, true)
	return true
}

func absoluteURL(parent string, u string) string {
	parsed, err := url.Parse(parent)
	if err != nil {
		log.Println(err)
	}
	if strings.HasPrefix(u, "http") {
		return u
	} else if strings.HasPrefix(u, "//") {
		return parsed.Scheme + ":" + u
	} else if !(strings.HasPrefix(u, "/")) {
		u = "/" + u
	}
	return parsed.Scheme + "://" + parsed.Host + u
}

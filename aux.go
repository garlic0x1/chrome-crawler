package main

import (
	"log"
	"net/url"
	"strings"
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

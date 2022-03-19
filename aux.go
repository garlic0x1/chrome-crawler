package main

import (
	"io/ioutil"
	"log"
	"strings"
)

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

func isUniqueURL(u string) bool {
	_, present := visited.Load(u)
	if present {
		return false
	}
	visited.Store(u, true)
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

// returns a random alphabetical string of provided length
func randomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

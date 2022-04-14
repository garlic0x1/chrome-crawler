package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func submitForm(f item, ctx context.Context, c1 chan int) string {
	var doc string

	err := chromedp.Run(ctx,
		chromedp.Navigate(f.Location),
	)
	if err != nil {
		log.Println(err)
		c1 <- 1
		return ""
	}

	sel := ""
	var injectedInputs []input
	for _, inp := range f.Inputs {
		if inp.Name != "" {
			if inp.Type == "textarea" {
				sel = fmt.Sprintf("textarea[name=%s]", inp.Name)
			} else {
				sel = fmt.Sprintf("input[name=%s]", inp.Name)
			}
			q := ""
			if inp.Type == "hidden" {
				continue
				q = inp.Value
			} else {
				injectedInputs = append(injectedInputs, input{
					Name:       inp.Name,
					Identifier: fmt.Sprintf(Canary, InjectionCounter),
				})
				if inp.Type != "email" {
					q = fmt.Sprintf(Canary, InjectionCounter)
					InjectionCounter++
				} else {
					q = fmt.Sprintf(Canary, InjectionCounter) + "@gmail.com"
					InjectionCounter++
				}
			}
			err = chromedp.Run(ctx,
				chromedp.SendKeys(sel, q),
			)
			if err != nil {
				log.Println(err)
				c1 <- 1
				return ""
			}
		}
	}
	mu.Lock()
	Injections = append(Injections, item{
		Type:     f.Type,
		URL:      f.URL,
		Level:    f.Level,
		Location: f.Location,
		Method:   f.Method,
		Inputs:   injectedInputs,
	})
	mu.Unlock()

	err = chromedp.Run(ctx,
		chromedp.Submit(sel),
		chromedp.Sleep(time.Duration(Wait)*time.Second),
		// read network values
		chromedp.ActionFunc(func(ctx context.Context) error {
			cookies, err := network.GetAllCookies().Do(ctx)
			if err != nil {
				log.Println(err)
				c1 <- 1
				return err
			}

			if len(cookies) > 0 {
				Cookies = cookies
			}

			return nil
		}),
		chromedp.Evaluate(`try { document.documentElement.innerHTML; } catch { "" }`, &doc),
	)
	if err != nil {
		log.Println("Error submitting form:", err)
		c1 <- 1
		return ""
	}

	return doc
}

func setCookies() chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			// add cookies to chrome
			for _, cookie := range Cookies {
				err := network.SetCookie(cookie.Name, cookie.Value).Do(ctx)
				if err != nil {
					return err
				}
			}
			return nil
		}),
	}
}

/*
func SetCookie(name, value, domain, path string, httpOnly, secure bool) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
		err := network.SetCookie(name, value).
			WithExpires(&expr).
			WithDomain(domain).
			WithPath(path).
			WithHTTPOnly(httpOnly).
			WithSecure(secure).
			Do(ctx)
		if err != nil {
			return err
		}
		return nil
	})
}

func ShowCookies() chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		cookies := network.GetCookiesParams.Do(ctx)
		log.Println(cookies)
		return nil
	})
}
*/

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

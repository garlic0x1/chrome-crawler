package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func getURL(l item, ctx context.Context, c1 chan int) string {
	var doc string
	err := chromedp.Run(ctx,
		setCookies(),
		chromedp.Navigate(l.URL),
		chromedp.Sleep(time.Duration(Wait)*time.Second),
		chromedp.Evaluate(`try { document.documentElement.outerHTML; } catch { "" }`, &doc),
	)
	if err != nil {
		c1 <- 1
		return ""
	}
	return doc
}

func submitForm(f item, ctx context.Context, c1 chan int) string {
	var doc string

	err := chromedp.Run(ctx,
		chromedp.Navigate(f.Location),
		setCookies,
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

	// make append thread safe (I think this is working? if not can switch to sync.Map)
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
		chromedp.Evaluate(`try { document.documentElement.outerHTML; } catch { "" }`, &doc),
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

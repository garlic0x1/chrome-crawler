package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func getURL(l item, ctx context.Context) string {
	parsed, _ := url.Parse(l.URL)
	var doc string
	err := chromedp.Run(ctx,
		network.Enable(),
		network.SetExtraHTTPHeaders(network.Headers(Headers)),
		setCookies(parsed.Host),
		chromedp.Navigate(l.URL),
		chromedp.Sleep(time.Duration(Wait)*time.Second),
		//chromedp.Evaluate(`document.documentElement.outerHTML;`, &doc),
		chromedp.OuterHTML(`html`, &doc),
	)
	if err != nil {
		log.Println(l, err)
		return ""
	}
	return doc
}

func submitForm(f item, ctx context.Context) string {
	parsed, _ := url.Parse(f.URL)
	var doc string
	err := chromedp.Run(ctx,
		network.Enable(),
		network.SetExtraHTTPHeaders(network.Headers(Headers)),
		setCookies(parsed.Host),
		chromedp.Navigate(f.Location),
	)
	if err != nil {
		log.Println("Error setting cookies:", err)
		return ""
	}

	sel := ""
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
				addInjection(fmt.Sprintf(Canary, InjectionCounter), f)
				if inp.Type != "email" {
					if inp.Pattern != "" {
						pattern := inp.Pattern + "(" + fmt.Sprintf(Canary, InjectionCounter) + ")"
						q = generateMatch(pattern)
						InjectionCounter++
					} else {
						q = fmt.Sprintf(Canary, InjectionCounter)
						InjectionCounter++
					}
				} else {
					q = fmt.Sprintf(Canary, InjectionCounter) + "@gmail.com"
					InjectionCounter++
				}
			}
			err = chromedp.Run(ctx,
				chromedp.SendKeys(sel, q),
			)
			if err != nil {
				log.Println("Error sending keys:", sel, q, err)
				return ""
			}
		}
	}

	err = chromedp.Run(ctx,
		chromedp.Submit(sel),
		chromedp.Sleep(time.Duration(Wait)*time.Second),
		getCookies(),
		//chromedp.Evaluate(`document.documentElement.outerHTML;`, &doc),
		chromedp.OuterHTML(`html`, &doc),
	)
	if err != nil {
		log.Println("Error submitting form:", f, err)
		return ""
	}

	return doc
}

func getCookies() chromedp.Tasks {
	return chromedp.Tasks{
		// read network values
		chromedp.ActionFunc(func(ctx context.Context) error {
			cookies, err := network.GetAllCookies().Do(ctx)
			if err != nil {
				log.Println("Error reading cookies:", err)
				return err
			}

			if len(cookies) > 0 {
				Cookies = cookies
			}

			return nil
		}),
	}
}

func setCookies(domain string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			// add cookies to chrome
			for _, cookie := range Cookies {
				err := network.SetCookie(cookie.Name, cookie.Value).WithDomain(domain).Do(ctx)
				if err != nil {
					return err
				}
			}
			return nil
		}),
	}
}

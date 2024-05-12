package scraper

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

func ScrapWebPage(url string, cssSelector string) ([]string, error) {
	paragraphs := []string{}

	c := colly.NewCollector()

	// c.OnRequest(func(r *colly.Request) {
	// 	fmt.Println("Visiting: ", r.URL)
	// })

	// c.OnError(func(_ *colly.Response, err error) {
	// 	log.Println("Something went wrong: ", err)
	// })

	// c.OnResponse(func(r *colly.Response) {
	// 	fmt.Println("Page visited: ", r.Request.URL)
	// })

	c.OnHTML(cssSelector, func(e *colly.HTMLElement) {
		text := e.Text
		text = strings.Trim(text, "\r\n")
		text = strings.Trim(text, "\t\n\v\f\r ")
		text = strings.TrimSpace(text)
		if text != "" {
			paragraphs = append(paragraphs, e.Text)
		}
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println(r.Request.URL, " scraped!")
	})

	c.Visit(url)

	return paragraphs, nil
}

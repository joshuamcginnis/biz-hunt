package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/runner"
)

const depth int = 2
const baseURL string = "https://www.bizbuysell.com"
const startURL string = "https://www.bizbuysell.com/california/los-angeles-county-businesses-for-sale"

// Listing represents the details from an individual listing
type Listing struct {
	// URL                string
	Title       string
	Location    string
	AskingPrice string
	CashFlow    string
	// GrossRevenue       uint
	// Ebitda             uint
	// FFE                uint
	// InventoryValue     uint
	// Rent               string
	// Established        uint
	Description string
	// Inventory          string
	// RealEstate         string
	// BuildingSF         uint
	// LeaseExpiration    string
	// Facilities         string
	// SupportAndTraining string
	// ReasonForSelling   string
}

func main() {
	var err error

	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := chromedp.NewPool()
	defer pool.Shutdown()

	c, err := pool.Allocate(ctxt)
	if err != nil {
		log.Fatalln("Error starting CDP instance.", err)
	}

	u := "https://www.bizbuysell.com/Business-Opportunity/Public-Works-Construction-7MM-Multi-Year-Contracts-Key-Employees/1577140/"

	err = c.Run(ctxt, loadListing(u))
	if err != nil {
		log.Fatalln("Error navigating to url", err)
	}

	var title string
	if err = c.Run(ctxt, chromedp.Text(".bfsTitle", &title, chromedp.BySearch)); err != nil {
		log.Fatalln("Could not get title.", err)
	}

	var description string
	if err = c.Run(ctxt, chromedp.Text(".businessDescription", &description, chromedp.BySearch)); err != nil {
		log.Fatalln("Could not get description.", err)
	}

	var askingPrice string
	if err = c.Run(ctxt, chromedp.Text(".price b", &askingPrice, chromedp.BySearch)); err != nil {
		log.Fatalln("Could not get askingPrice.", err)
	}

	var location string
	if err = c.Run(ctxt, chromedp.Text(".row-fluid h2.gray", &location, chromedp.BySearch)); err != nil {
		log.Fatalln("Could not get location.", err)
	}

	var cashFlow string
	if err = c.Run(ctxt, chromedp.Text("div.row-fluid.lgFinancials > div:nth-child(2) > p > b", &cashFlow, chromedp.BySearch)); err != nil {
		log.Fatalln("Could not get location.", err)
	}

	l := Listing{
		sanitizeInput(title),
		sanitizeInput(location),
		sanitizeInput(askingPrice),
		sanitizeInput(cashFlow),
		sanitizeInput(description),
	}

	fmt.Println(l)

	defer c.Release()

	// //poolLog := chromedp.PoolLog(log.Printf, log.Printf, log.Printf)
	// pool, err := chromedp.NewPool() //(poolLog)
	// defer pool.Shutdown()

	// var urlToVisit strings.Builder
	// var listingUrls []string

	// for i := 1; i < depth; i++ {
	// 	urlToVisit.Reset()
	// 	urlToVisit.WriteString(startURL)

	// 	if i > 1 {
	// 		urlToVisit.WriteString(fmt.Sprintf("/%v", i))
	// 	}

	// 	url := urlToVisit.String()

	// 	var listingNodes []*cdp.Node
	// 	err = c.Run(ctxt, getListingUrls(url, &listingNodes))
	// 	if err != nil {
	// 		log.Fatalln("Error running getListingUrls.", err)
	// 	}

	// 	for n := 0; n < len(listingNodes); n++ {
	// 		href := listingNodes[n].AttributeValue("href")
	// 		urlToVisit.Reset()
	// 		urlToVisit.WriteString(baseURL)
	// 		urlToVisit.WriteString(href)
	// 		formattedURL := strings.Split(urlToVisit.String(), "?")[0]
	// 		fmt.Println(formattedURL)
	// 		listingUrls = append(listingUrls, formattedURL)
	// 	}
	// }
	// fmt.Println(listingUrls)

	// chrome.Release()

	// var wg sync.WaitGroup
	// var listings []Listing

	// for _, urlStr := range listingUrls {
	// 	wg.Add(1)
	// 	go getListingDetails(ctxt, pool, &wg, urlStr, &listings)
	// }

	// wg.Wait()

	// err = pool.Shutdown()
	// fmt.Println("shutting down")
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

func sanitizeInput(input string) string {
	pattern := regexp.MustCompile(`[\n\s•]+`)
	input = pattern.ReplaceAllString(input, " ")
	input = strings.ReplaceAll(input, "’", "'")
	input = strings.TrimSpace(input)
	return input
}

func getListingDetails(ctxt context.Context, pool *chromedp.Pool, wg *sync.WaitGroup, urlStr string, listings *[]Listing) {
	defer wg.Done()

	chrome, err := pool.Allocate(ctxt,
		runner.Flag("headless", true),
		runner.Flag("disable-gpu", true),
		runner.Flag("no-first-run", true),
		runner.Flag("no-sandbox", true),
		runner.Flag("no-default-browser-check", true),
		//3runner.Flag("remote-debugging-port", 9222),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer chrome.Release()

	var description string
	err = chrome.Run(ctxt, getDescription(urlStr, &description))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(urlStr, "hi")
	fmt.Println(description, "hello")
}

func getDescription(url string, description *string) chromedp.Tasks {
	fmt.Println("running")
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Text(".businessDescription", description, chromedp.BySearch),
	}
}

func getListingUrls(url string, listingNodes *[]*cdp.Node) chromedp.Tasks {
	fmt.Println("heyoo", url)
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitReady(".pagination", chromedp.BySearch),
		chromedp.Nodes(".listingResult", listingNodes, chromedp.BySearch),
	}
}

func getListingInfo(url string, details *Listing, bodyNodes *[]*cdp.Node) chromedp.Action {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitReady("#footer", chromedp.BySearch),
		chromedp.Nodes(".pageContent", bodyNodes, chromedp.BySearch),
		chromedp.ActionFunc(func(ctxt context.Context, h cdp.Executor) error {
			fmt.Println(bodyNodes)
			return nil
		}),
	}
}

func loadListing(url string) chromedp.Action {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitReady("#footer"),
	}
}

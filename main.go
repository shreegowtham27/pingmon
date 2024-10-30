package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// XML structure for sitemap
type SitemapIndex struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []URL    `xml:"url"`
}

type URL struct {
	Loc string `xml:"loc"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	godotenv.Load("somerandomfile")
	// URL of the sitemap.xml
	sitemapURL := os.Getenv("URL")

	// Fetching the sitemap.xml file
	resp, err := http.Get(sitemapURL)
	if err != nil {
		fmt.Println("Error fetching sitemap:", err)
		return
	}
	defer resp.Body.Close()

	// Parsing XML
	var sitemap SitemapIndex
	err = xml.NewDecoder(resp.Body).Decode(&sitemap)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
		return
	}

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Channel to receive results from goroutines
	resultCh := make(chan string)

	// Number of concurrent goroutines
	numWorkers := 1

	// Start goroutines to fetch URLs concurrently
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range sitemap.URLs {
				start := time.Now()
				resp, err := http.Head(sitemap.URLs[url].Loc)
				if err != nil {
					fmt.Printf("Error checking URL %s: %s\n", sitemap.URLs[url].Loc, err)
					continue
				}
				duration := time.Since(start)
				resultCh <- fmt.Sprintf("URL: %s | Status Code: %d | Response Time: %s", sitemap.URLs[url].Loc, resp.StatusCode, duration)
			}
		}()
	}

	// Close result channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Print results from result channel
	for result := range resultCh {
		fmt.Println(result)
	}
}

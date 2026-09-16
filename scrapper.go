package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
)

// pollInterval controls how long to wait between stock checks once an item
// is confirmed sold out. Kept fairly conservative (3 minutes) to avoid
// hammering the target site with requests.
const pollInterval = 3 * time.Minute

// CheckSoldOutStatus takes an Item and continuously checks whether the
// product is back in stock, sending an email notification the moment it is.
func CheckSoldOutStatus(item Item, wg *sync.WaitGroup) {
	defer wg.Done() // Decrement the counter when the goroutine completes.

	// Start Playwright.
	pw, err := playwright.Run()
	if err != nil {
		log.Printf("Could not start playwright for URL %s: %v", item.URL, err)
		return
	}
	defer pw.Stop()

	// Launch the browser.
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		log.Printf("Could not launch browser for URL %s: %v", item.URL, err)
		return
	}
	defer browser.Close()

	// Create a new page and navigate to the provided URL.
	page, err := browser.NewPage()
	if err != nil {
		log.Printf("Could not create new page for URL %s: %v", item.URL, err)
		return
	}

	// Go to the product page.
	_, err = page.Goto(item.URL)
	if err != nil {
		log.Printf("Could not navigate to URL %s: %v", item.URL, err)
		return
	}

	// Select the element whose text tells us the current stock status.
	stockStatusLocator := page.Locator(item.Locator)

	// Keep checking until the item is back in stock.
	for {
		// Wait for the element to become visible.
		err = stockStatusLocator.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		})
		if err != nil {
			log.Printf("Error waiting for element for URL %s: %v", item.URL, err)
			return
		}

		// Extract the text content of the status element.
		status, err := stockStatusLocator.TextContent()
		if err != nil {
			log.Printf("Could not extract text content for URL %s: %v", item.URL, err)
			return
		}

		// Check if the product is sold out.
		if status == "Sold out" {
			log.Printf("The item at %s is still sold out. Checking again in %s...", item.URL, pollInterval)
			time.Sleep(pollInterval)
		} else {
			log.Printf("The item at %s is now available!", item.URL)
			body := fmt.Sprintf("The item %s is now in stock! Check it here: %s", item.Name, item.URL)
			subject := fmt.Sprintf("Item in Stock! %s", item.Name)
			if err := sendEmail(subject, body); err != nil {
				log.Printf("Failed to send notification email for %s: %v", item.Name, err)
			}
			break // Exit the loop when the item is available.
		}
	}
}

// CheckMultipleURLs checks a list of items concurrently, one goroutine per item.
func CheckMultipleURLs(items []Item) {
	var wg sync.WaitGroup

	for _, item := range items {
		wg.Add(1)
		go CheckSoldOutStatus(item, &wg)
	}

	wg.Wait()
}

// sendEmail sends a notification email via Gmail's SMTP server.
//
// Requires the following environment variables to be set:
//   - NOTIFY_EMAIL_FROM: the Gmail address to send from
//   - NOTIFY_EMAIL_TO: the address to send the notification to
//   - GMAIL_APP_PASSWORD: a Gmail App Password for NOTIFY_EMAIL_FROM
//     (https://myaccount.google.com/apppasswords) — never your real
//     account password, and never checked into source control.
func sendEmail(subject, body string) error {
	from := os.Getenv("NOTIFY_EMAIL_FROM")
	to := os.Getenv("NOTIFY_EMAIL_TO")
	password := os.Getenv("GMAIL_APP_PASSWORD")

	if from == "" || to == "" || password == "" {
		return fmt.Errorf("missing one or more required environment variables: NOTIFY_EMAIL_FROM, NOTIFY_EMAIL_TO, GMAIL_APP_PASSWORD")
	}

	const smtpHost = "smtp.gmail.com"
	const smtpPort = "587"

	message := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)
	auth := smtp.PlainAuth("", from, password, smtpHost)

	if err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(message)); err != nil {
		return err
	}

	fmt.Println("Email sent successfully!")
	return nil
}

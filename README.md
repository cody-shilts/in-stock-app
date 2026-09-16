# in-stock-app

A small Go tool that watches product pages for restocks and emails you the
moment an item comes back in stock. Built out of frustration with manually
refreshing product pages for items that sell out fast.

## How it works

- Each tracked item is a URL plus a locator (XPath or CSS selector) pointing
  at the element on the page whose text indicates stock status
  (e.g. an "Add to cart" button's label).
- Every item is checked concurrently in its own goroutine, using
  [Playwright](https://github.com/playwright-community/playwright-go) to
  drive a real (headless) browser, so it works on pages that render stock
  status client-side.
- If an item is sold out, the checker waits and tries again on an interval.
  As soon as the status text changes, it sends an email notification and
  stops checking that item.

## Setup

1. Install dependencies:

   ```bash
   go mod download
   ```

2. Install the Playwright browser binaries (one-time):

   ```bash
   go run github.com/playwright-community/playwright-go/cmd/playwright install
   ```

3. Set the required environment variables:

   | Variable              | Description                                                                 |
   |-----------------------|-------------------------------------------------------------------------------|
   | `NOTIFY_EMAIL_FROM`   | Gmail address to send notifications from                                    |
   | `NOTIFY_EMAIL_TO`     | Address to receive notifications                                            |
   | `GMAIL_APP_PASSWORD`  | A Gmail [App Password](https://myaccount.google.com/apppasswords) for the sending account — not your real account password |

4. Edit the `items` slice in `main.go` with the product pages and locators
   you want to track.

5. Run it:

   ```bash
   NOTIFY_EMAIL_FROM=you@gmail.com NOTIFY_EMAIL_TO=you@gmail.com GMAIL_APP_PASSWORD=xxxx go run .
   ```

## Notes

- The poll interval (3 minutes) is intentionally conservative to avoid
  hammering the target site. Adjust `pollInterval` in `scrapper.go` if needed.
- This is a personal-use tool for tracking a small handful of pages you
  care about — it's not built for high-frequency or large-scale scraping,
  and you should always check a site's terms of service before scraping it.

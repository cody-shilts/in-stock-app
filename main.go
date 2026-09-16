package main

// Item describes a product page to monitor and the locator used to read its
// stock-status text on the page.
type Item struct {
	Name    string
	URL     string
	Locator string
}

func main() {
	// Example items — replace with the product pages you want to track.
	// Locator is an XPath (or CSS selector, depending on how you call
	// page.Locator) pointing at the element whose text tells you whether
	// the item is sold out (e.g. an "Add to cart" button's label).
	items := []Item{
		{"Example Product A", "https://example.com/products/example-a", `//*[@id="AddToCart"]/span[1]`},
		{"Example Product B", "https://example.com/products/example-b", `//*[@id="AddToCart"]/span[1]`},
	}

	// Check all URLs concurrently.
	CheckMultipleURLs(items)
	select {} // Block forever to keep the program running.
}

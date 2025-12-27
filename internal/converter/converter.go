package converter

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

type StripConfig struct {
	ElementsToStrip   []string
	RemoveImagesNoAlt bool // If true, remove images without alt text entirely
}

// Default elements to strip - users can preserve with data-llm="keep"
var defaultStripElements = []string{"nav", "aside", "footer", "header", "script", "style", "noscript", "svg", "iframe"}

// ProcessScripts handles script tags with data-llm-description attribute.
// If a script has data-llm-description, it is replaced with a descriptive text node.
// Scripts with data-llm="keep" are preserved. Scripts without description are left
// for StripElements to remove.
func ProcessScripts(n *html.Node) {
	type scriptReplacement struct {
		node *html.Node
		desc string
	}

	var f func(*html.Node)
	f = func(parent *html.Node) {
		var toReplace []scriptReplacement

		for c := parent.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "script" {
				// Check for data-llm="keep" - if present, skip entirely
				shouldSkip := false
				var description string

				for _, attr := range c.Attr {
					if attr.Key == "data-llm" && attr.Val == "keep" {
						shouldSkip = true
						break
					}
					if attr.Key == "data-llm-description" {
						description = attr.Val
					}
				}

				if !shouldSkip && strings.TrimSpace(description) != "" {
					toReplace = append(toReplace, scriptReplacement{node: c, desc: description})
				}
				// Scripts without description are left for StripElements to remove
			} else {
				f(c) // Recurse into non-script elements
			}
		}

		// Replace scripts with text nodes
		for _, item := range toReplace {
			textNode := &html.Node{
				Type: html.TextNode,
				Data: "There's some javascript that has the following description: " + item.desc,
			}
			parent.InsertBefore(textNode, item.node)
			parent.RemoveChild(item.node)
		}
	}
	f(n)
}

// ProcessImages replaces img tags with their alt text.
// Format: "[Image: alt text]" or "[Image]" if no alt.
// If removeIfNoAlt is true and no alt text exists, the image is removed entirely.
func ProcessImages(n *html.Node, removeIfNoAlt bool) {
	type imageReplacement struct {
		node   *html.Node
		alt    string
		remove bool
	}

	var f func(*html.Node)
	f = func(parent *html.Node) {
		var toProcess []imageReplacement

		for c := parent.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "img" {
				var altText string
				for _, attr := range c.Attr {
					if attr.Key == "alt" {
						altText = attr.Val
						break
					}
				}

				shouldRemove := altText == "" && removeIfNoAlt
				toProcess = append(toProcess, imageReplacement{
					node:   c,
					alt:    altText,
					remove: shouldRemove,
				})
			} else {
				f(c) // Recurse
			}
		}

		for _, item := range toProcess {
			if item.remove {
				parent.RemoveChild(item.node)
			} else {
				var text string
				if item.alt != "" {
					text = "[Image: " + item.alt + "]"
				} else {
					text = "[Image]"
				}
				textNode := &html.Node{
					Type: html.TextNode,
					Data: text,
				}
				parent.InsertBefore(textNode, item.node)
				parent.RemoveChild(item.node)
			}
		}
	}
	f(n)
}

// StripElements removes specified HTML elements from the DOM
func StripElements(n *html.Node, tags ...string) {
	tagSet := make(map[string]bool, len(tags))
	for _, tag := range tags {
		tagSet[tag] = true
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		// Collect nodes to remove (can't remove while iterating)
		var toRemove []*html.Node

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && tagSet[c.Data] {
				shouldKeep := false
				for _, attr := range c.Attr {
					if attr.Key == "data-llm" && attr.Val == "keep" {
						shouldKeep = true
						break
					}
				}
				if !shouldKeep {
					toRemove = append(toRemove, c)
				}
			} else {
				shouldKeep := true
				for _, attr := range c.Attr {
					if attr.Key == "data-llm" && attr.Val == "drop" {
						shouldKeep = false
						break
					}
				}
				if !shouldKeep {
					toRemove = append(toRemove, c)
				} else {
					f(c)
				}
			}
		}

		// Remove collected nodes
		for _, node := range toRemove {
			n.RemoveChild(node)
		}
	}
	f(n)
}

// ProcessHTML strips specified tags from HTML based on options
func ProcessHTML(htmlContent []byte, stripConfig StripConfig) ([]byte, error) {
	doc, err := html.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	// Process scripts with data-llm-description FIRST (before stripping)
	// This extracts descriptions before scripts are removed
	ProcessScripts(doc)

	// Process images (replace with alt text)
	ProcessImages(doc, stripConfig.RemoveImagesNoAlt)

	// Combine user-specified elements with defaults
	elementsToStrip := append(stripConfig.ElementsToStrip, defaultStripElements...)

	// Strip specified tags
	StripElements(doc, elementsToStrip...)

	// Serialize back to HTML
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

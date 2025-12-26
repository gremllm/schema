package converter

import (
	"bytes"

	"golang.org/x/net/html"
)

type StripConfig struct {
	StripNav    bool
	StripAside  bool
	StripScript bool
}

// StripElements removes specified HTML elements from the DOM
func StripElements(n *html.Node, tags ...string) {
	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		// Collect nodes to remove (can't remove while iterating)
		var toRemove []*html.Node

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && tagSet[c.Data] {
				toRemove = append(toRemove, c)
			} else {
				f(c) // Recurse on children
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

	// Build list of tags to strip (always include header and footer)
	tags := []string{"header", "footer"}
	if stripConfig.StripNav {
		tags = append(tags, "nav")
	}
	if stripConfig.StripAside {
		tags = append(tags, "aside")
	}
	if stripConfig.StripScript {
		tags = append(tags, "script", "style")
	}

	// Strip specified tags
	StripElements(doc, tags...)

	// Serialize back to HTML
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

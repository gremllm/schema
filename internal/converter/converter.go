package converter

import (
	"bytes"
	"regexp"
	"strings"
	"sync"

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
				Data: "Javascript description: " + item.desc,
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

// Noise patterns to remove (attributions, credits, etc.)
var noisePatterns = []string{
	"photo by", "credit:", "source:", "©", "copyright",
	"all rights reserved",
}

// Pre-compiled regex for whitespace normalization
var multipleNewlines = regexp.MustCompile(`\n{3,}`)

// CondenseMarkdown optimizes markdown for LLM consumption by:
// - Removing noise (attributions, decorative text)
// - Fixing fragmented numbered lists
// - Normalizing whitespace
func CondenseMarkdown(md string) string {
	lines := strings.Split(md, "\n")
	lines = removeNoiseLines(lines)
	lines = fixFragmentedLists(lines)

	md = strings.Join(lines, "\n")

	// Collapse 3+ newlines to 2
	md = multipleNewlines.ReplaceAllString(md, "\n\n")

	// Remove trailing whitespace from lines
	resultLines := strings.Split(md, "\n")
	for i, line := range resultLines {
		resultLines[i] = strings.TrimRight(line, " \t")
	}
	md = strings.Join(resultLines, "\n")

	return strings.TrimSpace(md)
}

// removeNoiseLines filters out attribution, copyright, and decorative noise
func removeNoiseLines(lines []string) []string {
	var result []string
	for _, line := range lines {
		lower := strings.ToLower(strings.TrimSpace(line))
		if lower == "" {
			result = append(result, line)
			continue
		}
		isNoise := false
		for _, pattern := range noisePatterns {
			if strings.Contains(lower, pattern) {
				isNoise = true
				break
			}
		}
		if !isNoise {
			result = append(result, line)
		}
	}
	return result
}

// isStandaloneNumber checks if a line is just a number (1-99)
func isStandaloneNumber(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) == 0 || len(s) > 2 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// fixFragmentedLists converts standalone numbers followed by content into proper list items
func fixFragmentedLists(lines []string) []string {
	var result []string
	i := 0

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Check for standalone number (fragmented list)
		if isStandaloneNumber(trimmed) {
			num := trimmed
			i++

			// Skip empty lines
			for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
				i++
			}

			// Collect content until next number or heading
			var parts []string
			for i < len(lines) {
				next := strings.TrimSpace(lines[i])
				if next == "" {
					// Peek ahead
					j := i + 1
					for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
						j++
					}
					if j < len(lines) {
						peek := strings.TrimSpace(lines[j])
						if isStandaloneNumber(peek) || strings.HasPrefix(peek, "#") {
							break
						}
					} else {
						break
					}
					i++
					continue
				}
				if isStandaloneNumber(next) || strings.HasPrefix(next, "#") {
					break
				}
				parts = append(parts, next)
				i++
			}

			if len(parts) > 0 {
				// Join parts: "num. part1 - part2 - part3..."
				result = append(result, num+". "+strings.Join(parts, " - "))
			} else {
				result = append(result, line)
			}
			continue
		}

		result = append(result, line)
		i++
	}

	return result
}

// Buffer pool to reduce allocations
var bufPool = sync.Pool{
	New: func() any {
		return new(strings.Builder)
	},
}

func getBuffer() *strings.Builder {
	buf := bufPool.Get().(*strings.Builder)
	buf.Reset()
	return buf
}

func putBuffer(buf *strings.Builder) {
	bufPool.Put(buf)
}

// HTMLToMarkdown converts HTML to markdown in a single pass.
// It processes, strips, and converts in one tree walk.
func HTMLToMarkdown(htmlContent []byte, stripConfig StripConfig) (string, error) {
	doc, err := html.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	// Build strip set
	stripSet := make(map[string]bool)
	for _, tag := range defaultStripElements {
		stripSet[tag] = true
	}
	for _, tag := range stripConfig.ElementsToStrip {
		stripSet[tag] = true
	}

	buf := getBuffer()
	defer putBuffer(buf)

	ctx := &mdContext{
		buf:             buf,
		stripSet:        stripSet,
		removeImgNoAlt:  stripConfig.RemoveImagesNoAlt,
		inPre:           false,
		listDepth:       0,
		orderedListNums: make([]int, 10),
	}

	ctx.walk(doc)

	result := buf.String()
	return CondenseMarkdown(result), nil
}

// Markdown element rendering rules
type mdRule struct {
	prefix string
	suffix string
}

var (
	// Simple wrap rules: prefix + children + suffix
	wrapRules = map[string]mdRule{
		// Headings
		"h1": {"\n# ", "\n\n"},
		"h2": {"\n## ", "\n\n"},
		"h3": {"\n### ", "\n\n"},
		"h4": {"\n#### ", "\n\n"},
		"h5": {"\n##### ", "\n\n"},
		"h6": {"\n###### ", "\n\n"},

		// Block elements
		"p":          {"", "\n\n"},
		"blockquote": {"\n> ", "\n\n"},
		"address":    {"\n> ", "\n\n"}, // Treat like blockquote

		// Inline formatting
		"strong": {" **", "** "},
		"b":      {" **", "** "},
		"em":     {" *", "* "},
		"i":      {" *", "* "},
		"u":      {" _", "_ "},      // Underline as underscore
		"s":      {" ~~", "~~ "},    // Strikethrough
		"del":    {" ~~", "~~ "},    // Deleted text
		"ins":    {" __", "__ "},    // Inserted text
		"mark":   {" ==", "== "},    // Highlighted (some md flavors)
		"small":  {" ", " "},        // Just pass through
		"sub":    {"~", "~"},        // Subscript (some md flavors)
		"sup":    {"^", "^"},        // Superscript (some md flavors)
		"q":      {` "`, `" `},      // Inline quote

		// Code/technical
		"kbd":  {" `", "` "},  // Keyboard input
		"samp": {" `", "` "},  // Sample output
		"var":  {" _", "_ "},  // Variable
		"dfn":  {" *", "* "},  // Definition term
		"abbr": {"", ""},      // Abbreviation - just text
		"cite": {" *", "* "},  // Citation

		// Table elements
		"table":   {"\n", "\n"},
		"tr":      {"|", "\n"},
		"th":      {" **", "** |"},
		"td":      {" ", " |"},
		"caption": {"\n*", "*\n"}, // Table caption as italic

		// Description lists
		"dl": {"\n", "\n"},
		"dt": {"\n**", "**\n"}, // Term as bold
		"dd": {": ", "\n"},     // Description with colon

		// Details/summary
		"details": {"\n", "\n"},
		"summary": {"\n**", "**\n"},

		// Time
		"time": {"", ""},

		// Ruby annotations (East Asian)
		"ruby": {"", ""},
		"rt":   {" (", ")"},
		"rp":   {"", ""},
	}

	// Tags that just pass through to children (structural/container tags)
	passThroughTags = map[string]bool{
		// Document structure
		"html": true, "head": true, "body": true,

		// Sectioning
		"div": true, "section": true, "article": true, "main": true,
		"header": true, "footer": true, "aside": true, "nav": true,
		"hgroup": true, "search": true,

		// Grouping
		"span": true, "figure": true, "figcaption": true,

		// Forms (pass through, not interactive in markdown)
		"form": true, "fieldset": true, "legend": true,
		"label": true, "input": true, "button": true,
		"select": true, "optgroup": true, "option": true,
		"textarea": true, "output": true, "datalist": true,
		"meter": true, "progress": true,

		// Table grouping
		"thead": true, "tbody": true, "tfoot": true,
		"colgroup": true, "col": true,

		// Media (pass through, handled specially if needed)
		"picture": true, "source": true, "track": true,
		"map": true, "area": true,

		// Metadata (usually stripped but safe to pass)
		"meta": true, "title": true, "link": true, "base": true,

		// Text direction
		"bdi": true, "bdo": true,

		// Template (hidden content)
		"template": true,

		// Data
		"data": true,

		// Dialog
		"dialog": true,

		// Deprecated but may appear
		"center": true, "font": true, "big": true, "tt": true,
		"strike": true, "acronym": true, "dir": true,
	}

	// Tags to skip entirely (no output, no children)
	skipTags = map[string]bool{
		"canvas": true, // Graphics - not text
		"embed":  true, // External content
		"object": true, // External content
		"param":  true, // Object parameters
		"wbr":    true, // Word break opportunity
	}
)

type mdContext struct {
	buf             *strings.Builder
	stripSet        map[string]bool
	removeImgNoAlt  bool
	inPre           bool
	listDepth       int
	orderedListNums []int
	inOrderedList   []bool
}

func (ctx *mdContext) walk(n *html.Node) {
	switch n.Type {
	case html.TextNode:
		ctx.renderText(n.Data)
	case html.ElementNode:
		ctx.renderElement(n)
	case html.DocumentNode:
		ctx.children(n)
	}
}

func (ctx *mdContext) renderText(text string) {
	if !ctx.inPre {
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}
		text = strings.ReplaceAll(text, "\n", " ")
	}
	ctx.buf.WriteString(text)
}

func (ctx *mdContext) renderElement(n *html.Node) {
	// Check data-llm attributes
	if hasAttr(n, "data-llm", "drop") {
		return
	}

	// Check if should strip (unless data-llm="keep")
	if ctx.stripSet[n.Data] && !hasAttr(n, "data-llm", "keep") {
		if n.Data == "script" {
			if desc := getAttr(n, "data-llm-description"); desc != "" {
				ctx.buf.WriteString("\nJavascript description: ")
				ctx.buf.WriteString(desc)
				ctx.buf.WriteString("\n")
			}
		}
		return
	}

	// Skip tags that produce no useful output
	if skipTags[n.Data] {
		return
	}

	// Check simple wrap rules first
	if rule, ok := wrapRules[n.Data]; ok {
		ctx.buf.WriteString(rule.prefix)
		ctx.children(n)
		ctx.buf.WriteString(rule.suffix)
		return
	}

	// Check pass-through tags
	if passThroughTags[n.Data] {
		ctx.children(n)
		return
	}

	// Handle special cases
	switch n.Data {
	case "br":
		ctx.buf.WriteString("\n")
	case "hr":
		ctx.buf.WriteString("\n---\n\n")
	case "code":
		ctx.renderCode(n)
	case "pre":
		ctx.renderPre(n)
	case "a":
		ctx.renderLink(n)
	case "img":
		ctx.renderImage(n)
	case "ul", "menu": // menu is also unordered list
		ctx.renderList(n, false)
	case "ol":
		ctx.renderList(n, true)
	case "li":
		ctx.renderListItem(n)
	case "audio", "video":
		ctx.renderMedia(n)
	default:
		ctx.children(n)
	}
}

func (ctx *mdContext) renderCode(n *html.Node) {
	if ctx.inPre {
		ctx.children(n)
	} else {
		ctx.buf.WriteString("`")
		ctx.children(n)
		ctx.buf.WriteString("`")
	}
}

func (ctx *mdContext) renderPre(n *html.Node) {
	ctx.buf.WriteString("\n```\n")
	ctx.inPre = true
	ctx.children(n)
	ctx.inPre = false
	ctx.buf.WriteString("\n```\n\n")
}

func (ctx *mdContext) renderLink(n *html.Node) {
	ctx.buf.WriteString("[")
	ctx.children(n)
	ctx.buf.WriteString("](")
	ctx.buf.WriteString(getAttr(n, "href"))
	ctx.buf.WriteString(")")
}

func (ctx *mdContext) renderImage(n *html.Node) {
	alt := getAttr(n, "alt")
	if alt == "" && ctx.removeImgNoAlt {
		return
	}
	if alt != "" {
		ctx.buf.WriteString("[Image: ")
		ctx.buf.WriteString(alt)
		ctx.buf.WriteString("]")
	} else {
		ctx.buf.WriteString("[Image]")
	}
}

func (ctx *mdContext) renderMedia(n *html.Node) {
	// Output media type and any text content (like fallback text)
	mediaType := "Media"
	if n.Data == "audio" {
		mediaType = "Audio"
	} else if n.Data == "video" {
		mediaType = "Video"
	}

	src := getAttr(n, "src")
	if src != "" {
		ctx.buf.WriteString("[")
		ctx.buf.WriteString(mediaType)
		ctx.buf.WriteString(": ")
		ctx.buf.WriteString(src)
		ctx.buf.WriteString("]")
	} else {
		ctx.buf.WriteString("[")
		ctx.buf.WriteString(mediaType)
		ctx.buf.WriteString("]")
	}
	// Also render children (fallback content, source elements)
	ctx.children(n)
}

func (ctx *mdContext) renderList(n *html.Node, ordered bool) {
	ctx.buf.WriteString("\n")
	ctx.listDepth++
	ctx.inOrderedList = append(ctx.inOrderedList, ordered)
	if ordered && ctx.listDepth <= len(ctx.orderedListNums) {
		ctx.orderedListNums[ctx.listDepth-1] = 0
	}
	ctx.children(n)
	ctx.inOrderedList = ctx.inOrderedList[:len(ctx.inOrderedList)-1]
	ctx.listDepth--
	ctx.buf.WriteString("\n")
}

func (ctx *mdContext) renderListItem(n *html.Node) {
	ctx.buf.WriteString(strings.Repeat("  ", ctx.listDepth-1))
	if len(ctx.inOrderedList) > 0 && ctx.inOrderedList[len(ctx.inOrderedList)-1] {
		ctx.orderedListNums[ctx.listDepth-1]++
		ctx.buf.WriteString(itoa(ctx.orderedListNums[ctx.listDepth-1]))
		ctx.buf.WriteString(". ")
	} else {
		ctx.buf.WriteString("- ")
	}
	ctx.children(n)
	ctx.buf.WriteString("\n")
}

func (ctx *mdContext) children(n *html.Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		ctx.walk(c)
	}
}

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func hasAttr(n *html.Node, key, val string) bool {
	for _, attr := range n.Attr {
		if attr.Key == key && attr.Val == val {
			return true
		}
	}
	return false
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	if n < 100 {
		return string([]rune{rune('0' + n/10), rune('0' + n%10)})
	}
	var b strings.Builder
	b.Grow(3)
	for n > 0 {
		b.WriteByte(byte('0' + n%10))
		n /= 10
	}
	s := b.String()
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

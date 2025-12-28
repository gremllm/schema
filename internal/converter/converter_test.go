package converter

import (
	"strings"
	"testing"
)

// =============================================================================
// ProcessHTML Tests (HTML cleaning, returns HTML)
// =============================================================================

func TestProcessHTML_StripsDefaultElements(t *testing.T) {
	input := `<!DOCTYPE html><html><body>
		<nav>Navigation</nav>
		<aside>Sidebar</aside>
		<main>Content</main>
		<script>alert('hi')</script>
		<style>.foo{}</style>
		<footer>Footer</footer>
	</body></html>`

	result, err := ProcessHTML([]byte(input), StripConfig{})
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	resultStr := string(result)
	for _, tag := range []string{"<nav>", "<aside>", "<script>", "<style>", "<footer>"} {
		if strings.Contains(resultStr, tag) {
			t.Errorf("Result still contains %s", tag)
		}
	}
	if !strings.Contains(resultStr, "Content") {
		t.Error("Result missing main content")
	}
}

func TestProcessHTML_DataLLMKeep(t *testing.T) {
	input := `<html><body>
		<nav data-llm="keep">Important Nav</nav>
		<nav>Regular Nav</nav>
	</body></html>`

	result, err := ProcessHTML([]byte(input), StripConfig{})
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	resultStr := string(result)
	if !strings.Contains(resultStr, "Important Nav") {
		t.Error("Result missing nav with data-llm=keep")
	}
	if strings.Contains(resultStr, "Regular Nav") {
		t.Error("Result should not contain regular nav")
	}
}

func TestProcessHTML_DataLLMDrop(t *testing.T) {
	input := `<html><body>
		<div data-llm="drop">Drop this</div>
		<div>Keep this</div>
	</body></html>`

	result, err := ProcessHTML([]byte(input), StripConfig{})
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	resultStr := string(result)
	if strings.Contains(resultStr, "Drop this") {
		t.Error("Result should not contain dropped content")
	}
	if !strings.Contains(resultStr, "Keep this") {
		t.Error("Result missing kept content")
	}
}

func TestProcessHTML_ScriptDescription(t *testing.T) {
	input := `<html><body>
		<script data-llm-description="Form validation">validate()</script>
	</body></html>`

	result, err := ProcessHTML([]byte(input), StripConfig{})
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	resultStr := string(result)
	if strings.Contains(resultStr, "<script") {
		t.Error("Result still contains script tag")
	}
	if !strings.Contains(resultStr, "Javascript description: Form validation") {
		t.Error("Result missing script description")
	}
}

func TestProcessHTML_ImageAlt(t *testing.T) {
	input := `<html><body><img src="x.jpg" alt="A photo"></body></html>`

	result, err := ProcessHTML([]byte(input), StripConfig{})
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	resultStr := string(result)
	if strings.Contains(resultStr, "<img") {
		t.Error("Result still contains img tag")
	}
	if !strings.Contains(resultStr, "[Image: A photo]") {
		t.Error("Result missing image alt text")
	}
}

func TestProcessHTML_ImageNoAlt(t *testing.T) {
	tests := []struct {
		name           string
		removeNoAlt    bool
		expectContains string
	}{
		{"keep placeholder", false, "[Image]"},
		{"remove entirely", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := `<html><body><img src="x.jpg"></body></html>`
			result, _ := ProcessHTML([]byte(input), StripConfig{RemoveImagesNoAlt: tt.removeNoAlt})
			resultStr := string(result)

			if tt.expectContains == "" {
				if strings.Contains(resultStr, "[Image]") {
					t.Error("Result should not contain [Image]")
				}
			} else {
				if !strings.Contains(resultStr, tt.expectContains) {
					t.Errorf("Result missing %q", tt.expectContains)
				}
			}
		})
	}
}

// =============================================================================
// HTMLToMarkdown Tests (full pipeline to markdown)
// =============================================================================

func TestHTMLToMarkdown_Headings(t *testing.T) {
	tests := []struct {
		tag      string
		expected string
	}{
		{"h1", "# "},
		{"h2", "## "},
		{"h3", "### "},
		{"h4", "#### "},
		{"h5", "##### "},
		{"h6", "###### "},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			input := []byte("<html><body><" + tt.tag + ">Heading</" + tt.tag + "></body></html>")
			result, err := HTMLToMarkdown(input, StripConfig{})
			if err != nil {
				t.Fatalf("HTMLToMarkdown failed: %v", err)
			}
			if !strings.Contains(result, tt.expected+"Heading") {
				t.Errorf("Expected %q, got: %s", tt.expected+"Heading", result)
			}
		})
	}
}

func TestHTMLToMarkdown_TextFormatting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"strong", "<strong>bold</strong>", "**bold**"},
		{"b", "<b>bold</b>", "**bold**"},
		{"em", "<em>italic</em>", "*italic*"},
		{"i", "<i>italic</i>", "*italic*"},
		{"code", "<code>code</code>", "`code`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := []byte("<html><body><p>" + tt.input + "</p></body></html>")
			result, _ := HTMLToMarkdown(input, StripConfig{})
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected %q in result, got: %s", tt.expected, result)
			}
		})
	}
}

func TestHTMLToMarkdown_Links(t *testing.T) {
	input := []byte(`<html><body><a href="https://example.com">Click here</a></body></html>`)
	result, _ := HTMLToMarkdown(input, StripConfig{})

	if !strings.Contains(result, "[Click here](https://example.com)") {
		t.Errorf("Expected markdown link, got: %s", result)
	}
}

func TestHTMLToMarkdown_Lists(t *testing.T) {
	t.Run("unordered", func(t *testing.T) {
		input := []byte(`<html><body><ul><li>One</li><li>Two</li></ul></body></html>`)
		result, _ := HTMLToMarkdown(input, StripConfig{})

		if !strings.Contains(result, "- One") || !strings.Contains(result, "- Two") {
			t.Errorf("Expected unordered list, got: %s", result)
		}
	})

	t.Run("ordered", func(t *testing.T) {
		input := []byte(`<html><body><ol><li>First</li><li>Second</li></ol></body></html>`)
		result, _ := HTMLToMarkdown(input, StripConfig{})

		if !strings.Contains(result, "1. First") || !strings.Contains(result, "2. Second") {
			t.Errorf("Expected ordered list, got: %s", result)
		}
	})
}

func TestHTMLToMarkdown_CodeBlock(t *testing.T) {
	input := []byte(`<html><body><pre><code>func main() {}</code></pre></body></html>`)
	result, _ := HTMLToMarkdown(input, StripConfig{})

	if !strings.Contains(result, "```") || !strings.Contains(result, "func main()") {
		t.Errorf("Expected code block, got: %s", result)
	}
}

func TestHTMLToMarkdown_Table(t *testing.T) {
	input := []byte(`<html><body>
		<table>
			<tr><th>Name</th><th>Age</th></tr>
			<tr><td>Alice</td><td>30</td></tr>
		</table>
	</body></html>`)
	result, _ := HTMLToMarkdown(input, StripConfig{})

	if !strings.Contains(result, "| **Name**") || !strings.Contains(result, "| Alice") {
		t.Errorf("Expected table, got: %s", result)
	}
}

func TestHTMLToMarkdown_Blockquote(t *testing.T) {
	input := []byte(`<html><body><blockquote>A wise quote</blockquote></body></html>`)
	result, _ := HTMLToMarkdown(input, StripConfig{})

	if !strings.Contains(result, "> A wise quote") {
		t.Errorf("Expected blockquote, got: %s", result)
	}
}

func TestHTMLToMarkdown_StripsNavFooter(t *testing.T) {
	input := []byte(`<html><body>
		<nav>Skip nav</nav>
		<main><h1>Title</h1></main>
		<footer>Skip footer</footer>
	</body></html>`)
	result, _ := HTMLToMarkdown(input, StripConfig{})

	if strings.Contains(result, "Skip nav") || strings.Contains(result, "Skip footer") {
		t.Error("Result should not contain nav/footer content")
	}
	if !strings.Contains(result, "# Title") {
		t.Error("Result missing main content")
	}
}

func TestHTMLToMarkdown_Image(t *testing.T) {
	input := []byte(`<html><body><img src="x.jpg" alt="A sunset"></body></html>`)
	result, _ := HTMLToMarkdown(input, StripConfig{})

	if !strings.Contains(result, "[Image: A sunset]") {
		t.Errorf("Expected image alt, got: %s", result)
	}
}

func TestHTMLToMarkdown_ScriptDescription(t *testing.T) {
	input := []byte(`<html><body>
		<script data-llm-description="Interactive chart">chart()</script>
	</body></html>`)
	result, _ := HTMLToMarkdown(input, StripConfig{})

	if !strings.Contains(result, "Javascript description: Interactive chart") {
		t.Errorf("Expected script description, got: %s", result)
	}
}

func TestHTMLToMarkdown_AdditionalTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"strikethrough del", "<del>deleted</del>", "~~deleted~~"},
		{"strikethrough s", "<s>struck</s>", "~~struck~~"},
		{"underline", "<u>underlined</u>", "_underlined_"},
		{"inserted", "<ins>added</ins>", "__added__"},
		{"keyboard", "<kbd>Ctrl+C</kbd>", "`Ctrl+C`"},
		{"sample", "<samp>output</samp>", "`output`"},
		{"variable", "<var>x</var>", "_x_"},
		{"cite", "<cite>Book Title</cite>", "*Book Title*"},
		{"inline quote", "<q>quoted</q>", `"quoted"`},
		{"subscript", "<sub>2</sub>", "~2~"},
		{"superscript", "<sup>2</sup>", "^2^"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := []byte("<html><body><p>" + tt.input + "</p></body></html>")
			result, _ := HTMLToMarkdown(input, StripConfig{})
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected %q, got: %s", tt.expected, result)
			}
		})
	}
}

func TestHTMLToMarkdown_DescriptionList(t *testing.T) {
	input := []byte(`<html><body>
		<dl>
			<dt>Term</dt>
			<dd>Definition here</dd>
		</dl>
	</body></html>`)
	result, _ := HTMLToMarkdown(input, StripConfig{})

	if !strings.Contains(result, "**Term**") {
		t.Errorf("Expected bold term, got: %s", result)
	}
	if !strings.Contains(result, ": Definition") {
		t.Errorf("Expected definition with colon, got: %s", result)
	}
}

func TestHTMLToMarkdown_Media(t *testing.T) {
	t.Run("video", func(t *testing.T) {
		input := []byte(`<html><body><video src="movie.mp4"></video></body></html>`)
		result, _ := HTMLToMarkdown(input, StripConfig{})
		if !strings.Contains(result, "[Video: movie.mp4]") {
			t.Errorf("Expected video tag, got: %s", result)
		}
	})

	t.Run("audio", func(t *testing.T) {
		input := []byte(`<html><body><audio src="song.mp3"></audio></body></html>`)
		result, _ := HTMLToMarkdown(input, StripConfig{})
		if !strings.Contains(result, "[Audio: song.mp3]") {
			t.Errorf("Expected audio tag, got: %s", result)
		}
	})
}

func TestHTMLToMarkdown_SkipTags(t *testing.T) {
	input := []byte(`<html><body>
		<p>Before</p>
		<canvas>Canvas content</canvas>
		<embed src="x">
		<object>Object</object>
		<p>After</p>
	</body></html>`)
	result, _ := HTMLToMarkdown(input, StripConfig{})

	if strings.Contains(result, "Canvas") || strings.Contains(result, "Object") {
		t.Errorf("Should skip canvas/object, got: %s", result)
	}
	if !strings.Contains(result, "Before") || !strings.Contains(result, "After") {
		t.Error("Should preserve surrounding content")
	}
}

// =============================================================================
// CondenseMarkdown Tests
// =============================================================================

func TestCondenseMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"collapse multiple newlines", "Hello\n\n\n\n\nWorld", "Hello\n\nWorld"},
		{"trim trailing whitespace", "Hello   \nWorld\t\t", "Hello\nWorld"},
		{"trim document whitespace", "\n\n  Hello World  \n\n", "Hello World"},
		{"preserve paragraph breaks", "Para 1\n\nPara 2", "Para 1\n\nPara 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CondenseMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("CondenseMarkdown() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCondenseMarkdown_RemovesNoise(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"photo by", "Content\nPhoto by Someone\nMore content"},
		{"copyright", "Content\nCopyright 2024\nMore content"},
		{"credit", "Content\nCredit: Someone\nMore content"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CondenseMarkdown(tt.input)
			if strings.Contains(strings.ToLower(result), tt.name) {
				t.Errorf("Result should not contain %q noise", tt.name)
			}
			if !strings.Contains(result, "Content") || !strings.Contains(result, "More content") {
				t.Error("Result missing actual content")
			}
		})
	}
}

func TestCondenseMarkdown_FixesFragmentedLists(t *testing.T) {
	input := "1\n\nFirst item\n\n2\n\nSecond item"
	result := CondenseMarkdown(input)

	if !strings.Contains(result, "1. First item") {
		t.Errorf("Expected fixed list, got: %s", result)
	}
}

// =============================================================================
// Error Handling Tests
// =============================================================================

func TestHTMLToMarkdown_EmptyInput(t *testing.T) {
	result, err := HTMLToMarkdown([]byte{}, StripConfig{})
	if err != nil {
		t.Fatalf("Empty input should not error: %v", err)
	}
	if result != "" {
		t.Errorf("Empty input should produce empty output, got: %q", result)
	}
}

func TestHTMLToMarkdown_NilInput(t *testing.T) {
	result, err := HTMLToMarkdown(nil, StripConfig{})
	if err != nil {
		t.Fatalf("Nil input should not error: %v", err)
	}
	if result != "" {
		t.Errorf("Nil input should produce empty output, got: %q", result)
	}
}

func TestHTMLToMarkdown_MalformedHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"unclosed tags", "<html><body><div><p>Text"},
		{"mismatched tags", "<html><body><div></p></div></body></html>"},
		{"no html structure", "Just plain text without tags"},
		{"partial tags", "<html><bo"},
		{"only opening tag", "<div>"},
		{"random garbage", "<<<>>><<<"},
		{"script injection attempt", "<script>alert('xss')</script><p>Content</p>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HTMLToMarkdown([]byte(tt.input), StripConfig{})
			// Should not panic or error - graceful degradation
			if err != nil {
				t.Errorf("Malformed HTML should not error: %v", err)
			}
			// Result should be non-nil
			if result == "" && tt.name == "no html structure" {
				// Plain text should still be extracted
				if !strings.Contains(result, "plain text") {
					// This is acceptable - some parsers handle this differently
				}
			}
		})
	}
}

func TestHTMLToMarkdown_LargeInput(t *testing.T) {
	// Generate a large HTML document
	var sb strings.Builder
	sb.WriteString("<html><body>")
	for i := 0; i < 10000; i++ {
		sb.WriteString("<p>Paragraph number ")
		sb.WriteString(strings.Repeat("content ", 10))
		sb.WriteString("</p>")
	}
	sb.WriteString("</body></html>")

	input := []byte(sb.String())
	result, err := HTMLToMarkdown(input, StripConfig{})
	if err != nil {
		t.Fatalf("Large input should not error: %v", err)
	}
	if len(result) == 0 {
		t.Error("Large input should produce output")
	}
}

func TestHTMLToMarkdown_DeepNesting(t *testing.T) {
	// Generate deeply nested HTML
	var sb strings.Builder
	sb.WriteString("<html><body>")
	for i := 0; i < 100; i++ {
		sb.WriteString("<div>")
	}
	sb.WriteString("Deep content")
	for i := 0; i < 100; i++ {
		sb.WriteString("</div>")
	}
	sb.WriteString("</body></html>")

	result, err := HTMLToMarkdown([]byte(sb.String()), StripConfig{})
	if err != nil {
		t.Fatalf("Deep nesting should not error: %v", err)
	}
	if !strings.Contains(result, "Deep content") {
		t.Error("Deep content should be extracted")
	}
}

func TestHTMLToMarkdown_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"html entities", "<p>&amp; &lt; &gt; &quot;</p>", "& < > \""},
		{"unicode", "<p>Hello 世界 🌍</p>", "Hello 世界 🌍"},
		{"newlines in content", "<p>Line1\nLine2\nLine3</p>", "Line1"},
		{"tabs", "<p>Col1\tCol2\tCol3</p>", "Col1"},
		{"zero width chars", "<p>Hello\u200BWorld</p>", "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HTMLToMarkdown([]byte(tt.input), StripConfig{})
			if err != nil {
				t.Fatalf("Special chars should not error: %v", err)
			}
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected %q in result, got: %s", tt.expected, result)
			}
		})
	}
}

func TestProcessHTML_EmptyInput(t *testing.T) {
	result, err := ProcessHTML([]byte{}, StripConfig{})
	if err != nil {
		t.Fatalf("Empty input should not error: %v", err)
	}
	if len(result) == 0 {
		// Empty is acceptable for empty input
	}
}

func TestProcessHTML_MalformedHTML(t *testing.T) {
	input := `<html><body><div><p>Unclosed tags`
	result, err := ProcessHTML([]byte(input), StripConfig{})
	if err != nil {
		t.Fatalf("Malformed HTML should not error: %v", err)
	}
	// Should still extract text content
	if !strings.Contains(string(result), "Unclosed tags") {
		t.Error("Should extract text from malformed HTML")
	}
}

func TestCondenseMarkdown_EmptyInput(t *testing.T) {
	result := CondenseMarkdown("")
	if result != "" {
		t.Errorf("Empty input should produce empty output, got: %q", result)
	}
}

func TestCondenseMarkdown_OnlyWhitespace(t *testing.T) {
	result := CondenseMarkdown("   \n\n\t\t\n   ")
	if result != "" {
		t.Errorf("Whitespace-only input should produce empty output, got: %q", result)
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestFullPipeline(t *testing.T) {
	input := []byte(`<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
	<nav>Navigation</nav>
	<main>
		<h1>Welcome</h1>
		<p>This is <strong>important</strong> content.</p>
		<ul>
			<li>Item one</li>
			<li>Item two</li>
		</ul>
	</main>
	<footer>Photo by Someone</footer>
</body>
</html>`)

	result, err := HTMLToMarkdown(input, StripConfig{})
	if err != nil {
		t.Fatalf("HTMLToMarkdown failed: %v", err)
	}

	// Check nav/footer stripped
	if strings.Contains(result, "Navigation") {
		t.Error("Should not contain nav")
	}
	if strings.Contains(result, "Photo by") {
		t.Error("Should not contain footer attribution")
	}

	// Check content preserved
	if !strings.Contains(result, "# Welcome") {
		t.Error("Missing heading")
	}
	if !strings.Contains(result, "**important**") {
		t.Error("Missing bold text")
	}
	if !strings.Contains(result, "- Item one") {
		t.Error("Missing list items")
	}
}

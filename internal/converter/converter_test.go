package converter

import (
	"strings"
	"testing"
)

func TestProcessHTML(t *testing.T) {
	input := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
	<header><h1>Header Content</h1></header>
	<main><p>Main Content</p></main>
	<footer><p>Footer Content</p></footer>
</body>
</html>`

	result, err := ProcessHTML([]byte(input), StripConfig{ElementsToStrip: []string{}})
	if err != nil {
		t.Fatalf("processHTML failed: %v", err)
	}

	resultStr := string(result)

	// Check that header and footer are removed
	if strings.Contains(resultStr, "<header>") {
		t.Error("Result still contains <header> tag")
	}
	if strings.Contains(resultStr, "<footer>") {
		t.Error("Result still contains <footer> tag")
	}

	// Check that main content is preserved
	if !strings.Contains(resultStr, "<main>") {
		t.Error("Result missing <main> tag")
	}
	if !strings.Contains(resultStr, "Main Content") {
		t.Error("Result missing main content")
	}
}

func TestProcessHTMLWithElementsToStrip(t *testing.T) {
	input := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
	<div><p>Content to keep</p></div>
	<span><p>Content to strip</p></span>
</body>
</html>`

	result, err := ProcessHTML([]byte(input), StripConfig{ElementsToStrip: []string{"span"}})
	if err != nil {
		t.Fatalf("processHTML failed: %v", err)
	}

	resultStr := string(result)

	// Check that span is removed
	if strings.Contains(resultStr, "<span>") {
		t.Error("Result still contains <span> tag")
	}
}

func TestProcessHTMLWithDataLLMKeep(t *testing.T) {
	input := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
	<header data-llm="keep"><h1>Header Content</h1></header>
	<footer><p>Footer Content</p></footer>
</body>
</html>`

	result, err := ProcessHTML([]byte(input), StripConfig{ElementsToStrip: []string{}})
	if err != nil {
		t.Fatalf("processHTML failed: %v", err)
	}

	resultStr := string(result)
	// Check that header with data-llm="keep" is preserved
	if !strings.Contains(resultStr, "<header data-llm=\"keep\">") {
		t.Error("Result missing <header> tag with data-llm=\"keep\"")
	}
	// Check that footer is removed
	if strings.Contains(resultStr, "<footer>") {
		t.Error("Result still contains <footer> tag")
	}
}

func TestProcessHTMLWithDataLLMDrop(t *testing.T) {
	input := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
	<div data-llm="drop"><h1>drop this</h1></header>
	<footer><p>Footer Content</p></footer>
</body>
</html>`
	result, err := ProcessHTML([]byte(input), StripConfig{ElementsToStrip: []string{}})
	if err != nil {
		t.Fatalf("processHTML failed: %v", err)
	}

	resultStr := string(result)

	if strings.Contains(resultStr, "<div data-llm=\"drop\">") {
		t.Error("Result still contains <div> tag with data-llm=\"drop\"")
	}

	if strings.Contains(resultStr, "drop this") {
		t.Error("Result still contains <div> tag with data-llm=\"drop\"")
	}
}

func TestProcessHTMLExpandedDefaults(t *testing.T) {
	input := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
	<nav>Navigation</nav>
	<aside>Sidebar</aside>
	<main>Content</main>
	<script>alert('hi')</script>
	<style>.foo{}</style>
	<noscript>Enable JS</noscript>
	<svg><circle/></svg>
	<iframe src="x"></iframe>
</body>
</html>`

	result, err := ProcessHTML([]byte(input), StripConfig{})
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	resultStr := string(result)

	for _, tag := range []string{"<nav>", "<aside>", "<script>", "<style>", "<noscript>", "<svg>", "<iframe>"} {
		if strings.Contains(resultStr, tag) {
			t.Errorf("Result still contains %s tag", tag)
		}
	}

	if !strings.Contains(resultStr, "Content") {
		t.Error("Result missing main content")
	}
}

func TestProcessHTMLScriptWithDescription(t *testing.T) {
	input := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
	<p>Before</p>
	<script data-llm-description="Form validation for email input">
		function validate() { return true; }
	</script>
	<p>After</p>
</body>
</html>`

	result, err := ProcessHTML([]byte(input), StripConfig{})
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	resultStr := string(result)

	if strings.Contains(resultStr, "<script") {
		t.Error("Result still contains <script> tag")
	}

	if !strings.Contains(resultStr, "Javascript description: Form validation for email input") {
		t.Errorf("Result missing script description text. Got: %s", resultStr)
	}
}

func TestProcessHTMLScriptWithoutDescription(t *testing.T) {
	input := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
	<script>console.log('hi')</script>
	<p>Content</p>
</body>
</html>`

	result, err := ProcessHTML([]byte(input), StripConfig{})
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	resultStr := string(result)

	if strings.Contains(resultStr, "<script") {
		t.Error("Result still contains <script> tag")
	}
	if strings.Contains(resultStr, "javascript") {
		t.Error("Result should not contain javascript description text")
	}
	if !strings.Contains(resultStr, "Content") {
		t.Error("Result missing paragraph content")
	}
}

func TestProcessHTMLScriptWithKeep(t *testing.T) {
	input := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
	<script data-llm="keep">important()</script>
</body>
</html>`

	result, err := ProcessHTML([]byte(input), StripConfig{})
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	resultStr := string(result)

	if !strings.Contains(resultStr, "<script") {
		t.Error("Script with data-llm=\"keep\" should be preserved")
	}
	if !strings.Contains(resultStr, "important()") {
		t.Error("Script content should be preserved")
	}
}

func TestProcessHTMLImageWithAlt(t *testing.T) {
	input := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
	<p>Before</p>
	<img src="photo.jpg" alt="A sunset over mountains">
	<p>After</p>
</body>
</html>`

	result, err := ProcessHTML([]byte(input), StripConfig{})
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	resultStr := string(result)

	if strings.Contains(resultStr, "<img") {
		t.Error("Result still contains <img> tag")
	}
	if !strings.Contains(resultStr, "[Image: A sunset over mountains]") {
		t.Errorf("Result missing image alt text. Got: %s", resultStr)
	}
}

func TestProcessHTMLImageNoAltDefault(t *testing.T) {
	input := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
	<img src="photo.jpg">
</body>
</html>`

	result, err := ProcessHTML([]byte(input), StripConfig{RemoveImagesNoAlt: false})
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	resultStr := string(result)

	if strings.Contains(resultStr, "<img") {
		t.Error("Result still contains <img> tag")
	}
	if !strings.Contains(resultStr, "[Image]") {
		t.Errorf("Result missing [Image] placeholder. Got: %s", resultStr)
	}
}

func TestProcessHTMLImageNoAltRemove(t *testing.T) {
	input := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
	<p>Before</p>
	<img src="photo.jpg">
	<p>After</p>
</body>
</html>`

	result, err := ProcessHTML([]byte(input), StripConfig{RemoveImagesNoAlt: true})
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	resultStr := string(result)

	if strings.Contains(resultStr, "<img") {
		t.Error("Result still contains <img> tag")
	}
	if strings.Contains(resultStr, "[Image]") {
		t.Error("Result should not contain [Image] placeholder when RemoveImagesNoAlt is true")
	}
}

func TestProcessHTMLKeepNavWithDataLLM(t *testing.T) {
	input := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
	<nav data-llm="keep">Important Nav</nav>
	<nav>Regular Nav</nav>
</body>
</html>`

	result, err := ProcessHTML([]byte(input), StripConfig{})
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	resultStr := string(result)

	if !strings.Contains(resultStr, "Important Nav") {
		t.Error("Result missing nav with data-llm=\"keep\"")
	}
	if strings.Contains(resultStr, "Regular Nav") {
		t.Error("Result should not contain regular nav")
	}
}

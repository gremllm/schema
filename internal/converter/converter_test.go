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

	result, err := ProcessHTML([]byte(input), StripConfig{StripNav: true, StripAside: true, StripScript: true})
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

	t.Logf("Processed HTML:\n%s", resultStr)
}

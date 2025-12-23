# LLM-Optimized Content Schema Project

## Executive Summary

This project creates a standardized system for serving LLM-optimized versions of web content. Instead of forcing LLMs to parse verbose HTML with navigation, ads, and boilerplate, websites can serve token-minimized markdown versions at `.md` URLs (e.g., `example.com/page.md` instead of `example.com/page`).

The key innovation is **not just HTML-to-Markdown conversion** (that's solved), but **intelligent content extraction with aggressive token optimization** while preserving semantic meaning. This is achieved through a combination of smart defaults, site-owner annotations, and token-aware transformations.

---

## Project Goals

This project is designed with four core objectives, listed in priority order:

### 1. Aggressive Token Reduction (50-80%)
**Goal:** Minimize the number of tokens LLMs need to process while preserving semantic meaning.

- Strip all LLM-irrelevant content (navigation, ads, footers, analytics scripts)
- Collapse unnecessary whitespace and formatting
- Simplify complex structures (nested lists, tables) into minimal representations
- Optimize heading hierarchies and link text
- Extract only meaningful image descriptions

**Why it matters:** Token costs are real. A typical webpage might be 3000-4000 tokens of HTML but only 600-800 tokens of actual content. This project targets 50-80% reduction, making LLM interactions faster and cheaper.

### 2. Optimized for LLM Consumption
**Goal:** Format content in the way LLMs can most effectively use it.

- Markdown is more token-efficient than HTML
- Semantic meaning preserved (LLMs should understand the content relationships)
- Handle interactive components intelligently (describe what they do, not how they work)
- Maintain document structure (headings, lists, code blocks) in a consistent way
- Remove elements LLMs can't interact with (JavaScript, CSS, analytics)

**Why it matters:** LLMs don't need CSS styling or JavaScript functionality. They need clean, structured text that conveys meaning efficiently.

### 3. Single Source of Truth (Centralized Management)
**Goal:** Maintain one canonical implementation that generates or defines all others.

- One codebase to update when the spec evolves
- Consistent behavior across all language implementations
- Reduces maintenance burden (no need to fix bugs in 5+ languages)
- Clear versioning and evolution path
- Community can contribute to one place

**Why it matters:** Maintaining multiple implementations is error-prone and expensive. Divergent behavior breaks the standard. A single source of truth ensures consistency and reduces long-term costs.

### 4. Consistent Behavior Across Implementations
**Goal:** Guarantee that the same HTML input produces identical markdown output, regardless of language/platform.

- Go implementation produces same result as JavaScript implementation
- Python produces same result as .NET
- Site owners can rely on predictable behavior
- LLM applications get consistent input regardless of backend
- Comprehensive test suite validates consistency

**Why it matters:** If different implementations produce different output, the standard is broken. Site owners can't trust it. LLM applications get inconsistent data. Consistency is critical for adoption.

### Secondary Goals

- **Easy adoption for site owners:** Simple annotation system using standard HTML classes
- **Framework agnostic:** Works with Next.js, Django, Rails, ASP.NET, etc.
- **Self-describing:** Sites advertise their LLM-markdown capability
- **Extensible:** Site owners can override defaults for their specific needs
- **Open standard:** Can become a widely-adopted specification (like schema.org)

---

## The Problem We're Solving

### Why Existing Solutions Don't Work

**Standard HTML-to-Markdown converters** (turndown.js, html2text, pandoc) perform format conversion but don't:
- Extract semantically relevant content (they convert everything)
- Minimize tokens for LLM consumption
- Strip LLM-irrelevant elements (navigation, ads, footers)
- Optimize structure for context windows
- Handle modern interactive components intelligently

**Content extraction tools** (Mozilla Readability, Article Extractor) identify main content but:
- Don't optimize for token efficiency
- Can't handle site-specific requirements
- Don't provide a standardized format for LLM consumption
- Can't be customized by site owners

### The Token Problem

Modern websites are bloated:
```html
<!-- Typical page: ~15-30KB HTML -->
<nav>...</nav>                    <!-- LLM doesn't need this -->
<header>...</header>              <!-- Or this -->
<aside>Newsletter signup</aside>  <!-- Definitely not this -->
<article>
  <p>Actual content here</p>      <!-- THIS is what matters -->
</article>
<footer>...</footer>              <!-- Nope -->
<script>Analytics...</script>     <!-- Can't even use this -->
```

**LLMs pay token costs for all of it.** A 15KB HTML page might be 4000 tokens, but the actual content could be represented in 800 tokens of well-structured markdown.

---

## Core Concept

### The Flow

```
User Request: example.com/article.md
                    ↓
    Framework renders: example.com/article (normal SSR/SSG)
                    ↓
    Middleware intercepts fully-rendered HTML
                    ↓
    Schema conversion applies:
        - Smart defaults (drop nav, footer, etc.)
        - Site annotations (llm-keep, llm-drop)
        - Token optimizations (collapse whitespace, flatten structure)
        - Client-side placeholders (describe interactive components)
                    ↓
    Returns: Optimized markdown (~50-80% fewer tokens)
```

### Discovery Mechanism

Sites that support LLM-optimized content include a marker in their HTML:

```html
<!-- Option A: Meta tag -->
<meta name="llm-markdown" content="available" />

<!-- Option B: Hidden div -->
<div style="display:none" data-llm-markdown="true"></div>
```

**How it works:**
1. LLM makes initial request to `example.com/index.html`
2. Parses HTML, finds the marker
3. Knows it can now request `example.com/about.md`, `example.com/contact.md`, etc.
4. All subsequent requests use `.md` URLs

**Pros:**
- Self-describing (sites advertise their capability)
- No central registry needed
- Works with existing URL structures

**Cons:**
- Requires initial HTML parse
- LLM clients need to implement discovery
- Site owners must opt-in

---

## Architecture

### 1. Smart Defaults

Certain HTML elements are automatically handled:

| Element | Default Action | Reasoning |
|---------|---------------|-----------|
| `<nav>` | Drop | Navigation is for humans, not LLMs |
| `<footer>` | Drop | Contact info, copyright notices are rarely relevant |
| `<aside>` | Drop | Sidebar content is typically tangential |
| `<header>` | Drop (unless contains `<h1>`) | Usually just branding/navigation |
| `<script>` | Drop | LLMs can't execute JavaScript |
| `<style>` | Drop | Styling information is irrelevant |
| `<article>` | Keep | Primary content container |
| `<main>` | Keep | Explicit main content designation |
| `<p>`, `<h1>`-`<h6>`, `<ul>`, `<ol>` | Keep | Core content elements |
| `<code>`, `<pre>` | Keep | Code examples are often highly relevant |

### 2. Annotation System

Site owners can override defaults using standard HTML classes:

```html
<!-- Override: Keep a footer that has relevant content -->
<footer class="llm-keep">
  <p>Last updated: 2024-01-15</p>
  <p>Author: Jane Doe, PhD in Computer Science</p>
</footer>

<!-- Override: Drop a paragraph that's not relevant -->
<p class="llm-drop">
  Click here to subscribe to our newsletter!
</p>

<!-- Describe interactive components -->
<div class="llm-describe" data-llm-description="Interactive mortgage calculator. Input: loan amount, interest rate, term. Output: monthly payment.">
  <MortgageCalculator />  <!-- React component -->
</div>
```

**Why this approach:**
- Uses standard HTML attributes (no custom syntax to learn)
- Works with existing class-based styling
- Site owners have fine-grained control
- Backward compatible (classes don't break existing functionality)

**Pros:**
- Extremely flexible
- No schema lock-in for site owners
- Works with any framework/CMS

**Cons:**
- Requires manual annotation by site owners
- Could conflict with existing CSS classes (though unlikely with `llm-` prefix)

**Alternative approach to consider:**
```html
<!-- Using data attributes instead of classes -->
<footer data-llm="keep">...</footer>
<p data-llm="drop">...</p>
<div data-llm-describe="Interactive calculator">...</div>
```

This avoids any CSS class conflicts but is more verbose.

### 3. Token Optimization Strategies

Beyond just stripping content, aggressive optimizations are applied:

#### Whitespace Normalization
```html
<!-- Before -->
<p>
    This    has    excessive

    whitespace
</p>

<!-- After -->
This has excessive whitespace
```

#### Heading Flattening
```html
<!-- Before: 6 levels of headings -->
<h1>Title</h1>
<h2>Section</h2>
<h3>Subsection</h3>
<h4>Sub-subsection</h4>
<h5>Too deep</h5>
<h6>Way too deep</h6>

<!-- After: Max 3 levels -->
# Title
## Section
### Subsection
### Sub-subsection
### Too deep
### Way too deep
```

#### List Flattening
```html
<!-- Before: Deeply nested lists -->
<ul>
  <li>Item 1
    <ul>
      <li>Nested 1
        <ul>
          <li>Deeply nested</li>
        </ul>
      </li>
    </ul>
  </li>
</ul>

<!-- After: Flattened with indentation -->
- Item 1
  - Nested 1 - Deeply nested
```

#### Table Simplification
Maybe use something like tonl https://github.com/tonl-dev/tonl
```html
<!-- Before: Complex table for simple data -->
<table>
  <tr><td>Name</td><td>John</td></tr>
  <tr><td>Age</td><td>30</td></tr>
</table>

<!-- After: Simple list -->
- Name: John
- Age: 30
```

#### Link Optimization
```html
<!-- Before: Verbose link text -->
<a href="/about">Click here to learn more about our company</a>

<!-- After: Concise -->
[About](/about)
```

#### Image Handling
```html
<!-- Decorative images: Drop -->
<img src="divider.png" alt="" />
<!-- Dropped entirely -->

<!-- Content images: Alt text only -->
<img src="diagram.png" alt="System architecture showing three-tier design" />
<!-- Becomes: -->
[Image: System architecture showing three-tier design]
```

**Configurable optimization levels:**
```yaml
token_optimization:
  level: aggressive  # conservative | moderate | aggressive
  whitespace_collapse: true
  max_heading_depth: 3
  flatten_nested_lists: true
  simplify_tables: true
  link_text_optimization: true
  image_strategy: alt_text_only  # keep_all | alt_text_only | drop_decorative
```

**Pros:**
- Significant token reduction (50-80%)
- Preserves semantic meaning
- Configurable per site

**Cons:**
- Some information loss is inevitable
- May need site-specific tuning
- Could oversimplify complex structures

### 4. Handling Dynamic Content

Modern websites heavily use client-side JavaScript. Our approach:

#### Server-Side Rendered (SSR) Content
✅ Works perfectly. Middleware intercepts after rendering, so all dynamic data is already in HTML.

```
Request: /product.md?id=123
         ↓
Framework: Renders /product with id=123 (SSR)
         ↓
Middleware: Converts rendered HTML to markdown
         ↓
Returns: Optimized markdown with product details
```

#### Client-Side Rendered (CSR) Content
⚠️ Can't execute JavaScript. Use placeholders:

```html
<!-- Original React component -->
<div id="calculator">
  <Calculator
    type="mortgage"
    features={["amortization", "comparison"]}
  />
</div>

<!-- With annotation for LLM -->
<div
  id="calculator"
  class="llm-describe"
  data-llm-description="Interactive mortgage calculator with amortization schedule and loan comparison features. Users can input loan amount ($), interest rate (%), and term (years) to calculate monthly payments."
>
  <Calculator />
</div>

<!-- Becomes in markdown -->
[Interactive Component: Interactive mortgage calculator with amortization schedule and loan comparison features. Users can input loan amount ($), interest rate (%), and term (years) to calculate monthly payments.]
```

**Best practices for site owners:**
- Annotate major interactive components
- Describe inputs, outputs, and purpose
- Consider what an LLM agent would need to know to help a user

**Pros:**
- Transparent to LLMs (they know what's there)
- Enables LLMs to guide users ("Go to the calculator and enter...")
- Site owners control the description

**Cons:**
- Requires manual annotation
- LLMs can't interact directly with components
- Quality depends on description accuracy

---

## Implementation Options

This is the critical decision: How do we maintain a single source of truth (Go implementation) while supporting multiple languages/frameworks (JavaScript/Next.js, Python/Django, .NET/ASP.NET, etc.)?

### Option 1: Go → WebAssembly (WASM)

**Concept:** Write the core conversion logic in Go, compile to WASM, load from other languages.

#### How It Works

```
schema/               (This repo)
├── cmd/
│   └── schema/       Go CLI tool
├── internal/
│   └── converter/    Core conversion logic
├── wasm/
│   └── build.sh      Builds schema.wasm
└── schema.wasm       Compiled WASM module

schema-js/            (Separate repo)
├── index.js          Loads schema.wasm, provides JS API
└── package.json      npm package

schema-nextjs/        (Separate repo)
├── middleware.js     Next.js middleware using schema-js
└── package.json
```

**JavaScript usage:**
```javascript
// schema-js wraps the WASM
import { convertToMarkdown } from '@gremllm/schema-js'

export function middleware(request) {
  if (request.url.endsWith('.md')) {
    const html = await renderPage(request)
    const markdown = convertToMarkdown(html)
    return new Response(markdown)
  }
}
```

**Python usage (theoretically):**
```python
# Using wasmer-python or wasmtime
from wasmer import engine, Store, Module, Instance
from wasmer_compiler_cranelift import Compiler

store = Store(engine.JIT(Compiler))
module = Module(store, open('schema.wasm', 'rb').read())
instance = Instance(module)

markdown = instance.exports.convert(html)
```

#### Pros
- ✅ Single implementation (Go is source of truth)
- ✅ Consistent behavior across all languages
- ✅ No code duplication
- ✅ Native performance (WASM is fast)
- ✅ Works in browsers, Node.js, and server environments
- ✅ Easiest to maintain (only update Go code)

#### Cons
- ❌ WASM support varies by language:
  - **JavaScript/Node.js**: Excellent (native support)
  - **Python**: Decent (wasmer-python, wasmtime) but not idiomatic
  - **.NET**: Good (Wasmtime.NET) but uncommon
  - **Ruby/PHP**: Exists but clunky
- ❌ FFI overhead for string marshalling (HTML in, markdown out)
- ❌ Debugging is harder (stack traces cross WASM boundary)
- ❌ Not idiomatic in non-JS languages (Python devs don't expect WASM)
- ❌ Error handling across WASM boundary is complex
- ❌ Larger binary size than native code

#### Best For
- JavaScript/TypeScript ecosystems (Next.js, Express, Deno, Bun)
- Projects where consistency is paramount
- Teams that don't want to maintain multiple implementations

#### Unknowns / Research Needed
- What's the actual performance overhead of WASM FFI for HTML string passing?
- How well do WASM runtimes work in production Python/Ruby environments?
- What's the debugging experience like in practice?

---

### Option 2: Go CGO → C FFI

**Concept:** Compile Go to a C-compatible shared library, call via FFI from other languages.

#### How It Works

```
schema/
├── cmd/
│   └── libschema/    CGO build target
├── internal/
│   └── converter/    Core conversion logic
└── build/
    ├── libschema.so      (Linux)
    ├── libschema.dylib   (macOS)
    └── libschema.dll     (Windows)
```

**CGO export:**
```go
package main

import "C"
import (
    "github.com/gremllm/schema/internal/converter"
)

//export Convert
func Convert(html *C.char) *C.char {
    goHtml := C.GoString(html)
    markdown := converter.Convert(goHtml)
    return C.CString(markdown)
}

func main() {}
```

**Build:**
```bash
# Build shared library
go build -buildmode=c-shared -o libschema.so cmd/libschema/main.go
```

**JavaScript usage (via Node.js FFI):**
```javascript
const ffi = require('ffi-napi')

const lib = ffi.Library('./libschema.so', {
  'Convert': ['string', ['string']]
})

function convertToMarkdown(html) {
  return lib.Convert(html)
}
```

**Python usage:**
```python
from ctypes import cdll, c_char_p

lib = cdll.LoadLibrary('./libschema.so')
lib.Convert.argtypes = [c_char_p]
lib.Convert.restype = c_char_p

def convert_to_markdown(html):
    return lib.Convert(html.encode()).decode()
```

**.NET usage:**
```csharp
using System.Runtime.InteropServices;

public class Schema {
    [DllImport("libschema.dll")]
    private static extern string Convert(string html);

    public static string ConvertToMarkdown(string html) {
        return Convert(html);
    }
}
```

#### Pros
- ✅ Single implementation (Go is source of truth)
- ✅ Works with virtually every language (C FFI is universal)
- ✅ Better performance than WASM for some workloads
- ✅ More mature tooling than WASM
- ✅ Easier debugging than WASM (standard shared library)
- ✅ Can call Go's excellent HTML parsing libraries directly

#### Cons
- ❌ Platform-specific binaries (.so, .dylib, .dll)
- ❌ Must distribute separate binaries for Linux, macOS, Windows
- ❌ Cross-compilation complexity
- ❌ CGO disables some Go optimizations
- ❌ Memory management is complex (who frees strings?)
- ❌ Still not idiomatic in host languages
- ❌ Deployment is harder (must ship platform-specific .so/.dll files)
- ❌ CGO increases build times significantly

#### Best For
- Server-side environments where you control deployment
- Projects that need maximum performance
- Teams comfortable with FFI and memory management

#### Unknowns / Research Needed
- What's the cross-compilation story for all platforms?
- How do we handle memory management safely across FFI boundary?
- What's the deployment experience like in practice (Docker helps here)?

---

### Option 3: Hybrid Approach

**Concept:** Prioritize languages by usage, implement accordingly.

#### Strategy

**Tier 1 Languages (Native implementations):** Go, JavaScript
- Hand-written, idiomatic implementations
- These cover 80-90% of users
- Optimized for performance and developer experience

**Tier 2 Languages (WASM or FFI):** Python, .NET, Ruby
- Use WASM or FFI approach
- Provide thin wrapper libraries
- Good enough for less common use cases

```
schema/           (Go - Tier 1, source of truth)
schema-js/        (JavaScript - Tier 1, hand-written)
schema-python/    (Python - Tier 2, uses WASM/FFI)
schema-dotnet/    (C# - Tier 2, uses WASM/FFI)
```

#### How To Maintain Consistency

**1. Formal specification:**
```yaml
# spec.yaml - The canonical definition
version: 1.0.0

default_rules:
  nav: drop
  footer: drop
  article: keep
  main: keep
  # ...

annotations:
  keep: llm-keep
  drop: llm-drop
  describe: llm-describe

token_optimization:
  whitespace_collapse: aggressive
  max_heading_depth: 3
  # ...
```

**2. Comprehensive test fixtures:**
```
tests/
  fixtures/
    001-basic-html/
      input.html
      expected.md
    002-with-annotations/
      input.html
      expected.md
    003-complex-nesting/
      input.html
      expected.md
    # ... hundreds of test cases
```

**3. Cross-implementation test runner:**
```bash
# Run all implementations against same fixtures
./test.sh
  ✓ schema (Go): 150/150 tests passed
  ✓ schema-js (JavaScript): 150/150 tests passed
  ✓ schema-python (Python): 150/150 tests passed
```

Any divergence is immediately caught.

#### Pros
- ✅ Best of both worlds (native + universal)
- ✅ Optimal experience for primary users
- ✅ Still supports less common languages
- ✅ Formal spec prevents drift
- ✅ Can optimize per-language (e.g., JS uses browser APIs, Python uses lxml)

#### Cons
- ❌ Most maintenance burden (multiple implementations)
- ❌ Risk of drift if discipline lapses
- ❌ Need to update multiple repos for spec changes
- ❌ Requires strong testing infrastructure
- ❌ Team needs expertise in multiple languages

#### Best For
- Projects with long-term support horizon
- Teams that can maintain multiple codebases
- When developer experience is critical

---

### Option 4: Specification + Hand-Written Implementations

**Concept:** Go is just one implementation. The spec is the source of truth.

#### How It Works

```
schema-spec/      (The canonical specification)
├── spec.yaml     Core rules and behavior
├── tests/        Comprehensive test fixtures
└── docs/         Implementation guide

schema-go/        (Go implementation)
schema-js/        (JavaScript implementation)
schema-python/    (Python implementation)
# Each implements the spec independently
```

**The spec is language-agnostic:**
```yaml
# spec.yaml
version: 1.0.0
description: LLM-optimized content conversion specification

default_rules:
  - element: nav
    action: drop
    reason: Navigation is for human browsing

  - element: footer
    action: drop
    override_class: llm-keep
    reason: Footers typically contain boilerplate

  - element: article
    action: keep
    override_class: llm-drop

  - element: script
    action: drop
    reason: LLMs cannot execute JavaScript

# ... etc
```

Each implementation must:
1. Pass all test fixtures
2. Follow the spec exactly
3. Publish conformance report

#### Pros
- ✅ Each implementation can be fully idiomatic
- ✅ Best performance (native code, no FFI overhead)
- ✅ Best developer experience per language
- ✅ Spec evolution is language-neutral
- ✅ Community can contribute implementations
- ✅ Easier debugging (native stack traces)

#### Cons
- ❌ Highest maintenance burden
- ❌ Risk of behavioral divergence
- ❌ Bugs must be fixed in multiple places
- ❌ Requires maintainers for each language
- ❌ Spec changes require coordinated updates
- ❌ Most development effort upfront

#### Best For
- Open source projects seeking wide adoption
- When performance and developer experience are critical
- Projects with community contributors
- Long-term, stable specifications

---

## Implementation Comparison Matrix

| Aspect | WASM | CGO/FFI | Hybrid | Spec-Driven |
|--------|------|---------|--------|-------------|
| **Maintenance** | ⭐⭐⭐⭐⭐ Lowest | ⭐⭐⭐⭐ Low | ⭐⭐ Medium | ⭐ Highest |
| **Consistency** | ⭐⭐⭐⭐⭐ Perfect | ⭐⭐⭐⭐⭐ Perfect | ⭐⭐⭐⭐ Very Good | ⭐⭐⭐ Good |
| **Performance** | ⭐⭐⭐⭐ Good | ⭐⭐⭐⭐⭐ Best | ⭐⭐⭐⭐⭐ Best | ⭐⭐⭐⭐⭐ Best |
| **JS/Node Experience** | ⭐⭐⭐⭐⭐ Excellent | ⭐⭐⭐ OK | ⭐⭐⭐⭐⭐ Excellent | ⭐⭐⭐⭐⭐ Excellent |
| **Python Experience** | ⭐⭐ Poor | ⭐⭐⭐ OK | ⭐⭐⭐⭐⭐ Excellent | ⭐⭐⭐⭐⭐ Excellent |
| **Deployment Complexity** | ⭐⭐⭐⭐ Low | ⭐⭐ High | ⭐⭐⭐ Medium | ⭐⭐⭐⭐ Low |
| **Debugging** | ⭐⭐ Hard | ⭐⭐⭐ OK | ⭐⭐⭐⭐ Good | ⭐⭐⭐⭐⭐ Easy |
| **Community Contribution** | ⭐⭐ Limited | ⭐⭐ Limited | ⭐⭐⭐ Medium | ⭐⭐⭐⭐⭐ Easy |

---

## Recommended Path Forward

### Phase 1: Proof of Concept (Go Only)
**Goal:** Validate the core concept works

1. Build Go implementation with:
   - Core HTML → Markdown conversion
   - Smart defaults (drop nav, footer, etc.)
   - Annotation support (llm-keep, llm-drop, llm-describe)
   - Token optimization strategies
   - Comprehensive test suite

2. Create example websites demonstrating:
   - Blog posts
   - Documentation sites
   - E-commerce product pages
   - Complex nested structures

3. Measure token reduction (target: 50-80%)

**Success criteria:** Token reduction achieved, semantic meaning preserved, site owners can easily annotate.

### Phase 2: Choose Distribution Strategy
**Goal:** Support JavaScript/Next.js (largest user base)

**Decision point:** Based on Phase 1 learnings, choose:
- **WASM** if performance is good and developer experience acceptable
- **Hand-written JS** if need idiomatic code or WASM has issues
- **Hybrid** if both approaches have merits

Build:
- `schema-js` package
- `schema-nextjs` middleware
- Example Next.js site using the middleware

### Phase 3: Expand Language Support
**Goal:** Support Python and/or .NET based on demand

Options:
- Use same approach as Phase 2 (WASM or hand-written)
- Hybrid approach (WASM for less common languages)

### Phase 4: Community & Ecosystem
**Goal:** Drive adoption

- Documentation site
- Example implementations for popular frameworks
- Browser extension for testing (shows .md version alongside .html)
- Tools for site owners to validate their annotations

---

## Open Questions & Research Needed

### Technical
1. **WASM Performance:** What's the actual overhead of WASM FFI for large HTML strings?
2. **HTML Parsing:** Which Go HTML parser is best? (`golang.org/x/net/html`, `github.com/PuerkitoBio/goquery`)
3. **Markdown Generation:** Build custom or use library? (`github.com/JohannesKaufmann/html-to-markdown` as reference, not direct use)
4. **Configuration:** YAML files, JSON, Go structs, or CLI flags?

### User Experience
1. **Annotation Discoverability:** How do site owners learn about `llm-keep`/`llm-drop`?
2. **Debugging:** How do site owners see what got kept vs. dropped?
3. **Validation:** Tool to test annotations before deployment?

### Ecosystem
1. **Standard:** Should this become a formal spec (like schema.org)?
2. **Discovery:** Meta tag vs. hidden div vs. HTTP header?
3. **Versioning:** How to handle schema evolution?

### Business
1. **Target Audience:** Who adopts first? Documentation sites? Blogs? E-commerce?
2. **Adoption Path:** How do we convince site owners to add annotations?
3. **Competition:** Are there existing efforts in this space?

---

## Example Use Cases

### Documentation Site
**Before (HTML):** 2500 tokens
```html
<nav>Sidebar with 50 links...</nav>
<header>Logo, search, user menu...</header>
<main>
  <article>
    <h1>API Reference</h1>
    <p>Actual documentation content...</p>
    <code>Example code...</code>
  </article>
</main>
<footer>Copyright, social links...</footer>
<script>Analytics, search...</script>
```

**After (Markdown):** 600 tokens
```markdown
# API Reference

Actual documentation content...

```
Example code...
```
```

**Token reduction: 76%**

### E-commerce Product Page
**Before (HTML):** 3200 tokens (ads, recommendations, reviews section UI, etc.)

**After (Markdown):** 800 tokens
```markdown
# Product Name

Price: $XX.XX

Description...

Specifications:
- Spec 1
- Spec 2

[Customer Reviews: 4.5/5 stars from 234 reviews]
```

**Token reduction: 75%**

### Blog Post
**Before (HTML):** 2800 tokens (navigation, sidebar, related posts, comments form, etc.)

**After (Markdown):** 700 tokens
```markdown
# Blog Post Title

Author: Jane Doe | Published: 2024-01-15

Main content here...
```

**Token reduction: 75%**

---

## Success Metrics

### Technical Metrics
- **Token reduction:** 50-80% reduction vs. raw HTML
- **Semantic preservation:** LLM comprehension tests (can it answer questions about the content?)
- **Performance:** < 50ms conversion time for typical pages
- **Consistency:** 100% test pass rate across implementations

### Adoption Metrics
- Number of sites implementing the standard
- Number of framework integrations
- GitHub stars, npm downloads, etc.

### Quality Metrics
- Developer satisfaction (surveys, feedback)
- LLM application performance (can they use the content effectively?)
- Site owner ease of use (how hard is it to add annotations?)

---

## Next Steps

1. **Review this document** with your friend
2. **Choose initial implementation strategy** (WASM, FFI, hand-written, or hybrid)
3. **Set up Go project structure**
4. **Implement core conversion logic** (HTML parsing, smart defaults)
5. **Build test suite** (fixtures, expected outputs)
6. **Create proof-of-concept website** (simple blog or docs site)
7. **Measure token reduction** (validate the concept)
8. **Iterate on defaults and optimizations**
9. **Choose JavaScript distribution method** (WASM vs. hand-written)
10. **Build schema-js and schema-nextjs** (first framework integration)

---

## Contributing & Feedback

*This section will be filled in once the project is public*

**Questions to answer:**
- How do people contribute new language implementations?
- How do we handle spec evolution?
- What's the governance model?

---

## Appendix: Alternative Approaches Considered

### HTTP API Service (Rejected)
Every language makes HTTP calls to a Go service.

**Why rejected:**
- Network overhead (even localhost adds latency)
- Deployment complexity (need to run a service)
- Adds infrastructure burden for site owners
- Makes development harder (need service running)

While this is the most universal approach, it moves too much complexity to runtime. For a conversion task that happens on every request, in-process execution is better.

### Browser Extension Only (Rejected)
End-users install extension to get .md versions.

**Why rejected:**
- Doesn't solve the LLM agent use case (they can't install extensions)
- Puts burden on end-users instead of site owners
- Can't optimize based on site-specific semantics
- Doesn't create a standard

### LLM-Specific HTML Tags (Rejected)
Custom HTML like `<llm-keep>`, `<llm-drop>`.

**Why rejected:**
- Requires custom HTML parser (breaks W3C validation)
- Tooling doesn't support custom tags
- CMS/frameworks would need updates
- Classes/attributes are more compatible

---

## License & Legal

*To be determined*

Considerations:
- Open source license (MIT, Apache 2.0, BSD?)
- Trademark for "LLM Schema" or similar?
- Patent considerations (unlikely, but worth thinking about)

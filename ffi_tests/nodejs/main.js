const koffi = require('koffi')

// Load the shared library
const lib = koffi.load('./build/libschema.so')

// Define function signatures - using char* for auto string conversion
const Convert = lib.func('char* Convert(char* htmlInput, int stripNav, int stripAside, int stripScript)')

// Note: Not using Free() due to koffi memory management complexity
// In production, you'd need a proper memory management strategy

// Test HTML
const htmlInput = `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
    <header><h1>This should be stripped</h1></header>
    <nav><a href="/">Home</a></nav>
    <main><p>This content should remain</p></main>
    <footer><p>This should also be stripped</p></footer>
</body>
</html>`

console.log('Testing Convert()...\n')
console.log('Input HTML:')
console.log(htmlInput)

// Call with all options enabled (strip nav, aside, script)
const result = Convert(htmlInput, 1, 1, 1)

console.log('\nOutput HTML:')
console.log(result)

let errors = 0;

// Basic checks
if (!result.includes('<header>')) {
    console.log('\n✓ Header tag successfully stripped')
} else {
    console.log('\n✗ Header tag still present')
    errors++;
}

if (!result.includes('<footer>')) {
    console.log('✓ Footer tag successfully stripped')
} else {
    console.log('✗ Footer tag still present')
    errors++;
}

if (result.includes('<main>')) {
    console.log('✓ Main content preserved')
} else {
    console.log('✗ Main content missing')
    errors++;
}

if (errors > 0) {
    console.log(`\n${errors} errors found`)
    process.exit(1)
}
console.log('\nCGO library is working correctly!')

#!/bin/bash
# Test script to verify C11 macro compilation
# Usage: ./test_macros.sh

set -e

echo "=== Testing cgocopy2 C11 Macros ==="
echo ""

# Check for C11 compiler
if ! command -v gcc &> /dev/null; then
    echo "Error: gcc not found. Please install a C11-compatible compiler."
    exit 1
fi

# Check GCC version (need 4.9+ for good C11 support)
GCC_VERSION=$(gcc -dumpversion | cut -d. -f1)
if [ "$GCC_VERSION" -lt 5 ]; then
    echo "Warning: GCC version may not have complete C11 support (found $GCC_VERSION, recommend 5+)"
fi

echo "Compiler: $(gcc --version | head -n1)"
echo ""

# Compile example.c
echo "Compiling example.c with C11 standard..."
gcc -std=c11 -Wall -Wextra -DSTANDALONE_TEST -o example_test example.c

if [ $? -ne 0 ]; then
    echo "Error: Compilation failed"
    exit 1
fi

echo "✓ Compilation successful"
echo ""

# Run the example
echo "Running example to verify metadata generation..."
echo ""
./example_test

if [ $? -ne 0 ]; then
    echo "Error: Example execution failed"
    exit 1
fi

# Cleanup
rm -f example_test

echo ""
echo "✓ All tests passed!"
echo ""
echo "The C11 macros are working correctly."
echo "You can now use them in your cgo project with:"
echo "  // #cgo CFLAGS: -std=c11"
echo "  // #include \"native2/cgocopy_macros.h\""

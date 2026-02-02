#!/bin/bash
# run.sh - Compile C to ARM64 and run on macOS Apple Silicon
#
# Usage: ./scripts/run.sh <input.c>
#
# This script:
# 1. Compiles C source to ARM64 assembly using ralph-cc
# 2. Converts Linux/ELF assembly to macOS/Mach-O format
# 3. Assembles with `as`
# 4. Links with `ld`
# 5. Runs the executable

set -e

if [ $# -lt 1 ]; then
    echo "Usage: $0 <input.c>" >&2
    exit 1
fi

INPUT="$1"
BASENAME="${INPUT%.c}"
ASM_FILE="${BASENAME}.s"
MACOS_ASM="${BASENAME}_macos.s"
OBJ_FILE="${BASENAME}.o"
EXEC_FILE="${BASENAME}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
RALPH_CC="${SCRIPT_DIR}/../bin/ralph-cc"

# Step 1: Generate ARM64 assembly
echo "==> Compiling $INPUT to ARM64 assembly..."
"$RALPH_CC" -dasm "$INPUT"

if [ ! -f "$ASM_FILE" ]; then
    echo "Error: Assembly file $ASM_FILE not generated" >&2
    exit 1
fi

# Step 2: Convert Linux/ELF assembly to macOS/Mach-O format
# - Remove .type and .size directives (ELF-specific)
# - Add underscore prefix to global symbols and labels (macOS ABI)
echo "==> Converting to macOS format..."
perl -pe '
    s/^\s*\.type.*//;       # Remove .type directive
    s/^\s*\.size.*//;       # Remove .size directive
    s/\.global\s+([a-zA-Z_][a-zA-Z0-9_]*)/.global _\1/;  # Prefix global symbols
    s/^([a-zA-Z_][a-zA-Z0-9_]*):/_\1:/;                   # Prefix label definitions
    s/\bbl\s+([a-zA-Z_][a-zA-Z0-9_]*)/bl _\1/;           # Prefix bl targets
' "$ASM_FILE" > "$MACOS_ASM"

# Step 3: Assemble
echo "==> Assembling..."
as -o "$OBJ_FILE" "$MACOS_ASM"

# Step 4: Link
# Note: We need to link against system libraries for a proper executable
echo "==> Linking..."
SDK_PATH=$(xcrun --show-sdk-path)
ld -o "$EXEC_FILE" "$OBJ_FILE" -lSystem -L"$SDK_PATH/usr/lib"

# Step 5: Run
echo "==> Running $EXEC_FILE..."
echo "---"
"$EXEC_FILE"
EXIT_CODE=$?
echo "---"
echo "Exit code: $EXIT_CODE"

# Cleanup intermediate files
rm -f "$MACOS_ASM" "$OBJ_FILE"

exit $EXIT_CODE

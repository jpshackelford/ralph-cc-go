#!/bin/bash
# run.sh - Compile C to ARM64 and run on macOS Apple Silicon
#
# Usage: ./scripts/run.sh <input.c>
#
# This script:
# 1. Compiles C source to ARM64 assembly using ralph-cc (macOS format)
# 2. Assembles with `as`
# 3. Links with `ld`
# 4. Runs the executable

set -e

if [ $# -lt 1 ]; then
    echo "Usage: $0 <input.c>" >&2
    exit 1
fi

INPUT="$1"
BASENAME="${INPUT%.c}"
INPUTNAME="$(basename "$BASENAME")"
ASM_FILE="${BASENAME}.s"
OBJ_FILE="${BASENAME}.o"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
RALPH_CC="${SCRIPT_DIR}/../bin/ralph-cc"
OUT_DIR="${SCRIPT_DIR}/../out"

# Create output directory
mkdir -p "$OUT_DIR"
EXEC_FILE="${OUT_DIR}/${INPUTNAME}"

# Step 1: Generate ARM64 assembly (already in macOS format)
echo "==> Compiling $INPUT to ARM64 assembly..."
"$RALPH_CC" -dasm "$INPUT"

if [ ! -f "$ASM_FILE" ]; then
    echo "Error: Assembly file $ASM_FILE not generated" >&2
    exit 1
fi

# Step 2: Assemble
echo "==> Assembling..."
as -o "$OBJ_FILE" "$ASM_FILE"

# Step 3: Link
# Note: We need to link against system libraries for a proper executable
echo "==> Linking..."
SDK_PATH=$(xcrun --show-sdk-path)
ld -o "$EXEC_FILE" "$OBJ_FILE" -lSystem -L"$SDK_PATH/usr/lib"

# Step 4: Run
echo "==> Running $EXEC_FILE..."
echo "---"
"$EXEC_FILE"
EXIT_CODE=$?
echo "---"
echo "Exit code: $EXIT_CODE"

# Cleanup intermediate files
rm -f "$OBJ_FILE"

exit $EXIT_CODE

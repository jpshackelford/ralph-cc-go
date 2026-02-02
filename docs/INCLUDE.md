# Include Directive Support

ralph-cc supports C preprocessor directives including `#include` by delegating to the system's C preprocessor, following the same approach as CompCert.

## How It Works

1. When compiling a `.c` file, ralph-cc first runs the file through the system's C preprocessor (`cc -E`)
2. The preprocessor expands all `#include` directives, macros, and conditional compilation
3. The preprocessed output (with `#line` directives) is then parsed by ralph-cc
4. The lexer already handles `#line` directives to track source locations

## File Extensions

Following CompCert conventions:

| Extension | Preprocessed? |
|-----------|--------------|
| `.c`      | Yes - runs through preprocessor |
| `.i`      | No - assumed already preprocessed |
| `.p`      | No - assumed already preprocessed |

## Command Line Options

### Include Paths

Use `-I` to add directories to the include search path:

```bash
ralph-cc -I./include -I/usr/local/include source.c -dparse
```

Multiple `-I` flags can be specified.

## Example Usage

### Basic Include

Create a header file `myheader.h`:
```c
#define MY_CONSTANT 42
int helper_function(int x);
```

Create a source file `main.c`:
```c
#include "myheader.h"

int main() {
    return MY_CONSTANT;
}
```

Compile with:
```bash
ralph-cc main.c -dparse
```

### System Includes

To include system headers like `<stdio.h>`:

```c
#include <stdio.h>

int main() {
    return 0;
}
```

Note: While the preprocessor expands system headers, ralph-cc's parser may not support all constructs found in system headers (e.g., `__attribute__`, complex type specifiers). For practical use, consider using a subset of standard library functionality or providing simplified header stubs.

## Limitations

1. **System Header Compatibility**: Standard library headers often contain compiler-specific extensions that ralph-cc may not parse. For testing purposes, consider:
   - Using simple custom headers
   - Creating stub headers with only the declarations you need
   - Pre-processing with a compatible compiler and using the `.i` extension

2. **Preprocessor Selection**: ralph-cc uses the first available preprocessor from: `cc`, `gcc`, `clang`. The preprocessor must be in the system PATH.

3. **Preprocessor Flags**: Currently only `-I` paths are supported. Future versions may add:
   - `-D` for macro definitions
   - `-U` for macro undefinitions

## Implementation Details

The preprocessing is handled by `pkg/preproc/preproc.go`:

- `Preprocess(filename, opts)` - Runs preprocessor on a file
- `NeedsPreprocessing(filename)` - Checks if file needs preprocessing based on extension
- Preprocessor output includes `#line` directives which the lexer handles in `skipLineDirective()`

## Comparison with CompCert

Like CompCert, ralph-cc:
- Delegates preprocessing to an external tool
- Does not implement its own preprocessor
- Handles `#line` directives from preprocessor output
- Supports include path specification via `-I`

Unlike CompCert, ralph-cc:
- Does not yet support all preprocessing options (-D, -U, -Wp, etc.)
- Uses a simpler preprocessor detection mechanism

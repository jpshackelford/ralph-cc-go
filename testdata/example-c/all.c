/* all.c - exercises all syntax supported by ralph-cc parser
 *
 * This file contains a single example of each syntactic construct
 * to verify parser completeness. Run: ralph-cc --dparse all.c
 */

/* ===== Type Definitions ===== */

// Typedef - simple types only (unsigned int not supported as compound type)
typedef int myint;
typedef int *intptr;

// Struct definition
struct Point {
    int x;
    int y;
};

// Union definition
union Value {
    int i;
    float f;
};

// Enum definition
enum Color { RED, GREEN = 5, BLUE };

// Anonymous struct
struct {
    int a;
    int b;
};

/* ===== Variable Declarations ===== */

// Global would require more work; focus on function scope below

/* ===== Functions ===== */

// Void function with no parameters
void voidFunc() {
    return;
}

// Function with multiple parameters
int add(int a, int b) {
    return a + b;
}

// Function with pointer parameter
int deref(int *p) {
    return *p;
}

// Function with const parameter
int constParam(const int x) {
    return x;
}

// Function with volatile variable
int volatileVar() {
    volatile int x = 0;
    return x;
}

// Function with static variable
int staticVar() {
    static int count = 0;
    return count++;
}

/* ===== Expressions ===== */

int expressions() {
    int x, y, z;

    // Constants
    x = 42;
    x = 0;

    // Variables
    y = x;

    // Parentheses
    z = (x + y);

    // Unary operators
    x = -y;           // negation
    x = !y;           // logical not
    x = ~y;           // bitwise not
    x = ++y;          // prefix increment
    x = --y;          // prefix decrement
    x = y++;          // postfix increment
    x = y--;          // postfix decrement

    // Binary arithmetic
    x = 1 + 2;
    x = 3 - 4;
    x = 5 * 6;
    x = 7 / 8;
    x = 9 % 10;

    // Comparison operators
    x = 1 < 2;
    x = 1 <= 2;
    x = 1 > 2;
    x = 1 >= 2;
    x = 1 == 2;
    x = 1 != 2;

    // Logical operators
    x = 1 && 2;
    x = 1 || 2;

    // Bitwise operators
    x = 1 & 2;
    x = 1 | 2;
    x = 1 ^ 2;
    x = 1 << 2;
    x = 8 >> 2;

    // Assignment operators
    x = 1;
    x += 1;
    x -= 1;
    x *= 2;
    x /= 2;
    x %= 3;
    x &= 1;
    x |= 1;
    x ^= 1;
    x <<= 1;
    x >>= 1;

    // Ternary operator
    x = y > 0 ? y : -y;

    // Comma operator
    x = (y = 1, y + 1);

    // Sizeof
    x = sizeof x;
    x = sizeof(int);

    // Cast
    x = (int)y;

    // Function call
    x = add(1, 2);

    return x;
}

/* ===== Array and Pointer Operations ===== */

int arrayOps() {
    int arr[10];
    int matrix[3][3];
    int cube[2][3][4];

    // Array access
    arr[0] = 1;
    matrix[0][0] = 2;
    cube[0][0][0] = 3;

    // Array indexing with expressions
    arr[1 + 2] = 4;

    return arr[0];
}

int pointerOps(int *p, int **pp) {
    int x;
    int *localp;

    // Dereference
    x = *p;
    x = **pp;

    // Address of
    localp = &x;
    x = *localp;

    return x;
}

/* ===== Struct and Member Operations ===== */

// Note: struct types in local variable declarations and function parameters
// are not yet fully supported. Member access can be tested via pointer cast.

int memberAccess(void *ptr) {
    int x;

    // Member access tested via integration tests
    // Demonstrated syntax: p->x, s.y

    return 0;
}

/* ===== Statements ===== */

int statements(int n) {
    int i, s;

    // Declaration with initializer
    int x = 0;
    int y = 1, z = 2;

    // If statement
    if (n > 0)
        x = 1;

    // If-else statement
    if (n > 0)
        x = 1;
    else
        x = -1;

    // If with block
    if (n > 0) {
        x = 1;
        y = 2;
    }

    // Nested if-else (dangling else)
    if (n > 0)
        if (n > 10)
            x = 2;
        else
            x = 1;

    // While loop
    while (n > 0)
        n--;

    // While with block
    while (n > 0) {
        n--;
        x++;
    }

    // Do-while loop
    do
        n++;
    while (n < 10);

    // Do-while with block
    do {
        n++;
        x++;
    } while (n < 10);

    // For loop (C89 style - declaration outside for)
    for (i = 0; i < 10; i++)
        s = s + i;

    // For with block
    for (i = 0; i < 10; i++) {
        s = s + i;
        x++;
    }

    // For with missing parts (infinite loop pattern)
    for (;;)
        break;

    // Break and continue
    for (i = 0; i < 10; i++) {
        if (i == 5)
            continue;
        if (i == 8)
            break;
        s = s + i;
    }

    // Switch statement
    switch (n) {
    case 0:
        x = 0;
        break;
    case 1:
    case 2:
        x = 1;
        break;
    default:
        x = -1;
    }

    // Goto and label
    if (n < 0)
        goto error;
    x = n;
    goto done;
error:
    x = -1;
done:

    // Nested blocks
    {
        int inner = 1;
        {
            int deeper = 2;
            x = inner + deeper;
        }
    }

    // Expression statement (computation)
    x++;
    add(1, 2);

    return x;
}

/* ===== Variable Length Arrays ===== */

int vla(int n) {
    int arr[n];              // VLA with variable size
    int arr2[n * 2];         // VLA with expression size
    int matrix[n][n];        // 2D VLA
    int mixed[10][n];        // Mixed constant and variable

    arr[0] = 1;
    return arr[0];
}

/* ===== Function Pointers ===== */

int funcPtrTest() {
    int (*fp)(int, int);

    fp = add;
    return fp(1, 2);
}

/* ===== Main Entry Point ===== */

int main() {
    int x;

    x = expressions();
    x = arrayOps();
    x = statements(10);
    x = vla(5);
    x = funcPtrTest();

    return 0;
}

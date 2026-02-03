#include <stdio.h>

int main() {
    int n = 30;
    long long first = 0, second = 1, next;

    printf("First 30 Fibonacci numbers:\n");

    for (int i = 1; i <= n; i++) {
        printf("%lld ", first);
        next = first + second;
        first = second;
        second = next;
    }

    return 0;
}

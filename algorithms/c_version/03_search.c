#include <math.h>
#include <stdbool.h>

#include "algorithms.h"

bool LinearSearch(const int *haystack, int n, int needle) {

    if (haystack == NULL || n <= 0) {
            return false;
    }

    for (int i = 0; i < n; i++) {
        if (haystack[i] == needle) {
            return true;
        }
    }

    return false;
}

bool BinarySearch(const int *haystack, int n, int needle) {

    if (haystack == NULL || n <= 0) {
            return false;
    }

    int low = 0;
    int high = n;

    while (low < high) {
        int midpoint = low + (high - low) / 2;
        int value = haystack[midpoint];

        if (value == needle) {
            return true;
        } else if (value < needle) {
            low = midpoint + 1;
        } else {
            high = midpoint;
        }
    }

    return false;
}

bool BinarySearchRange(const int *haystack, int low, int high, int needle) {

    while (low < high) {
        int mid = low + (high - low) / 2;
        if (haystack[mid] == needle) {
            return true;
        }
        if (haystack[mid] < needle) {
            low = mid + 1;
        } else {
            high = mid;
        }
    }

    return false;
}

bool InterpolationSearch(const int *haystack, int n, int needle) {

    if (haystack == NULL || n <= 0) {
            return false;
    }

    int low = 0;
    int high = n;

    while (low < high && needle >= haystack[low] && needle <= haystack[high - 1]) {

        if (haystack[low] == haystack[high - 1]) {
            return haystack[low] == needle;
        }

        int pos = low + (int)(((double)needle - haystack[low]) * ((high - 1) - low) / ((double)haystack[high - 1] - haystack[low]));

        if (haystack[pos] == needle) {
            return true;
        }

        if (haystack[pos] < needle) {
            low = pos + 1;
        } else {
            high = pos;
        }
    }

    return false;
}

bool ExponentialSearch(const int *haystack, int n, int needle) {

    if (haystack == NULL || n <= 0) {
            return false;
    }
    if (haystack[0] == needle) {
        return true;
    }

    int high = 1;
    while (high < n && haystack[high] <= needle) {
        high *= 2;
    }

    return BinarySearchRange(haystack, high / 2, high < n ? high : n, needle);
}

bool JumpSearch(const int *haystack, int n, int needle) {

    if (haystack == NULL || n <= 0) {
            return false;
    }

    int step = (int)sqrt((double)n);
    int low = 0;
    int high = step;

    while (high < n && haystack[high - 1] < needle) {
        low = high;
        high += step;
    }

    if (high > n) {
        high = n;
    }

    for (int i = low; i < high; i++) {
        if (haystack[i] == needle) {
            return true;
        }
        if (haystack[i] > needle) {
            return false;
        }
    }

    return false;
}

int  TwoCrystalBalls(const bool *breaks, int n) {

    if (breaks == NULL || n <= 0) {
            return -1;
    }

    int jumpAmount = (int)sqrt((double)n);
    int i = jumpAmount;

    for (; i < n; i += jumpAmount) {
        if (breaks[i]) {
            break;
        }
    }

    i -= jumpAmount;
    for (int j = 0; j <= jumpAmount && i < n; j++) {
        if (breaks[i]) {
            return i;
        }
        i++;
    }

    return -1;
}

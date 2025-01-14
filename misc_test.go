package synapsecleaner_test

import (
	"fmt"
	"reflect"
	"testing"

	synapsecleaner "github.com/erdnaxeli/synapse-cleaner"
)

func TestDiffSlices(t *testing.T) {
	testCases := []struct {
		a        []int
		b        []int
		expected []int
	}{
		{
			a:        []int{1, 2, 3, 4, 5},
			b:        []int{},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			a:        []int{},
			b:        []int{1, 2, 3, 4, 5},
			expected: []int{},
		},
		{
			a:        []int{1, 2, 3, 4, 5},
			b:        []int{6, 7, 8},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			a:        []int{1, 2, 3, 4, 5},
			b:        []int{1, 3, 5, 7},
			expected: []int{2, 4},
		},
		{
			a:        []int{1, 2, 3, 4, 5},
			b:        []int{0, 1, 2, 3, 4, 5, 6},
			expected: []int{},
		},
	}

	for _, testCase := range testCases {
		result := synapsecleaner.DiffSlices(testCase.a, testCase.b)
		if !reflect.DeepEqual(result, testCase.expected) {
			t.Errorf("Expected %v but got %v", testCase.expected, result)
		}
	}
}

func TestDiffSlicesFunc(t *testing.T) {
	testCases := []struct {
		a        []int
		b        []string
		keyA     func(e int) string
		keyB     func(e string) string
		expected []int
	}{
		{
			a:        []int{1, 2, 3, 4, 5},
			b:        []string{},
			keyA:     func(e int) string { return fmt.Sprintf("%d-", e) },
			keyB:     func(e string) string { return fmt.Sprintf("%s-", e) },
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			a:        []int{},
			b:        []string{"1", "2", "3", "4", "5"},
			keyA:     func(e int) string { return fmt.Sprintf("%d-", e) },
			keyB:     func(e string) string { return fmt.Sprintf("%s-", e) },
			expected: []int{},
		},
		{
			a:        []int{1, 2, 3, 4, 5},
			b:        []string{"6", "7", "8"},
			keyA:     func(e int) string { return fmt.Sprintf("%d-", e) },
			keyB:     func(e string) string { return fmt.Sprintf("%s-", e) },
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			a:        []int{1, 2, 3, 4, 5},
			b:        []string{"1", "3", "5", "7"},
			keyA:     func(e int) string { return fmt.Sprintf("%d-", e) },
			keyB:     func(e string) string { return fmt.Sprintf("%s-", e) },
			expected: []int{2, 4},
		},
		{
			a:        []int{1, 2, 3, 4, 5},
			b:        []string{"0", "1", "2", "3", "4", "5", "6"},
			keyA:     func(e int) string { return fmt.Sprintf("%d-", e) },
			keyB:     func(e string) string { return fmt.Sprintf("%s-", e) },
			expected: []int{},
		},
	}

	for _, testCase := range testCases {
		result := synapsecleaner.DiffSlicesFunc(testCase.a, testCase.b, testCase.keyA, testCase.keyB)
		if !reflect.DeepEqual(result, testCase.expected) {
			t.Errorf("Expected %v but got %v", testCase.expected, result)
		}
	}
}

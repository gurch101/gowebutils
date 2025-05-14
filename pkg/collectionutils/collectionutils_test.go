package collectionutils_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/gurch101/gowebutils/pkg/collectionutils"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name      string
		slice     []int
		predicate func(int) bool
		expected  bool
	}{
		{
			name:  "empty slice",
			slice: []int{},
			predicate: func(n int) bool {
				return n > 0
			},
			expected: false,
		},
		{
			name:  "contains matching element",
			slice: []int{1, 2, 3, 4, 5},
			predicate: func(n int) bool {
				return n == 3
			},
			expected: true,
		},
		{
			name:  "no matching elements",
			slice: []int{1, 2, 3, 4, 5},
			predicate: func(n int) bool {
				return n > 5
			},
			expected: false,
		},
		{
			name:  "multiple matching elements",
			slice: []int{1, 2, 3, 4, 5},
			predicate: func(n int) bool {
				return n%2 == 0
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collectionutils.Contains(tt.slice, tt.predicate)
			if result != tt.expected {
				t.Errorf("Contains() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContainsAll(t *testing.T) {
	tests := []struct {
		name      string
		slice     []string
		predicate func(string) bool
		expected  bool
	}{
		{
			name:  "empty slice",
			slice: []string{},
			predicate: func(s string) bool {
				return len(s) > 0
			},
			expected: true, // vacuously true for empty set
		},
		{
			name:  "all elements match",
			slice: []string{"apple", "avocado"},
			predicate: func(s string) bool {
				return s[0] == 'a'
			},
			expected: true,
		},
		{
			name:  "not all elements match",
			slice: []string{"apple", "banana", "avocado"},
			predicate: func(s string) bool {
				return s[0] == 'a'
			},
			expected: false,
		},
		{
			name:  "single element that matches",
			slice: []string{"apple"},
			predicate: func(s string) bool {
				return len(s) > 3
			},
			expected: true,
		},
		{
			name:  "single element that doesn't match",
			slice: []string{"pie"},
			predicate: func(s string) bool {
				return len(s) > 5
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collectionutils.ContainsAll(tt.slice, tt.predicate)
			if result != tt.expected {
				t.Errorf("ContainsAll() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWithCustomType(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	people := []Person{
		{"Alice", 25},
		{"Bob", 30},
		{"Charlie", 35},
	}

	t.Run("Contains with custom type", func(t *testing.T) {
		result := collectionutils.Contains(people, func(p Person) bool {
			return p.Name == "Bob"
		})
		if !result {
			t.Error("Expected to find Bob in the slice")
		}
	})

	t.Run("ContainsAll with custom type", func(t *testing.T) {
		result := collectionutils.ContainsAll(people, func(p Person) bool {
			return p.Age > 20
		})
		if !result {
			t.Error("Expected all people to be over 20")
		}

		result = collectionutils.ContainsAll(people, func(p Person) bool {
			return p.Age > 30
		})
		if result {
			t.Error("Not all people are over 30")
		}
	})
}

func TestFindFirst(t *testing.T) {
	type person struct {
		name string
		age  int
	}

	tests := []struct {
		name      string
		slice     []int
		predicate func(int) bool
		wantVal   int
		wantFound bool
	}{
		{
			name:      "empty slice",
			slice:     []int{},
			predicate: func(n int) bool { return n > 0 },
			wantVal:   0,
			wantFound: false,
		},
		{
			name:      "found first matching element",
			slice:     []int{1, 2, 3, 4, 5},
			predicate: func(n int) bool { return n > 2 },
			wantVal:   3,
			wantFound: true,
		},
		{
			name:      "no matching elements",
			slice:     []int{1, 2, 3, 4, 5},
			predicate: func(n int) bool { return n > 5 },
			wantVal:   0,
			wantFound: false,
		},
		{
			name:      "first of multiple matches",
			slice:     []int{1, 2, 3, 4, 5},
			predicate: func(n int) bool { return n%2 == 0 },
			wantVal:   2,
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotFound := collectionutils.FindFirst(tt.slice, tt.predicate)
			if gotVal != tt.wantVal || gotFound != tt.wantFound {
				t.Errorf("FindFirst() = (%v, %v), want (%v, %v)", gotVal, gotFound, tt.wantVal, tt.wantFound)
			}
		})
	}

	t.Run("custom struct type", func(t *testing.T) {
		people := []person{
			{"Alice", 25},
			{"Bob", 30},
			{"Charlie", 35},
		}

		gotPerson, found := collectionutils.FindFirst(people, func(p person) bool {
			return p.age > 28
		})

		wantPerson := person{"Bob", 30}
		if !found || gotPerson != wantPerson {
			t.Errorf("FindFirst() = (%v, %v), want (%v, true)", gotPerson, found, wantPerson)
		}
	})

	t.Run("zero value when not found", func(t *testing.T) {
		gotPerson, found := collectionutils.FindFirst([]person{}, func(p person) bool {
			return p.age > 28
		})

		if found || gotPerson != (person{}) {
			t.Errorf("FindFirst() = (%v, %v), want (%v, false)", gotPerson, found, person{})
		}
	})
}

func TestEquals(t *testing.T) {
	t.Parallel()

	t.Run("int equals", func(t *testing.T) {
		tests := []struct {
			name     string
			a, b     []int
			expected bool
		}{
			{
				name:     "equal sets, same order",
				a:        []int{1, 2, 3},
				b:        []int{1, 2, 3},
				expected: true,
			},
			{
				name:     "equal sets, different order",
				a:        []int{1, 2, 3},
				b:        []int{3, 1, 2},
				expected: true,
			},
			{
				name:     "different values",
				a:        []int{1, 2, 3},
				b:        []int{4, 5, 6},
				expected: false,
			},
			{
				name:     "one extra element",
				a:        []int{1, 2},
				b:        []int{1, 2, 3},
				expected: false,
			},
			{
				name:     "duplicates in one slice",
				a:        []int{1, 2, 2, 3},
				b:        []int{3, 1, 2},
				expected: true,
			},
			{
				name:     "both empty",
				a:        []int{},
				b:        []int{},
				expected: true,
			},
			{
				name:     "one empty",
				a:        []int{},
				b:        []int{1},
				expected: false,
			},
		}

		for _, test := range tests {
			result := collectionutils.Equals(test.a, test.b)
			if result != test.expected {
				t.Errorf("Test '%s' failed: Equals(%v, %v) = %v, expected %v",
					test.name, test.a, test.b, result, test.expected)
			}
		}
	})

	t.Run("string equals", func(t *testing.T) {
		a := []string{"apple", "banana", "cherry"}
		b := []string{"banana", "cherry", "apple"}
		c := []string{"apple", "banana"}

		if !collectionutils.Equals(a, b) {
			t.Errorf("Expected Equals(a, b) to be true")
		}

		if collectionutils.Equals(a, c) {
			t.Errorf("Expected Equals(a, c) to be false")
		}
	})
}

func TestMap(t *testing.T) {
	t.Parallel()

	t.Run("int to int", func(t *testing.T) {
		input := []int{1, 2, 3}
		expected := []int{2, 4, 6}

		result := collectionutils.Map(input, func(x int) int {
			return x * 2
		})

		if len(result) != len(expected) {
			t.Errorf("Expected length %d, got %d", len(expected), len(result))
		}

		for i := range expected {
			if result[i] != expected[i] {
				t.Errorf("At index %d: expected %d, got %d", i, expected[i], result[i])
			}
		}
	})

	t.Run("int to string", func(t *testing.T) {
		input := []int{1, 2, 3}
		expected := []string{"1", "2", "3"}

		result := collectionutils.Map(input, strconv.Itoa)

		for i := range expected {
			if result[i] != expected[i] {
				t.Errorf("At index %d: expected %s, got %s", i, expected[i], result[i])
			}
		}
	})

	t.Run("string to upper", func(t *testing.T) {
		input := []string{"apple", "banana", "cherry"}
		expected := []string{"APPLE", "BANANA", "CHERRY"}

		result := collectionutils.Map(input, strings.ToUpper)

		for i := range expected {
			if result[i] != expected[i] {
				t.Errorf("At index %d: expected %s, got %s", i, expected[i], result[i])
			}
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		input := []int{}
		result := collectionutils.Map(input, func(x int) int {
			return x * 2
		})

		if len(result) != 0 {
			t.Errorf("Expected empty slice, got length %d", len(result))
		}
	})
}

package memory

import (
	"testing"
)

func TestHasMatchingTag(t *testing.T) {
	testcases := []struct {
		tags     map[string]string
		itemTags map[string]string
		expected bool
	}{
		{
			tags:     map[string]string{},
			itemTags: map[string]string{"k1": "v1"},
			expected: true,
		},
		{
			tags:     map[string]string{"k1": "v1"},
			itemTags: map[string]string{},
			expected: false,
		},
		{
			tags:     map[string]string{"k1": "v1"},
			itemTags: map[string]string{"k1": "v1"},
			expected: true,
		},
		{
			tags:     map[string]string{"k2": "v2"},
			itemTags: map[string]string{"k1": "v1"},
			expected: false,
		},
		{
			tags:     map[string]string{"k1": "v1"},
			itemTags: map[string]string{"k1": "v2"},
			expected: false,
		},
	}

	for _, tc := range testcases {
		res := hasMatchingTag(tc.tags, tc.itemTags)
		if res != tc.expected {
			t.Errorf("expected hasMatchingTag to return '%v' but got '%v'", tc.expected, res)
		}
	}
}

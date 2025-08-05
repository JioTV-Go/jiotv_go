package television

import (
	"reflect"
	"testing"
)

func TestFilterChannelsByDefaults(t *testing.T) {
	// Create test channels with different categories and languages
	testChannels := []Channel{
		{ID: "1", Name: "Channel1", Category: 1, Language: 1}, // Hindi Entertainment
		{ID: "2", Name: "Channel2", Category: 2, Language: 1}, // Hindi News
		{ID: "3", Name: "Channel3", Category: 1, Language: 2}, // English Entertainment
		{ID: "4", Name: "Channel4", Category: 3, Language: 1}, // Hindi Sports
		{ID: "5", Name: "Channel5", Category: 2, Language: 2}, // English News
	}

	tests := []struct {
		name       string
		channels   []Channel
		categories []int
		languages  []int
		expected   []Channel
	}{
		{
			name:       "No filters returns all channels",
			channels:   testChannels,
			categories: []int{},
			languages:  []int{},
			expected:   testChannels,
		},
		{
			name:       "Filter by single category",
			channels:   testChannels,
			categories: []int{1}, // Entertainment
			languages:  []int{},
			expected: []Channel{
				{ID: "1", Name: "Channel1", Category: 1, Language: 1},
				{ID: "3", Name: "Channel3", Category: 1, Language: 2},
			},
		},
		{
			name:       "Filter by single language",
			channels:   testChannels,
			categories: []int{},
			languages:  []int{1}, // Hindi
			expected: []Channel{
				{ID: "1", Name: "Channel1", Category: 1, Language: 1},
				{ID: "2", Name: "Channel2", Category: 2, Language: 1},
				{ID: "4", Name: "Channel4", Category: 3, Language: 1},
			},
		},
		{
			name:       "Filter by multiple categories",
			channels:   testChannels,
			categories: []int{1, 2}, // Entertainment and News
			languages:  []int{},
			expected: []Channel{
				{ID: "1", Name: "Channel1", Category: 1, Language: 1},
				{ID: "2", Name: "Channel2", Category: 2, Language: 1},
				{ID: "3", Name: "Channel3", Category: 1, Language: 2},
				{ID: "5", Name: "Channel5", Category: 2, Language: 2},
			},
		},
		{
			name:       "Filter by multiple languages",
			channels:   testChannels,
			categories: []int{},
			languages:  []int{1, 2}, // Hindi and English
			expected:   testChannels, // All channels match
		},
		{
			name:       "Filter by category AND language",
			channels:   testChannels,
			categories: []int{1}, // Entertainment
			languages:  []int{1}, // Hindi
			expected: []Channel{
				{ID: "1", Name: "Channel1", Category: 1, Language: 1},
			},
		},
		{
			name:       "Filter by multiple categories AND multiple languages",
			channels:   testChannels,
			categories: []int{1, 2}, // Entertainment and News
			languages:  []int{2},    // English
			expected: []Channel{
				{ID: "3", Name: "Channel3", Category: 1, Language: 2},
				{ID: "5", Name: "Channel5", Category: 2, Language: 2},
			},
		},
		{
			name:       "No matches",
			channels:   testChannels,
			categories: []int{99}, // Non-existent category
			languages:  []int{},
			expected:   []Channel{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterChannelsByDefaults(tt.channels, tt.categories, tt.languages)
			
			// Handle nil vs empty slice comparison
			if len(result) == 0 && len(tt.expected) == 0 {
				return // Both are empty, test passes
			}
			
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("FilterChannelsByDefaults() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
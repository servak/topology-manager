package grouping

import (
	"testing"
)

func TestGroupByLongestCommonPrefix_Basic(t *testing.T) {
	deviceNames := []string{"access-001", "access-002", "access-003"}
	minGroupSize := 3

	groups := GroupByLongestCommonPrefix(deviceNames, minGroupSize)
	
	t.Logf("Input: %v", deviceNames)
	t.Logf("MinGroupSize: %d", minGroupSize)
	t.Logf("Groups count: %d", len(groups))
	
	for i, group := range groups {
		t.Logf("Group %d: prefix=%s, count=%d, devices=%v", i, group.Prefix, group.Count, group.DeviceIDs)
	}

	if len(groups) == 0 {
		t.Error("Expected at least one group, but got 0")
	}

	if len(groups) > 0 {
		group := groups[0]
		if group.Prefix != "access-" {
			t.Errorf("Expected prefix 'access-', got '%s'", group.Prefix)
		}
		if group.Count != 3 {
			t.Errorf("Expected count 3, got %d", group.Count)
		}
	}
}

func TestGroupByLongestCommonPrefix_NotEnoughDevices(t *testing.T) {
	deviceNames := []string{"access-001", "access-002"}
	minGroupSize := 3

	groups := GroupByLongestCommonPrefix(deviceNames, minGroupSize)
	
	t.Logf("Input: %v", deviceNames)
	t.Logf("MinGroupSize: %d", minGroupSize)
	t.Logf("Groups count: %d", len(groups))

	if len(groups) != 0 {
		t.Errorf("Expected 0 groups when not enough devices, got %d", len(groups))
	}
}

func TestGroupByLongestCommonPrefix_ExactMinimum(t *testing.T) {
	deviceNames := []string{"dist-100", "dist-101", "dist-102"}
	minGroupSize := 3

	groups := GroupByLongestCommonPrefix(deviceNames, minGroupSize)
	
	t.Logf("Input: %v", deviceNames)
	t.Logf("MinGroupSize: %d", minGroupSize)
	t.Logf("Groups count: %d", len(groups))
	
	for i, group := range groups {
		t.Logf("Group %d: prefix=%s, count=%d, devices=%v", i, group.Prefix, group.Count, group.DeviceIDs)
	}

	if len(groups) == 0 {
		t.Error("Expected at least one group when exactly meeting minimum, but got 0")
	}
}

func TestLongestCommonPrefix(t *testing.T) {
	tests := []struct {
		str1     string
		str2     string
		expected string
	}{
		{"access-001", "access-002", "access-00"},
		{"dist-100", "dist-101", "dist-10"},
		{"core-001", "core-002", "core-00"},
		{"different", "other", ""},
		{"same", "same", "same"},
	}

	for _, test := range tests {
		result := longestCommonPrefix(test.str1, test.str2)
		if result != test.expected {
			t.Errorf("longestCommonPrefix(%s, %s) = %s, expected %s", 
				test.str1, test.str2, result, test.expected)
		}
	}
}

func TestFindGroupPrefix(t *testing.T) {
	tests := []struct {
		devices  []string
		expected string
	}{
		{[]string{"access-001", "access-002", "access-003"}, "access-"},
		{[]string{"dist-100", "dist-101", "dist-102"}, "dist-"},
		{[]string{"core-001", "core-002"}, "core-"},
		{[]string{"single"}, "single"},
		{[]string{}, ""},
	}

	for _, test := range tests {
		result := findGroupPrefix(test.devices)
		if result != test.expected {
			t.Errorf("findGroupPrefix(%v) = %s, expected %s", 
				test.devices, result, test.expected)
		}
	}
}
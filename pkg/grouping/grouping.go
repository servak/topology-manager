package grouping

import (
	"sort"
	"strings"
)

// Group represents a group of devices with common prefix
type Group struct {
	Prefix    string   `json:"prefix"`
	Count     int      `json:"count"`
	DeviceIDs []string `json:"device_ids"`
}

// GroupByLongestCommonPrefix groups device names by their longest common prefix
func GroupByLongestCommonPrefix(deviceNames []string, minGroupSize int) []Group {
	if len(deviceNames) < minGroupSize {
		return []Group{}
	}

	// Sort device names for consistent processing
	sortedNames := make([]string, len(deviceNames))
	copy(sortedNames, deviceNames)
	sort.Strings(sortedNames)

	groups := make(map[string][]string)
	processed := make(map[string]bool)

	// Find groups with common prefixes
	for i := 0; i < len(sortedNames); i++ {
		if processed[sortedNames[i]] {
			continue
		}

		currentGroup := []string{sortedNames[i]}
		processed[sortedNames[i]] = true

		// Find devices with common prefix
		for j := i + 1; j < len(sortedNames); j++ {
			if processed[sortedNames[j]] {
				continue
			}

			prefix := longestCommonPrefix(sortedNames[i], sortedNames[j])
			if len(prefix) >= 3 { // Minimum prefix length
				// Check if this device has common prefix with any device in current group
				hasCommonPrefix := false
				for _, groupDevice := range currentGroup {
					if len(longestCommonPrefix(groupDevice, sortedNames[j])) >= 3 {
						hasCommonPrefix = true
						break
					}
				}

				if hasCommonPrefix {
					currentGroup = append(currentGroup, sortedNames[j])
					processed[sortedNames[j]] = true
				}
			}
		}

		// If group has enough members, add it
		if len(currentGroup) >= minGroupSize {
			prefix := findGroupPrefix(currentGroup)
			groups[prefix] = currentGroup
		}
	}

	// Convert map to slice
	result := make([]Group, 0, len(groups))
	for prefix, devices := range groups {
		result = append(result, Group{
			Prefix:    prefix,
			Count:     len(devices),
			DeviceIDs: devices,
		})
	}

	// Sort groups by count (descending)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	return result
}

// longestCommonPrefix finds the longest common prefix between two strings
func longestCommonPrefix(str1, str2 string) string {
	minLen := len(str1)
	if len(str2) < minLen {
		minLen = len(str2)
	}

	commonLength := 0
	for i := 0; i < minLen; i++ {
		if str1[i] == str2[i] {
			commonLength++
		} else {
			break
		}
	}

	return str1[:commonLength]
}

// findGroupPrefix finds the common prefix for a group of devices
func findGroupPrefix(devices []string) string {
	if len(devices) == 0 {
		return ""
	}
	if len(devices) == 1 {
		return devices[0]
	}

	prefix := devices[0]
	for i := 1; i < len(devices); i++ {
		prefix = longestCommonPrefix(prefix, devices[i])
		if len(prefix) == 0 {
			break
		}
	}

	// Ensure prefix ends at a logical boundary (e.g., after dash, dot, underscore)
	if len(prefix) > 0 {
		lastChar := prefix[len(prefix)-1]
		
		// If it already ends with a separator, keep it
		if lastChar == '-' || lastChar == '.' || lastChar == '_' {
			return prefix
		}

		// Find the last logical separator and include it
		for i := len(prefix) - 1; i >= 0; i-- {
			char := prefix[i]
			if char == '-' || char == '.' || char == '_' {
				return prefix[:i+1]
			}
		}
	}

	return prefix
}

// GroupByDepth groups devices by their depth from root
func GroupByDepth(deviceDepths map[string]int, targetDepth int) []Group {
	groups := make(map[int][]string)

	for deviceID, depth := range deviceDepths {
		if depth == targetDepth {
			groups[depth] = append(groups[depth], deviceID)
		}
	}

	result := make([]Group, 0, len(groups))
	for depth, devices := range groups {
		if len(devices) > 1 {
			result = append(result, Group{
				Prefix:    "Depth-" + string(rune(depth+'0')),
				Count:     len(devices),
				DeviceIDs: devices,
			})
		}
	}

	return result
}

// GroupByType groups devices by their type
func GroupByType(deviceTypes map[string]string) []Group {
	groups := make(map[string][]string)

	for deviceID, deviceType := range deviceTypes {
		groups[deviceType] = append(groups[deviceType], deviceID)
	}

	result := make([]Group, 0, len(groups))
	for deviceType, devices := range groups {
		if len(devices) > 1 {
			result = append(result, Group{
				Prefix:    strings.Title(deviceType),
				Count:     len(devices),
				DeviceIDs: devices,
			})
		}
	}

	// Sort by count
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	return result
}
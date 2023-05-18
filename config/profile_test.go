package config

import (
	"reflect"
	"testing"
)

func TestCreateToSearchPaths(t *testing.T) {
	profileDir := "test_profile_dir"
	cfg := Config{
		CurrentProfile: "cfg_profile.yaml",
	}

	testCases := []struct {
		name     string
		overload string
		expected []string
	}{
		{
			name:     "Overload is empty",
			overload: "",
			expected: []string{profileDir + "/cfg_profile.yaml"},
		},
		{
			name:     "Overload is a relative path without extension",
			overload: "custom_profile",
			expected: []string{"custom_profile", profileDir + "/custom_profile.yaml", profileDir + "/custom_profile.yml"},
		},
		{
			name:     "Overload is a relative path with extension",
			overload: "custom_profile.yaml",
			expected: []string{"custom_profile.yaml", profileDir + "/custom_profile.yaml"},
		},
		{
			name:     "Overload is current directory",
			overload: "./custom_profile.yaml",
			expected: []string{"./custom_profile.yaml"},
		},
		{
			name:     "Overload is an absolute path",
			overload: "/absolute/custom_profile.yaml",
			expected: []string{"/absolute/custom_profile.yaml"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := createToSearchPaths(profileDir, cfg, tc.overload)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, but got %v", tc.expected, result)
			}
		})
	}
}

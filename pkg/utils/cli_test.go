package utils

import (
	"testing"

	"github.com/urfave/cli/v2"
)

func TestStringFlag(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		value    string
		usage    string
		aliases  []string
		expected *cli.StringFlag
	}{
		{
			name:     "String flag with aliases",
			flagName: "host",
			value:    "localhost",
			usage:    "Host to listen on",
			aliases:  []string{"H"},
			expected: &cli.StringFlag{
				Name:    "host",
				Aliases: []string{"H"},
				Value:   "localhost",
				Usage:   "Host to listen on",
			},
		},
		{
			name:     "String flag without aliases",
			flagName: "config",
			value:    "",
			usage:    "Path to config file",
			aliases:  nil,
			expected: &cli.StringFlag{
				Name:    "config",
				Aliases: nil,
				Value:   "",
				Usage:   "Path to config file",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringFlag(tt.flagName, tt.value, tt.usage, tt.aliases...)
			
			if result.Name != tt.expected.Name {
				t.Errorf("Expected name %s, got %s", tt.expected.Name, result.Name)
			}
			if result.Value != tt.expected.Value {
				t.Errorf("Expected value %s, got %s", tt.expected.Value, result.Value)
			}
			if result.Usage != tt.expected.Usage {
				t.Errorf("Expected usage %s, got %s", tt.expected.Usage, result.Usage)
			}
			if len(result.Aliases) != len(tt.expected.Aliases) {
				t.Errorf("Expected %d aliases, got %d", len(tt.expected.Aliases), len(result.Aliases))
			}
			for i, alias := range result.Aliases {
				if alias != tt.expected.Aliases[i] {
					t.Errorf("Expected alias %s, got %s", tt.expected.Aliases[i], alias)
				}
			}
		})
	}
}

func TestBoolFlag(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		usage    string
		aliases  []string
		expected *cli.BoolFlag
	}{
		{
			name:     "Bool flag with aliases",
			flagName: "public",
			usage:    "Open server to public",
			aliases:  []string{"P"},
			expected: &cli.BoolFlag{
				Name:    "public",
				Aliases: []string{"P"},
				Usage:   "Open server to public",
			},
		},
		{
			name:     "Bool flag without aliases",
			flagName: "tls",
			usage:    "Enable TLS",
			aliases:  nil,
			expected: &cli.BoolFlag{
				Name:    "tls",
				Aliases: nil,
				Usage:   "Enable TLS",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BoolFlag(tt.flagName, tt.usage, tt.aliases...)
			
			if result.Name != tt.expected.Name {
				t.Errorf("Expected name %s, got %s", tt.expected.Name, result.Name)
			}
			if result.Usage != tt.expected.Usage {
				t.Errorf("Expected usage %s, got %s", tt.expected.Usage, result.Usage)
			}
			if len(result.Aliases) != len(tt.expected.Aliases) {
				t.Errorf("Expected %d aliases, got %d", len(tt.expected.Aliases), len(result.Aliases))
			}
		})
	}
}

func TestConfigFlag(t *testing.T) {
	result := ConfigFlag()
	
	expected := &cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Value:   "",
		Usage:   "Path to config file",
	}
	
	if result.Name != expected.Name {
		t.Errorf("Expected name %s, got %s", expected.Name, result.Name)
	}
	if result.Value != expected.Value {
		t.Errorf("Expected value %s, got %s", expected.Value, result.Value)
	}
	if result.Usage != expected.Usage {
		t.Errorf("Expected usage %s, got %s", expected.Usage, result.Usage)
	}
	if len(result.Aliases) != 1 || result.Aliases[0] != "c" {
		t.Errorf("Expected aliases [c], got %v", result.Aliases)
	}
}

func TestVersionFlag(t *testing.T) {
	result := VersionFlag()
	
	if result.Name != "version" {
		t.Errorf("Expected name version, got %s", result.Name)
	}
	if len(result.Aliases) != 1 || result.Aliases[0] != "v" {
		t.Errorf("Expected aliases [v], got %v", result.Aliases)
	}
}

func TestCommonServerFlags(t *testing.T) {
	flags := CommonServerFlags()
	
	expectedFlagNames := []string{"host", "port", "public", "tls", "tls-cert", "tls-key"}
	
	if len(flags) != len(expectedFlagNames) {
		t.Errorf("Expected %d flags, got %d", len(expectedFlagNames), len(flags))
	}
	
	for i, flag := range flags {
		var name string
		switch f := flag.(type) {
		case *cli.StringFlag:
			name = f.Name
		case *cli.BoolFlag:
			name = f.Name
		}
		
		if name != expectedFlagNames[i] {
			t.Errorf("Expected flag %s at position %d, got %s", expectedFlagNames[i], i, name)
		}
	}
}

func TestNewCommand(t *testing.T) {
	config := CommandConfig{
		Name:        "test",
		Aliases:     []string{"t"},
		Usage:       "Test command",
		Description: "A test command",
		Action:      func(c *cli.Context) error { return nil },
		Flags:       []cli.Flag{StringFlag("test-flag", "default", "Test flag")},
	}
	
	result := NewCommand(config)
	
	if result.Name != config.Name {
		t.Errorf("Expected name %s, got %s", config.Name, result.Name)
	}
	if len(result.Aliases) != len(config.Aliases) {
		t.Errorf("Expected %d aliases, got %d", len(config.Aliases), len(result.Aliases))
	}
	if result.Usage != config.Usage {
		t.Errorf("Expected usage %s, got %s", config.Usage, result.Usage)
	}
	if result.Description != config.Description {
		t.Errorf("Expected description %s, got %s", config.Description, result.Description)
	}
	if len(result.Flags) != len(config.Flags) {
		t.Errorf("Expected %d flags, got %d", len(config.Flags), len(result.Flags))
	}
}
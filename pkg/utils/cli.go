package utils

import (
	"github.com/urfave/cli/v2"
)

// CLI flag utility functions to reduce repetition

// StringFlag creates a standard string flag with common properties
func StringFlag(name, value, usage string, aliases ...string) *cli.StringFlag {
	return &cli.StringFlag{
		Name:    name,
		Aliases: aliases,
		Value:   value,
		Usage:   usage,
	}
}

// BoolFlag creates a standard bool flag with common properties
func BoolFlag(name, usage string, aliases ...string) *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:    name,
		Aliases: aliases,
		Usage:   usage,
	}
}

// ConfigFlag creates a standardized config flag
func ConfigFlag() *cli.StringFlag {
	return StringFlag("config", "", "Path to config file", "c")
}

// VersionFlag creates a standardized version flag
func VersionFlag() *cli.StringFlag {
	return StringFlag("version", "", "Update to a custom specific version that is not latest", "v")
}

// CommonServerFlags returns common server-related flags
func CommonServerFlags() []cli.Flag {
	return []cli.Flag{
		StringFlag("host", "localhost", "Host to listen on", "H"),
		StringFlag("port", "5001", "Port to listen on", "p"),
		BoolFlag("public", "Open server to public. This will expose your server outside your local network. Equivalent to passing --host [::]", "P"),
		BoolFlag("tls", "Enable TLS. This will enable HTTPS for the server.", "https"),
		StringFlag("tls-cert", "", "Path to TLS certificate file", "cert"),
		StringFlag("tls-key", "", "Path to TLS key file", "cert-key"),
	}
}

// Command creates a standardized command structure
type CommandConfig struct {
	Name        string
	Aliases     []string
	Usage       string
	Description string
	Action      cli.ActionFunc
	Flags       []cli.Flag
	Subcommands []*cli.Command
}

// NewCommand creates a new CLI command with standardized structure
func NewCommand(config CommandConfig) *cli.Command {
	return &cli.Command{
		Name:        config.Name,
		Aliases:     config.Aliases,
		Usage:       config.Usage,
		Description: config.Description,
		Action:      config.Action,
		Flags:       config.Flags,
		Subcommands: config.Subcommands,
	}
}
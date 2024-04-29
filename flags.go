package main

import (
	"fmt"

	"github.com/alecthomas/kong"
)

type VersionFlag bool

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(appName + ": " + appVersion)
	app.Exit(0)
	return nil
}

// Flags supplied to cli.
type Flags struct {
	Config      string      `help:"Location of config file" type:"existingfile"`
	IPAddresses []string    `help:"List of IP addresses to find."`
	DNSBLFiles  []string    `help:"List of DNSBL files to scan." type:"path"`
	Version     VersionFlag `name:"version" help:"Print version information and quit"`
}

// Parse the supplied flags.
func (a *App) ParseFlags() *kong.Context {
	app.flags = &Flags{}

	ctx := kong.Parse(app.flags,
		kong.Name(appName),
		kong.Description(appDescription),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)
	return ctx
}

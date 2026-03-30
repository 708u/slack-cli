// Package tz holds the process-wide timezone for formatting and
// parsing. A package-level variable is used instead of dependency
// injection because the location is set once at startup and then
// read from dozens of call sites across util, format, and cmd
// packages -- threading it through every function signature would
// add significant noise for no practical benefit in a CLI tool.
package tz

import "time"

var loc = time.Local

// Set sets the timezone used for all formatting and parsing.
func Set(l *time.Location) { loc = l }

// Location returns the current timezone.
func Location() *time.Location { return loc }

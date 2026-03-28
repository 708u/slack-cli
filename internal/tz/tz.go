package tz

import "time"

var loc = time.Local

// Set sets the timezone used for all formatting and parsing.
func Set(l *time.Location) { loc = l }

// Location returns the current timezone.
func Location() *time.Location { return loc }

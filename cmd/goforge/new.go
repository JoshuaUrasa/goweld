package main

import "github.com/JoshuaUrasa/goforge/internal/scaffold"

// New creates a project using the default standard-library HTTP template.
func New(name string) error {
	return scaffold.Create(".", scaffold.Options{Name: name, Framework: "std", Database: "none"})
}

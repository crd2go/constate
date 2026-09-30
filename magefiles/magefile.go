//go:build mage

//mage:multiline

// Set the general description you want to have displayed with mage -l here.
package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Run all CI steps: build, test, coverage, and vulnerability check.
func CI() {
	mg.Deps(Build, Test, Cover, Vulncheck)
}

// Compile all packages in the module.
func Build() error {
	return sh.Run("go", "build", "./...")
}

// Run all tests, writing a coverage profile to coverage.out.
func Test() error {
	return sh.Run("go", "test", "-coverprofile=coverage.out", "./...")
}

// Report test coverage.
func Cover() error {
	mg.Deps(Test)
	return sh.Run("go", "tool", "cover", "-func=coverage.out")
}

// Scan the project for vulnerable dependencies.
func Vulncheck() error {
	return sh.Run("go", "tool", "govulncheck", "./...")
}

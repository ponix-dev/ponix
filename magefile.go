//go:build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Stack mg.Namespace

// starts dependencies for stack
func (Stack) Up() error {
	err := sh.Run("docker", "compose", "-f", "docker-compose.yaml", "up", "-d")
	if err != nil {
		return err
	}

	return nil
}

// stops dependencies for stack
func (Stack) Down() error {
	err := sh.Run("docker", "compose", "-f", "docker-compose.yaml", "down")
	if err != nil {
		return err
	}

	return nil
}

type DB mg.Namespace

// generates database code
func (DB) Gen() error {
	err := sh.Run("sqlc", "generate")
	if err != nil {
		return err
	}

	return nil
}

// generates a database migration with a given name
func (DB) Migrate(migration string) error {
	err := sh.Run("atlas", "migrate", "diff", migration, "--dir", "file://internal/postgres/atlas", "--to", "file://schema/schema.sql", "--dev-url", "docker://postgres?search_path=public")
	if err != nil {
		return err
	}

	return nil
}

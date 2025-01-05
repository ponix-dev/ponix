//go:build mage

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	baseDir = mageDir()
)

func mageDir() string {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		panic("wtf no base dir?")
	}

	return dir
}

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
	err := sh.Run("docker", "run", "--rm", "-v", baseDir+":/src", "-w", "/src", "sqlc/sqlc", "generate")
	if err != nil {
		return err
	}

	return nil
}

// generates a database migration with a given name
func (DB) Migrate(migration string) error {
	err := sh.Run("atlas", "migrate", "diff", migration, "--dir", "file://internal/postgres/atlas", "--to", "file://schema/schema.sql", "--dev-url", "docker://postgres/15/dev?search_path=public")
	if err != nil {
		return err
	}

	return nil
}

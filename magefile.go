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
func (DB) Migrate(db string, migration string) error {
	err := sh.Run("goose", "create", migration, "sql", "-dir", fmt.Sprintf("internal/%s/goose", db))
	if err != nil {
		return err
	}

	return nil
}

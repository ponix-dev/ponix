//go:build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Stack mg.Namespace

func (Stack) Up() error {
	err := sh.Run("docker", "compose", "-f", "docker-compose.yaml", "up", "-d")
	if err != nil {
		return err
	}

	return nil
}

func (Stack) Down() error {
	err := sh.Run("docker", "compose", "-f", "docker-compose.yaml", "down")
	if err != nil {
		return err
	}

	return nil
}

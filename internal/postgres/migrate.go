package postgres

import (
	"context"
	"embed"
	"log/slog"
	"path/filepath"

	"ariga.io/atlas-go-sdk/atlasexec"
)

//go:embed atlas/*.sql
//go:embed atlas/*.sum
var migrations embed.FS

func migrationsPath() string {
	return "file://" + filepath.Join("migrations", "atlas")
}

func RunMigrations(ctx context.Context, connUrl ConnUrl) error {
	workdir, err := atlasexec.NewWorkingDir(
		atlasexec.WithMigrations(
			migrations,
		),
	)
	if err != nil {
		return err
	}

	defer workdir.Close()

	// Initialize the client.
	client, err := atlasexec.NewClient(workdir.Path(), "atlas")
	if err != nil {
		return err
	}

	slog.Default().Info("applying migrations...")

	resp, err := client.MigrateApply(ctx, &atlasexec.MigrateApplyParams{
		DirURL: migrationsPath(),
		URL:    string(connUrl),
	})
	if err != nil {
		return err
	}

	for _, applied := range resp.Applied {
		slog.Default().Info("migration applied", slog.String("name", applied.Name), slog.Time("start", applied.Start), slog.Time("end", applied.End))
	}

	slog.Default().Info("migrations applied", slog.Int("count", len(resp.Applied)))

	return nil
}

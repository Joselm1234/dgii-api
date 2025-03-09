package cmd

import (
	"context"
	"fmt"
	"my-dgii-api/contributors"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

var Migrations = migrate.NewMigrations()

func init() {
    Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
        // Crear la tabla contributors
        _, err := db.NewCreateTable().Model((*contributors.Contributor)(nil)).IfNotExists().Exec(ctx)
        if err != nil {
            return fmt.Errorf("error creating contributors table: %w", err)
        }
        return nil
    }, func(ctx context.Context, db *bun.DB) error {
        // Eliminar la tabla contributors
        _, err := db.NewDropTable().Model((*contributors.Contributor)(nil)).IfExists().Exec(ctx)
        if err != nil {
            return fmt.Errorf("error dropping contributors table: %w", err)
        }
        return nil
    })
}

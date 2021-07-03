package main

import (
	"embed"
	"fmt"
	"github.com/graaphscom/dbmigrat"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
)

//go:embed auth/migrations
var auth embed.FS

//go:embed billing/migrations
var billing embed.FS

//go:embed inventory/migrations
var inventory embed.FS

func main() {
	db, err := sqlx.Open("postgres", "postgres://dbmigrat:dbmigrat@localhost:5432/dbmigrat?sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}

	pgStore := &dbmigrat.PostgresStore{DB: db}
	err = pgStore.CreateLogTable()
	if err != nil {
		log.Fatalln(err)
	}

	authMigrations, err := dbmigrat.ReadDir(auth, "auth/migrations")
	if err != nil {
		log.Fatalln(err)
	}
	billingMigrations, err := dbmigrat.ReadDir(billing, "billing/migrations")
	if err != nil {
		log.Fatalln(err)
	}
	inventoryMigrations, err := dbmigrat.ReadDir(inventory, "inventory/migrations")
	if err != nil {
		log.Fatalln(err)
	}
	migrations := dbmigrat.Migrations{
		"auth":      authMigrations,
		"billing":   billingMigrations,
		"inventory": inventoryMigrations,
	}

	//logsCount, err := dbmigrat.Migrate(pgStore, migrations, dbmigrat.RepoOrder{"auth", "inventory", "billing"})
	logsCount, err := dbmigrat.Rollback(pgStore, migrations, dbmigrat.RepoOrder{"billing", "inventory", "auth"}, -1)
	if err != nil {
		log.Fatalln(err)
	}
	//fmt.Printf("[dbmigrat] applied %d migrations\n", logsCount)
	fmt.Printf("[dbmigrat] rolled back %d migrations\n", logsCount)
}

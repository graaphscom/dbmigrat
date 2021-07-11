[![CI](https://github.com/graaphscom/dbmigrat/actions/workflows/ci.yml/badge.svg)](https://github.com/graaphscom/dbmigrat/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/graaphscom/dbmigrat/branch/master/graph/badge.svg?token=T8UAQ93Z3T)](https://codecov.io/gh/graaphscom/dbmigrat)
[![Go Reference](https://pkg.go.dev/badge/github.com/graaphscom/dbmigrat.svg)](https://pkg.go.dev/github.com/graaphscom/dbmigrat)

# dbmigrat
Library for maintaining database migrations in several locations (repos).

## Installation
```
go get github.com/graaphscom/dbmigrat
```
## Example code
- [example test](example_test.go)
- [example ecommerce app](internal/docs/ecommerceapp)
## Usage
Let's assume we're building a simple e-commerce app.
It's split into the following modules:
- auth (responsible for managing users and access)
- billing (contains logic related to placing order and payment)
- inventory (responsible for managing available products)

Each of the above modules needs to store data in a database.
The below diagram represents tables we need to have at the first iteration of our project:

![](./internal/docs/er_diagram.png)

That's how our project's directory layout looks like:  
```
|-- ecommerceapp
|   |-- auth
|   |   `-- migrations
|   |       |-- 0.init.down.sql
|   |       `-- 0.init.up.sql
|   |-- billing
|   |   `-- migrations
|   |       |-- 0.init.down.sql
|   |       `-- 0.init.up.sql
|   |-- inventory
|   |   `-- migrations
|   |       |-- 0.init.down.sql
|   |       `-- 0.init.up.sql
|   `-- main.go
```
We have separate `0.init.up.sql` files containing SQL create table statements for each module.

Now let's build `main.go` file:

1. Add required imports and load migrations directories (`go:embed` is available since go 1.16)
```
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

```
2. Connect to the database, create migrations log table
(`CreateLogTable` does nothing when the table already exists):
```diff
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

+ func main() {
+ 	db, err := sqlx.Open("postgres", "postgres://dbmigrat:dbmigrat@localhost:5432/dbmigrat?sslmode=disable")
+ 	if err != nil {
+ 		log.Fatalln(err)
+ 	}
+ 
+ 	pgStore := &dbmigrat.PostgresStore{DB: db}
+ 	err = pgStore.CreateLogTable()
+ 	if err != nil {
+ 		log.Fatalln(err)
+ 	}
+
```
3. Read migrations from filesystems and build migrations map:
```diff
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

+ 	authMigrations, err := dbmigrat.ReadDir(auth, "auth/migrations")
+ 	if err != nil {
+ 		log.Fatalln(err)
+ 	}
+ 	billingMigrations, err := dbmigrat.ReadDir(billing, "billing/migrations")
+ 	if err != nil {
+ 		log.Fatalln(err)
+ 	}
+ 	inventoryMigrations, err := dbmigrat.ReadDir(inventory, "inventory/migrations")
+ 	if err != nil {
+ 		log.Fatalln(err)
+ 	}
+ 	migrations := dbmigrat.Migrations{
+ 		"auth":    authMigrations,
+ 		"billing": billingMigrations,
+ 		"inventory": inventoryMigrations,
+ 	}
+ 
```
4. Apply migrations. Notice that `billing` module has references to
`auth` and `inventory`. For that reason third argument to `dbmigrat.Migrate`
is `dbmigrat.RepoOrder{"auth", "inventory", "billing"}`
```diff
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
		"auth":    authMigrations,
		"billing": billingMigrations,
		"inventory": inventoryMigrations,
	}

+ 	logsCount, err := dbmigrat.Migrate(pgStore, migrations, dbmigrat.RepoOrder{"auth", "inventory", "billing"})
+ 	if err != nil {
+ 		log.Fatalln(err)
+ 	}
+ 	fmt.Printf("[dbmigrat] applied %d migrations\n", logsCount)
+ }
```

After running `main.go` three migrations should have been applied.
Now, `dbmigrat_log` table has these rows:
![](./internal/docs/migrate_serial_0.png)
Notice that every migration has `migration_serial` equal to zero.
That's because all these migrations were applied by a single run.

### Let's add product's description
1. create migrations files with SQL alter table statements: 
```diff
  |-- ecommerceapp
  |   |-- auth
  |   |   `-- migrations
  |   |       |-- 0.init.down.sql
  |   |       `-- 0.init.up.sql
  |   |-- billing
  |   |   `-- migrations
  |   |       |-- 0.init.down.sql
  |   |       `-- 0.init.up.sql
  |   |-- inventory
  |   |   `-- migrations
  |   |       |-- 0.init.down.sql
  |   |       |-- 0.init.up.sql
+ |   |       |-- 1.product_description.down.sql
+ |   |       `-- 1.product_description.up.sql
  |   `-- main.go
```
2. run the `main.go` which we have already built

Now, `dbmigrat_log` table has additional fourth row:
![](./internal/docs/migrate_serial_1.png)
Notice that this time migration serial is equal to one.

### Rollback
Rollback allows for reverting migrations to a specified migration serial.

Let's roll back the addition of the product's description. Edit the `main.go`:
```diff
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

-	logsCount, err := dbmigrat.Migrate(pgStore, migrations, dbmigrat.RepoOrder{"auth", "inventory", "billing"})
+	logsCount, err := dbmigrat.Rollback(pgStore, migrations, dbmigrat.RepoOrder{"billing", "inventory", "auth"}, 0)
	if err != nil {
		log.Fatalln(err)
	}
-	fmt.Printf("[dbmigrat] applied %d migrations\n", logsCount)
+	fmt.Printf("[dbmigrat] rolled back %d migrations\n", logsCount)
}
```
Notice that comparing to the `Migrate` function, `RepoOrder` argument is reversed.

If we would like to roll back all migrations, we would provide `-1` as the last argument to the `Rollback`.

## Credits
ER diagram built with https://staruml.io
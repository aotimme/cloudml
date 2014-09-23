package db

import (
  "os"
  "log"
  "database/sql"
  "github.com/coopernurse/gorp"
  _ "github.com/lib/pq"
)

var DATA_DIR="/tmp/cloudml"
var DBMAP *gorp.DbMap

func init() {
  if _, err := os.Stat(DATA_DIR); err != nil {
    err := os.Mkdir(DATA_DIR, 0755)
    if err != nil {
      log.Fatal(err)
    }
  }
  DBMAP = initDb()
}

func initDb() *gorp.DbMap {
  // connect to db using standard Go database/sql API
  // use whatever database/sql driver you wish
  db, err := sql.Open("postgres", "dbname=cloudml sslmode=disable")
  if err != nil {
    log.Fatal(err)
  }

  // construct a gorp DbMap
  dbmap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}

  // add a table, setting the table name to 'posts' and
  // specifying that the Id property is an auto incrementing PK
  dbmap.AddTableWithName(Model{}, "models").SetKeys(false, "Id")
  dbmap.AddTableWithName(Coefficient{}, "coefficients").SetKeys(false, "Id")
  dbmap.AddTableWithName(Covariate{}, "covariates").SetKeys(false, "Id")
  dbmap.AddTableWithName(Datum{}, "data").SetKeys(false, "Id")

  // create the table. in a production system you'd generally
  // use a migration tool, or create the tables via scripts
  err = dbmap.CreateTablesIfNotExists()
  if err != nil {
    log.Fatal(err)
  }

  return dbmap
}

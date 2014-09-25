package db

import (
  "os"
  "log"
  "database/sql"
  "github.com/coopernurse/gorp"
  _ "github.com/lib/pq"
)

var DBMAP *gorp.DbMap

func init() {
  DBMAP = initDb()
}

func initDb() *gorp.DbMap {
  user := os.Getenv("USER")
  password := os.Getenv("PASS")
  dbname := os.Getenv("DBNAME")
  sqlOptionsString := "sslmode=disable"
  if dbname != "" {
    sqlOptionsString += " dbname=" + dbname
  }
  if user != "" {
    sqlOptionsString += " user=" + user
  }
  if password != "" {
    sqlOptionsString += " password=" + password
  }
  // connect to db using standard Go database/sql API
  // use whatever database/sql driver you wish
  db, err := sql.Open("postgres", sqlOptionsString)
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

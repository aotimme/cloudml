package db

import (
  "os"
  "log"
)

var DATA_DIR="/tmp/cloudml"

func init() {
  if _, err := os.Stat(DATA_DIR); err != nil {
    err := os.Mkdir(DATA_DIR, 0755)
    if err != nil {
      log.Fatal(err)
    }
  }
}

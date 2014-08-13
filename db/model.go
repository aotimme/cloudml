package db

import (
  "path"
  "encoding/json"
  "io/ioutil"
  "log"
  "os"
)

func (m *Model) getDirectoryName() string {
  return path.Join(DATA_DIR, m.Id)
}
func (m *Model) getDataDirectoryName() string {
  return path.Join(DATA_DIR, m.Id, "data")
}

func (m *Model) getFileName() string {
  return path.Join(DATA_DIR, m.Id, "model.json")
}

func GetModelFromJSON(jsonData []byte) (*Model, error) {
  var m Model
  err := json.Unmarshal(jsonData, &m)
  if err != nil {
    log.Printf("Error unmarshaling json in `GetModelFromJSON`\n")
    return nil, err
  }
  return &m, nil
}

func GetModelById(id string) (*Model, error) {
  filename := path.Join(DATA_DIR, id, "model.json")
  jsonData, err := ioutil.ReadFile(filename)
  if err != nil {
    log.Printf("Error getting model for id %v\n", id)
    return nil, err
  }
  return GetModelFromJSON(jsonData)
}

func (m *Model) Save() error {
  if m.Id == "" {
    id, err := newUUID()
    if err == nil {
      m.Id = id
    } else {
      return err
    }
  }
  if _, err := os.Stat(m.getDirectoryName()); err != nil {
    err := os.Mkdir(m.getDirectoryName(), 0755)
    if err != nil {
      return err
    }
  }
  if _, err := os.Stat(m.getDataDirectoryName()); err != nil {
    err := os.Mkdir(m.getDataDirectoryName(), 0755)
    if err != nil {
      return err
    }
  }
  jsonData, err := json.Marshal(m)
  if err != nil {
    return err
  }
  return ioutil.WriteFile(m.getFileName(), jsonData, 0644)
}

func (m *Model) Delete() error {
  return os.RemoveAll(m.getDirectoryName())
}

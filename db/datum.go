package db

import (
  "os"
  "io/ioutil"
  "encoding/json"
  "fmt"
  "log"
  "path"
)

func (d *Datum) getFileName() string {
  return path.Join(DATA_DIR, d.Model, "data", fmt.Sprintf("%v.json", d.Id))
}

func (m *Model) CreateDatum(covariates map[string]float64, value float64) (*Datum, error) {
  id, err := newUUID()
  if err != nil {
    return nil, err
  }
  d := &Datum{
    Id: id,
    Model: m.Id,
    Covariates: covariates,
    Value: value,
  }
  return d, nil
}

func (m *Model) GetData() ([]*Datum, error) {
  dir, err := os.Open(m.getDataDirectoryName())
  if err != nil {
    return nil, err
  }
  files, err := dir.Readdir(0)
  if err != nil {
    return nil, err
  }
  data := make([]*Datum, len(files))
  for i, file := range files {
    d, err := GetDatumFromFile(path.Join(dir.Name(), file.Name()))
    if err != nil {
      return nil, err
    }
    data[i] = d
  }
  return data, nil
}

func GetDatumFromJSON(jsonData []byte) (d *Datum, err error) {
  err = json.Unmarshal(jsonData, &d)
  if err != nil {
    log.Printf("Error unmarshaling json in `GetModelFromJSON`\n")
  }
  return
}

func GetDatumFromFile(filename string) (*Datum, error) {
  jsonData, err := ioutil.ReadFile(filename)
  if err != nil {
    return nil, err
  }
  return GetDatumFromJSON(jsonData)
}

func (d *Datum) Save() error {
  jsonData, err := json.Marshal(d)
  if err != nil {
    return err
  }
  return ioutil.WriteFile(d.getFileName(), jsonData, 0644)
}

func (m *Model) CreateAndSaveDatum(covariates map[string]float64, value float64) (*Datum, error) {
  d, err := m.CreateDatum(covariates, value)
  if err != nil {
    return nil, err
  }
  err = d.Save()
  if err != nil {
    return nil, err
  }
  m.N++
  err = m.Save()
  if err != nil {
    return nil, err
  }
  return d, nil
}

func (m *Model) DeleteData() error {
  err := os.RemoveAll(m.getDataDirectoryName())
  if err != nil {
    return err
  }
  for key, _ := range m.Coefficients {
    m.Coefficients[key] = 0.0
  }
  m.N = 0
  return m.Save()
}

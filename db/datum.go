package db

import (
  "log"
)

func (m *Model) CreateDatum(covMap map[string]float64, value float64) (*Datum, error) {
  datumId, err := newUUID()
  if err != nil {
    log.Printf("Error creating UUID: %v\n", err)
    return nil, err
  }
  coefficients, err := m.GetCoefficients()
  if err != nil {
    return nil, err
  }
  covariates := make([]Covariate, len(coefficients))
  for i, coefficient := range coefficients {
    covId, err := newUUID()
    if err != nil {
      log.Printf("Error creating UUID: %v\n", err)
      return nil, err
    }
    covariates[i].Label = coefficient.Label
    covariates[i].Value = covMap[coefficient.Label]
    covariates[i].Datum = datumId
    covariates[i].Id = covId
  }
  d := &Datum{
    Id: datumId,
    Value: value,
    Model: m.Id,
  }
  m.NumTrainingData++
  txn, err := DBMAP.Begin()
  if err != nil {
    return nil, err
  }
  txn.Insert(d)
  txn.Update(m)
  for _, covariate := range covariates {
    txn.Insert(&covariate)
  }
  err = txn.Commit()
  if err != nil {
    return nil, err
  }
  return d, nil
}

func (m *Model) GetData() ([]*Datum, error) {
  var data []*Datum
  _, err := DBMAP.Select(&data, "select * from data where model=:model", map[string]interface{} {"model": m.Id})
  if err != nil {
    return nil, err
  }
  return data, nil
}

func (d *Datum) GetCovariates() ([]Covariate, error) {
  var covariates []Covariate
  _, err := DBMAP.Select(&covariates, "select * from covariates where datum=:datum order by label", map[string]interface{} {"datum": d.Id})
  if err != nil {
    return nil, err
  }
  return covariates, nil
}
func GetDatumById(id string) (*Datum, error) {
  obj, err := DBMAP.Get(Datum{}, id)
  if err != nil {
    log.Fatal("Error getting datum", err)
    return nil, err
  }
  if obj == nil {
    return nil, nil
  }
  datum := obj.(*Datum)
  return datum, nil
}

func (m *Model) DeleteData() error {
  var dataIds []string
  _, err := DBMAP.Select(&dataIds, "select id from data where model=:model", map[string]interface{}{"model": m.Id})
  if err != nil {
    return err
  }
  txn, err := DBMAP.Begin()
  if err != nil {
    return err
  }
  for _, datumId := range dataIds {
    txn.Exec("delete from covariates where datum=:datum",  map[string]interface{} {"datum": datumId})
  }
  txn.Exec("delete from data where model=:model", map[string]interface{} {"model": m.Id})
  err = txn.Commit()
  if err != nil {
    return err
  }

  m.NumTrainingData = 0
  coefficients, err := m.GetCoefficients()
  if err != nil {
    return err
  }
  for _, coefficient := range coefficients {
    coefficient.Value = 0.0
  }
  return m.SaveWithCoefficients(coefficients)
}

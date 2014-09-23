package db

import (
  "log"
  "errors"
)

func GetAllModelIds() ([]string, error) {
  var modelIds []string
  _, err := DBMAP.Select(&modelIds, "select id from models")
  if err != nil {
    return nil, err
  }
  return modelIds, nil
}

//func GetAllModels() ([]Model, error) {
//  modelIds, err := GetAllModelIds()
//  if err != nil {
//    return nil, err
//  }
//  models := make([]Model, len(modelIds))
//  for i, modelId := range modelIds {
//    models[i], err = GetModelById(modelId)
//    if err != nil {
//      return nil, err
//    }
//  }
//  return models, nil
//}

func GetModelById(id string) (*Model, error) {
  //filename := path.Join(DATA_DIR, id, "model.json")
  obj, err := DBMAP.Get(Model{}, id)
  if err != nil {
    log.Fatal("Error getting model", err)
    return nil, err
  }
  if obj == nil {
    return nil, nil
  }
  model := obj.(*Model)
  return model, nil
}

func GetModelAndCoefficientsById(id string) (*Model, []Coefficient, error) {
  model, err := GetModelById(id)
  if err != nil {
    log.Fatalln("Error getting model", err)
    return nil, nil, err
  }
  if model == nil {
    return nil, nil, nil
  }
  coefficients, err := model.GetCoefficients()
  if err != nil {
    return nil, nil, err
  }
  return model, coefficients, nil
}

func (m *Model) GetCoefficients() ([]Coefficient, error) {
  var coefficients []Coefficient
  _, err := DBMAP.Select(&coefficients, "select * from coefficients where model=:model order by label", map[string]interface{} { "model": m.Id})
  if err != nil {
    return nil, err
  }
  return coefficients, nil
}

func (model *Model) Update() error {
  if model.Id == "" {
    return errors.New("Cannot update model without id")
  }
  _, err := DBMAP.Update(model)
  return err
}

func (model *Model) SaveWithCoefficients(coefficients []Coefficient) error {
  isNew := model.Id == ""
  if isNew {
    modelId, err := newUUID()
    if err != nil {
      log.Fatal("UUID error", err)
      return err
    }
    model.Id = modelId
    for i, _ := range coefficients {
      id, err := newUUID()
      if err != nil {
        log.Fatal("UUID error", err)
        return err
      }
      coefficients[i].Id = id
      coefficients[i].Model = modelId
    }
  }
  model.NumCovariates = len(coefficients)
  txn, err := DBMAP.Begin()
  if err != nil {
    log.Fatal("Transaction Begin error", err)
    return err
  }
  if isNew {
    txn.Insert(model)
    for _, coefficient := range coefficients {
      txn.Insert(&coefficient)
    }
  } else {
    txn.Update(model)
    for _, coefficient := range coefficients {
      txn.Update(&coefficient)
    }
  }
  err = txn.Commit()
  if err != nil {
    log.Fatal("failed insert", err)
    return err
  }
  return nil
}

func DeleteModelById(modelId string) error {
  paramMap := map[string]interface{} {"model": modelId}

  var dataIds []string
  _, err := DBMAP.Select(&dataIds, "select id from data where model=:model", paramMap)
  if err != nil {
    return err
  }

  txn, err := DBMAP.Begin()
  if err != nil {
    return err
  }
  for _, datumId := range dataIds {
    _, err = txn.Exec("delete from covariates where datum=$1",  datumId)
    if err != nil {
      log.Printf("err on delete covariates (datum = %v): %v\n", datumId, err)
      return txn.Rollback()
    }
  }
  _, err = txn.Exec("delete from data where model=$1", modelId)
  if err != nil {
    log.Printf("err on delete data: %v\n", err)
    return txn.Rollback()
  }
  _, err = txn.Exec("delete from coefficients where model=$1", modelId)
  if err != nil {
    log.Printf("err on delete coefficients: %v\n", err)
    return txn.Rollback()
  }
  _, err = txn.Exec("delete from models where id=$1", modelId)
  if err != nil {
    log.Printf("err on delete model: %v\n", err)
    return txn.Rollback()
  }
  return txn.Commit()
}
//func (m *Model) Delete() error {
//  return os.RemoveAll(m.getDirectoryName())
//}

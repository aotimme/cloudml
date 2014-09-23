package db

import (
  "github.com/aotimme/cloudml/logistic"
  "github.com/aotimme/cloudml/linear"
  "log"
  "errors"
)

func (m *Model) GetDataArray() ([][]float64, []float64, error) {
  data, err := m.GetData()
  if err != nil {
    return nil, nil, err
  }
  n := m.NumTrainingData
  p := m.NumCovariates
  dataArray := make([][]float64, n)
  values := make([]float64, n)
  for i, datum := range data {
    dataArray[i] = make([]float64, p)
    covariates, err := datum.GetCovariates()
    if err != nil {
      return nil, nil, err
    }
    for j, cov := range covariates {
      dataArray[i][j] = cov.Value
    }
    values[i] = datum.Value
  }
  return dataArray, values, nil
}

func GetCoefficientsArrayFromCoefficients(coefficients []Coefficient) []float64 {
  array := make([]float64, len(coefficients))
  for j, coef := range coefficients {
    array[j] = coef.Value
  }
  return array
}

func (m *Model) Learn() error {
  dataArray, values, err := m.GetDataArray()
  if err != nil {
    return err
  }
  coefficients, err := m.GetCoefficients()
  if err != nil {
    return err
  }
  var coefArray []float64
  if m.Type == "logistic" {
    coefArray, err = logistic.Learn(dataArray, values, m.Lambda, GetCoefficientsArrayFromCoefficients(coefficients), 100)
    if err != nil {
      log.Printf("Error running regression: %v\n", err)
      return err
    }
    m.TrainRmse = logistic.RMSE(coefArray, dataArray, values)
  } else if m.Type == "linear" {
    coefArray, err = linear.Learn(dataArray, values, m.Lambda)
    if err != nil {
      log.Printf("Error running regression\n")
      return err
    }
    m.TrainRmse = linear.RMSE(coefArray, dataArray, values)
  }
  for j, value := range coefArray {
    coefficients[j].Value = value
  }
  err = m.SaveWithCoefficients(coefficients)
  if err != nil {
    log.Printf("Error saving model\n")
    return err
  }
  return nil
}

func (m *Model) CV() error {
  dataArray, values, err := m.GetDataArray()
  if err != nil {
    return err
  }
  var cv float64
  if m.Type == "logistic" {
    cv, err = logistic.CV(dataArray, values, m.Lambda)
    if err != nil {
      log.Printf("Error running cv: %v\n", err)
      return err
    }
  } else if m.Type == "linear" {
    cv, err = linear.CV(dataArray, values, m.Lambda)
    if err != nil {
      log.Printf("Error running cv: %v\n", err)
      return err
    }
  }
  m.CvRmse = cv
  err = m.Update()
  if err != nil {
    log.Printf("Error saving model\n")
    return err
  }
  return nil
}

func (m *Model) Predict(covariates map[string]float64) (float64, error) {
  coefficients, err := m.GetCoefficients()
  if err != nil {
    return 0.0, err
  }
  dot := 0.0
  for _, coef := range coefficients {
    dot += coef.Value * covariates[coef.Label]
  }
  coefArray := GetCoefficientsArrayFromCoefficients(coefficients)
  covs := make([]float64, len(coefficients))
  for j, coef := range coefficients {
    covs[j] = covariates[coef.Label]
  }
  if m.Type == "logistic" {
    result := logistic.Predict(coefArray, covs)
    return result, nil
  } else if m.Type == "linear" {
    result := linear.Predict(coefArray, covs)
    return result, nil
  }
  return 0.0, errors.New("Unknown model type")
}

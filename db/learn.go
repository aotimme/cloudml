package db

import (
  "github.com/aotimme/cloudml/logistic"
  "github.com/aotimme/cloudml/linear"
  "log"
  "sort"
)

func (m *Model) GetCovariates() []string {
  p := len(m.Coefficients)
  covKeys := make([]string, p)
  i := 0
  for cov, _ := range m.Coefficients {
    covKeys[i] = cov
    i++
  }
  sort.Strings(covKeys)
  return covKeys
}

func (m *Model) GetDataArray() ([][]float64, []float64, error) {
  data, err := m.GetData()
  if err != nil {
    return nil, nil, err
  }
  n := len(data)
  p := len(m.Coefficients)
  dataArray := make([][]float64, n)
  values := make([]float64, n)
  covKeys := m.GetCovariates()
  for i, datum := range data {
    dataArray[i] = make([]float64, p)
    for j, cov := range covKeys {
      dataArray[i][j] = datum.Covariates[cov]
    }
    values[i] = datum.Value
  }
  return dataArray, values, nil
}

func (m *Model) GetCoefficientsArray() []float64 {
  covKeys := m.GetCovariates()
  p := len(m.Coefficients)
  coefficients := make([]float64, p)
  for j, cov := range covKeys {
    coefficients[j] = m.Coefficients[cov]
  }
  return coefficients
}

func (m *Model) SetCoefficientsFromArray(coefficients []float64) {
  covKeys := m.GetCovariates()
  for j, cov := range covKeys {
    m.Coefficients[cov] = coefficients[j]
  }
}

func (m *Model) Learn() error {
  dataArray, values, err := m.GetDataArray()
  if err != nil {
    return err
  }
  var coefficients []float64
  if m.Type == "logistic" {
    coefficients, err = logistic.Learn(dataArray, values, m.GetCoefficientsArray(), 100)
    if err != nil {
      log.Printf("Error running regression: %v\n", err)
      return err
    }
    m.TrainRmse = logistic.RMSE(coefficients, dataArray, values)
  } else if m.Type == "linear" {
    coefficients, err = linear.Learn(dataArray, values)
    if err != nil {
      log.Printf("Error running regression\n")
      return err
    }
    m.TrainRmse = linear.RMSE(coefficients, dataArray, values)
  }
  m.SetCoefficientsFromArray(coefficients)
  err = m.Save()
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
    cv, err = logistic.CV(dataArray, values)
    if err != nil {
      log.Printf("Error running cv: %v\n", err)
      return err
    }
  } else if m.Type == "linear" {
    cv, err = linear.CV(dataArray, values)
    if err != nil {
      log.Printf("Error running cv: %v\n", err)
      return err
    }
  }
  m.CvRmse = cv
  err = m.Save()
  if err != nil {
    log.Printf("Error saving model\n")
    return err
  }
  return nil
}

func (m *Model) Predict(covariates map[string]float64) float64 {
  dot := 0.0
  for key, val := range m.Coefficients {
    dot += val * covariates[key]
  }
  coefficients := m.GetCoefficientsArray()
  covKeys := m.GetCovariates()
  p := len(m.Coefficients)
  covs := make([]float64, p)
  for j, cov := range covKeys {
    covs[j] = covariates[cov]
  }
  if m.Type == "logistic" {
    return logistic.Predict(coefficients, covs)
  } else if m.Type == "linear" {
    return linear.Predict(coefficients, covs)
  }
  return 0.0
}

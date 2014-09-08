package db

import (
  "github.com/aotimme/cloudml/logistic"
  "github.com/aotimme/cloudml/linear"
  "log"
)

func (m *Model) GetCovariates() []string {
  p := len(m.Coefficients)
  covKeys := make([]string, p)
  for i, variable := range m.Coefficients {
    covKeys[i] = variable.Label
  }
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
  for i, datum := range data {
    dataArray[i] = make([]float64, p)
    for j, cov := range datum.Covariates {
      dataArray[i][j] = cov.Value
    }
    values[i] = datum.Value
  }
  return dataArray, values, nil
}

func (m *Model) GetCoefficientsArray() []float64 {
  p := len(m.Coefficients)
  coefficients := make([]float64, p)
  for j, variable := range m.Coefficients {
    coefficients[j] = variable.Value
  }
  return coefficients
}

func (m *Model) SetCoefficientsFromArray(coefficients []float64) {
  for j, value := range coefficients {
    m.Coefficients[j].Value = value
  }
}

func (m *Model) Learn() error {
  // TODO(Alden): Make this a parameter to the model
  lambda := 0.001
  dataArray, values, err := m.GetDataArray()
  if err != nil {
    return err
  }
  var coefficients []float64
  if m.Type == "logistic" {
    coefficients, err = logistic.Learn(dataArray, values, lambda, m.GetCoefficientsArray(), 100)
    if err != nil {
      log.Printf("Error running regression: %v\n", err)
      return err
    }
    m.TrainRmse = logistic.RMSE(coefficients, dataArray, values)
  } else if m.Type == "linear" {
    coefficients, err = linear.Learn(dataArray, values, lambda)
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
  // TODO(Alden): as above, make this a parameter on the model
  lambda := 0.001
  dataArray, values, err := m.GetDataArray()
  if err != nil {
    return err
  }
  var cv float64
  if m.Type == "logistic" {
    cv, err = logistic.CV(dataArray, values, lambda)
    if err != nil {
      log.Printf("Error running cv: %v\n", err)
      return err
    }
  } else if m.Type == "linear" {
    cv, err = linear.CV(dataArray, values, lambda)
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
  for _, variable := range m.Coefficients {
    dot += variable.Value * covariates[variable.Label]
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

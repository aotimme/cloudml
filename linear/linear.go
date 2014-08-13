package linear

import (
  "math"
  "math/rand"
  "github.com/skelterjohn/go.matrix"
)

func dot(vec1, vec2 []float64) (val float64) {
  for i, v := range vec1 {
    val += v * vec2[i]
  }
  return
}

func Learn(data [][]float64, values []float64) ([]float64, error) {
  X := matrix.MakeDenseMatrixStacked(data)
  Y := matrix.MakeDenseMatrix(values, len(values), 1)
  Xt := X.Transpose()
  XtX, err := Xt.TimesDense(X)
  if err != nil {
    return nil, err
  }
  XtXInv, err := XtX.Inverse()
  if err != nil {
    return nil, err
  }
  XtY, err := Xt.TimesDense(Y)
  if err != nil {
    return nil, err
  }
  coefficients, err := XtXInv.TimesDense(XtY)
  if err != nil {
    return nil, err
  }
  return coefficients.Array(), nil
}

func Predict(beta []float64, covariates []float64) float64 {
  return dot(beta, covariates)
}

func RMSE(beta []float64, data [][]float64, values []float64) float64 {
  rmse := 0.0
  for i, datum := range data {
    rmse += math.Pow(values[i] - Predict(beta, datum), 2.0)
  }
  return math.Sqrt(rmse)
}

func CV(data [][]float64, values []float64) ([]float64, error) {
  fold := 5
  cvRMSE := make([]float64, fold)
  n := len(data)
  // NOTE: stop if p > fold*n
  perm := rand.Perm(n)
  for i, j := range perm {
    data[i], data[j] = data[j], data[i]
    values[i], values[j] = values[j], values[i]
  }
  numPer := n / fold
  mod := n % fold
  for i := 0; i < fold; i++ {
    num := numPer
    if i < mod {
      num++
    }
    tmpValues := make([]float64, num)
    tmpData := make([][]float64, num)
    for j := 0; j < num; j++ {
      tmpData[j] = data[j*fold+i]
      tmpValues[j] = values[j*fold+i]
    }
    betas, err := Learn(tmpData, tmpValues)
    if err != nil {
      return nil, err
    }
    cvRMSE[i] = RMSE(betas, tmpData, tmpValues)
  }
  return cvRMSE, nil
}

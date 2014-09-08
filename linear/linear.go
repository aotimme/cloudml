package linear

import (
  "math"
  "math/rand"
  "log"
  "github.com/skelterjohn/go.matrix"
)

func dot(vec1, vec2 []float64) (val float64) {
  for i, v := range vec1 {
    val += v * vec2[i]
  }
  return
}

func Learn(data [][]float64, values []float64, lambda float64) ([]float64, error) {
  n := len(data)
  p := len(data[0])
  X := matrix.MakeDenseMatrixStacked(data)
  Y := matrix.MakeDenseMatrix(values, n, 1)
  Xt := X.Transpose()
  XtX, err := Xt.TimesDense(X)
  if err != nil {
    return nil, err
  }
  lambdaMatrix := matrix.Eye(p)
  lambdaMatrix.Scale(lambda)
  err = XtX.AddDense(lambdaMatrix)
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
  rmse /= float64(len(data))
  return math.Sqrt(rmse)
}

func CV(data [][]float64, values []float64, lambda float64) (float64, error) {
  fold := 5
  n := len(data)
  // NOTE: stop if p > fold*n
  perm := rand.Perm(n)
  for i, j := range perm {
    data[i], data[j] = data[j], data[i]
    values[i], values[j] = values[j], values[i]
  }
  numPer := n / fold
  mod := n % fold
  cv := 0.0
  numRun := 0
  for i := 0; i < fold; i++ {
    minBreakVal := 0
    for j := 0; j < i; j++ {
      minBreakVal += numPer
      if j < mod {
        minBreakVal++
      }
    }
    maxBreakVal := 0
    for j := 0; j < i + 1; j++ {
      maxBreakVal += numPer
      if j < mod {
        maxBreakVal++
      }
    }
    num := maxBreakVal - minBreakVal
    trainValues := make([]float64, n - num)
    trainData := make([][]float64, n - num)
    testValues := make([]float64, num)
    testData := make([][]float64, num)
    for j := 0; j < n; j++ {
      if j < minBreakVal {
        trainData[j] = data[j]
        trainValues[j] = values[j]
      } else if j < maxBreakVal {
        testData[j - minBreakVal] = data[j]
        testValues[j - minBreakVal] = values[j]
      } else {
        trainData[j - num] = data[j]
        trainValues[j - num] = values[j]
      }
    }
    betas, err := Learn(trainData, trainValues, lambda)
    if err != nil {
      log.Printf("CV error: %v\n", err)
    } else {
      numRun++
    }
    cv += RMSE(betas, testData, testValues)
  }
  cv /= float64(numRun)
  return cv, nil
}

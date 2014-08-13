package db

import (
  "github.com/skelterjohn/go.matrix"
)

func Linear(data [][]float64, values []float64) ([]float64, error) {
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

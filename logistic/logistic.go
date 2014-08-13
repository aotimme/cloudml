package logistic

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

func expit(val float64) float64 {
  return 1.0 / (1.0 + math.Exp(-val))
}

func l2(vals []float64) float64 {
  sum := 0.0
  for _, val := range vals {
    sum += val * val
  }
  return math.Sqrt(sum)
}

func Predict(beta []float64, covariates []float64) float64 {
  return expit(dot(beta, covariates))
}

func Learn(data [][]float64, values []float64, betaStart []float64, iterations int) ([]float64, error) {
  n := len(data)
  p := len(betaStart)
  iter := 0
  X := matrix.MakeDenseMatrixStacked(data)
  //Y := matrix.MakeDenseMatrix(values, n, 1)
  beta := matrix.MakeDenseMatrix(betaStart, p, 1)
  if p >= n {
    return beta.Array(), nil
  }
  for {
    iter++
    //log.Printf("Iteration: %v\n", iter)
    //log.Printf("beta = %v\n", beta)
    hessian := matrix.Zeros(p, p)
    gradient := matrix.Zeros(p, 1)
    lin, err := X.TimesDense(beta)
    if err != nil {
      return nil, err
    }
    exp := matrix.Zeros(n, 1)
    for i := 0; i < n; i++ {
      exp.Set(i, 0, expit(lin.Get(i, 0)))
    }
    for i := 0; i < n; i++ {
      e := exp.Get(i, 0)
      //log.Printf("expit[%v, %v] = %v\n", data[i], values[i], e)
      for j := 0; j < p; j++ {
        toAdd := (values[i] - e) * X.Get(i,j)
        //log.Printf("toAdd (%v, %v) = %v\n", i, j, toAdd)
        gradient.Set(j, 0, gradient.Get(j,0) + toAdd)
      }
      x := X.GetRowVector(i).Transpose()
      xTx, err := x.TimesDense(x.Transpose())
      if err != nil {
        return nil, err
      }
      xTx.Scale(e * (1 - e))
      hessian.AddDense(xTx)
    }
    //log.Printf("gradient = %v\n", gradient.Array())
    //log.Printf("hessian  = %v\n", hessian)
    hessInv, err := hessian.Inverse()
    if err != nil {
      //log.Printf("Hessian inverse error: %v\n", err)
      return nil, err
    }
    diff, err := hessInv.TimesDense(gradient)
    if err != nil {
      return nil, err
    }
    beta.AddDense(diff)
    if diff.TwoNorm() < 1e-6 {
      log.Printf("Converged after %v iterations\n", iter)
      break
    }
    if iter >= iterations {
      log.Printf("Did not converge after %v iterations\n", iter)
      break
    }
    // objective
    //objective := 0.0
    //for i, datum := range data {
    //  e := expit(dot(datum, coefficients))
    //  if values[i] == 1 {
    //    objective += math.Log(e)
    //  } else if values[i] == 0 {
    //    objective += math.Log(1.0 - e)
    //  } else {
    //    log.Printf("Bad value (%v): %v\n", i, values[i])
    //  }
    //}
    //log.Printf("Grad: %v\n", gradient)
    //log.Printf("Coef: %v\n", coefficients)
    //log.Printf("Obj : %v\n", objective)
  }
  return beta.Array(), nil
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
  p := len(data[0])
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
    betaStart := make([]float64, p)
    betas, err := Learn(tmpData, tmpValues, betaStart, 100)
    if err != nil {
      return nil, err
    }
    cvRMSE[i] = RMSE(betas, tmpData, tmpValues)
  }
  return cvRMSE, nil
}

package db

type Model struct {
  Id string `json:"id"`
  Type string `json:"type"`
  Coefficients map[string]float64 `json:"coefficients"`
  N int `json:"n"`
  CVError float64 `json:"cv_error"`
}

type Datum struct {
  Id string `json:"id"`
  Model string `json:"model"`
  Value float64 `json:"value"`
  Covariates map[string]float64 `json:"covariates"`
}


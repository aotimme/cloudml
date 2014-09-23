package main

type Model struct {
  Id string `json:"id"`
  Type string `json:"type"`
  Lambda float64 `json:"lambda"`
  NumTrainingData int `json:"num_training_data"`
  NumCovariates int `json:"num_covariates"`
  TrainRmse float64 `json:"train_rmse"`
  CvRmse float64 `json:"cv_rmse"`
  Coefficients []Coefficient `json:"coefficients"`
}

type Coefficient struct {
  Id string `json:"id"`
  Model string `json:"model"`
  Label string `json:"label"`
  Value float64 `json:"value"`
}

type Covariate struct {
  Id string `json:"id"`
  Datum string `json:"datum"`
  Label string `json:"label"`
  Value float64 `json:"value"`
}

type Datum struct {
  Id string `json:"id"`
  Model string `json:"model"`
  Value float64 `json:"value"`
  Covariates []Covariate `json:"covariates"`
}


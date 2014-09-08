package db

type Variable struct {
  Label string `json:"label"`
  Value float64 `json:"value"`
}

type Model struct {
  Id string `json:"id"`
  Type string `json:"type"`
  Coefficients []Variable `json:"coefficients"`
  NumTrainingData int `json:"num_training_data"`
  TrainRmse float64 `json:"train_rmse"`
  CvRmse float64 `json:"cv_rmse"`
}

type Datum struct {
  Id string `json:"id"`
  Model string `json:"model"`
  Value float64 `json:"value"`
  Covariates []Variable `json:"covariates"`
}


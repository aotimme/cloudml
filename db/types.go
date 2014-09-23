package db

type Model struct {
  Id string `db:"id"`
  Type string `db:"type"`
  Lambda float64 `db:"lambda"`
  NumTrainingData int `db:"num_training_data"`
  NumCovariates int `db:"num_covariates"`
  TrainRmse float64 `db:"train_rmse"`
  CvRmse float64 `db:"cv_rmse"`
}
type Coefficient struct {
  Id string `db:"id"`
  Label string `db:"label"`
  Value float64 `db:"value"`
  Model string `db:"model"`
}
type Covariate struct {
  Id string `db:"id"`
  Label string `db:"label"`
  Value float64 `db:"value"`
  Datum string `db:"datum"`
}
type Datum struct {
  Id string `db:"id"`
  Value float64 `db:"value"`
  //Covariates []Covariate `db:"covariates"`
  Model string `db:"model"`
}

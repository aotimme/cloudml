package main

import (
  "github.com/aotimme/cloudml/db"
  "net/http"
  "github.com/gorilla/mux"
  "log"
  "encoding/json"
  "fmt"
)

var learnChannel chan string = make(chan string, 1000)

type ErrorResponse struct {
  Error string `json:"error"`
}
type PreModel struct {
  Type string `json:"type"`
  Covariates []string `json:"covariates"`
  Lambda float64 `json:"lambda"`
}
type PreDatum struct {
  Value float64 `json:"value"`
  Covariates map[string]float64 `json:"covariates"`
}


func SendError(rw http.ResponseWriter, message string, statusCode int) {
  response := &ErrorResponse{Error: message}
  jsonData, err := json.Marshal(response)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  log.Printf("sending error %v", string(jsonData))
  rw.Header().Set("Content-Type", "application/json")
  rw.WriteHeader(statusCode)
  rw.Write(jsonData)
}

func GetDatumById(datumId string) (*Datum, error) {
  datum, err := db.GetDatumById(datumId)
  if err != nil {
    return nil, err
  }
  covariates, err := datum.GetCovariates()
  if err != nil {
    return nil, err
  }
  covs := make([]Covariate, len(covariates))
  for i, cov := range covariates {
    covs[i] = Covariate{
      Id: cov.Id,
      Datum: cov.Datum,
      Label: cov.Label,
      Value: cov.Value,
    }
  }
  d := &Datum{
    Id: datum.Id,
    Model: datum.Model,
    Value: datum.Value,
    Covariates: covs,
  }
  return d, nil
}
func SendDatumById(rw http.ResponseWriter, datumId string) {
  d, err := GetDatumById(datumId)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  jsonData, err := json.Marshal(d)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  rw.Header().Set("Content-Type", "application/json")
  rw.Write(jsonData)
}

func GetModelFromDBModelAndCoefficients(m *db.Model, cs []db.Coefficient) (*Model) {
  coefficients := make([]Coefficient, len(cs))
  for i, c := range cs {
    coefficients[i] = Coefficient{
      Id: c.Id,
      Model: c.Model,
      Label: c.Label,
      Value: c.Value,
    }
  }
  return &Model{
    Id: m.Id,
    Type: m.Type,
    Lambda: m.Lambda,
    NumTrainingData: m.NumTrainingData,
    NumCovariates: m.NumCovariates,
    TrainRmse: m.TrainRmse,
    CvRmse: m.CvRmse,
    Coefficients: coefficients,
  }
}
func GetModelById(modelId string) (*Model, error) {
  model, coefficients, err := db.GetModelAndCoefficientsById(modelId)
  if err != nil {
    return nil, err
  }
  if model == nil {
    return nil, nil
  }
  m := GetModelFromDBModelAndCoefficients(model, coefficients)
  return m, nil
}
func SendModelById(rw http.ResponseWriter, modelId string) {
  m, err := GetModelById(modelId)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  if m == nil {
    http.Error(rw, "Not Found", http.StatusNotFound)
    return
  }
  jsonData, err := json.Marshal(m)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  rw.Header().Set("Content-Type", "application/json")
  rw.Write(jsonData)
}
func SendAllModelsByIds(rw http.ResponseWriter, modelIds []string) {
  models := make([]*Model, len(modelIds))
  for i, modelId := range modelIds {
    m, err := GetModelById(modelId)
    if err != nil {
      http.Error(rw, err.Error(), http.StatusInternalServerError)
      return
    }
    models[i] = m
  }
  jsonData, err := json.Marshal(models)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  rw.Header().Set("Content-Type", "application/json")
  rw.Write(jsonData)
}

func SendDatumJSON(rw http.ResponseWriter, d *db.Datum) {
  jsonData, err := json.Marshal(d)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  rw.Header().Set("Content-Type", "application/json")
  rw.Write(jsonData)
}
func SendDataJSON(rw http.ResponseWriter, ds []*Datum) {
  jsonData, err := json.Marshal(ds)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  rw.Header().Set("Content-Type", "application/json")
  rw.Write(jsonData)
}

func IndexHandler(rw http.ResponseWriter, req *http.Request) {
  log.Println("Handling GET \"/\"")
  rw.Write([]byte("OK"))
}

func CreateModelHandler(rw http.ResponseWriter, req *http.Request) {
  log.Printf("Handling POST \"/api/models\"\n")
  decoder := json.NewDecoder(req.Body)
  var pre PreModel
  err := decoder.Decode(&pre)
  if err != nil {
    log.Printf(err.Error());
    SendError(rw, "Malformed model data", http.StatusBadRequest);
    //http.Error(rw, err.Error(), http.StatusBadRequest)
    return
  }
  m := &db.Model{
    Type: pre.Type,
    Lambda: pre.Lambda,
  }
  coefficients := make([]db.Coefficient, len(pre.Covariates))
  for i, covariate := range pre.Covariates {
    coefficients[i].Label = covariate
    coefficients[i].Value = 0.0
  }
  if err != nil {
    http.Error(rw, err.Error(), http.StatusBadRequest)
    return
  }
  log.Printf("Creating model: %v\n", m)
  err = m.SaveWithCoefficients(coefficients)
  if err == nil {
    log.Printf("Successfully created model: %v\n", m)
  } else {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }

  SendModelById(rw, m.Id)
}

func GetModelsHandler(rw http.ResponseWriter, req *http.Request) {
  log.Printf("Handling GET \"/api/models\"\n")
  modelIds, err := db.GetAllModelIds()
  if err != nil {
    http.Error(rw, err.Error(), http.StatusNotFound)
    return
  }
  SendAllModelsByIds(rw, modelIds)
}


func GetModelHandler(rw http.ResponseWriter, req *http.Request) {
  vars := mux.Vars(req)
  id := vars["id"]
  log.Printf("Handling GET \"/api/models/%v\"\n", id)
  SendModelById(rw, id)
  //SendError(rw, fmt.Sprintf("Could not find model with id %v", id), http.StatusNotFound)
}

func DeleteModelHandler(rw http.ResponseWriter, req *http.Request) {
  vars := mux.Vars(req)
  id := vars["id"]
  log.Printf("Handling DELETE \"/api/models/%v\"\n", id)
  err := db.DeleteModelById(id)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  rw.Header().Set("Content-Type", "application/json")
  rw.Write([]byte("{}"))
}

func CreateDatumHandler(rw http.ResponseWriter, req *http.Request) {
  vars := mux.Vars(req)
  id := vars["id"]
  log.Printf("Handling POST \"/api/models/%v/datum\"\n", id)
  m, err := db.GetModelById(id)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusNotFound)
    return
  }
  decoder := json.NewDecoder(req.Body)
  var pre PreDatum
  err = decoder.Decode(&pre)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusBadRequest)
    return
  }
  d, err := m.CreateDatum(pre.Covariates, pre.Value)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  //XXX(Alden): enable async learning?
  //learnChannel <- m.Id
  SendDatumById(rw, d.Id)
}

func CreateDataHandler(rw http.ResponseWriter, req *http.Request) {
  vars := mux.Vars(req)
  id := vars["id"]
  log.Printf("Handling POST \"/api/models/%v/data\"\n", id)
  m, err := db.GetModelById(id)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusNotFound)
    return
  }
  decoder := json.NewDecoder(req.Body)
  var pres []PreDatum
  err = decoder.Decode(&pres)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusBadRequest)
    return
  }
  ds := make([]*db.Datum, len(pres))
  for i, pre := range pres {
    d, err := m.CreateDatum(pre.Covariates, pre.Value)
    if err != nil {
      http.Error(rw, err.Error(), http.StatusBadRequest)
      return
    }
    ds[i] = d
  }
  //XXX(Alden): enable async learning?
  //learnChannel <- m.Id
  // TODO: Actually get data (with covariates) and send
  data := make([]*Datum, len(ds))
  for i, datum := range ds {
    data[i], err = GetDatumById(datum.Id)
    if err != nil {
      http.Error(rw, err.Error(), http.StatusInternalServerError)
      return
    }
  }
  SendDataJSON(rw, data)
}

func LearnModelHandler(rw http.ResponseWriter, req *http.Request) {
  vars := mux.Vars(req)
  id := vars["id"]
  log.Printf("Handling POST \"/api/models/%v/learn\"\n", id)
  m, err := db.GetModelById(id)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusNotFound)
    return
  }
  err = m.Learn()
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  SendModelById(rw, m.Id)
}

func CVModelHandler(rw http.ResponseWriter, req *http.Request) {
  vars := mux.Vars(req)
  id := vars["id"]
  log.Printf("Handling POST \"/api/models/%v/cv\"\n", id)
  m, err := db.GetModelById(id)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusNotFound)
    return
  }
  err = m.CV()
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  SendModelById(rw, m.Id)
}


func PredictModelHandler(rw http.ResponseWriter, req *http.Request) {
  vars := mux.Vars(req)
  id := vars["id"]
  log.Printf("Handling PUT \"/api/models/%v/learn\"\n", id)
  m, err := db.GetModelById(id)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusNotFound)
    return
  }
  decoder := json.NewDecoder(req.Body)
  var pre PreDatum
  err = decoder.Decode(&pre)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusBadRequest)
    return
  }

  prediction, err := m.Predict(pre.Covariates)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  resp := make(map[string]float64)
  resp["value"] = prediction

  jsonData, err := json.Marshal(resp)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  rw.Header().Set("Content-Type", "application/json")
  rw.Write(jsonData)
}

func GetDataHandler(rw http.ResponseWriter, req *http.Request) {
  vars := mux.Vars(req)
  id := vars["id"]
  log.Printf("Handling GET \"/api/models/%v/data\"\n", id)
  m, err := db.GetModelById(id)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusNotFound)
    return
  }
  ds, err := m.GetData()
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  // TODO: Send data with covariates
  data := make([]*Datum, len(ds))
  for i, datum := range ds {
    data[i], err = GetDatumById(datum.Id)
    if err != nil {
      http.Error(rw, err.Error(), http.StatusInternalServerError)
      return
    }
  }
  SendDataJSON(rw, data)
}

func RemoveDataHandler(rw http.ResponseWriter, req *http.Request) {
  vars := mux.Vars(req)
  id := vars["id"]
  log.Printf("Handling DELETE \"/api/models/%v/data\"\n", id)
  m, err := db.GetModelById(id)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusNotFound)
    return
  }
  err = m.DeleteData()
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  rw.Header().Set("Content-Type", "application/json")
  rw.Write([]byte("{}"))
}

func main() {
  // TODO(Alden): if we really enable this, we should debounce the calls to `Learn`
  go func(ch <-chan string) {
    for id := range ch {
      m, err := db.GetModelById(id)
      if err != nil {
        log.Printf("Learn error: %v\n", err)
      } else {
        m.Learn()
      }
    }
  }(learnChannel)

  r := mux.NewRouter()
  r.HandleFunc("/api/models", CreateModelHandler).Methods("POST")
  r.HandleFunc("/api/models", GetModelsHandler).Methods("GET")
  r.HandleFunc("/api/models/{id}", GetModelHandler).Methods("GET")
  r.HandleFunc("/api/models/{id}", DeleteModelHandler).Methods("DELETE")
  r.HandleFunc("/api/models/{id}/datum", CreateDatumHandler).Methods("POST")
  r.HandleFunc("/api/models/{id}/data", CreateDataHandler).Methods("POST")
  r.HandleFunc("/api/models/{id}/data", GetDataHandler).Methods("GET")
  r.HandleFunc("/api/models/{id}/data", RemoveDataHandler).Methods("DELETE")
  // XXX(Alden): Remove and do learning async?
  r.HandleFunc("/api/models/{id}/learn", LearnModelHandler).Methods("POST")
  r.HandleFunc("/api/models/{id}/predict", PredictModelHandler).Methods("POST")
  r.HandleFunc("/api/models/{id}/cv", CVModelHandler).Methods("POST")
  r.HandleFunc("/", IndexHandler).Methods("GET")
  http.Handle("/", r)
  port := "6060"
  log.Printf("CloudML: port %v\n", port)
  log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

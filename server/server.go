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
func SendModelJSON(rw http.ResponseWriter, model *db.Model) {
  jsonData, err := json.Marshal(model)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  rw.Header().Set("Content-Type", "application/json")
  rw.Write(jsonData)
}
func SendModelsJSON(rw http.ResponseWriter, models []*db.Model) {
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
func SendDataJSON(rw http.ResponseWriter, ds []*db.Datum) {
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
  m := &db.Model{Type: pre.Type, Coefficients: make(map[string]float64)}
  for _, covariate := range pre.Covariates {
    m.Coefficients[covariate] = 0
  }
  if err != nil {
    http.Error(rw, err.Error(), http.StatusBadRequest)
    return
  }
  log.Printf("Creating model: %v\n", m)
  err = m.Save()
  if err == nil {
    log.Printf("Successfully created model: %v\n", m)
  } else {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }

  SendModelJSON(rw, m)
}

func GetModelsHandler(rw http.ResponseWriter, req *http.Request) {
  log.Printf("Handling GET \"/api/models\"\n")
  models, err := db.GetAllModels()
  if err != nil {
    http.Error(rw, err.Error(), http.StatusNotFound)
    return
  }
  SendModelsJSON(rw, models)
}


func GetModelHandler(rw http.ResponseWriter, req *http.Request) {
  vars := mux.Vars(req)
  id := vars["id"]
  log.Printf("Handling GET \"/api/models/%v\"\n", id)
  m, err := db.GetModelById(id)
  if err != nil {
    SendError(rw, fmt.Sprintf("Could not find model with id %v", id), http.StatusNotFound)
    return
  }
  SendModelJSON(rw, m)
}

func DeleteModelHandler(rw http.ResponseWriter, req *http.Request) {
  vars := mux.Vars(req)
  id := vars["id"]
  log.Printf("Handling GET \"/api/models/%v\"\n", id)
  m, err := db.GetModelById(id)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusNotFound)
    return
  }
  err = m.Delete()
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
  d, err := m.CreateAndSaveDatum(pre.Covariates, pre.Value)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
    return
  }
  learnChannel <- m.Id
  SendDatumJSON(rw, d)
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
    d, err := m.CreateAndSaveDatum(pre.Covariates, pre.Value)
    if err != nil {
      http.Error(rw, err.Error(), http.StatusBadRequest)
      return
    }
    ds[i] = d
  }
  learnChannel <- m.Id
  SendDataJSON(rw, ds)
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
  SendModelJSON(rw, m)
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
  SendModelJSON(rw, m)
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

  prediction := m.Predict(pre.Covariates)
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
  SendDataJSON(rw, ds)
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
  // TODO: Remove? Not actually necessary (learn on each data save)
  r.HandleFunc("/api/models/{id}/learn", LearnModelHandler).Methods("POST")
  r.HandleFunc("/api/models/{id}/predict", PredictModelHandler).Methods("POST")
  r.HandleFunc("/api/models/{id}/cv", CVModelHandler).Methods("POST")
  r.HandleFunc("/", IndexHandler).Methods("GET")
  http.Handle("/", r)
  port := "6060"
  log.Printf("CloudML: port %v\n", port)
  log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

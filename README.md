cloudml
=======

Solve ML problems as a web service

* `POST /models` => create a model
* `GET /models/:id` => get the statistics on a model
* `POST /models/:id/datum` => send a data point
* `POST /models/:id/data` => send multiple data points
* `POST /models/:id/learn` => train the model
* `POST /models/:id/predict` => evaluate on a data point

Types of models (all regression):

* "logistic"
* "linear"
* "poisson" (later?) 

```
POST /models
```

```json
{
  "id": "xxx",
  "type": "logistic",
  "covariates": ["age", "gender", "age-gender"],
  "trained": false
}
```

```
GET /models/:id
```

```json
{
  "id": "xxx",
  "type": "logistic",
  "coefficients": {
    "age": 2.52,
    "gender": 3.112,
    "age-gender": -0.91,
  },
  "training_error": "0.01",
  "regularization": "0.0001"
}
```

```
DELETE /models/:id
```

OK (deletes model and all data (?))

```
POST /models/:id/datum
```

```json
{
  "id": "yyy",
  "value": 1,
  "covariates": {
    "age": 78,
    "gender": 1,
    "age-gender": 78
  }
}
```

```
POST /models/:id/data
```

```
POST /models/:id/learn
```

Responds as `GET /models/:id`, OR sends 200 and then has the user check
`GET /models/:id`

```
POST /models/:id/predict
```

```json
{
  "prediction": 0.23,
  "value": 0
}
```

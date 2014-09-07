var should  = require('should');
var request = require('request').defaults({jar: false});
var async   = require('async');
var _       = require('underscore');

var CLOUDML_URL = 'http://localhost:6060'

var LOGISTIC_MODEL = {
  type: 'logistic',
  covariates: ['intercept', 'age', 'gender']
};

describe('logistic', function() {
  var LOGISTIC_MODEL_ID = undefined;
  var LOGISTIC_MODEL_RESPONSE = undefined;

  it('should correctly save the model', function(done) {
    request({
      url: CLOUDML_URL + '/api/models',
      method: 'POST',
      json: LOGISTIC_MODEL
    }, function(err, resp, body) {
      should.not.exist(err);
      body.should.have.property('id');
      body.should.have.property('coefficients');
      _.keys(body.coefficients).should.have.lengthOf(3);
      body.should.have.property('num_training_data', 0);
      LOGISTIC_MODEL_RESPONSE = body;
      LOGISTIC_MODEL_ID = body.id;
      done();
    });
  });

  it('should correctly get the saved model', function(done) {
    request({
      url: CLOUDML_URL + '/api/models/' + LOGISTIC_MODEL_ID,
      method: 'GET',
      json: true
    }, function(err, resp, body) {
      should.not.exist(err);
      body.should.eql(LOGISTIC_MODEL_RESPONSE);
      done();
    });
  });

  it('should correctly delete the saved model', function(done) {
    request({
      url: CLOUDML_URL + '/api/models/' + LOGISTIC_MODEL_ID,
      method: 'DELETE',
      json: true
    }, function(err, resp, body) {
      should.not.exist(err);
      JSON.stringify(body).should.equal("{}");
      done();
    });
  });

  it('should receive a 404 on deleted model', function(done) {
    request({
      url: CLOUDML_URL + '/api/models/' + LOGISTIC_MODEL_ID,
      method: 'GET',
      json: true
    }, function(err, resp, body) {
      should.not.exist(err);
      resp.should.have.property('statusCode', 404);
      done();
    });
  });

  it('should receive not receive model among all models', function(done) {
    request({
      url: CLOUDML_URL + '/api/models',
      method: 'GET',
      json: true
    }, function(err, resp, body) {
      should.not.exist(err);
      ids = _.map(body, function(model) {
        return model.id;
      });
      ids.should.not.containEql(LOGISTIC_MODEL_ID);
      done();
    });
  });

});

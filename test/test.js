var should  = require('should');
var request = require('request').defaults({jar: false});
var async   = require('async');
var _       = require('underscore');
var fs      = require('fs');

var CLOUDML_URL = 'http://localhost:6060'

// Looks like:
// {type: "logistic",  covariates: ['intercept', 'age', 'gender']}
var LOGISTIC_MODEL = undefined;
// Looks like:
//[{
//  value: 0,
//  covariates: {
//    intercept: 1,
//    age: 10,
//    gender: 0
//  }
//}]
var LOGISTIC_DATA = undefined;

describe('logistic', function() {
  this.timeout(10000);

  var LOGISTIC_MODEL_ID = undefined;
  var LOGISTIC_MODEL_RESPONSE = undefined;

  before(function(done) {
    fs.readFile('./data/binary.csv', function(err, data) {
      if (err) {
        return done(err);
      }
      data = data.toString();
      var lines = data.split('\r\n');
      var covariates = lines[0].split(',').splice(1);
      LOGISTIC_MODEL = {type: 'logistic', covariates: covariates};
      LOGISTIC_DATA = _.chain(lines.splice(1))
        .map(function(line) {
          if (!line) {
            return null;
          }
          var datum = line.split(',');
          var value = parseFloat(datum[0]);
          var covs = {};
          _.each(datum.splice(1), function(val, i) {
            covs[covariates[i]] = parseFloat(val);
          });
          return {value: value, covariates: covs};
        })
        .compact()
        .value();
      done();
    });
  });

  it('should correctly save the model', function(done) {
    request({
      url: CLOUDML_URL + '/api/models',
      method: 'POST',
      json: LOGISTIC_MODEL
    }, function(err, resp, body) {
      should.not.exist(err);
      LOGISTIC_MODEL_RESPONSE = body;
      body.should.have.property('id');
      LOGISTIC_MODEL_ID = body.id;
      body.should.have.property('coefficients');
      body.coefficients.should.have.lengthOf(3);
      var labels = _.map(body.coefficients, function(coef) {
        return coef.label;
      });
      labels.should.eql(LOGISTIC_MODEL.covariates);
      _.each(body.coefficients, function(coef) {
        coef.should.have.property('value', 0);
      });
      body.should.have.property('num_training_data', 0);
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

  it('should correctly add the data', function(done) {
    request({
      url: CLOUDML_URL + '/api/models/' + LOGISTIC_MODEL_ID + '/data',
      method: 'POST',
      json: LOGISTIC_DATA
    }, function(err, resp, data) {
      should.not.exist(err);
      data.should.have.lengthOf(LOGISTIC_DATA.length);
      _.each(data, function(datum) {
        datum.should.have.property('model', LOGISTIC_MODEL_ID);
      });
      // Let it learn...
      setTimeout(done, 1000);
    });
  });

  it('should correctly get the model', function(done) {
    request({
      url: CLOUDML_URL + '/api/models/' + LOGISTIC_MODEL_ID,
      method: 'GET',
      json: true
    }, function(err, resp, model) {
      should.not.exist(err);
      model.should.have.property('id', LOGISTIC_MODEL_ID);
      model.should.have.property('num_training_data', LOGISTIC_DATA.length);
      var coefMap = {};
      _.each(model.coefficients, function(coef) {
        coefMap[coef.label] = coef.value;
      });
      _.keys(coefMap).should.eql(LOGISTIC_MODEL.covariates);
      //console.log('coefficients', model.coefficients);
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

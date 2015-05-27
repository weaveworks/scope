var flux = require('flux');
var _ = require('lodash');

var AppDispatcher = new flux.Dispatcher();

AppDispatcher.dispatch = _.wrap(flux.Dispatcher.prototype.dispatch, function(func) {
  var args = Array.prototype.slice.call(arguments, 1);
  // console.log(args[0]);
  func.apply(this, args);
});

module.exports = AppDispatcher;

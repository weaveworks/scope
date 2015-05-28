const flux = require('flux');
const _ = require('lodash');

const AppDispatcher = new flux.Dispatcher();

AppDispatcher.dispatch = _.wrap(flux.Dispatcher.prototype.dispatch, function(func) {
  const args = Array.prototype.slice.call(arguments, 1);
  // console.log(args[0]);
  func.apply(this, args);
});

module.exports = AppDispatcher;

import { Dispatcher } from 'flux';
import _ from 'lodash';
import debug from 'debug';
const log = debug('scope:dispatcher');

const instance = new Dispatcher();

instance.dispatch = _.wrap(Dispatcher.prototype.dispatch, function(func) {
  const args = Array.prototype.slice.call(arguments, 1);
  const type = args[0] && args[0].type;
  log(type, args[0]);
  func.apply(this, args);
});

export default instance;
export const dispatch = instance.dispatch.bind(instance);

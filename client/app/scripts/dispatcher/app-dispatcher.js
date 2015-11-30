import { Dispatcher } from 'flux';
import _ from 'lodash';

const instance = new Dispatcher();

instance.dispatch = _.wrap(Dispatcher.prototype.dispatch, function(func) {
  const args = Array.prototype.slice.call(arguments, 1);
  // console.log(args[0]);
  func.apply(this, args);
});

export default instance;
export const dispatch = instance.dispatch.bind(instance);

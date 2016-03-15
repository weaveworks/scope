import { Dispatcher } from 'flux';
import _ from 'lodash';
import debug from 'debug';
const log = debug('scope:dispatcher');

const instance = new Dispatcher();

instance.dispatch = _.wrap(Dispatcher.prototype.dispatch, (func, payload) => {
  log(payload.type, payload);
  func.call(instance, payload);
});

export default instance;
export const dispatch = instance.dispatch.bind(instance);

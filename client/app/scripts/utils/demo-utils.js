import _ from 'lodash';
import { clickNode, clickTopology, clickCloseDetails } from '../actions/app-actions';

const INTERVAL = 5000;

function getRandomNode(getState) {
  const { nodes } = getState().toJS();
  return _.sample(nodes);
}

function runAction(action, args, dispatch, getState) {
  const payload = action.apply(this, args);
  if (typeof payload === 'function') {
    return payload(dispatch, getState);
  }

  return dispatch(payload);
}

function openNodeDetails(dispatch, getState, run, next) {
  const { id, label } = getRandomNode(getState);
  run(clickNode, [id, label]);
  setTimeout(() => {
    run(clickCloseDetails, [id]);
    next();
  }, INTERVAL * 2);
}

function changeTopology(dispatch, getState, run, next) {
  const blacklist = ['hosts', 'weave', 'containers-by-hostname'];
  const { topologies } = getState().toJS();
  const subTopologies = _.reduce(topologies, (result, t) => {
    if (t.sub_topologies) {
      return result.concat(t.sub_topologies);
    }
    return result;
  }, []);
  const all = _.filter(topologies.concat(subTopologies), i => !_.includes(blacklist, i.id));
  run(clickTopology, [_.sample(all).id]);
  next();
}

const demos = [openNodeDetails, changeTopology];

export default function runDemo(dispatch, getState) {
  setTimeout(() => {
    const fn = _.sample(demos);
    const run = _.partialRight(runAction, dispatch, getState);
    const next = runDemo.bind(runDemo, dispatch, getState);
    fn(dispatch, getState, run, next);
  }, INTERVAL);
}

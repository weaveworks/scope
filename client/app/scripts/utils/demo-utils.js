import _ from 'lodash';
import { clickNode } from '../actions/app-actions';

const INTERVAL = 3000;

function openNodeDetails(getState, dispatch) {
  const action = clickNode();
  if (typeof action === 'function') {
    return action(dispatch, getState);
  }

  return dispatch(action);
}

const actions = [openNodeDetails];

export default function runDemo(getState) {
  setInterval(() => {
    const state = getState().toJS();
    const fn = _.sample(actions);

  }, INTERVAL);
}

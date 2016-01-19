import React from 'react';
import _ from 'lodash';

import debug from 'debug';
const log = debug('scope:debug-panel');

import { receiveNodesDelta } from '../actions/app-actions';
import AppStore from '../stores/app-store';


const sample = function(collection) {
  return _.range(_.random(4)).map(() => _.sample(collection));
};

const deltaAdd = function(name, adjacency = []) {
  return {
    'adjacency': adjacency,
    'controls': {},
    'id': name,
    'label_major': name,
    'label_minor': 'weave-1',
    'latest': {},
    'metadata': {},
    'origins': [],
    'rank': 'alpine'
  };
};

function addNodes(n) {
  const ns = AppStore.getNodes();
  const nodeNames = ns.keySeq().toJS();
  const newNodeNames = _.range(ns.size, ns.size + n).map((i) => 'zing' + i);
  const allNodes = _(nodeNames).concat(newNodeNames).value();

  receiveNodesDelta({
    add: newNodeNames.map((name) => deltaAdd(name, sample(allNodes)))
  });
}

export function showingDebugToolbar() {
  return Boolean(localStorage.debugToolbar);
}

export class DebugToolbar extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.onChange = this.onChange.bind(this);
    this.state = {
      nodesToAdd: 30
    };
  }

  onChange(ev) {
    this.setState({nodesToAdd: parseInt(ev.target.value, 10)});
  }

  render() {
    log('rending debug panel');

    return (
      <div className="debug-panel">
        <label>Add nodes </label>
        <button onClick={() => addNodes(1)}>+1</button>
        <button onClick={() => addNodes(10)}>+10</button>
        <input type="number" onChange={this.onChange} value={this.state.nodesToAdd} />
        <button onClick={() => addNodes(this.state.nodesToAdd)}>+</button>
      </div>
    );
  }
}

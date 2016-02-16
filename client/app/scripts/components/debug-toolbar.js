import React from 'react';
import _ from 'lodash';

import debug from 'debug';
const log = debug('scope:debug-panel');

import { receiveNodesDelta } from '../actions/app-actions';
import AppStore from '../stores/app-store';

const SHAPES = ['circle', 'hexagon', 'square'];
const NODE_COUNTS = [1, 2, 3];
const STACK_VARIANTS = [true, false];

const sample = function(collection) {
  return _.range(_.random(4)).map(() => _.sample(collection));
};

const deltaAdd = function(name, adjacency = [], shape = 'circle', stack = false, nodeCount = 1) {
  return {
    'adjacency': adjacency,
    'controls': {},
    'shape': shape,
    'stack': stack,
    'node_count': nodeCount,
    'id': name,
    'label_major': name,
    'label_minor': 'weave-1',
    'latest': {},
    'metadata': {},
    'origins': [],
    'rank': 'alpine'
  };
};

function addAllVariants() {
  const newNodes = _.flattenDeep(SHAPES.map(s => {
    return STACK_VARIANTS.map(stack => {
      if (!stack) return [deltaAdd([s, 1, stack].join('-'), [], s, stack, 1)];
      return NODE_COUNTS.map(n => {
        return deltaAdd([s, n, stack].join('-'), [], s, stack, n);
      });
    });
  }));

  console.log(newNodes);

  receiveNodesDelta({
    add: newNodes
  });
}

function addNodes(n) {
  const ns = AppStore.getNodes();
  const nodeNames = ns.keySeq().toJS();
  const newNodeNames = _.range(ns.size, ns.size + n).map((i) => 'zing' + i);
  const allNodes = _(nodeNames).concat(newNodeNames).value();

  receiveNodesDelta({
    add: newNodeNames.map((name) => deltaAdd(name,
                                             sample(allNodes)),
                                             _.sample(SHAPES),
                                             _.sample(STACK_VARIANTS),
                                             _.sample(NODE_COUNTS))
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
        <button onClick={() => addAllVariants()}>Variants</button>
      </div>
    );
  }
}

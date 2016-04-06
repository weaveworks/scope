/* eslint react/jsx-no-bind: "off" */
import React from 'react';
import _ from 'lodash';
import Perf from 'react-addons-perf';

import debug from 'debug';
const log = debug('scope:debug-panel');

import { receiveNodesDelta } from '../actions/app-actions';
import AppStore from '../stores/app-store';
import { getNodeColor, getNodeColorDark } from '../utils/color-utils';


const SHAPES = ['square', 'hexagon', 'heptagon', 'circle'];
const NODE_COUNTS = [1, 2, 3];
const STACK_VARIANTS = [false, true];
const METRIC_FILLS = [0, 0.1, 50, 99.9, 100];


const sample = (collection) => _.range(_.random(4)).map(() => _.sample(collection));


const shapeTypes = {
  square: ['Process', 'Processes'],
  hexagon: ['Container', 'Containers'],
  heptagon: ['Pod', 'Pods'],
  circle: ['Host', 'Hosts']
};


const LABEL_PREFIXES = _.range('A'.charCodeAt(), 'Z'.charCodeAt() + 1)
  .map(n => String.fromCharCode(n));


// const randomLetter = () => _.sample(LABEL_PREFIXES);


const deltaAdd = (name, adjacency = [], shape = 'circle', stack = false, nodeCount = 1) => ({
  adjacency,
  controls: {},
  shape,
  stack,
  node_count: nodeCount,
  id: name,
  label: name,
  label_minor: name,
  latest: {},
  metadata: {},
  origins: [],
  rank: name
});


function addMetrics(node, v) {
  const availableMetrics = AppStore.getAvailableCanvasMetrics().toJS();
  const metrics = availableMetrics.length > 0 ? availableMetrics : [
    {id: 'host_cpu_usage_percent', label: 'CPU'}
  ];

  return Object.assign({}, node, {
    metrics: metrics.map(m => Object.assign({}, m, {max: 100, value: v}))
  });
}


function label(shape, stacked) {
  const type = shapeTypes[shape];
  return stacked ? `Group of ${type[1]}` : type[0];
}


function addAllVariants() {
  const newNodes = _.flattenDeep(STACK_VARIANTS.map(stack => (SHAPES.map(s => {
    if (!stack) return [deltaAdd(label(s, stack), [], s, stack, 1)];
    return NODE_COUNTS.map(n => deltaAdd(label(s, stack), [], s, stack, n));
  }))));

  receiveNodesDelta({
    add: newNodes
  });
}


function addAllMetricVariants() {
  const newNodes = _.flattenDeep(METRIC_FILLS.map((v, i) => (
    SHAPES.map(s => [addMetrics(deltaAdd(label(s) + i, [], s), v)])
  )));

  receiveNodesDelta({
    add: newNodes
  });
}


function stopPerf() {
  Perf.stop();
  const measurements = Perf.getLastMeasurements();
  Perf.printInclusive(measurements);
  Perf.printWasted(measurements);
}

function startPerf(delay) {
  Perf.start();
  setTimeout(stopPerf, delay * 1000);
}


function addNodes(n) {
  const ns = AppStore.getNodes();
  const nodeNames = ns.keySeq().toJS();
  const newNodeNames = _.range(ns.size, ns.size + n).map(i => (
    // `${randomLetter()}${randomLetter()}-zing`
    `zing${i}`
  ));
  const allNodes = _(nodeNames).concat(newNodeNames).value();

  receiveNodesDelta({
    add: newNodeNames.map((name) => deltaAdd(
      name,
      sample(allNodes),
      _.sample(SHAPES),
      _.sample(STACK_VARIANTS),
      _.sample(NODE_COUNTS)
    ))
  });
}

export function showingDebugToolbar() {
  return (('debugToolbar' in localStorage && JSON.parse(localStorage.debugToolbar))
    || location.pathname.indexOf('debug') > -1);
}


export function toggleDebugToolbar() {
  if ('debugToolbar' in localStorage) {
    localStorage.debugToolbar = !showingDebugToolbar();
  }
}


function enableLog(ns) {
  debug.enable(`scope:${ns}`);
  window.location.reload();
}

function disableLog() {
  debug.disable();
  window.location.reload();
}

export class DebugToolbar extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.onChange = this.onChange.bind(this);
    this.toggleColors = this.toggleColors.bind(this);
    this.state = {
      nodesToAdd: 30,
      showColors: false
    };
  }

  onChange(ev) {
    this.setState({nodesToAdd: parseInt(ev.target.value, 10)});
  }

  toggleColors() {
    this.setState({
      showColors: !this.state.showColors
    });
  }

  render() {
    log('rending debug panel');

    return (
      <div className="debug-panel">
        <div>
          <label>Add nodes </label>
          <button onClick={() => addNodes(1)}>+1</button>
          <button onClick={() => addNodes(10)}>+10</button>
          <input type="number" onChange={this.onChange} value={this.state.nodesToAdd} />
          <button onClick={() => addNodes(this.state.nodesToAdd)}>+</button>
          <button onClick={() => addAllVariants()}>Variants</button>
          <button onClick={() => addAllMetricVariants()}>Metric Variants</button>
        </div>

        <div>
          <label>Logging</label>
          <button onClick={() => enableLog('*')}>scope:*</button>
          <button onClick={() => enableLog('dispatcher')}>scope:dispatcher</button>
          <button onClick={() => enableLog('app-key-press')}>scope:app-key-press</button>
          <button onClick={() => disableLog()}>Disable log</button>
        </div>

        <div>
          <label>Colors</label>
          <button onClick={this.toggleColors}>toggle</button>
        </div>

        {this.state.showColors && [getNodeColor, getNodeColorDark].map(fn => (
          <table>
            <tbody>
              {LABEL_PREFIXES.map(r => (
                <tr key={r}>
                  {LABEL_PREFIXES.map(c => (
                    <td key={c} title={`(${r}, ${c})`} style={{backgroundColor: fn(r, c)}}></td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        ))}

        <div>
          <label>Measure React perf for </label>
          <button onClick={() => startPerf(2)}>2s</button>
          <button onClick={() => startPerf(5)}>5s</button>
          <button onClick={() => startPerf(10)}>10s</button>
        </div>
      </div>
    );
  }
}

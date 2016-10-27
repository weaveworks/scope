/* eslint react/jsx-no-bind: "off" */
import React from 'react';
import d3 from 'd3';
import _ from 'lodash';
import Perf from 'react-addons-perf';
import { connect } from 'react-redux';
import { fromJS } from 'immutable';

import debug from 'debug';
const log = debug('scope:debug-panel');

import ActionTypes from '../constants/action-types';
import { receiveNodesDelta } from '../actions/app-actions';
import { getNodeColor, getNodeColorDark, text2degree } from '../utils/color-utils';


const SHAPES = ['square', 'hexagon', 'heptagon', 'circle'];
const NODE_COUNTS = [1, 2, 3];
const STACK_VARIANTS = [false, true];
const METRIC_FILLS = [0, 0.1, 50, 99.9, 100];
const NETWORKS = [
  'be', 'fe', 'zb', 'db', 're', 'gh', 'jk', 'lol', 'nw'
].map(n => ({id: n, label: n, colorKey: n}));

const LOREM = `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation
ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in
voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non
proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`;

const sample = (collection, n = 4) => _.sampleSize(collection, _.random(n));


const shapeTypes = {
  square: ['Process', 'Processes'],
  hexagon: ['Container', 'Containers'],
  heptagon: ['Pod', 'Pods'],
  circle: ['Host', 'Hosts']
};


const LABEL_PREFIXES = _.range('A'.charCodeAt(), 'Z'.charCodeAt() + 1)
  .map(n => String.fromCharCode(n));


const deltaAdd = (
  name, adjacency = [], shape = 'circle', stack = false, nodeCount = 1,
    networks = NETWORKS
) => ({
  adjacency,
  controls: {},
  shape,
  stack,
  node_count: nodeCount,
  id: name,
  label: name,
  labelMinor: name,
  latest: {},
  origins: [],
  rank: name,
  networks
});


function addMetrics(availableMetrics, node, v) {
  const metrics = availableMetrics.size > 0 ? availableMetrics : fromJS([
    {id: 'host_cpu_usage_percent', label: 'CPU'}
  ]);

  return Object.assign({}, node, {
    metrics: metrics.map(m => Object.assign({}, m, {label: 'zing', max: 100, value: v})).toJS()
  });
}


function label(shape, stacked) {
  const type = shapeTypes[shape];
  return stacked ? `Group of ${type[1]}` : type[0];
}


function addAllVariants(dispatch) {
  const newNodes = _.flattenDeep(STACK_VARIANTS.map(stack => (SHAPES.map(s => {
    if (!stack) return [deltaAdd(label(s, stack), [], s, stack, 1)];
    return NODE_COUNTS.map(n => deltaAdd(label(s, stack), [], s, stack, n));
  }))));

  dispatch(receiveNodesDelta({
    add: newNodes
  }));
}


function addAllMetricVariants(availableMetrics) {
  const newNodes = _.flattenDeep(METRIC_FILLS.map((v, i) => (
    SHAPES.map(s => [addMetrics(availableMetrics, deltaAdd(label(s) + i, [], s), v)])
  )));

  return (dispatch) => {
    dispatch(receiveNodesDelta({
      add: newNodes
    }));
  };
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


function setAppState(fn) {
  return (dispatch) => {
    dispatch({
      type: ActionTypes.DEBUG_TOOLBAR_INTERFERING,
      fn
    });
  };
}


class DebugToolbar extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.onChange = this.onChange.bind(this);
    this.toggleColors = this.toggleColors.bind(this);
    this.addNodes = this.addNodes.bind(this);
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

  asyncDispatch(v) {
    setTimeout(() => this.props.dispatch(v), 0);
  }

  setLoading(loading) {
    this.asyncDispatch(setAppState(state => state.set('topologiesLoaded', !loading)));
  }

  updateAdjacencies() {
    const ns = this.props.nodes;
    const nodeNames = ns.keySeq().toJS();
    this.asyncDispatch(receiveNodesDelta({
      add: this._addNodes(7),
      update: sample(nodeNames).map(n => ({
        id: n,
        adjacency: sample(nodeNames),
      }), nodeNames.length),
      remove: this._removeNode(),
    }));
  }

  _addNodes(n, prefix = 'zing') {
    const ns = this.props.nodes;
    const nodeNames = ns.keySeq().toJS();
    const newNodeNames = _.range(ns.size, ns.size + n).map(i => (
      // `${randomLetter()}${randomLetter()}-zing`
      `${prefix}${i}`
    ));
    const allNodes = _(nodeNames).concat(newNodeNames).value();
    return newNodeNames.map((name) => deltaAdd(
      name,
      sample(allNodes),
      _.sample(SHAPES),
      _.sample(STACK_VARIANTS),
      _.sample(NODE_COUNTS),
      sample(NETWORKS, 10)
    ));
  }

  addNodes(n, prefix = 'zing') {
    setTimeout(() => {
      this.asyncDispatch(receiveNodesDelta({
        add: this._addNodes(n, prefix)
      }));
      log('added nodes', n);
    }, 0);
  }

  _removeNode() {
    const ns = this.props.nodes;
    const nodeNames = ns.keySeq().toJS();
    return [nodeNames[_.random(nodeNames.length - 1)]];
  }

  removeNode() {
    this.asyncDispatch(receiveNodesDelta({
      remove: this._removeNode()
    }));
  }

  render() {
    const { availableCanvasMetrics } = this.props;

    return (
      <div className="debug-panel">
        <div>
          <label>Add nodes </label>
          <button onClick={() => this.addNodes(1)}>+1</button>
          <button onClick={() => this.addNodes(10)}>+10</button>
          <input type="number" onChange={this.onChange} value={this.state.nodesToAdd} />
          <button onClick={() => this.addNodes(this.state.nodesToAdd)}>+</button>
          <button onClick={() => this.asyncDispatch(addAllVariants)}>Variants</button>
          <button onClick={() => this.asyncDispatch(addAllMetricVariants(availableCanvasMetrics))}>
            Metric Variants
          </button>
          <button onClick={() => this.addNodes(1, LOREM)}>Long name</button>
          <button onClick={() => this.removeNode()}>Remove node</button>
          <button onClick={() => this.updateAdjacencies()}>Update adj.</button>
        </div>

        <div>
          <label>Logging</label>
          <button onClick={() => enableLog('*')}>scope:*</button>
          <button onClick={() => enableLog('dispatcher')}>scope:dispatcher</button>
          <button onClick={() => enableLog('app-key-press')}>scope:app-key-press</button>
          <button onClick={() => enableLog('terminal')}>scope:terminal</button>
          <button onClick={() => disableLog()}>Disable log</button>
        </div>

        <div>
          <label>Colors</label>
          <button onClick={this.toggleColors}>toggle</button>
        </div>

        {this.state.showColors &&
        <table>
          <tbody>
            {LABEL_PREFIXES.map(r => (
              <tr key={r}>
                <td
                  title={`${r}`}
                  style={{backgroundColor: d3.hsl(text2degree(r), 0.5, 0.5).toString()}} />
              </tr>
            ))}
          </tbody>
        </table>}

        {this.state.showColors && [getNodeColor, getNodeColorDark].map((fn, i) => (
          <table key={i}>
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
          <label>state</label>
          <button onClick={() => this.setLoading(true)}>Set doing initial load</button>
          <button onClick={() => this.setLoading(false)}>Stop</button>
        </div>

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


function mapStateToProps(state) {
  return {
    nodes: state.get('nodes'),
    availableCanvasMetrics: state.get('availableCanvasMetrics')
  };
}


export default connect(
  mapStateToProps
)(DebugToolbar);

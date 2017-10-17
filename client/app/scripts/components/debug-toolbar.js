/* eslint react/jsx-no-bind: "off" */
import React from 'react';
import Perf from 'react-addons-perf';
import { connect } from 'react-redux';
import { sampleSize, sample, random, range, flattenDeep, times } from 'lodash';
import { fromJS, Set as makeSet } from 'immutable';
import { hsl } from 'd3-color';
import debug from 'debug';

import ActionTypes from '../constants/action-types';
import { receiveNodesDelta } from '../actions/app-actions';
import { getNodeColor, getNodeColorDark, text2degree } from '../utils/color-utils';
import { availableMetricsSelector } from '../selectors/node-metric';


const SHAPES = ['square', 'hexagon', 'heptagon', 'circle'];
const STACK_VARIANTS = [false, true];
const METRIC_FILLS = [0, 0.1, 50, 99.9, 100];
const NETWORKS = [
  'be', 'fe', 'zb', 'db', 're', 'gh', 'jk', 'lol', 'nw'
].map(n => ({id: n, label: n, colorKey: n}));

const INTERNET = 'the-internet';
const LOREM = `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation
ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in
voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non
proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`;

const sampleArray = (collection, n = 4) => sampleSize(collection, random(n));
const log = debug('scope:debug-panel');

const shapeTypes = {
  square: ['Process', 'Processes'],
  hexagon: ['Container', 'Containers'],
  heptagon: ['Pod', 'Pods'],
  circle: ['Host', 'Hosts']
};


const LABEL_PREFIXES = range('A'.charCodeAt(), 'Z'.charCodeAt() + 1)
  .map(n => String.fromCharCode(n));


const deltaAdd = (name, adjacency = [], shape = 'circle', stack = false, networks = NETWORKS) => ({
  adjacency,
  controls: {},
  shape,
  stack,
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
    metrics: metrics.map(m => Object.assign({}, m, {
      id: 'zing', label: 'zing', max: 100, value: v
    })).toJS()
  });
}


function label(shape, stacked) {
  const type = shapeTypes[shape];
  return stacked ? `Group of ${type[1]}` : type[0];
}


function addAllVariants(dispatch) {
  const newNodes = flattenDeep(STACK_VARIANTS.map(stack => (SHAPES.map((s) => {
    if (!stack) return [deltaAdd(label(s, stack), [], s, stack)];
    return times(3).map(() => deltaAdd(label(s, stack), [], s, stack));
  }))));

  dispatch(receiveNodesDelta({
    add: newNodes
  }));
}


function addAllMetricVariants(availableMetrics) {
  const newNodes = flattenDeep(METRIC_FILLS.map((v, i) => (
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
    this.intermittentTimer = null;
    this.intermittentNodes = makeSet();
    this.shortLivedTimer = null;
    this.shortLivedNodes = makeSet();
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

  setIntermittent() {
    // simulate epheremal nodes
    if (this.intermittentTimer) {
      clearInterval(this.intermittentTimer);
      this.intermittentTimer = null;
    } else {
      this.intermittentTimer = setInterval(() => {
        // add new node
        this.addNodes(1);

        // remove random node
        const ns = this.props.nodes;
        const nodeNames = ns.keySeq().toJS();
        const randomNode = sample(nodeNames);
        this.asyncDispatch(receiveNodesDelta({
          remove: [randomNode]
        }));
      }, 1000);
    }
  }

  setShortLived() {
    // simulate nodes with same ID popping in and out
    if (this.shortLivedTimer) {
      clearInterval(this.shortLivedTimer);
      this.shortLivedTimer = null;
    } else {
      this.shortLivedTimer = setInterval(() => {
        // filter random node
        const ns = this.props.nodes;
        const nodeNames = ns.keySeq().toJS();
        const randomNode = sample(nodeNames);
        if (randomNode) {
          let nextNodes = ns.setIn([randomNode, 'filtered'], true);
          this.shortLivedNodes = this.shortLivedNodes.add(randomNode);
          // bring nodes back after a bit
          if (this.shortLivedNodes.size > 5) {
            const returningNode = this.shortLivedNodes.first();
            this.shortLivedNodes = this.shortLivedNodes.rest();
            nextNodes = nextNodes.setIn([returningNode, 'filtered'], false);
          }
          this.asyncDispatch(setAppState(state => state.set('nodes', nextNodes)));
        }
      }, 1000);
    }
  }

  updateAdjacencies() {
    const ns = this.props.nodes;
    const nodeNames = ns.keySeq().toJS();
    this.asyncDispatch(receiveNodesDelta({
      add: this.createRandomNodes(7),
      update: sampleArray(nodeNames).map(n => ({
        id: n,
        adjacency: sampleArray(nodeNames),
      }), nodeNames.length),
      remove: this.randomExistingNode(),
    }));
  }

  createRandomNodes(n, prefix = 'zing') {
    const ns = this.props.nodes;
    const nodeNames = ns.keySeq().toJS();
    const newNodeNames = range(ns.size, ns.size + n).map(i => (
      // `${randomLetter()}${randomLetter()}-zing`
      `${prefix}${i}`
    ));
    const allNodes = nodeNames.concat(newNodeNames);
    return newNodeNames.map(name => deltaAdd(
      name,
      sampleArray(allNodes),
      sample(SHAPES),
      sample(STACK_VARIANTS),
      sampleArray(NETWORKS, 10)
    ));
  }

  addInternetNode() {
    setTimeout(() => {
      this.asyncDispatch(receiveNodesDelta({
        add: [{id: INTERNET, label: INTERNET, pseudo: true, labelMinor: 'Outgoing packets', shape: 'cloud'}]
      }));
    }, 0);
  }

  addNodes(n, prefix = 'zing') {
    setTimeout(() => {
      this.asyncDispatch(receiveNodesDelta({
        add: this.createRandomNodes(n, prefix)
      }));
      log('added nodes', n);
    }, 0);
  }

  randomExistingNode() {
    const ns = this.props.nodes;
    const nodeNames = ns.keySeq().toJS();
    return [nodeNames[random(nodeNames.length - 1)]];
  }

  removeNode() {
    this.asyncDispatch(receiveNodesDelta({
      remove: this.randomExistingNode()
    }));
  }

  render() {
    const { availableMetrics } = this.props;

    return (
      <div className="debug-panel">
        <div>
          <strong>Add nodes </strong>
          <button onClick={() => this.addNodes(1)}>+1</button>
          <button onClick={() => this.addNodes(10)}>+10</button>
          <input type="number" onChange={this.onChange} value={this.state.nodesToAdd} />
          <button onClick={() => this.addNodes(this.state.nodesToAdd)}>+</button>
          <button onClick={() => this.asyncDispatch(addAllVariants)}>Variants</button>
          <button onClick={() => this.asyncDispatch(addAllMetricVariants(availableMetrics))}>
            Metric Variants
          </button>
          <button onClick={() => this.addNodes(1, LOREM)}>Long name</button>
          <button onClick={() => this.addInternetNode()}>Internet</button>
          <button onClick={() => this.removeNode()}>Remove node</button>
          <button onClick={() => this.updateAdjacencies()}>Update adj.</button>
        </div>

        <div>
          <strong>Logging </strong>
          <button onClick={() => enableLog('*')}>scope:*</button>
          <button onClick={() => enableLog('dispatcher')}>scope:dispatcher</button>
          <button onClick={() => enableLog('app-key-press')}>scope:app-key-press</button>
          <button onClick={() => enableLog('terminal')}>scope:terminal</button>
          <button onClick={() => disableLog()}>Disable log</button>
        </div>

        <div>
          <strong>Colors </strong>
          <button onClick={this.toggleColors}>toggle</button>
        </div>

        {this.state.showColors &&
        <table>
          <tbody>
            {LABEL_PREFIXES.map(r => (
              <tr key={r}>
                <td
                  title={`${r}`}
                  style={{backgroundColor: hsl(text2degree(r), 0.5, 0.5).toString()}} />
              </tr>
            ))}
          </tbody>
        </table>}

        {this.state.showColors && [getNodeColor, getNodeColorDark].map(fn => (
          <table key={fn}>
            <tbody>
              {LABEL_PREFIXES.map(r => (
                <tr key={r}>
                  {LABEL_PREFIXES.map(c => (
                    <td key={c} title={`(${r}, ${c})`} style={{backgroundColor: fn(r, c)}} />
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        ))}

        <div>
          <strong>State </strong>
          <button onClick={() => this.setLoading(true)}>Set doing initial load</button>
          <button onClick={() => this.setLoading(false)}>Stop</button>
        </div>

        <div>
          <strong>Short-lived nodes </strong>
          <button onClick={() => this.setShortLived()}>Toggle short-lived nodes</button>
          <button onClick={() => this.setIntermittent()}>Toggle intermittent nodes</button>
        </div>

        <div>
          <strong>Measure React perf for </strong>
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
    availableMetrics: availableMetricsSelector(state),
  };
}


export default connect(mapStateToProps)(DebugToolbar);

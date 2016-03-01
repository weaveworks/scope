import _ from 'lodash';
import d3 from 'd3';


// Inspired by Lee Byron's test data generator.
function bumpLayer(n, maxValue) {
  function bump(a) {
    const x = 1 / (0.1 + Math.random());
    const y = 2 * Math.random() - 0.5;
    const z = 10 / (0.1 + Math.random());
    for (let i = 0; i < n; i++) {
      const w = (i / n - y) * z;
      a[i] += x * Math.exp(-w * w);
    }
  }

  const a = [];
  let i;
  for (i = 0; i < n; ++i) a[i] = 0;
  for (i = 0; i < 5; ++i) bump(a);
  const values = a.map(function(d) { return Math.max(0, d * maxValue); });
  const s = d3.scale.linear().domain(d3.extent(values)).range([0, maxValue]);
  return values.map(s);
}


const nodeData = {};


function getNextValue(keyValues, maxValue) {
  const key = keyValues.join('-');
  let series = nodeData[key];
  if (!series || !series.length) {
    series = nodeData[key] = bumpLayer(100, maxValue);
  }
  const v = series.shift();
  return v;
}


function mergeMetrics(node) {
  return Object.assign({}, node, {
    metrics: {
      'process_cpu_usage_percent': {
        samples: [{value: getNextValue([node.id, 'cpu'], 100)}],
        max: 100
      },
      'memory': {
        samples: [{value: getNextValue([node.id, 'memory'], 1024)}],
        max: 1024
      }
    }
  });
}


function handleAdd(nodes) {
  if (!nodes) {
    return nodes;
  }
  return nodes.map(mergeMetrics);
}


function handleUpdated(updatedNodes, prevNodes) {
  const modifiedNodesIndex = _.zipObject((updatedNodes || []).map(n => [n.id, n]));
  return prevNodes.toIndexedSeq().toJS().map(n => {
    return Object.assign({}, mergeMetrics(n), modifiedNodesIndex[n.id]);
  });
}


export function addMetrics(delta, prevNodes) {
  return Object.assign({}, delta, {
    add: handleAdd(delta.add),
    update: handleUpdated(delta.update, prevNodes)
  });
}


export function getMetricValue(metrics, size) {
  if (!metrics) {
    return {height: 0, vp: null};
  }

  const max = 100;
  const v = metrics.getIn(['process_cpu_usage_percent', 'samples', 0, 'value']);
  const vp = v === 0 ? 0 : v / max;
  const baseline = 0.00;
  const displayedValue = vp * (1 - baseline) + baseline;
  const height = size * displayedValue;

  return {height, vp};
}

const formatLargeValue = d3.format('s');
export function formatCanvasMetric(v) {
  if (v === null) {
    return 'n/a';
  }
  return formatLargeValue(Number(v * 100).toFixed(1)) + '%';
}

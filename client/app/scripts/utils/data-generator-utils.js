import { zipObject } from 'lodash';
import { scaleLinear } from 'd3-scale';
import { extent } from 'd3-array';

// Inspired by Lee Byron's test data generator.
/* eslint-disable */
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
  const s = scaleLinear().domain(extent(values)).range([0, maxValue]);
  return values.map(s);
}
/* eslint-enable */


const nodeData = {};


function getNextValue(keyValues, maxValue) {
  const key = keyValues.join('-');
  let series = nodeData[key];
  if (!series || !series.length) {
    nodeData[key] = bumpLayer(100, maxValue);
    series = nodeData[key];
  }
  const v = series.shift();
  return v;
}

export const METRIC_LABELS = {
  docker_cpu_total_usage: 'CPU',
  docker_memory_usage: 'Memory',
  host_cpu_usage_percent: 'CPU',
  host_mem_usage_bytes: 'Memory',
  load1: 'Load 1',
  load5: 'Load 5',
  load15: 'Load 15',
  open_files_count: 'Open files',
  process_cpu_usage_percent: 'CPU',
  process_memory_usage_bytes: 'Memory'
};


export function label(m) {
  return METRIC_LABELS[m.id];
}


const memoryMetric = (node, name, max = 1024 * 1024 * 1024) => ({
  max,
  samples: [{value: getNextValue([node.id, name], max)}]
});

const cpuMetric = (node, name, max = 100) => ({
  max,
  samples: [{value: getNextValue([node.id, name], max)}]
});

const fileMetric = (node, name, max = 1000) => ({
  max,
  samples: [{value: getNextValue([node.id, name], max)}]
});

const loadMetric = (node, name, max = 10) => ({
  max,
  samples: [{value: getNextValue([node.id, name], max)}]
});

const metrics = {
  // host
  circle: {
    host_cpu_usage_percent: cpuMetric,
    host_mem_usage_bytes: memoryMetric,
    load5: loadMetric
  },
  // container
  hexagon: {
    docker_cpu_total_usage: cpuMetric,
    docker_memory_usage: memoryMetric
  },
  // process
  square: {
    open_files_count: fileMetric,
    process_cpu_usage_percent: cpuMetric,
    process_memory_usage_bytes: memoryMetric
  }
};


function mergeMetrics(node) {
  if (node.pseudo || node.stack) {
    return node;
  }
  return Object.assign({}, node, {
    metrics: (metrics[node.shape] || [])
      .map((fn, name) => [name, fn(node)])
      .fromPairs()
  });
}


function handleAdd(nodes) {
  if (!nodes) {
    return nodes;
  }
  return nodes.map(mergeMetrics);
}


function handleUpdated(updatedNodes, prevNodes) {
  const modifiedNodesIndex = zipObject((updatedNodes || []).map(n => [n.id, n]));
  return prevNodes.toIndexedSeq().toJS().map(n => (
    Object.assign({}, mergeMetrics(n), modifiedNodesIndex[n.id])
  ));
}


export function addMetrics(delta, prevNodes) {
  return Object.assign({}, delta, {
    add: handleAdd(delta.add),
    update: handleUpdated(delta.update, prevNodes)
  });
}

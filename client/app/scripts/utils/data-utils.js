import _ from 'lodash';
import d3 from 'd3';
import { formatMetric } from './string-utils';
import AppStore from '../stores/app-store';


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

const METRIC_FORMATS = {
  docker_cpu_total_usage: 'percent',
  docker_memory_usage: 'filesize',
  host_cpu_usage_percent: 'percent',
  host_mem_usage_bytes: 'filesize',
  load1: 'number',
  load15: 'number',
  load5: 'number',
  open_files_count: 'number',
  process_cpu_usage_percent: 'percent',
  process_memory_usage_bytes: 'filesize'
};

const openFilesScale = d3.scale.log().domain([1, 100000]).range([0, 1]);
//
// loadScale(1) == 0.5; E.g. a nicely balanced system :).
const loadScale = d3.scale.log().domain([0.01, 100]).range([0, 1]);

export function getMetricValue(metric, size) {
  if (!metric) {
    return {height: 0, value: null, formattedValue: 'n/a'};
  }

  const max = metric.getIn(['max']);
  const value = metric.getIn(['samples', 0, 'value']);
  const selectedMetric = AppStore.getSelectedMetric();

  let valuePercentage = value === 0 ? 0 : value / max;
  if (selectedMetric === 'open_files_count') {
    valuePercentage = openFilesScale(value);
  } else if (_.includes(['load1', 'load5', 'load15'], selectedMetric)) {
    valuePercentage = loadScale(value);
  }

  let displayedValue = value;
  if (displayedValue > 0) {
    const baseline = 0.1;
    displayedValue = valuePercentage * (1 - baseline) + baseline;
  }
  const height = size * displayedValue;
  const metricWithFormat = Object.assign(
    {}, {format: METRIC_FORMATS[selectedMetric]}, metric.toJS());

  return {
    height: height,
    value: value,
    formattedValue: formatMetric(value, metricWithFormat, true)
  };
}

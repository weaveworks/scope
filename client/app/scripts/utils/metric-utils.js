import _ from 'lodash';
import d3 from 'd3';
import { formatMetric } from './string-utils';
import { colors } from './color-utils';


const openFilesScale = d3.scale.log().domain([1, 100000]).range([0, 1]);
//
// loadScale(1) == 0.5; E.g. a nicely balanced system :).
const loadScale = d3.scale.log().domain([0.01, 100]).range([0, 1]);

export function getMetricValue(metric, size) {
  if (!metric) {
    return {height: 0, value: null, formattedValue: 'n/a'};
  }
  const m = metric.toJS();
  const value = m.value;

  let valuePercentage = value === 0 ? 0 : value / m.max;
  if (m.id === 'open_files_count') {
    valuePercentage = openFilesScale(value);
  } else if (_.includes(['load1', 'load5', 'load15'], m.id)) {
    valuePercentage = loadScale(value);
  }

  let displayedValue = Number(value).toFixed(1);
  if (displayedValue > 0) {
    const baseline = 0.1;
    displayedValue = valuePercentage * (1 - baseline) + baseline;
  }
  const height = size * displayedValue;

  return {
    height,
    value,
    formattedValue: formatMetric(value, m, true)
  };
}


export function getMetricColor(metric) {
  const selectedMetric = metric && metric.get('id');
  // bluey
  if (/memory/.test(selectedMetric)) {
    return '#1f77b4';
  } else if (/cpu/.test(selectedMetric)) {
    return colors('cpu');
  } else if (/files/.test(selectedMetric)) {
    // return colors('files');
    // purple
    return '#9467bd';
  } else if (/load/.test(selectedMetric)) {
    return colors('load');
  }
  return 'steelBlue';
}

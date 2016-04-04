import _ from 'lodash';
import d3 from 'd3';
import { formatMetricSvg } from './string-utils';
import { getNodeColorDark as colors } from './color-utils';
import React from 'react';


export function getClipPathDefinition(clipId, size, height,
                                      x = -size * 0.5, y = size * 0.5 - height) {
  return (
    <defs>
      <clipPath id={clipId}>
        <rect
          width={size}
          height={size}
          x={x}
          y={y}
          />
      </clipPath>
    </defs>
  );
}


//
// Open files, 100k should be enought for anyone?
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
  let max = m.max;
  if (m.id === 'open_files_count') {
    valuePercentage = openFilesScale(value);
    max = null;
  } else if (_.includes(['load1', 'load5', 'load15'], m.id)) {
    valuePercentage = loadScale(value);
    max = null;
  }

  let displayedValue = Number(value).toFixed(1);
  if (displayedValue > 0 && (!max || displayedValue < max)) {
    const baseline = 0.1;
    displayedValue = valuePercentage * (1 - baseline * 2) + baseline;
  } else if (displayedValue >= m.max && displayedValue > 0) {
    displayedValue = 1;
  }
  const height = size * displayedValue;

  return {
    height,
    hasMetric: value !== null,
    formattedValue: formatMetricSvg(value, m)
  };
}


export function getMetricColor(metric) {
  const selectedMetric = metric && metric.get('id');
  if (/mem/.test(selectedMetric)) {
    return colors('p', 'a');
  } else if (/cpu/.test(selectedMetric)) {
    return colors('z', 'a');
  } else if (/files/.test(selectedMetric)) {
    return colors('t', 'a');
  } else if (/load/.test(selectedMetric)) {
    return colors('a', 'a');
  }
  return 'steelBlue';
}

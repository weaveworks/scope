import { includes } from 'lodash';
import { scaleLog } from 'd3-scale';
import React from 'react';

import { NODE_BASE_SIZE, NODE_SHAPE_DOT_RADIUS } from '../constants/styles';
import { formatMetricSvg } from './string-utils';
import { colors } from './color-utils';

export function getClipPathDefinition(clipId, height) {
  return (
    <defs>
      <clipPath id={clipId} transform={`scale(${NODE_BASE_SIZE})`}>
        <rect width={1} height={1} x={-0.5} y={0.5 - height} />
      </clipPath>
    </defs>
  );
}

export function renderMetricValue(value, condition) {
  return condition ? <text>{value}</text> : <circle className="node" r={NODE_SHAPE_DOT_RADIUS} />;
}

//
// loadScale(1) == 0.5; E.g. a nicely balanced system :).
const loadScale = scaleLog().domain([0.01, 100]).range([0, 1]);

// Used in the graph view
export function getMetricValue(metric) {
  if (!metric) {
    return {height: 0, value: null, formattedValue: 'n/a'};
  }
  const m = metric.toJS();
  const value = m.value;

  let valuePercentage = value === 0 ? 0 : value / m.max;
  let max = m.max;
  if (includes(['load1', 'load5', 'load15'], m.id)) {
    valuePercentage = loadScale(value);
    max = null;
  }

  let displayedValue = Number(value).toFixed(1);
  if (displayedValue > 0 && (!max || displayedValue < max)) {
    const baseline = 0.1;
    displayedValue = (valuePercentage * (1 - (baseline * 2))) + baseline;
  } else if (displayedValue >= m.max && displayedValue > 0) {
    displayedValue = 1;
  }

  return {
    height: displayedValue,
    hasMetric: value !== null,
    formattedValue: formatMetricSvg(value, m)
  };
}

// Used in the resource view
export function getHumanizedMetricInfo(metric) {
  const showExtendedInfo = metric.get('withCapacity') && metric.get('format') !== 'percent';
  const totalCapacity = formatMetricSvg(metric.get('totalCapacity'), metric.toJS());
  const absoluteConsumption = formatMetricSvg(metric.get('absoluteConsumption'), metric.toJS());
  const relativeConsumption = formatMetricSvg(100.0 * metric.get('relativeConsumption'),
    { format: 'percent' });
  return (
    <span>
      <strong>
        {showExtendedInfo ? relativeConsumption : absoluteConsumption}
      </strong> consumed
      {showExtendedInfo && <i>{' - '}
        ({absoluteConsumption} / <strong>{totalCapacity}</strong>)
      </i>}
    </span>
  );
}

export function getMetricColor(metric) {
  const selectedMetric = metric && metric.get('id');
  if (/mem/.test(selectedMetric)) {
    return 'steelBlue';
  } else if (/cpu/.test(selectedMetric)) {
    return colors('cpu');
  } else if (/files/.test(selectedMetric)) {
    // purple
    return '#9467bd';
  } else if (/load/.test(selectedMetric)) {
    return colors('load');
  }
  return 'steelBlue';
}

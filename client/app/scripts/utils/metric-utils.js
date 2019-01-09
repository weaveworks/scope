import { includes } from 'lodash';
import { scaleLog } from 'd3-scale';
import React from 'react';

import { formatMetricSvg } from './string-utils';
import { colors } from './color-utils';

export function getClipPathDefinition(clipId, height, radius) {
  const barHeight = 1 - (2 * height); // in the interval [-1, 1]
  return (
    <defs>
      <clipPath id={clipId} transform={`scale(${2 * radius})`}>
        <rect width={2} height={2} x={-1} y={barHeight} />
      </clipPath>
    </defs>
  );
}

//
// loadScale(1) == 0.5; E.g. a nicely balanced system :).
const loadScale = scaleLog().domain([0.01, 100]).range([0, 1]);


export function getMetricValue(metric) {
  if (!metric) {
    return { formattedValue: 'n/a', height: 0, value: null };
  }
  const m = metric.toJS();
  const { value } = m;

  let valuePercentage = value === 0 ? 0 : value / m.max;
  let { max } = m;
  if (includes(['load1', 'load5', 'load15'], m.id)) {
    valuePercentage = loadScale(value);
    max = null;
  }

  let displayedValue = Number(value);
  if (displayedValue > 0 && (!max || displayedValue < max)) {
    const baseline = 0.1;
    displayedValue = (valuePercentage * (1 - (baseline * 2))) + baseline;
  } else if (displayedValue >= m.max && displayedValue > 0) {
    displayedValue = 1;
  }

  return {
    formattedValue: formatMetricSvg(value, m),
    hasMetric: value !== null,
    height: displayedValue
  };
}


export function getMetricColor(metric) {
  const metricId = typeof metric === 'string'
    ? metric
    : metric && metric.get('id');
  if (/mem/.test(metricId)) {
    return 'steelBlue';
  } else if (/cpu/.test(metricId)) {
    return colors('cpu').toString();
  } else if (/files/.test(metricId)) {
    // purple
    return '#9467bd';
  } else if (/load/.test(metricId)) {
    return colors('load').toString();
  }
  return 'steelBlue';
}

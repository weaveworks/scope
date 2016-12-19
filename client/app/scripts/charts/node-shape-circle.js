import React from 'react';
import classNames from 'classnames';

import { getMetricValue, getMetricColor, getClipPathDefinition } from '../utils/metric-utils';
import { CANVAS_METRIC_FONT_SIZE } from '../constants/styles';
import { getBorderProps } from './node';

export default function NodeShapeCircle({id, highlighted, size, color, metric}) {
  const clipId = `mask-${id}`;
  const {height, hasMetric, formattedValue} = getMetricValue(metric, size);
  const metricStyle = { fill: getMetricColor(metric) };
  const className = classNames('shape', { metrics: hasMetric });
  const fontSize = size * CANVAS_METRIC_FONT_SIZE;

  return (
    <g className={className}>
      {hasMetric && getClipPathDefinition(clipId, size, height)}
      {highlighted && <circle r={size * 0.7} className="highlighted" />}
      <circle r={size * 0.5} className="border" stroke={color} {...getBorderProps()} />
      <circle r={size * 0.45} className="shadow" />
      {hasMetric && <circle
        r={size * 0.45}
        className="metric-fill"
        style={metricStyle}
        clipPath={`url(#${clipId})`}
      />}
      {highlighted && hasMetric ?
        <text style={{fontSize}}>{formattedValue}</text> :
        <circle className="node" r={Math.max(2, (size * 0.125))} />}
    </g>
  );
}

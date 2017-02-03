import React from 'react';
import classNames from 'classnames';

import {
  getMetricValue,
  getMetricColor,
  getClipPathDefinition,
  renderMetricValue,
} from '../utils/metric-utils';
import {
  NODE_SHAPE_HIGHLIGHT_RADIUS,
  NODE_SHAPE_BORDER_RADIUS,
  NODE_SHAPE_SHADOW_RADIUS,
} from '../constants/styles';


export default function NodeShapeCircle({ id, highlighted, color, metric }) {
  const { height, hasMetric, formattedValue } = getMetricValue(metric);
  const metricStyle = { fill: getMetricColor(metric) };

  const className = classNames('shape', 'shape-circle', { metrics: hasMetric });
  const clipId = `mask-${id}`;

  return (
    <g className={className}>
      {hasMetric && getClipPathDefinition(clipId, height)}
      {highlighted && <circle className="highlighted" r={NODE_SHAPE_HIGHLIGHT_RADIUS} />}
      <circle className="border" stroke={color} r={NODE_SHAPE_BORDER_RADIUS} />
      <circle className="shadow" r={NODE_SHAPE_SHADOW_RADIUS} />
      {hasMetric && <circle
        className="metric-fill"
        clipPath={`url(#${clipId})`}
        style={metricStyle}
        r={NODE_SHAPE_SHADOW_RADIUS}
      />}
      {renderMetricValue(formattedValue, highlighted && hasMetric)}
    </g>
  );
}

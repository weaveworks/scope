import React from 'react';
import classNames from 'classnames';
import {getMetricValue, getMetricColor, getClipPathDefinition} from '../utils/metric-utils.js';
import {CANVAS_METRIC_FONT_SIZE} from '../constants/styles.js';


export default function NodeShapeCircle({
  id, highlighted, size, metric, color, lightColor, darkColor
}) {
  const clipId = `mask-${id}`;
  const {height, hasMetric, formattedValue} = getMetricValue(metric, size);
  const metricStyle = { fill: getMetricColor(metric) };
  const className = classNames('shape', { metrics: hasMetric });
  const fontSize = size * CANVAS_METRIC_FONT_SIZE;

  return (
    <g className={className}>
      {hasMetric && getClipPathDefinition(clipId, size, height)}
      {highlighted && <circle r={size * 0.7} className="highlighted" />}
      <circle r={size * 0.535} className="outline" fill="none" stroke={darkColor} />
      <circle r={size * 0.5} className="border" fill={lightColor} stroke={color} />
      {hasMetric && <circle r={size * 0.45} className="metric-fill" style={metricStyle}
        clipPath={`url(#${clipId})`} />}
      {highlighted && hasMetric ?
        <text style={{fontSize}}>{formattedValue}</text> :
        <circle className="node" r={Math.max(1.33333, (size * 0.08333))} fill={darkColor} />}
    </g>
  );
}

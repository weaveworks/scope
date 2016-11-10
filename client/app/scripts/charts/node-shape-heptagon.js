import React from 'react';
import d3 from 'd3';
import classNames from 'classnames';
import {getMetricValue, getMetricColor, getClipPathDefinition} from '../utils/metric-utils.js';
import {CANVAS_METRIC_FONT_SIZE} from '../constants/styles.js';


const line = d3.svg.line()
  .interpolate('cardinal-closed')
  .tension(0.25);


function polygon(r, sides) {
  const a = (Math.PI * 2) / sides;
  const points = [[r, 0]];
  for (let i = 1; i < sides; i++) {
    points.push([r * Math.cos(a * i), r * Math.sin(a * i)]);
  }
  return points;
}


export default function NodeShapeHeptagon({
  id, highlighted, size, metric, color, lightColor, darkColor
}) {
  const scaledSize = size * 1.0;
  const pathProps = v => ({
    d: line(polygon(scaledSize * v, 7)),
    transform: 'rotate(90)'
  });

  const clipId = `mask-${id}`;
  const {height, hasMetric, formattedValue} = getMetricValue(metric, size);
  const metricStyle = { fill: getMetricColor(metric) };
  const className = classNames('shape', { metrics: hasMetric });
  const fontSize = size * CANVAS_METRIC_FONT_SIZE;

  return (
    <g className={className}>
      {hasMetric && getClipPathDefinition(clipId, size, height, size * 0.5 - height, -size * 0.5)}
      {highlighted && <path className="highlighted" {...pathProps(0.7)} />}
      <path className="outline" stroke={darkColor} fill="none" {...pathProps(0.535)} />
      <path className="border" stroke={color} fill={lightColor} {...pathProps(0.5)} />
      {hasMetric && <path className="metric-fill" clipPath={`url(#${clipId})`}
        style={metricStyle} {...pathProps(0.45)} />}
      {highlighted && hasMetric ?
        <text style={{fontSize}}>{formattedValue}</text> :
        <circle className="node" r={Math.max(1.33333, (size * 0.08333))} fill={darkColor} />}
    </g>
  );
}

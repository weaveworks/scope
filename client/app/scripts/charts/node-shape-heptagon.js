import React from 'react';
import classNames from 'classnames';
import { line, curveCardinalClosed } from 'd3-shape';
import { getMetricValue, getMetricColor, getClipPathDefinition } from '../utils/metric-utils.js';
import { CANVAS_METRIC_FONT_SIZE } from '../constants/styles.js';


const spline = line()
  .curve(curveCardinalClosed.tension(0.65));


function polygon(r, sides) {
  const a = (Math.PI * 2) / sides;
  const points = [];
  for (let i = 0; i < sides; i++) {
    points.push([r * Math.sin(a * i), -r * Math.cos(a * i)]);
  }
  return points;
}


export default function NodeShapeHeptagon({id, highlighted, size, color, metric}) {
  const scaledSize = size * 1.0;
  const pathProps = v => ({
    d: spline(polygon(scaledSize * v, 7))
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
      <path className="border" stroke={color} {...pathProps(0.5)} />
      <path className="shadow" {...pathProps(0.45)} />
      {hasMetric && <path className="metric-fill" clipPath={`url(#${clipId})`}
        style={metricStyle} {...pathProps(0.45)} />}
      {highlighted && hasMetric ?
        <text style={{fontSize}}>{formattedValue}</text> :
        <circle className="node" r={Math.max(2, (size * 0.125))} />}
    </g>
  );
}

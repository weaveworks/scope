import React from 'react';
import {formatCanvasMetric, getMetricValue} from '../utils/data-utils.js';

export default function NodeShapeSquare({
  highlighted, size, color, rx = 0, ry = 0, metric
}) {
  const rectProps = v => ({
    width: v * size * 2,
    height: v * size * 2,
    rx: v * size * rx,
    ry: v * size * ry,
    x: -size * v,
    y: -size * v
  });

  const clipId = `mask-${Math.random()}`;
  const {height, v} = getMetricValue(metric, size);

  return (
    <g className="shape">
      <defs>
        <clipPath id={clipId}>
          <rect
            width={size}
            height={size}
            x={-size * 0.5}
            y={size * 0.5 - height}
            />
        </clipPath>
      </defs>
      {highlighted && <rect className="highlighted" {...rectProps(0.7)} />}
      <rect className="border" stroke={color} {...rectProps(0.5)} />
      <rect className="shadow" {...rectProps(0.45)} />
      <rect className="metric-fill" clipPath={`url(#${clipId})`} {...rectProps(0.45)} />
      {highlighted ?
        <text dy="0.35em" style={{'textAnchor': 'middle'}}>
          {formatCanvasMetric(v)}
        </text> :
        <circle className="node" r={Math.max(2, (size * 0.125))} />}
    </g>
  );
}

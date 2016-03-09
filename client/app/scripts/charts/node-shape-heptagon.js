import React from 'react';
import d3 from 'd3';
import classNames from 'classnames';
import {getMetricValue} from '../utils/data-utils.js';

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

export default function NodeShapeHeptagon({highlighted, size, color, metric}) {
  const scaledSize = size * 1.0;
  const pathProps = v => ({
    d: line(polygon(scaledSize * v, 7)),
    transform: 'rotate(90)'
  });

  const clipId = `mask-${Math.random()}`;
  const {height, value, formattedValue} = getMetricValue(metric, size);

  const className = classNames('shape', {
    'metrics': value !== null
  });

  return (
    <g className={className}>
      <defs>
        <clipPath id={clipId}>
          <rect
            width={size}
            height={size}
            y={-size * 0.5}
            x={size * 0.5 - height}
            />
        </clipPath>
      </defs>
      {highlighted && <path className="highlighted" {...pathProps(0.7)} />}
      <path className="border" stroke={color} {...pathProps(0.5)} />
      <path className="shadow" {...pathProps(0.45)} />
      <path className="metric-fill" clipPath={`url(#${clipId})`} {...pathProps(0.45)} />
      {highlighted && value !== null ?
        <text dy="0.35em" style={{'textAnchor': 'middle'}}>{formattedValue}</text> :
        <circle className="node" r={Math.max(2, (size * 0.125))} />}
    </g>
  );
}

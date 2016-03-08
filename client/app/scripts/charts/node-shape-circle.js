import React from 'react';
import {getMetricValue} from '../utils/data-utils.js';

export default function NodeShapeCircle({highlighted, size, color, metric}) {
  const hightlightNode = <circle r={size * 0.7} className="highlighted" />;
  const clipId = `mask-${Math.random()}`;
  const {height, formattedValue} = getMetricValue(metric, size);

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
      {highlighted && hightlightNode}
      <circle r={size * 0.5} className="border" stroke={color} />
      <circle r={size * 0.45} className="shadow" />
      <circle r={size * 0.45} className="metric-fill" clipPath={`url(#${clipId})`} />
      {highlighted ?
        <text dy="0.35em" style={{'textAnchor': 'middle'}}>{formattedValue}</text> :
        <circle className="node" r={Math.max(2, (size * 0.125))} />}
    </g>
  );
}

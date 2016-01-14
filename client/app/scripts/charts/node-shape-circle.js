import React from 'react';

export default function NodeShapeCircle({highlighted, size, color}) {
  return (
    <g className="shape">
      {highlighted &&
        <circle r={size * 0.7} className="highlighted"></circle>}

      <circle r={size * 0.5} className="border" stroke={color}></circle>
      <circle r={size * 0.45} className="shadow"></circle>
      <circle r={Math.max(2, size * 0.125)} className="node"></circle>
    </g>
  );
}

import React from 'react';
import d3 from 'd3';

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

export default function NodeShapeSeptagon({onlyHighlight, highlighted, size, color}) {
  const scaledSize = size * 1.0;
  const pathProps = (v) => {
    return {
      d: line(polygon(scaledSize * v, 7)),
      transform: `rotate(90)`
    };
  };

  const hightlightNode = <path className="highlighted" {...pathProps(0.7)} />;

  if (onlyHighlight) {
    return (
      <g className="shape">
        {highlighted && hightlightNode}
      </g>
    );
  }

  return (
    <g className="shape">
      {highlighted && hightlightNode}
      <path className="border" stroke={color} {...pathProps(0.5)} />
      <path className="shadow" {...pathProps(0.45)} />
      <circle className="node" r={Math.max(2, (scaledSize * 0.125))} />
    </g>
  );
}


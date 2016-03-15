import React from 'react';

export default function NodeShapeSquare({onlyHighlight, highlighted, size, color, rx = 0, ry = 0}) {
  const rectProps = v => ({
    width: v * size * 2,
    height: v * size * 2,
    rx: v * size * rx,
    ry: v * size * ry,
    transform: `translate(-${size * v}, -${size * v})`
  });

  const hightlightNode = <rect className="highlighted" {...rectProps(0.7)} />;

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
      <rect className="border" stroke={color} {...rectProps(0.5)} />
      <rect className="shadow" {...rectProps(0.45)} />
      <circle className="node" r={Math.max(2, (size * 0.125))} />
    </g>
  );
}

import React from 'react';

export default function NodeShapeCircle({onlyHighlight, highlighted, size, color}) {
  const hightlightNode = <circle r={size * 0.7} className="highlighted" />;

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
      <circle r={size * 0.5} className="border" stroke={color} />
      <circle r={size * 0.45} className="shadow" />
      <circle r={Math.max(2, size * 0.125)} className="node" />
    </g>
  );
}

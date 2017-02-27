import React from 'react';

export default function Marker({id, offset, size, children, viewBox}) {
  return (
    <marker
      className="edge-marker"
      id={id}
      viewBox={viewBox}
      refX={offset}
      refY="3.5"
      markerWidth={size}
      markerHeight={size}
      orient="auto"
    >
      {children}
    </marker>
  );
}

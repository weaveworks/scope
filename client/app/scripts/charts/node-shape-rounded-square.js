import React from 'react';
import NodeShapeSquare from './node-shape-square';

// TODO how to express a cmp in terms of another cmp? (Rather than a sub-cmp as here).
// HOC!

export default function NodeShapeRoundedSquare(props) {
  return (
    <NodeShapeSquare {...props} rx="0.4" ry="0.4" />
  );
}

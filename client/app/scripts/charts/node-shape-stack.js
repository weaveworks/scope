import React from 'react';

import { NODE_BASE_SIZE } from '../constants/styles';

export default function NodeShapeStack(props) {
  const verticalDistance = NODE_BASE_SIZE * (props.contrastMode ? 0.12 : 0.1);
  const verticalTranslate = t => `translate(0, ${t * verticalDistance})`;
  const Shape = props.shape;

  // Stack three shapes on top of one another pretending they are never highlighted.
  // Instead, fake the highlight of the whole stack with a vertically stretched shape
  // drawn in the background. This seems to give a good approximation of the stack
  // highlight and prevents us from needing to do some render-heavy SVG clipping magic.
  return (
    <g transform={verticalTranslate(-2.5)} className="stack">
      <g transform={`${verticalTranslate(1)} scale(1, 1.14)`}>
        <Shape className="highlight-only" {...props} />
      </g>
      <g transform={verticalTranslate(2)}><Shape {...props} highlighted={false} /></g>
      <g transform={verticalTranslate(1)}><Shape {...props} highlighted={false} /></g>
      <g transform={verticalTranslate(0)}><Shape {...props} highlighted={false} /></g>
    </g>
  );
}

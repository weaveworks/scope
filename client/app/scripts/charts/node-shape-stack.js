import React from 'react';
import _ from 'lodash';

export default function NodeShapeCircleStack(props) {
  const propsNoHighlight = _.clone(props);
  const Shape = props.shape;
  delete propsNoHighlight.highlighted;
  return (
    <g className="stack">
      <g transform="translate(0, 4)">
        <Shape {...propsNoHighlight} />
      </g>
      <Shape {...props} />
      <g transform="translate(0, -4)">
        <Shape {...propsNoHighlight} />
      </g>
    </g>
  );
}

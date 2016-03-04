import React from 'react';
import _ from 'lodash';

import { isContrastMode } from '../utils/contrast-utils';

function dissoc(obj, key) {
  const newObj = _.clone(obj);
  delete newObj[key];
  return newObj;
}

export default function NodeShapeStack(props) {
  const propsNoHighlight = dissoc(props, 'highlighted');
  const propsOnlyHighlight = Object.assign({}, props, {onlyHighlight: true});
  const contrastMode = isContrastMode();

  const Shape = props.shape;
  const [dx, dy] = contrastMode ? [0, 8] : [0, 5];
  const dsx = (props.size * 2 + (dx * 2)) / (props.size * 2);
  const dsy = (props.size * 2 + (dy * 2)) / (props.size * 2);
  const hls = [dsx, dsy];

  return (
    <g transform={`translate(${dx * -1}, ${dy * -2.5})`} className="stack">
      <g transform={`scale(${hls})translate(${dx}, ${dy}) `}>
        <Shape {...propsOnlyHighlight} />
      </g>
      <g transform={`translate(${dx * 2}, ${dy * 2})`}>
        <Shape {...propsNoHighlight} />
      </g>
      <g transform={`translate(${dx * 1}, ${dy * 1})`}>
        <Shape {...propsNoHighlight} />
      </g>
      <Shape {...propsNoHighlight} />
    </g>
  );
}

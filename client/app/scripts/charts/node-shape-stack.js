import React from 'react';
import _ from 'lodash';

function dissoc(obj, key) {
  const newObj = _.clone(obj);
  delete newObj[key];
  return newObj;
}

export default function NodeShapeStack(props) {
  const propsNoHighlight = dissoc(props, 'highlighted');
  const propsOnlyHighlight = Object.assign({}, props, {onlyHighlight: true});

  const Shape = props.shape;
  const nodeCount = props.nodeCount;
  const [dx, dy] = [0, 6];
  const ds = 0.075;
  const dsx = (props.size * 2 + dx) / (props.size * 2);
  const dsy = (props.size * 2 + dy) / (props.size * 2);
  const hls = [dsx, dsy];

  const propsWithGrey = Object.assign({}, propsNoHighlight, {color: '#aaa', className: 'mock'});
  const propsForIndex = i => (nodeCount < i ? propsWithGrey : propsNoHighlight);

  return (
    <g transform={`translate(${dx * -1}, ${dy * -1})`} className="stack">
      <g transform={`translate(${dx * 2}, ${dy * 2}) scale(${1 - (2 * ds)}, ${1 - (2 * ds)})`}>
        <Shape {...propsForIndex(3)} />
      </g>
      <g transform={`translate(${dx * 1}, ${dy * 1}) scale(${1 - (1 * ds)}, ${1 - (1 * ds)})`}>
        <Shape {...propsForIndex(2)} />
      </g>
      <g transform={`translate(${dx * 0.5}, ${dy * 0.5}) scale(${hls})`} className="stack">
        <Shape {...propsOnlyHighlight} />
      </g>
      <Shape {...propsForIndex(1)} />
    </g>
  );
}

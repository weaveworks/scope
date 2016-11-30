import React from 'react';
import { scaleBand } from 'd3-scale';
import { List as makeList } from 'immutable';
import { getNetworkColor } from '../utils/color-utils';
import { isContrastMode } from '../utils/contrast-utils';


// Gap size between bar segments.
const minBarHeight = 3;
const padding = 0.05;
const rx = 1;
const ry = rx;
const x = scaleBand();

function NodeNetworksOverlay({offset, size, stack, networks = makeList()}) {
  // Min size is about a quarter of the width, feels about right.
  const minBarWidth = (size / 4);
  const barWidth = Math.max(size, minBarWidth * networks.size);
  const barHeight = Math.max(size * 0.085, minBarHeight);

  // Update singleton scale.
  x.domain(networks.map((n, i) => i).toJS());
  x.range([barWidth * -0.5, barWidth * 0.5]);
  x.paddingInner(padding);

  const bars = networks.map((n, i) => (
    <rect
      x={x(i)}
      y={offset - barHeight * 0.5}
      width={x.bandwidth()}
      height={barHeight}
      rx={rx}
      ry={ry}
      className="node-network"
      style={{
        fill: getNetworkColor(n.get('colorKey', n.get('id')))
      }}
      key={n.get('id')}
    />
  ));

  let transform = '';
  if (stack) {
    const contrastMode = isContrastMode();
    const [dx, dy] = contrastMode ? [0, 8] : [0, 0];
    transform = `translate(${dx}, ${dy * -1.5})`;
  }

  return (
    <g transform={transform}>
      {bars.toJS()}
    </g>
  );
}

export default NodeNetworksOverlay;

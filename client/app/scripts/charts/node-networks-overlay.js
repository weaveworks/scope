import React from 'react';
import d3 from 'd3';
import { List as makeList } from 'immutable';
import { getNetworkColor } from '../utils/color-utils';
import { isContrastMode } from '../utils/contrast-utils';


const barHeight = 5;
const barMarginTop = 6;
const labelHeight = 32;
// Gap size between bar segments.
const padding = 0.05;
const rx = 1;
const ry = rx;
const x = d3.scale.ordinal();

function NodeNetworksOverlay({labelOffsetY, size, stack, networks = makeList()}) {
  const offset = labelOffsetY + labelHeight + barMarginTop;

  // Min size is about a quarter of the width, feels about right.
  const minBarWidth = (size / 4);
  const barWidth = Math.max(size, minBarWidth * networks.size);

  // Update singleton scale.
  x.domain(networks.map((n, i) => i).toJS());
  x.rangeBands([barWidth * -0.5, barWidth * 0.5], padding, 0);

  const bars = networks.map((n, i) => (
    <rect
      x={x(i)}
      y={offset}
      width={x.rangeBand()}
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

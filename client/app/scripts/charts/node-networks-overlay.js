import React from 'react';
import d3 from 'd3';
import { List as makeList } from 'immutable';
import { getNodeColor } from '../utils/color-utils';
import { isContrastMode } from '../utils/contrast-utils';


const padding = 0.05;
const offset = Math.PI;
const arc = d3.svg.arc()
  .startAngle(d => d.startAngle + offset)
  .endAngle(d => d.endAngle + offset);
const arcScale = d3.scale.linear()
  .range([Math.PI * 0.25 + padding, Math.PI * 1.75 - padding]);


function NodeNetworksOverlay({size, stack, networks = makeList()}) {
  arcScale.domain([0, networks.size]);
  const radius = size * 0.9;

  const paths = networks.map((n, i) => {
    const d = arc({
      padAngle: 0.05,
      innerRadius: radius,
      outerRadius: radius + 4,
      startAngle: arcScale(i),
      endAngle: arcScale(i + 1)
    });

    return (<path d={d} style={{fill: getNodeColor(n)}} key={n} />);
  });

  let transform = '';
  if (stack) {
    const contrastMode = isContrastMode();
    const [dx, dy] = contrastMode ? [0, 8] : [0, 5];
    transform = `translate(${dx}, ${dy * -1.5})`;
  }

  return (
    <g transform={transform}>
      {paths.toJS()}
    </g>
  );
}

export default NodeNetworksOverlay;

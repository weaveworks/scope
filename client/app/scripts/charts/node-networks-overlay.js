import React from 'react';
import { scaleBand } from 'd3-scale';
import { List as makeList } from 'immutable';
import { connect } from 'react-redux';

import { getNetworkColor } from '../utils/color-utils';
import { NODE_BASE_SIZE } from '../constants/styles';

// Min size is about a quarter of the width, feels about right.
const minBarWidth = 0.25;
const barHeight = 0.08;
const innerPadding = 0.04;
const borderRadius = 0.01;
const offset = 0.67;
const x = scaleBand();

function NodeNetworksOverlay({ stack, networks = makeList(), contrastMode }) {
  const barWidth = Math.max(1, minBarWidth * networks.size);
  const yPosition = offset - (barHeight * 0.5);

  // Update singleton scale.
  x.domain(networks.map((n, i) => i).toJS());
  x.range([barWidth * -0.5, barWidth * 0.5]);
  x.paddingInner(innerPadding);

  const bandwidth = x.bandwidth();
  const bars = networks.map((n, i) => (
    <rect
      className="node-network"
      key={n.get('id')}
      x={x(i)}
      y={yPosition}
      width={bandwidth}
      height={barHeight}
      rx={borderRadius}
      ry={borderRadius}
      style={{ fill: getNetworkColor(n.get('colorKey', n.get('id'))) }}
    />
  ));

  const translateY = stack && contrastMode ? 0.15 : 0;
  return (
    <g transform={`translate(0, ${translateY}) scale(${NODE_BASE_SIZE})`}>
      {bars.toJS()}
    </g>
  );
}

function mapStateToProps(state) {
  return {
    contrastMode: state.get('contrastMode')
  };
}

export default connect(mapStateToProps)(NodeNetworksOverlay);

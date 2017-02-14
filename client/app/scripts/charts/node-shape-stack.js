import React from 'react';
import { connect } from 'react-redux';

import { NODE_BASE_SIZE } from '../constants/styles';

function NodeShapeStack(props) {
  const shift = props.contrastMode ? 0.15 : 0.1;
  const highlightScale = [1, 1 + shift];
  const dy = NODE_BASE_SIZE * shift;

  const Shape = props.shape;
  return (
    <g transform={`translate(0, ${dy * -2.5})`} className="stack">
      <g transform={`scale(${highlightScale}) translate(0, ${dy})`} className="highlight">
        <Shape {...props} />
      </g>
      <g transform={`translate(0, ${dy * 2})`}>
        <Shape {...props} />
      </g>
      <g transform={`translate(0, ${dy * 1})`}>
        <Shape {...props} />
      </g>
      <g className="only-metrics">
        <Shape {...props} />
      </g>
    </g>
  );
}

function mapStateToProps(state) {
  return {
    contrastMode: state.get('contrastMode')
  };
}

export default connect(mapStateToProps)(NodeShapeStack);

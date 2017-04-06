import React from 'react';
import { connect } from 'react-redux';

import NodesChartNodes from './nodes-chart-nodes';
import { graphExceedsComplexityThreshSelector } from '../selectors/topology';
import {
  selectedScaleSelector,
  layoutNodesSelector,
  layoutEdgesSelector
} from '../selectors/graph-view/layout';


class NodesChartElements extends React.Component {
  render() {
    const { layoutNodes, layoutEdges, selectedScale, isAnimated } = this.props;

    return (
      <g className="nodes-chart-elements">
        <NodesChartNodes
          layoutNodes={layoutNodes}
          layoutEdges={layoutEdges}
          selectedScale={selectedScale}
          isAnimated={isAnimated} />
      </g>
    );
  }
}


function mapStateToProps(state) {
  return {
    layoutNodes: layoutNodesSelector(state),
    layoutEdges: layoutEdgesSelector(state),
    selectedScale: selectedScaleSelector(state),
    isAnimated: !graphExceedsComplexityThreshSelector(state),
  };
}

export default connect(
  mapStateToProps
)(NodesChartElements);

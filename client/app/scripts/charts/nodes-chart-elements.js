import React from 'react';
import { connect } from 'react-redux';

import { completeNodesSelector } from '../selectors/chartSelectors';
import NodesChartEdges from './nodes-chart-edges';
import NodesChartNodes from './nodes-chart-nodes';

class NodesChartElements extends React.Component {
  render() {
    const props = this.props;
    return (
      <g className="nodes-chart-elements" transform={props.transform}>
        <NodesChartEdges
          layoutEdges={props.layoutEdges}
          selectedScale={props.selectedScale}
          isAnimated={props.isAnimated} />
        <NodesChartNodes
          layoutNodes={props.completeNodes}
          selectedScale={props.selectedScale}
          isAnimated={props.isAnimated} />
      </g>
    );
  }
}

function mapStateToProps(state, props) {
  return {
    completeNodes: completeNodesSelector(state, props)
  };
}

export default connect(mapStateToProps)(NodesChartElements);

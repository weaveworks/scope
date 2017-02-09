import React from 'react';
import { connect } from 'react-redux';

import { completeNodesSelector } from '../selectors/nodes-chart';
import NodesChartEdges from './nodes-chart-edges';
import NodesChartNodes from './nodes-chart-nodes';

class NodesChartElements extends React.Component {
  render() {
    const props = this.props;
    return (
      <g className="nodes-chart-elements" transform={props.transform}>
        <NodesChartNodes
          layoutNodes={props.layoutNodes}
          selectedScale={props.selectedScale}
          isAnimated={props.isAnimated} />
      </g>
    );
  }
}

function mapStateToProps(state, props) {
  return {
    // completeNodes: completeNodesSelector(state, props)
  };
}

export default connect(mapStateToProps)(NodesChartElements);

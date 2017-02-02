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
          layoutPrecision={props.layoutPrecision} />
        <NodesChartNodes
          layoutNodes={props.completeNodes}
          nodeScale={props.nodeScale}
          scale={props.scale}
          selectedNodeScale={props.selectedNodeScale}
          layoutPrecision={props.layoutPrecision} />
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

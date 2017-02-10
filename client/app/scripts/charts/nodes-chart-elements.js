import React from 'react';
import { connect } from 'react-redux';

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
          layoutNodes={props.layoutNodes}
          selectedScale={props.selectedScale}
          isAnimated={props.isAnimated} />
      </g>
    );
  }
}

export default connect()(NodesChartElements);

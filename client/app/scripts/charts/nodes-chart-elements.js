import React from 'react';

import NodesChartEdges from './nodes-chart-edges';
import NodesChartNodes from './nodes-chart-nodes';

export default class NodesChartElements extends React.PureComponent {
  render() {
    const { transform, layoutEdges, layoutNodes, selectedScale, isAnimated } = this.props;
    return (
      <g className="nodes-chart-elements" transform={transform}>
        <NodesChartEdges
          layoutEdges={layoutEdges}
          selectedScale={selectedScale}
          isAnimated={isAnimated} />
        <NodesChartNodes
          layoutNodes={layoutNodes}
          selectedScale={selectedScale}
          isAnimated={isAnimated} />
      </g>
    );
  }
}

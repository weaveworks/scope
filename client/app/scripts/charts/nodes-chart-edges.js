import React from 'react';
import PureRenderMixin from 'react-addons-pure-render-mixin';
import reactMixin from 'react-mixin';

import EdgeContainer from './edge-container';

export default class NodesChartEdges extends React.Component {
  render() {
    const {hasSelectedNode, highlightedEdgeIds, layoutEdges, layoutPrecision,
      selectedNodeId} = this.props;

    return (
      <g className="nodes-chart-edges">
        {layoutEdges.toIndexedSeq().map(edge => <EdgeContainer key={edge.get('id')}
          id={edge.get('id')} source={edge.get('source')} target={edge.get('target')}
          points={edge.get('points')} layoutPrecision={layoutPrecision}
          highlightedEdgeIds={highlightedEdgeIds} hasSelectedNode={hasSelectedNode}
          selectedNodeId={selectedNodeId} />)}
      </g>
    );
  }
}

reactMixin.onClass(NodesChartEdges, PureRenderMixin);

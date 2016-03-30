import React from 'react';

import EdgeContainer from './edge-container';

export default function NodesChartEdges({hasSelectedNode, highlightedEdgeIds,
  layoutEdges, layoutPrecision, selectedNodeId}) {
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

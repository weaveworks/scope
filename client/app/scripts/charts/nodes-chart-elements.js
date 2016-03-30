import React from 'react';

import NodesChartEdges from './nodes-chart-edges';
import NodesChartNodes from './nodes-chart-nodes';

export default function NodesChartElements(props) {
  return (
    <g className="nodes-chart-elements" transform={props.transform}>
      <NodesChartEdges layoutEdges={props.edges} selectedNodeId={props.selectedNodeId}
        highlightedEdgeIds={props.highlightedEdgeIds}
        hasSelectedNode={props.hasSelectedNode}
        layoutPrecision={props.layoutPrecision} />
      <NodesChartNodes layoutNodes={props.nodes} selectedNodeId={props.selectedNodeId}
        selectedMetric={props.selectedMetric}
        highlightedNodeIds={props.highlightedNodeIds}
        hasSelectedNode={props.hasSelectedNode}
        adjacentNodes={props.adjacentNodes}
        nodeScale={props.nodeScale} onNodeClick={props.onNodeClick}
        scale={props.scale} selectedNodeScale={props.selectedNodeScale}
        topologyId={props.topologyId} layoutPrecision={props.layoutPrecision} />
    </g>
  );
}

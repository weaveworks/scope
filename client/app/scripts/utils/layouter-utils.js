import { Map as makeMap } from 'immutable';

import { EDGE_ID_SEPARATOR } from '../constants/naming';


export function getNodesFromEdgeId(edgeId) {
  return edgeId.split(EDGE_ID_SEPARATOR);
}

export function constructEdgeId(source, target) {
  return [source, target].join(EDGE_ID_SEPARATOR);
}

// Constructs the edges for the layout engine from the nodes' adjacency table.
// We don't collapse edge pairs (A->B, B->A) here as we want to let the layout
// engine decide how to handle bidirectional edges.
export function initEdgesFromNodes(nodes) {
  let edges = makeMap();

  nodes.forEach((node, nodeId) => {
    (node.get('adjacency') || []).forEach((adjacentId) => {
      const source = nodeId;
      const target = adjacentId;

      if (nodes.has(target)) {
        // The direction source->target is important since dagre takes
        // directionality into account when calculating the layout.
        const edgeId = constructEdgeId(source, target);
        const edge = makeMap({
          id: edgeId, source, target, value: 1
        });
        edges = edges.set(edgeId, edge);
      }
    });
  });

  return edges;
}

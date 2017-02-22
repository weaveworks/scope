import { Map as makeMap } from 'immutable';

import { EDGE_ID_SEPARATOR } from '../constants/naming';


function constructEdgeId(source, target) {
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
        const edge = makeMap({ id: edgeId, value: 1, source, target });
        edges = edges.set(edgeId, edge);
      }
    });
  });

  return edges;
}

// Replaces all pairs of edges (A->B, B->A) with a single A->B edge that is marked as
// bidirectional. We do this to prevent double rendering of edges between the same nodes.
export function collapseMultiEdges(directedEdges) {
  let collapsedEdges = makeMap();

  directedEdges.forEach((edge, edgeId) => {
    const source = edge.get('source');
    const target = edge.get('target');
    const reversedEdgeId = constructEdgeId(target, source);

    if (collapsedEdges.has(reversedEdgeId)) {
      // If the edge between the same nodes with the opposite direction already exists,
      // mark it as bidirectional and don't add any other edges (making graph simple).
      collapsedEdges = collapsedEdges.setIn([reversedEdgeId, 'bidirectional'], true);
    } else {
      // Otherwise just copy the edge.
      collapsedEdges = collapsedEdges.set(edgeId, edge);
    }
  });

  return collapsedEdges;
}

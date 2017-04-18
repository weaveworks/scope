import { createSelector } from 'reselect';
import { List as makeList, Map as makeMap, Set as makeSet } from 'immutable';

import { EDGE_ID_SEPARATOR } from '../../constants/naming';


const nodeAdjacenciesSelector = createSelector(
  [
    state => state.get('nodes'),
  ],
  (nodes) => {
    let nodeAdjacencies = makeMap();

    nodes.forEach((node, nodeId) => {
      node.get('adjacency', makeList()).forEach((adjacentNodeId) => {
        nodeAdjacencies = nodeAdjacencies.setIn([nodeId, adjacentNodeId], true);
        nodeAdjacencies = nodeAdjacencies.setIn([adjacentNodeId, nodeId], true);
      });
    });

    return nodeAdjacencies;
  }
);

export const highlightedNodeIdsSelector = createSelector(
  [
    nodeAdjacenciesSelector,
    state => state.get('mouseOverNodeId'),
    state => state.get('mouseOverEdgeId'),
  ],
  (nodeAdjacencies, mouseOverNodeId, mouseOverEdgeId) => {
    let nodeIds = makeSet();

    // If a node if hovered, highlight it together with
    // all its neighbours (not minding directionality).
    if (mouseOverNodeId) {
      const adjacencies = nodeAdjacencies.get(mouseOverNodeId, makeMap()).keySeq();
      nodeIds = nodeIds.union(adjacencies);
      nodeIds = nodeIds.add(mouseOverNodeId);
    }

    // When an edge is hovered, highlight both of its endpoint nodes.
    if (mouseOverEdgeId) {
      nodeIds = nodeIds.union(mouseOverEdgeId.split(EDGE_ID_SEPARATOR));
    }

    return nodeIds;
  }
);

export const highlightedEdgeIdsSelector = createSelector(
  [
    nodeAdjacenciesSelector,
    state => state.get('mouseOverNodeId'),
    state => state.get('mouseOverEdgeId'),
  ],
  (nodeAdjacencies, mouseOverNodeId, mouseOverEdgeId) => {
    let edgeIds = makeSet();

    // If a node is hovered, highlight all the edges that go into or out from the node.
    if (mouseOverNodeId) {
      const adjacencies = nodeAdjacencies.get(mouseOverNodeId, makeMap()).keySeq();
      edgeIds = edgeIds.union(adjacencies.flatMap(adjacentId => [
        [adjacentId, mouseOverNodeId].join(EDGE_ID_SEPARATOR),
        [mouseOverNodeId, adjacentId].join(EDGE_ID_SEPARATOR),
      ]));
    }

    // When an edge is hovered, highlight it together with its opposite direction counterpart.
    if (mouseOverEdgeId) {
      const oppositeId = mouseOverEdgeId.split(EDGE_ID_SEPARATOR).reverse().join(EDGE_ID_SEPARATOR);
      edgeIds = edgeIds.add(mouseOverEdgeId);
      edgeIds = edgeIds.add(oppositeId);
    }

    return edgeIds;
  }
);

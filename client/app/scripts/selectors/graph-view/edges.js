import { createSelector } from 'reselect';
import { Map as makeMap, Set as makeSet } from 'immutable';

import { nodeAdjacenciesSelector } from '../node-adjacencies';
import { EDGE_ID_SEPARATOR } from '../../constants/naming';


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

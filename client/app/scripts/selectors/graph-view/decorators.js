import { createSelector } from 'reselect';
import { Set as makeSet } from 'immutable';

import { getNodesFromEdgeId } from '../../utils/layouter-utils';


const adjacentToHoveredNodeIdsSelector = createSelector(
  [
    state => state.get('mouseOverNodeId'),
    state => state.get('nodes'),
  ],
  (mouseOverNodeId, nodes) => {
    let nodeIds = makeSet();

    if (mouseOverNodeId) {
      nodeIds = makeSet(nodes.getIn([mouseOverNodeId, 'adjacency']));
      // fill up set with reverse edges
      nodes.forEach((node, id) => {
        if (node.get('adjacency') && node.get('adjacency').includes(mouseOverNodeId)) {
          nodeIds = nodeIds.add(id);
        }
      });
    }

    return nodeIds;
  }
);

export const highlightedNodeIdsSelector = createSelector(
  [
    adjacentToHoveredNodeIdsSelector,
    state => state.get('mouseOverNodeId'),
    state => state.get('mouseOverEdgeId'),
  ],
  (adjacentToHoveredNodeIds, mouseOverNodeId, mouseOverEdgeId) => {
    let highlightedNodeIds = makeSet();

    if (mouseOverEdgeId) {
      highlightedNodeIds = highlightedNodeIds.union(getNodesFromEdgeId(mouseOverEdgeId));
    }

    if (mouseOverNodeId) {
      highlightedNodeIds = highlightedNodeIds.add(mouseOverNodeId);
      highlightedNodeIds = highlightedNodeIds.union(adjacentToHoveredNodeIds);
    }

    return highlightedNodeIds;
  }
);

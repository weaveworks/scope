import { createSelector } from 'reselect';
import { Map as makeMap, Set as makeSet } from 'immutable';

import { searchNodeMatchesSelector } from '../search';
import { nodeAdjacenciesSelector } from '../node-adjacencies';
import { selectedNetworkNodesIdsSelector } from '../node-networks';
import { EDGE_ID_SEPARATOR } from '../../constants/naming';


export const highlightedNodeIdsSelector = createSelector(
  [
    nodeAdjacenciesSelector,
    state => state.get('selectedNodeId'),
    state => state.get('mouseOverNodeId'),
    state => state.get('mouseOverEdgeId'),
  ],
  (nodeAdjacencies, selectedNodeId, mouseOverNodeId, mouseOverEdgeId) => {
    let nodeIds = makeSet();

    // Selected node is always highlighted.
    if (selectedNodeId) {
      nodeIds = nodeIds.add(selectedNodeId);
    }

    // If a node if hovered, highlight it together with
    // all its neighbours (not minding directionality).
    if (mouseOverNodeId) {
      const adjacencies = nodeAdjacencies.get(mouseOverNodeId, makeMap());
      nodeIds = nodeIds.union(adjacencies.keySeq());
      nodeIds = nodeIds.add(mouseOverNodeId);
    }

    // When an edge is hovered, highlight both of its endpoint nodes.
    if (mouseOverEdgeId) {
      nodeIds = nodeIds.union(mouseOverEdgeId.split(EDGE_ID_SEPARATOR));
    }

    return nodeIds;
  }
);

export const focusedNodeIdsSelector = createSelector(
  [
    nodeAdjacenciesSelector,
    state => state.get('selectedNodeId'),
  ],
  (nodeAdjacencies, selectedNodeId) => {
    let nodeIds = makeSet();

    // Selected node and all its neighbours (not minding
    // directionality) are always in focus.
    if (selectedNodeId) {
      const adjacencies = nodeAdjacencies.get(selectedNodeId, makeMap());
      nodeIds = nodeIds.union(adjacencies.keySeq());
      nodeIds = nodeIds.add(selectedNodeId);
    }

    return nodeIds;
  }
);

export const blurredNodeIdsSelector = createSelector(
  [
    focusedNodeIdsSelector,
    highlightedNodeIdsSelector,
    searchNodeMatchesSelector,
    selectedNetworkNodesIdsSelector,
    state => state.get('selectedNodeId'),
    state => state.get('selectedNetwork'),
    state => state.get('searchQuery'),
    state => state.get('nodes'),
  ],
  (focusedNodeIds, highlightedNodeIds, searchNodeMatches, selectedNetworkNodesIds,
    selectedNodeId, selectedNetwork, searchQuery, allNodes) => {
    const allNodesIds = allNodes.keySeq().toSet();
    let nodeIds = makeSet();

    if (selectedNodeId) {
      const outsideOfFocus = allNodesIds.subtract(focusedNodeIds);
      nodeIds = nodeIds.union(outsideOfFocus);
    }

    if (selectedNetwork) {
      const notInNetwork = allNodesIds.subtract(selectedNetworkNodesIds);
      nodeIds = nodeIds.union(notInNetwork);
    }

    if (searchQuery) {
      const notHighlighted = allNodesIds.subtract(highlightedNodeIds);
      const notMatched = searchNodeMatches.filter(matches => matches.isEmpty()).keySeq();
      nodeIds = nodeIds.union(notHighlighted.intersect(notMatched));
    }

    return nodeIds;
  }
);

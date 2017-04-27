import { createSelector } from 'reselect';
import { Map as makeMap, List as makeList } from 'immutable';


export const nodeAdjacenciesSelector = createSelector(
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

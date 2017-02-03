import { createSelector } from 'reselect';
import { Map as makeMap } from 'immutable';


const allNodesSelector = state => state.get('nodes');

export const nodesSelector = createSelector(
  [
    allNodesSelector,
  ],
  allNodes => allNodes.filter(node => !node.get('filtered'))
);


export const nodeAdjacenciesSelector = createSelector(
  [
    nodesSelector,
  ],
  nodes => nodes.map(node => makeMap({
    id: node.get('id'),
    adjacency: node.get('adjacency'),
    label: node.get('label'),
    pseudo: node.get('pseudo'),
    subLabel: node.get('labelMinor'),
    nodeCount: node.get('node_count'),
    metrics: node.get('metrics'),
    rank: node.get('rank'),
    shape: node.get('shape'),
    stack: node.get('stack'),
    networks: node.get('networks'),
  }))
);

import { createSelector } from 'reselect';
import { Map as makeMap } from 'immutable';


export const shownNodesSelector = createSelector(
  [
    state => state.get('nodes'),
  ],
  nodes => nodes.filter(node => !node.get('filtered'))
);

export const nodeAdjacenciesSelector = createSelector(
  [
    shownNodesSelector,
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

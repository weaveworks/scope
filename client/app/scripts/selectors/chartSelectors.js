import { createSelector } from 'reselect';
import { Map as makeMap } from 'immutable';


const allNodesSelector = state => state.get('nodes');


export const nodesSelector = createSelector(
  allNodesSelector,
  (allNodes) => allNodes.filter(node => !node.get('filtered'))
);


export const nodeAdjacenciesSelector = createSelector(
  nodesSelector,
  (nodes) => nodes.map(n => makeMap({
    id: n.get('id'),
    adjacency: n.get('adjacency'),
  }))
);


export const layoutNodesSelector = (_, props) => props.layoutNodes;


export const dataNodesSelector = createSelector(
  nodesSelector,
  (nodes) => nodes.map((node, id) => makeMap({
    id,
    label: node.get('label'),
    pseudo: node.get('pseudo'),
    subLabel: node.get('label_minor'),
    nodeCount: node.get('node_count'),
    metrics: node.get('metrics'),
    rank: node.get('rank'),
    shape: node.get('shape'),
    stack: node.get('stack'),
    networks: node.get('networks'),
  }))
);


function mergeDeepIfExists(mapA, mapB) {
  //
  // Does a deep merge on any key that exists in the first map
  //
  return mapA.map((v, k) => v.mergeDeep(mapB.get(k)));
}


export const completeNodesSelector = createSelector(
  layoutNodesSelector,
  dataNodesSelector,
  (layoutNodes, dataNodes) => {
    if (!layoutNodes || layoutNodes.size === 0) {
      return makeMap();
    }

    return mergeDeepIfExists(dataNodes, layoutNodes);
  }
);

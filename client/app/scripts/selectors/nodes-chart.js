import { createSelector } from 'reselect';
import { Map as makeMap, Set } from 'immutable';


const allNodesSelector = state => state.get('nodes');


const nodesSelector = createSelector(
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


function mergeDeepKeyIntersection(mapA, mapB) {
  const commonKeys = Set.fromKeys(mapA).intersect(mapB.keySeq());
  return makeMap(commonKeys.map(k => [k, mapA.get(k).mergeDeep(mapB.get(k))]));
}

const layoutNodesSelector = (_, props) => props.layoutNodes || makeMap();
const dataNodesSelector = createSelector(
  [
    nodesSelector,
  ],
  nodes => nodes.map((node, id) => makeMap({
    id,
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

export const completeNodesSelector = createSelector(
  [
    layoutNodesSelector,
    dataNodesSelector,
  ],
  (layoutNodes, dataNodes) => {
    console.log('Recalculated complete nodes');
    return mergeDeepKeyIntersection(dataNodes, layoutNodes);
  }
);

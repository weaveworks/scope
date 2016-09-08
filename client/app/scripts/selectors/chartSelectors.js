import { createSelector, createSelectorCreator, defaultMemoize } from 'reselect';
import { Map as makeMap, is } from 'immutable';

import { getAdjacentNodes } from '../utils/topology-utils';


//
// "immutable" createSelector
//
const createDeepEqualSelector = createSelectorCreator(
  defaultMemoize,
  is
);


const identity = v => v;


const allNodesSelector = state => state.get('nodes');


export const nodesSelector = createSelector(
  allNodesSelector,
  (allNodes) => allNodes.filter(node => !node.get('filtered'))
);


//
// This is like an === cache...
//
// - getAdjacentNodes is run on every state change and can generate a new immutable object each
// time:
//   - v1 = getAdjacentNodes(a)
//   - v2 = getAdjacentNodes(a)
//   - v1 !== v2
//   - is(v1, v2) === true
//
// - createDeepEqualSelector will wrap those calls with a: is(v1, v2) ? v1 : v2
//   - Thus you can compare consecutive calls to adjacentNodesSelector(state) with === (which is
//     what redux is doing with connect()
//
// Note: this feels like the wrong way to be using reselect...
//
export const adjacentNodesSelector = createDeepEqualSelector(
  getAdjacentNodes,
  identity
);


//
// You what? What is going on here?
//
// We wrap the result of nodes.map in another equality test which discards the new value
// if it was the same as the old one. Again preserving ===
//
export const nodeAdjacenciesSelector = createDeepEqualSelector(
  createSelector(
    nodesSelector,
    (nodes) => nodes.map(n => makeMap({
      id: n.get('id'),
      adjacency: n.get('adjacency'),
    }))
  ),
  identity
);


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


// FIXME: this is a bit of a hack...
export const layoutNodesSelector = (_, props) => props.layoutNodes || makeMap();


function mergeDeepIfExists(mapA, mapB) {
  //
  // Does a deep merge on any key that exists in the first map
  //
  return mapA.map((v, k) => v.mergeDeep(mapB.get(k)));
}


const _completeNodesSelector = createSelector(
  layoutNodesSelector,
  dataNodesSelector,
  (layoutNodes, dataNodes) => {
    if (layoutNodes.size === 0 || dataNodes.size === 0) {
      return makeMap();
    }

    // dataNodes might get updated before layoutNodes when a node is removed from the topo.
    return mergeDeepIfExists(dataNodes, layoutNodes);
  }
);


export const completeNodesSelector = createDeepEqualSelector(
  _completeNodesSelector,
  identity
);

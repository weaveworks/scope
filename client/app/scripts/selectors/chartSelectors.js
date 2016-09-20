import debug from 'debug';
import { createSelector, createSelectorCreator, defaultMemoize } from 'reselect';
import { Map as makeMap, is, Set } from 'immutable';

import { getAdjacentNodes } from '../utils/topology-utils';


const log = debug('scope:selectors');


//
// `mergeDeepKeyIntersection` does a deep merge on keys that exists in both maps
//
function mergeDeepKeyIntersection(mapA, mapB) {
  const commonKeys = Set.fromKeys(mapA).intersect(mapB.keySeq());
  return makeMap(commonKeys.map(k => [k, mapA.get(k).mergeDeep(mapB.get(k))]));
}


//
// `returnPreviousRefIfEqual` is a helper function that checks the new computed of a selector
// against the previously computed value. If they are deeply equal return the previous result. This
// is important for things like connect() which tests whether componentWillReceiveProps should be
// called by doing a '===' on the values you return from mapStateToProps.
//
// e.g.
//
// const filteredThings = createSelector(
//   state => state.things,
//   (things) => things.filter(t => t > 2)
// );
//
// // This will trigger componentWillReceiveProps on every store change:
// connect(s => { things: filteredThings(s) }, ThingComponent);
//
// // But if we wrap it, the result will be === if it `is()` equal and...
// const filteredThingsWrapped = returnPreviousRefIfEqual(filteredThings);
//
// // ...We're safe!
// connect(s => { things: filteredThingsWrapped(s) }, ThingComponent);
//
// Note: This is a slightly stange way to use reselect. Selectors memoize their *arguments* not
// "their results", so use the result of the wrapped selector as the argument to another selector
// here to memoize it and get what we want.
//
const _createDeepEqualSelector = createSelectorCreator(defaultMemoize, is);
const _identity = v => v;
const returnPreviousRefIfEqual = (selector) => _createDeepEqualSelector(selector, _identity);


//
// Selectors!
//


const allNodesSelector = state => state.get('nodes');


export const nodesSelector = returnPreviousRefIfEqual(
  createSelector(
    allNodesSelector,
    (allNodes) => allNodes.filter(node => !node.get('filtered'))
  )
);


export const adjacentNodesSelector = returnPreviousRefIfEqual(getAdjacentNodes);


export const nodeAdjacenciesSelector = returnPreviousRefIfEqual(
  createSelector(
    nodesSelector,
    (nodes) => nodes.map(n => makeMap({
      id: n.get('id'),
      adjacency: n.get('adjacency'),
    }))
  )
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


//
// FIXME: this is a bit of a hack...
//
export const layoutNodesSelector = (_, props) => props.layoutNodes || makeMap();


export const completeNodesSelector = createSelector(
  layoutNodesSelector,
  dataNodesSelector,
  (layoutNodes, dataNodes) => {
    //
    // There are no guarantees whether this selector will be computed first (when
    // node-chart-elements.mapStateToProps is called by store.subscribe before
    // nodes-chart.mapStateToProps is called), and component render batching and yadada.
    //
    if (layoutNodes.size !== dataNodes.size) {
      log('Obviously mismatched node data', layoutNodes.size, dataNodes.size);
    }
    return mergeDeepKeyIntersection(dataNodes, layoutNodes);
  }
);

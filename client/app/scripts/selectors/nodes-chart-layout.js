import debug from 'debug';
import { createSelector, createStructuredSelector } from 'reselect';
import { Map as makeMap } from 'immutable';
import timely from 'timely';

import { EDGE_ID_SEPARATOR } from '../constants/naming';
import { doLayout } from '../charts/nodes-layout';

const log = debug('scope:nodes-chart');


function initEdgesFromNodes(nodes) {
  let edges = makeMap();

  nodes.forEach((node, nodeId) => {
    const adjacency = node.get('adjacency');
    if (adjacency) {
      adjacency.forEach((adjacent) => {
        const edge = nodeId < adjacent ? [nodeId, adjacent] : [adjacent, nodeId];
        const edgeId = edge.join(EDGE_ID_SEPARATOR);

        if (!edges.has(edgeId)) {
          const source = edge[0];
          const target = edge[1];
          if (nodes.has(source) && nodes.has(target)) {
            edges = edges.set(edgeId, makeMap({
              id: edgeId,
              value: 1,
              source,
              target
            }));
          }
        }
      });
    }
  });

  return edges;
}

// TODO: Make all the selectors below pure (so that they only depend on the global state).

const layoutOptionsSelector = createStructuredSelector({
  width: state => state.width,
  height: state => state.height,
  margins: (_, props) => props.margins,
  forceRelayout: (_, props) => props.forceRelayout,
  topologyId: (_, props) => props.topologyId,
  topologyOptions: (_, props) => props.topologyOptions,
});

export const graphLayout = createSelector(
  [
    // TODO: Instead of sending the nodes with all the information (metrics, metadata, etc...)
    // to the layout engine, it would suffice to forward it just the nodes adjacencies map, which
    // we could get with another selector like:
    //
    // const nodesAdjacenciesSelector = createMapSelector(
    //   [ (_, props) => props.nodes ],
    //   node => node.get('adjacency') || makeList()
    // );
    //
    // That would enable us to use smarter caching, so that the layout doesn't get recalculated
    // if adjacencies don't change but e.g. metrics gets updated. We also don't need to init
    // edges here as the adjacencies data is enough to reconstruct them in the layout engine (this
    // might enable us to simplify the caching system there since we really only need to cache
    // the adjacencies map in that case and not nodes and edges).
    (_, props) => props.nodes,
    layoutOptionsSelector,
  ],
  (nodes, options) => {
    // If the graph is empty, skip computing the layout.
    if (nodes.size === 0) {
      return {
        layoutNodes: makeMap(),
        layoutEdges: makeMap(),
      };
    }

    const edges = initEdgesFromNodes(nodes);
    const timedLayouter = timely(doLayout);
    const graph = timedLayouter(nodes, edges, options);

    // NOTE: We probably shouldn't log anything in a
    // computed property, but this is still useful.
    log(`graph layout calculation took ${timedLayouter.time}ms`);

    return {
      layoutNodes: graph.nodes,
      layoutEdges: graph.edges,
    };
  }
);

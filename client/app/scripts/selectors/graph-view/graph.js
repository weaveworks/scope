import debug from 'debug';
import { createSelector, createStructuredSelector } from 'reselect';
import { Map as makeMap } from 'immutable';

import { initEdgesFromNodes } from '../../utils/layouter-utils';
import { canvasWidthSelector, canvasHeightSelector } from '../canvas';
import { activeTopologyOptionsSelector } from '../topology';
import { shownNodesSelector } from '../node-filters';
import { doLayout } from '../../charts/nodes-layout';
import { timer } from '../../utils/time-utils';

const log = debug('scope:nodes-chart');


const layoutOptionsSelector = createStructuredSelector({
  forceRelayout: state => state.get('forceRelayout'),
  height: canvasHeightSelector,
  topologyId: state => state.get('currentTopologyId'),
  topologyOptions: activeTopologyOptionsSelector,
  width: canvasWidthSelector,
});

const graphLayoutSelector = createSelector(
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
    shownNodesSelector,
    layoutOptionsSelector,
  ],
  (nodes, options) => {
    // If the graph is empty, skip computing the layout.
    if (nodes.size === 0) {
      return {
        edges: makeMap(),
        nodes: makeMap(),
      };
    }

    const edges = initEdgesFromNodes(nodes);
    const timedLayouter = timer(doLayout);
    const graph = timedLayouter(nodes, edges, options);

    // NOTE: We probably shouldn't log anything in a
    // computed property, but this is still useful.
    log(`graph layout calculation took ${timedLayouter.time}ms`);

    return graph;
  }
);

export const graphNodesSelector = createSelector(
  [
    graphLayoutSelector,
  ],
  // NOTE: This might be a good place to add (some of) nodes/edges decorators.
  graph => graph.nodes
);

export const graphEdgesSelector = createSelector(
  [
    graphLayoutSelector,
  ],
  graph => graph.edges
);

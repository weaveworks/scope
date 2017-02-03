import debug from 'debug';
import { createSelector } from 'reselect';
import { Map as makeMap } from 'immutable';
import timely from 'timely';

import { EDGE_ID_SEPARATOR } from '../constants/naming';
import { doLayout } from '../charts/nodes-layout';

const log = debug('scope:nodes-chart');


const stateWidthSelector = state => state.width;
const stateHeightSelector = state => state.height;
const inputNodesSelector = (_, props) => props.nodes;
const propsMarginsSelector = (_, props) => props.margins;
const forceRelayoutSelector = (_, props) => props.forceRelayout;
const topologyIdSelector = (_, props) => props.topologyId;
const topologyOptionsSelector = (_, props) => props.topologyOptions;


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

const layoutOptionsSelector = createSelector(
  [
    stateWidthSelector,
    stateHeightSelector,
    propsMarginsSelector,
    forceRelayoutSelector,
    topologyIdSelector,
    topologyOptionsSelector,
  ],
  (width, height, margins, forceRelayout, topologyId, topologyOptions) => (
    { width, height, margins, forceRelayout, topologyId, topologyOptions }
  )
);

export const graphLayout = createSelector(
  [
    inputNodesSelector,
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

    const layoutEdges = graph.edges;
    const layoutNodes = graph.nodes.map(node => makeMap({
      x: node.get('x'),
      y: node.get('y'),
      id: node.get('id'),
      label: node.get('label'),
      pseudo: node.get('pseudo'),
      subLabel: node.get('labelMinor'),
      nodeCount: node.get('node_count'),
      metrics: node.get('metrics'),
      rank: node.get('rank'),
      shape: node.get('shape'),
      stack: node.get('stack'),
      // networks: node.get('networks'),
    }));

    return { layoutNodes, layoutEdges };
  }
);

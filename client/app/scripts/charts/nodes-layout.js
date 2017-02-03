import dagre from 'dagre';
import debug from 'debug';
import { fromJS, Map as makeMap, Set as ImmSet } from 'immutable';

import { NODE_BASE_SIZE } from '../constants/styles';
import { EDGE_ID_SEPARATOR } from '../constants/naming';
import { featureIsEnabledAny } from '../utils/feature-utils';
import { buildTopologyCacheId, updateNodeDegrees } from '../utils/topology-utils';

const log = debug('scope:nodes-layout');

const topologyCaches = {};
export const DEFAULT_MARGINS = {top: 0, left: 0};
const NODE_SIZE_FACTOR = NODE_BASE_SIZE;
const NODE_SEPARATION_FACTOR = 1.5 * NODE_BASE_SIZE;
const RANK_SEPARATION_FACTOR = 2.5 * NODE_BASE_SIZE;
let layoutRuns = 0;
let layoutRunsTrivial = 0;

function graphNodeId(id) {
  return id.replace('.', '<DOT>');
}

function fromGraphNodeId(encodedId) {
  return encodedId.replace('<DOT>', '.');
}

/**
 * Layout engine runner
 * After the layout engine run nodes and edges have x-y-coordinates. Engine is
 * not run if the number of nodes is bigger than `MAX_NODES`.
 * @param  {Object} graph dagre graph instance
 * @param  {Map} imNodes new node set
 * @param  {Map} imEdges new edge set
 * @return {Object}         Layout with nodes, edges, dimensions
 */
function runLayoutEngine(graph, imNodes, imEdges) {
  let nodes = imNodes;
  let edges = imEdges;

  const ranksep = RANK_SEPARATION_FACTOR;
  const nodesep = NODE_SEPARATION_FACTOR;
  const nodeWidth = NODE_SIZE_FACTOR;
  const nodeHeight = NODE_SIZE_FACTOR;

  // configure node margins
  graph.setGraph({
    nodesep,
    ranksep
  });

  // add nodes to the graph if not already there
  nodes.forEach((node) => {
    const gNodeId = graphNodeId(node.get('id'));
    if (!graph.hasNode(gNodeId)) {
      graph.setNode(gNodeId, {
        width: nodeWidth,
        height: nodeHeight
      });
    }
  });

  // remove nodes that are no longer there or are 0-degree nodes
  graph.nodes().forEach((gNodeId) => {
    const nodeId = fromGraphNodeId(gNodeId);
    if (!nodes.has(nodeId) || nodes.get(nodeId).get('degree') === 0) {
      graph.removeNode(gNodeId);
    }
  });

  // add edges to the graph if not already there
  edges.forEach((edge) => {
    const s = graphNodeId(edge.get('source'));
    const t = graphNodeId(edge.get('target'));
    if (!graph.hasEdge(s, t)) {
      const virtualNodes = s === t ? 1 : 0;
      graph.setEdge(s, t, {id: edge.get('id'), minlen: virtualNodes});
    }
  });

  // remove edges that are no longer there
  graph.edges().forEach((edgeObj) => {
    const edge = [fromGraphNodeId(edgeObj.v), fromGraphNodeId(edgeObj.w)];
    const edgeId = edge.join(EDGE_ID_SEPARATOR);
    if (!edges.has(edgeId)) {
      graph.removeEdge(edgeObj.v, edgeObj.w);
    }
  });

  dagre.layout(graph, { debugTiming: false });
  const layout = graph.graph();

  // apply coordinates to nodes and edges

  graph.nodes().forEach((gNodeId) => {
    const graphNode = graph.node(gNodeId);
    const nodeId = fromGraphNodeId(gNodeId);
    nodes = nodes.setIn([nodeId, 'x'], graphNode.x);
    nodes = nodes.setIn([nodeId, 'y'], graphNode.y);
  });

  graph.edges().forEach((graphEdge) => {
    const graphEdgeMeta = graph.edge(graphEdge);
    const edge = edges.get(graphEdgeMeta.id);
    let points = fromJS(graphEdgeMeta.points);

    // set beginning and end points to node coordinates to ignore node bounding box
    const source = nodes.get(fromGraphNodeId(edge.get('source')));
    const target = nodes.get(fromGraphNodeId(edge.get('target')));
    points = points.mergeIn([0], {x: source.get('x'), y: source.get('y')});
    points = points.mergeIn([points.size - 1], {x: target.get('x'), y: target.get('y')});

    edges = edges.setIn([graphEdgeMeta.id, 'points'], points);
  });

  // return object with the width and height of layout
  return {
    graphWidth: layout.width,
    graphHeight: layout.height,
    width: layout.width,
    height: layout.height,
    nodes,
    edges
  };
}

/**
 * Adds `points` array to edge based on location of source and target
 * @param {Map} edge           new edge
 * @param {Map} nodeCache      all nodes
 * @returns {Map}              modified edge
 */
function setSimpleEdgePoints(edge, nodeCache) {
  const source = nodeCache.get(edge.get('source'));
  const target = nodeCache.get(edge.get('target'));
  return edge.set('points', fromJS([
    {x: source.get('x'), y: source.get('y')},
    {x: target.get('x'), y: target.get('y')}
  ]));
}

/**
 * Layout nodes that have rank that already exists.
 * Relies on only nodes being added that have a connection to an existing node
 * while having a rank of an existing node. They will be laid out in the same
 * line as the latter, with a direct connection between the existing and the new node.
 * @param  {object} layout    Layout with nodes and edges
 * @param  {Map} nodeCache    previous nodes
 * @param  {object} opts      Options
 * @return {object}           new layout object
 */
export function doLayoutNewNodesOfExistingRank(layout, nodeCache) {
  const result = Object.assign({}, layout);
  const nodesep = NODE_SEPARATION_FACTOR;
  const nodeWidth = NODE_SIZE_FACTOR;

  // determine new nodes
  const oldNodes = ImmSet.fromKeys(nodeCache);
  const newNodes = ImmSet.fromKeys(layout.nodes.filter(n => n.get('degree') > 0))
    .subtract(oldNodes);
  result.nodes = layout.nodes.map((n) => {
    if (newNodes.contains(n.get('id'))) {
      const nodesSameRank = nodeCache.filter(nn => nn.get('rank') === n.get('rank'));
      if (nodesSameRank.size > 0) {
        const y = nodesSameRank.first().get('y');
        const x = nodesSameRank.maxBy(nn => nn.get('x')).get('x') + nodesep + nodeWidth;
        return n.merge({ x, y });
      }
      return n;
    }
    return n;
  });

  result.edges = layout.edges.map((edge) => {
    if (!edge.has('points')) {
      return setSimpleEdgePoints(edge, layout.nodes);
    }
    return edge;
  });

  return result;
}

/**
 * Add coordinates to 0-degree nodes using a square layout
 * Depending on the previous layout run's graph aspect ratio, the square will be
 * placed on the right side or below the graph.
 * @param  {Object} layout Layout with nodes and edges
 * @param  {Object} opts   Options with node distances
 * @return {Object}        modified layout
 */
function layoutSingleNodes(layout, opts) {
  const result = Object.assign({}, layout);
  const options = opts || {};
  const margins = options.margins || DEFAULT_MARGINS;
  const ranksep = RANK_SEPARATION_FACTOR / 2; // dagre splits it in half
  const nodesep = NODE_SEPARATION_FACTOR;
  const nodeWidth = NODE_SIZE_FACTOR;
  const nodeHeight = NODE_SIZE_FACTOR;
  const graphHeight = layout.graphHeight || layout.height;
  const graphWidth = layout.graphWidth || layout.width;
  const aspectRatio = graphHeight ? graphWidth / graphHeight : 1;

  let nodes = layout.nodes;

  // 0-degree nodes
  const singleNodes = nodes.filter(node => node.get('degree') === 0);

  if (singleNodes.size) {
    let offsetX;
    let offsetY;
    const nonSingleNodes = nodes.filter(node => node.get('degree') !== 0);
    if (nonSingleNodes.size > 0) {
      if (aspectRatio < 1) {
        log('laying out single nodes to the right', aspectRatio);
        offsetX = nonSingleNodes.maxBy(node => node.get('x')).get('x');
        offsetY = nonSingleNodes.minBy(node => node.get('y')).get('y');
        if (offsetX) {
          offsetX += nodeWidth + nodesep;
        }
      } else {
        log('laying out single nodes below', aspectRatio);
        offsetX = nonSingleNodes.minBy(node => node.get('x')).get('x');
        offsetY = nonSingleNodes.maxBy(node => node.get('y')).get('y');
        if (offsetY) {
          offsetY += nodeHeight + ranksep;
        }
      }
    }

    // default margins
    offsetX = offsetX || (margins.left + nodeWidth) / 2;
    offsetY = offsetY || (margins.top + nodeHeight) / 2;

    const columns = Math.ceil(Math.sqrt(singleNodes.size));
    let row = 0;
    let col = 0;
    let singleX;
    let singleY;
    nodes = nodes.sortBy(node => node.get('rank')).map((node) => {
      if (singleNodes.has(node.get('id'))) {
        if (col === columns) {
          col = 0;
          row += 1;
        }
        singleX = (col * (nodesep + nodeWidth)) + offsetX;
        singleY = (row * (ranksep + nodeHeight)) + offsetY;
        col += 1;
        return node.merge({
          x: singleX,
          y: singleY
        });
      }
      return node;
    });

    // adjust layout dimensions if graph is now bigger
    result.width = Math.max(layout.width, singleX + (nodeWidth / 2) + nodesep);
    result.height = Math.max(layout.height, singleY + (nodeHeight / 2) + ranksep);
    result.nodes = nodes;
  }

  return result;
}

/**
 * Determine if nodes were added between node sets
 * @param  {Map} nodes     new Map of nodes
 * @param  {Map} cache     old Map of nodes
 * @return {Boolean}       True if nodes had node ids that are not in cache
 */
export function hasUnseenNodes(nodes, cache) {
  const hasUnseen = nodes.size > cache.size
    || !ImmSet.fromKeys(nodes).isSubset(ImmSet.fromKeys(cache));
  if (hasUnseen) {
    log('unseen nodes:', ...ImmSet.fromKeys(nodes).subtract(ImmSet.fromKeys(cache)).toJS());
  }
  return hasUnseen;
}

/**
 * Determine if all new nodes are 0-degree nodes
 * Requires cached nodes (implies a previous layout run).
 * @param  {Map} nodes     new Map of nodes
 * @param  {Map} cache     old Map of nodes
 * @return {Boolean} True if all new nodes are 0-nodes
 */
function hasNewSingleNode(nodes, cache) {
  const oldNodes = ImmSet.fromKeys(cache);
  const newNodes = ImmSet.fromKeys(nodes).subtract(oldNodes);
  const hasNewSingleNodes = newNodes.every(key => nodes.getIn([key, 'degree']) === 0);
  return oldNodes.size > 0 && hasNewSingleNodes;
}

/**
 * Determine if all new nodes are of existing ranks
 * Requires cached nodes (implies a previous layout run).
 * @param  {Map} nodes     new Map of nodes
 * @param  {Map} edges     new Map of edges
 * @param  {Map} cache     old Map of nodes
 * @return {Boolean} True if all new nodes have a rank that already exists
 */
export function hasNewNodesOfExistingRank(nodes, edges, cache) {
  const oldNodes = ImmSet.fromKeys(cache);
  const newNodes = ImmSet.fromKeys(nodes).subtract(oldNodes);

  // if new there are edges that connect 2 new nodes, need a full layout
  const bothNodesNew = edges.find(edge => newNodes.contains(edge.get('source'))
    && newNodes.contains(edge.get('target')));
  if (bothNodesNew) {
    return false;
  }

  const oldRanks = cache.filter(n => n.get('rank')).map(n => n.get('rank')).toSet();
  const hasNewNodesOfExistingRankOrSingle = newNodes.every(key => nodes.getIn([key, 'degree']) === 0
    || oldRanks.contains(nodes.getIn([key, 'rank'])));
  return oldNodes.size > 0 && hasNewNodesOfExistingRankOrSingle;
}

/**
 * Determine if edge has same endpoints in new nodes as well as in the nodeCache
 * @param  {Map}  edge      Edge with source and target
 * @param  {Map}  nodes     new node set
 * @return {Boolean}           True if old and new endpoints have same coordinates
 */
function hasSameEndpoints(cachedEdge, nodes) {
  const oldPoints = cachedEdge.get('points');
  const oldSourcePoint = oldPoints.first();
  const oldTargetPoint = oldPoints.last();
  const newSource = nodes.get(cachedEdge.get('source'));
  const newTarget = nodes.get(cachedEdge.get('target'));
  return (oldSourcePoint && oldTargetPoint && newSource && newTarget
    && oldSourcePoint.get('x') === newSource.get('x')
    && oldSourcePoint.get('y') === newSource.get('y')
    && oldTargetPoint.get('x') === newTarget.get('x')
    && oldTargetPoint.get('y') === newTarget.get('y'));
}

/**
 * Clones a previous layout
 * @param  {Object} layout Layout object
 * @param  {Map} nodes  new nodes
 * @param  {Map} edges  new edges
 * @return {Object}        layout clone
 */
function cloneLayout(layout, nodes, edges) {
  const clone = Object.assign({}, layout, {nodes, edges});
  return clone;
}

/**
 * Copies node properties from previous layout runs to new nodes.
 * This assumes the cache has data for all new nodes.
 * @param  {Object} layout Layout
 * @param  {Object} nodeCache  cache of all old nodes
 * @param  {Object} edgeCache  cache of all old edges
 * @return {Object}        modified layout
 */
function copyLayoutProperties(layout, nodeCache, edgeCache) {
  const result = Object.assign({}, layout);
  result.nodes = layout.nodes.map(node => (nodeCache.has(node.get('id'))
    ? node.merge(nodeCache.get(node.get('id'))) : node));
  result.edges = layout.edges.map((edge) => {
    if (edgeCache.has(edge.get('id'))
      && hasSameEndpoints(edgeCache.get(edge.get('id')), result.nodes)) {
      return edge.merge(edgeCache.get(edge.get('id')));
    } else if (nodeCache.get(edge.get('source')) && nodeCache.get(edge.get('target'))) {
      return setSimpleEdgePoints(edge, nodeCache);
    }
    return edge;
  });
  return result;
}

/**
 * Layout of nodes and edges
 * If a previous layout was given and not too much changed, the previous layout
 * is changed and returned. Otherwise does a new layout engine run.
 * @param  {Map} immNodes All nodes
 * @param  {Map} immEdges All edges
 * @param  {object} opts  width, height, margins, etc...
 * @return {object} graph object with nodes, edges, dimensions
 */
export function doLayout(immNodes, immEdges, opts) {
  const options = opts || {};
  const cacheId = buildTopologyCacheId(options.topologyId, options.topologyOptions);

  // one engine and node and edge caches per topology, to keep renderings similar
  if (options.noCache || !topologyCaches[cacheId]) {
    topologyCaches[cacheId] = {
      nodeCache: makeMap(),
      edgeCache: makeMap(),
      graph: new dagre.graphlib.Graph({})
    };
  }

  const cache = topologyCaches[cacheId];
  const cachedLayout = options.cachedLayout || cache.cachedLayout;
  const nodeCache = options.nodeCache || cache.nodeCache;
  const edgeCache = options.edgeCache || cache.edgeCache;
  const useCache = !options.forceRelayout && cachedLayout && nodeCache && edgeCache;
  let layout;

  layoutRuns += 1;
  if (useCache && !hasUnseenNodes(immNodes, nodeCache)) {
    layoutRunsTrivial += 1;
    log('skip layout, trivial adjustment', layoutRunsTrivial, layoutRuns);
    layout = cloneLayout(cachedLayout, immNodes, immEdges);
    // copy old properties, works also if nodes get re-added
    layout = copyLayoutProperties(layout, nodeCache, edgeCache);
  } else {
    const nodesWithDegrees = updateNodeDegrees(immNodes, immEdges);
    if (useCache
      && featureIsEnabledAny('layout-dance', 'layout-dance-single')
      && hasNewSingleNode(nodesWithDegrees, nodeCache)) {
      // special case: new nodes are 0-degree nodes, no need for layout run,
      // they will be laid out further below
      log('skip layout, only 0-degree node(s) added');
      layout = cloneLayout(cachedLayout, nodesWithDegrees, immEdges);
      layout = copyLayoutProperties(layout, nodeCache, edgeCache);
    } else if (useCache
      && featureIsEnabledAny('layout-dance', 'layout-dance-rank')
      && hasNewNodesOfExistingRank(nodesWithDegrees, immEdges, nodeCache)) {
      // special case: few new nodes were added, no need for layout run,
      // they will inserted according to ranks
      log('skip layout, used rank-based insertion');
      layout = cloneLayout(cachedLayout, nodesWithDegrees, immEdges);
      layout = copyLayoutProperties(layout, nodeCache, edgeCache);
      layout = doLayoutNewNodesOfExistingRank(layout, nodeCache);
    } else {
      const graph = cache.graph;
      layout = runLayoutEngine(graph, nodesWithDegrees, immEdges);
      if (!layout) {
        return layout;
      }
    }

    layout = layoutSingleNodes(layout, opts);
  }

  // cache results
  if (layout) {
    cache.cachedLayout = layout;
    cache.nodeCache = cache.nodeCache.merge(layout.nodes);
    cache.edgeCache = cache.edgeCache.merge(layout.edges);
  }

  return layout;
}

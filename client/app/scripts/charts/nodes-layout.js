// polyfill
Object.assign = require('object-assign');

const dagre = require('dagre');
const debug = require('debug')('scope:nodes-layout');
const makeMap = require('immutable').Map;
const ImmSet = require('immutable').Set;
const Naming = require('../constants/naming');

const MAX_NODES = 100;
const topologyCaches = {};
let layoutRuns = 0;
let layoutRunsTrivial = 0;

/**
 * Layout engine runner
 * After the layout engine run nodes and edges have x-y-coordinates. Engine is
 * not run if the number of nodes is bigger than `MAX_NODES`.
 * @param  {Object} graph dagre graph instance
 * @param  {Map} imNodes new node set
 * @param  {Map} imEdges new edge set
 * @param  {Object} opts    dimensions, scales, etc.
 * @return {Object}         Layout with nodes, edges, dimensions
 */
function runLayoutEngine(graph, imNodes, imEdges, opts) {
  let nodes = imNodes;
  let edges = imEdges;

  if (nodes.size > MAX_NODES) {
    debug('Too many nodes for graph layout engine. Limit: ' + MAX_NODES);
    return null;
  }

  const options = opts || {};
  const margins = options.margins || {top: 0, left: 0};
  const width = options.width || 800;
  const height = options.height || width / 2;
  const scale = options.scale || (val => val * 2);

  // configure node margins
  graph.setGraph({
    nodesep: scale(2.5),
    ranksep: scale(2.5)
  });

  // add nodes to the graph if not already there
  nodes.forEach(node => {
    if (!graph.hasNode(node.get('id'))) {
      graph.setNode(node.get('id'), {
        id: node.get('id'),
        width: scale(1),
        height: scale(1)
      });
    }
  });

  // remove nodes that are no longer there
  graph.nodes().forEach(nodeid => {
    if (!nodes.has(nodeid)) {
      graph.removeNode(nodeid);
    }
  });

  // add edges to the graph if not already there
  edges.forEach(edge => {
    if (!graph.hasEdge(edge.get('source'), edge.get('target'))) {
      const virtualNodes = edge.get('source') === edge.get('target') ? 1 : 0;
      graph.setEdge(
        edge.get('source'),
        edge.get('target'),
        {id: edge.get('id'), minlen: virtualNodes}
      );
    }
  });

  // remove edges that are no longer there
  graph.edges().forEach(edgeObj => {
    const edge = [edgeObj.v, edgeObj.w];
    const edgeId = edge.join(Naming.EDGE_ID_SEPARATOR);
    if (!edges.has(edgeId)) {
      graph.removeEdge(edgeObj.v, edgeObj.w);
    }
  });

  dagre.layout(graph);
  const layout = graph.graph();

  // shifting graph coordinates to center

  let offsetX = 0 + margins.left;
  let offsetY = 0 + margins.top;

  if (layout.width < width) {
    offsetX = (width - layout.width) / 2 + margins.left;
  }
  if (layout.height < height) {
    offsetY = (height - layout.height) / 2 + margins.top;
  }

  // apply coordinates to nodes and edges

  graph.nodes().forEach(id => {
    const graphNode = graph.node(id);
    nodes = nodes.setIn([id, 'x'], graphNode.x + offsetX);
    nodes = nodes.setIn([id, 'y'], graphNode.y + offsetY);
  });

  graph.edges().forEach(id => {
    const graphEdge = graph.edge(id);
    const edge = edges.get(graphEdge.id);
    const points = graphEdge.points.map(point => ({
      x: point.x + offsetX,
      y: point.y + offsetY
    }));

    // set beginning and end points to node coordinates to ignore node bounding box
    const source = nodes.get(edge.get('source'));
    const target = nodes.get(edge.get('target'));
    points[0] = {x: source.get('x'), y: source.get('y')};
    points[points.length - 1] = {x: target.get('x'), y: target.get('y')};

    edges = edges.setIn([graphEdge.id, 'points'], points);
  });

  // return object with the width and height of layout
  layout.nodes = nodes;
  layout.edges = edges;
  return layout;
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
  return edge.set('points', [
    {x: source.get('x'), y: source.get('y')},
    {x: target.get('x'), y: target.get('y')}
  ]);
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
    debug('unseen nodes:', ...ImmSet.fromKeys(nodes).subtract(ImmSet.fromKeys(cache)).toJS());
  }
  return hasUnseen;
}

/**
 * Determine if edge has same endpoints in new nodes as well as in the nodeCache
 * @param  {Map}  edge      Edge with source and target
 * @param  {Map}  nodes     new node set
 * @return {Boolean}           True if old and new endpoints have same coordinates
 */
function hasSameEndpoints(cachedEdge, nodes) {
  const oldPoints = cachedEdge.get('points');
  const oldSourcePoint = oldPoints[0];
  const oldTargetPoint = oldPoints[oldPoints.length - 1];
  const newSource = nodes.get(cachedEdge.get('source'));
  const newTarget = nodes.get(cachedEdge.get('target'));
  return (oldSourcePoint.x === newSource.get('x')
    && oldSourcePoint.y === newSource.get('y')
    && oldTargetPoint.x === newTarget.get('x')
    && oldTargetPoint.y === newTarget.get('y'));
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
  layout.nodes = layout.nodes.map(node => {
    return node.merge(nodeCache.get(node.get('id')));
  });
  layout.edges = layout.edges.map(edge => {
    if (edgeCache.has(edge.get('id')) && hasSameEndpoints(edgeCache.get(edge.get('id')), layout.nodes)) {
      return edge.merge(edgeCache.get(edge.get('id')));
    }
    return setSimpleEdgePoints(edge, nodeCache);
  });
  return layout;
}

/**
 * Layout of nodes and edges
 * If a previous layout was given and not too much changed, the previous layout
 * is changed and returned. Otherwise does a new layout engine run.
 * @param  {Map} nodes All nodes
 * @param  {Map} edges All edges
 * @param  {object} opts  width, height, margins, etc...
 * @return {object} graph object with nodes, edges, dimensions
 */
export function doLayout(nodes, edges, opts) {
  const options = opts || {};
  const topologyId = options.topologyId || 'noId';

  // one engine and node and edge caches per topology, to keep renderings similar
  if (!topologyCaches[topologyId]) {
    topologyCaches[topologyId] = {
      nodeCache: makeMap(),
      edgeCache: makeMap(),
      graph: new dagre.graphlib.Graph({})
    };
  }

  const cache = topologyCaches[topologyId];
  const cachedLayout = options.cachedLayout || cache.cachedLayout;
  const nodeCache = options.nodeCache || cache.nodeCache;
  const edgeCache = options.edgeCache || cache.edgeCache;
  let layout;

  ++layoutRuns;
  if (cachedLayout && nodeCache && edgeCache && !hasUnseenNodes(nodes, nodeCache)) {
    debug('skip layout, trivial adjustment', ++layoutRunsTrivial, layoutRuns);
    layout = cloneLayout(cachedLayout, nodes, edges);
    // copy old properties, works also if nodes get re-added
    layout = copyLayoutProperties(layout, nodeCache, edgeCache);
  } else {
    const graph = cache.graph;
    layout = runLayoutEngine(graph, nodes, edges, opts);
  }

  // cache results
  if (layout) {
    cache.cachedLayout = layout;
    cache.nodeCache = cache.nodeCache.merge(layout.nodes);
    cache.edgeCache = cache.edgeCache.merge(layout.edges);
  }

  return layout;
}

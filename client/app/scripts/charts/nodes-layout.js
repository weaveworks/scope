const dagre = require('dagre');
const debug = require('debug')('scope:nodes-layout');
const ImmSet = require('immutable').Set;
const Naming = require('../constants/naming');

const MAX_NODES = 100;
const topologyGraphs = {};

/**
 * Wrapper around layout engine
 * After the layout engine run nodes and edges have x-y-coordinates. Creates and
 * reuses one engine per topology. Engine is not run if the number of nodes is
 * bigger than `MAX_NODES`.
 * @param  {Map} imNodes new node set
 * @param  {Map} imEdges new edge set
 * @param  {Object} opts    dimensions, scales, etc.
 * @return {Object}         Layout with nodes, edges, dimensions
 */
function runLayoutEngine(imNodes, imEdges, opts) {
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
  const topologyId = options.topologyId || 'noId';

  // one engine per topology, to keep renderings similar
  if (!topologyGraphs[topologyId]) {
    topologyGraphs[topologyId] = new dagre.graphlib.Graph({});
  }
  const graph = topologyGraphs[topologyId];

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
 * Modifies add/remove edges to a previous layout based on what is present in
 * the new edge set
 * @param {Map} nodes          new node set
 * @param {Map} edges          new edges
 * @param {Object} previousLayout modified layout
 */
function addRemoveLayoutEdges(nodes, edges, previousLayout) {
  const previousEdges = previousLayout.edges;

  // remove old edges
  let layoutEdges = previousEdges.filter(edge => {
    return edges.has(edge.get('id'));
  });

  // add new edges with points from source and target
  let source;
  let target;
  let layoutEdge;
  edges.forEach(edge => {
    if (!layoutEdges.has(edge.get('id'))) {
      source = nodes.get(edge.get('source'));
      target = nodes.get(edge.get('target'));
      layoutEdge = edge.set('points', [
        {x: source.get('x'), y: source.get('y')},
        {x: target.get('x'), y: target.get('y')}
      ]);
      layoutEdges = layoutEdges.set(layoutEdge.get('id'), layoutEdge);
    }
  });

  previousLayout.edges = layoutEdges;
  return previousLayout;
}

/**
 * Removes nodes from `previousLayout.nodes` that are not in `nodes` and returns
 * the modified `previousLayout`.
 * @param  {Map} nodes          new set of nodes
 * @param  {Map} edges          new set of edges
 * @param  {object} previousLayout old layout
 * @return {Object}                Layout with nodes and and edges
 */
function removeOldLayoutNodes(nodes, edges, previousLayout) {
  const previousNodes = previousLayout.nodes;
  let layoutNodes = previousNodes.filter(node => {
    return nodes.has(node.get('id'));
  });
  previousLayout.nodes = layoutNodes;
  return previousLayout;
}

/**
 * Determine if two node sets have the same nodes
 * @param  {Map}  nodes     new node set
 * @param  {Map}  prevNodes old node set
 * @return {Boolean}           True if node ids of both sets are the same
 */
function hasSameNodes(nodes, prevNodes) {
  return ImmSet.fromKeys(nodes).equals(ImmSet.fromKeys(prevNodes));
}

/**
 * Determine if nodes were removed between node sets
 * @param  {Map} nodes     new Map of nodes
 * @param  {Map} prevNodes old Map of nodes
 * @return {Boolean}           True if nodes had no new node ids
 */
function wereNodesOnlyRemoved(nodes, prevNodes) {
  return (nodes.size < prevNodes.size
    && ImmSet.fromKeys(nodes).isSubset(ImmSet.fromKeys(prevNodes)));
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
  const previous = options.history && options.history.first();
  let layout;

  if (previous) {
    if (hasSameNodes(nodes, previous.nodes)) {
      debug('skip layout, only edges changed', edges.size, previous.edges.size);
      layout = addRemoveLayoutEdges(nodes, edges, previous);
    } else if (wereNodesOnlyRemoved(nodes, previous.nodes)) {
      debug('skip layout, only nodes removed', nodes.size, previous.nodes.size);
      layout = removeOldLayoutNodes(nodes, edges, previous);
      layout = addRemoveLayoutEdges(nodes, edges, layout);
    }
  }

  if (layout === undefined) {
    layout = runLayoutEngine(nodes, edges, opts);
  }

  return layout;
}

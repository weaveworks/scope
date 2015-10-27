const dagre = require('dagre');
const debug = require('debug')('scope:nodes-layout');
const Naming = require('../constants/naming');

const MAX_NODES = 100;
const topologyGraphs = {};

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
 * Layout of nodes and edges
 * @param  {Map} nodes All nodes
 * @param  {Map} edges All edges
 * @param  {object} opts  width, height, margins, etc...
 * @return {object} graph object with nodes, edges, dimensions
 */
export function doLayout(nodes, edges, opts) {
  // const options = opts || {};
  // const history = options.history || [];

  return runLayoutEngine(nodes, edges, opts);
}

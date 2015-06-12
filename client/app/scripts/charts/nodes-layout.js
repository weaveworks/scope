const dagre = require('dagre');
const debug = require('debug')('nodes-layout');
const _ = require('lodash');

const MAX_NODES = 100;

const doLayout = function(nodes, edges, width, height, scale, margins) {
  let offsetX = 0 + margins.left;
  let offsetY = 0 + margins.top;
  const g = new dagre.graphlib.Graph({});

  if (_.size(nodes) > MAX_NODES) {
    debug('Too many nodes for graph layout engine. Limit: ' + MAX_NODES);
    return null;
  }

  // configure node margins

  g.setGraph({
    nodesep: scale(2.5),
    ranksep: scale(2.5)
  });

  // add nodes and edges to layout engine

  _.each(nodes, function(node) {
    g.setNode(node.id, {id: node.id, width: scale(0.75), height: scale(0.75)});
  });

  _.each(edges, function(edge) {
    const virtualNodes = edge.source.id === edge.target.id ? 1 : 0;
    g.setEdge(edge.source.id, edge.target.id, {id: edge.id, minlen: virtualNodes});
  });

  dagre.layout(g);

  const graph = g.graph();

  // shifting graph coordinates to center

  if (graph.width < width) {
    offsetX = (width - graph.width) / 2 + margins.left;
  }
  if (graph.height < height) {
    offsetY = (height - graph.height) / 2 + margins.top;
  }

  // apply coordinates to nodes and edges

  g.nodes().forEach(function(id) {
    const node = nodes[id];
    const graphNode = g.node(id);
    node.x = graphNode.x + offsetX;
    node.y = graphNode.y + offsetY;
  });

  g.edges().forEach(function(id) {
    const graphEdge = g.edge(id);
    const edge = edges[graphEdge.id];
    _.each(graphEdge.points, function(point) {
      point.x += offsetX;
      point.y += offsetY;
    });
    edge.points = graphEdge.points;
  });

  // return object with width and height of layout

  return graph;
};

module.exports = {
  doLayout: doLayout
};

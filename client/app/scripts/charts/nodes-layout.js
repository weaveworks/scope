var dagre = require('dagre');
var _ = require('lodash');

var MAX_NODES = 100;

var doLayout = function(nodes, edges, width, height, scale, margins) {
  var offsetX = 0 + margins.left;
  var offsetY = 0 + margins.top;
  var g = new dagre.graphlib.Graph({});

  if (_.size(nodes) > MAX_NODES) {
    console.error('Too many nodes for graph layout engine. Limit: ' + MAX_NODES);
    return;
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
    var virtualNodes = edge.source.id === edge.target.id ? 1 : 0;
    g.setEdge(edge.source.id, edge.target.id, {id: edge.id, minlen: virtualNodes});
  });

  dagre.layout(g);

  var graph = g.graph();

  // shifting graph coordinates to center

  if (graph.width < width) {
    offsetX = (width - graph.width) / 2 + margins.left;
  }
  if (graph.height < height) {
    offsetY = (height - graph.height) / 2 + margins.top;
  }

  // apply coordinates to nodes and edges

  g.nodes().forEach(function(id) {
    var node = nodes[id];
    var graphNode = g.node(id);
    node.x = graphNode.x + offsetX;
    node.y = graphNode.y + offsetY;
  });

  g.edges().forEach(function(id) {
    var graphEdge = g.edge(id);
    var edge = edges[graphEdge.id];
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

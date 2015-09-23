const dagre = require('dagre');
const debug = require('debug')('scope:nodes-layout');
const Naming = require('../constants/naming');
const _ = require('lodash');

const MAX_NODES = 100;
const topologyGraphs = {};

const doLayout = function(nodes, edges, width, height, scale, margins, topologyId) {
  let offsetX = 0 + margins.left;
  let offsetY = 0 + margins.top;
  let graph;

  if (_.size(nodes) > MAX_NODES) {
    debug('Too many nodes for graph layout engine. Limit: ' + MAX_NODES);
    return null;
  }

  // one engine per topology, to keep renderings similar
  if (!topologyGraphs[topologyId]) {
    topologyGraphs[topologyId] = new dagre.graphlib.Graph({
      compound: true
    });
  }
  graph = topologyGraphs[topologyId];

  // configure node margins
  graph.setGraph({
    // nodesep: scale(1.5),
    ranksep: scale(1.5)
  });

  // add nodes to the graph if not already there, and collect ranks
  _.each(nodes, function(node) {
    if (!graph.hasNode(node.id)) {
      graph.setNode(node.id, {
        id: node.id,
        width: scale(1.0),
        height: scale(1.0)
      });
    }
  });

  // remove nodes that are no longer there
  _.each(graph.nodes(), function(nodeid) {
    if (!_.has(nodes, nodeid)) {
      graph.removeNode(nodeid);
    }
  });

  // add node ranks for clustering
  const ranks = _.uniq(_.pluck(nodes, 'rank'));
  _.each(ranks, function(rank) {
    const rankId = 'rank_' + rank;
    if (!graph.hasNode(rankId)) {
      graph.setNode(rankId, {});
    }
  });
  _.each(nodes, function(node) {
    if (node.rank) {
      graph.setParent(node.id, 'rank_' + node.rank);
    }
  });

  // TODO remove rank nodes

  // add edges to the graph if not already there
  _.each(edges, function(edge) {
    if (!graph.hasEdge(edge.source.id, edge.target.id)) {
      const virtualNodes = edge.source.id === edge.target.id ? 1 : 0;
      graph.setEdge(edge.source.id, edge.target.id, {id: edge.id, minlen: virtualNodes});
    }
  });

  // remoed egdes that are no longer there
  _.each(graph.edges(), function(edgeObj) {
    const edge = [edgeObj.v, edgeObj.w];
    const edgeId = edge.join(Naming.EDGE_ID_SEPARATOR);
    if (!_.has(edges, edgeId)) {
      graph.removeEdge(edgeObj.v, edgeObj.w);
    }
  });

  dagre.layout(graph);

  const layout = graph.graph();

  // shifting graph coordinates to center

  if (layout.width < width) {
    offsetX = (width - layout.width) / 2 + margins.left;
  }
  if (layout.height < height) {
    offsetY = (height - layout.height) / 2 + margins.top;
  }

  // apply coordinates to nodes and edges

  graph.nodes().forEach(function(id) {
    const node = nodes[id];
    if (node) {
      const graphNode = graph.node(id);
      node.x = graphNode.x + offsetX;
      node.y = graphNode.y + offsetY;
    }
  });

  graph.edges().forEach(function(id) {
    const graphEdge = graph.edge(id);
    const edge = edges[graphEdge.id];
    _.each(graphEdge.points, function(point) {
      point.x += offsetX;
      point.y += offsetY;
    });
    edge.points = graphEdge.points;
    // set beginning and end points to node coordinates to ignore node bounding box
    edge.points[0] = {x: edge.source.x, y: edge.source.y};
    edge.points[edge.points.length - 1] = {x: edge.target.x, y: edge.target.y};
  });

  // return object with the width and height of layout

  return layout;
};

module.exports = {
  doLayout: doLayout
};

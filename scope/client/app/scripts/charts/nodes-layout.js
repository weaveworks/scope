var d3 = require('d3');
var dagre = require('dagre');
var _ = require('lodash');

var MAX_NODES = 100;

var doLayout = function(nodes, edges, width, height, scale) {
    var offsetX = 0;
    var offsetY = 80;
    var g = new dagre.graphlib.Graph({
    });

    var line = d3.svg.line()
        .interpolate("cardinal")
        .x(function(d) { return d.x; })
        .y(function(d) { return d.y; });

    g.setGraph({
        nodesep: scale(1.75),
        ranksep: scale(1.5)
    });

    _.each(nodes, function(node) {
        g.setNode(node.id, {id: node.id, width: scale(0.75), height: scale(0.75)});
    });

    _.each(edges, function(edge) {
        g.setEdge(edge.source.id, edge.target.id, {id: edge.id});
    });

    dagre.layout(g);

    var graph = g.graph();
    if (graph.width < width) {
        offsetX = (width - graph.width) / 2;
    }
    if (graph.height < height) {
        offsetY = (height - graph.height) / 2;
    }

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
        edge.path = line(graphEdge.points);
    });
};

module.exports = {
    doLayout: doLayout
};

var d3 = require('d3');
var _ = require('lodash');
var EventEmitter = require('events').EventEmitter;
var Node = require('./node');


function determineAllEdges(nodes, allNodes) {
    var edges = [],
        edgeIds = {},
        nodeIds = _.pluck(nodes, 'id');

    _.each(nodes, function(node) {
        var nodeId = node.id;

        _.each(node.adjacency, function(adjacent) {
            var edge = [nodeId, adjacent],
                edgeId = edge.join('-');

            if (!edgeIds[edgeId] && _.contains(nodeIds, nodeId) && _.contains(nodeIds, adjacent)) {
                edges.push({
                    id: edgeId,
                    value: 5,
                    source: allNodes[edge[0]],
                    target: allNodes[edge[1]]
                });
                edgeIds[edgeId] = true;
            }
        });
    });

    return edges;
}

function getAdjacentNodes(nodes, rootNodeIds) {
    var rootNodes = _.compact(_.map(rootNodeIds, function(nodeId) {
        return nodes[nodeId];
    }));

    var allAdjacentNodeIds = _.union(_.flatten(_.map(rootNodes, function(node) {
        return node.adjacency;
    })), rootNodeIds);

    return _.compact(_.map(allAdjacentNodeIds, function(nodeId) {
        return nodes[nodeId];
    }));
}

function getAdjacentEdges(nodes, root) {
    return _.map(root.adjacency, function(nodeId) {
        var edge = [root.id, nodeId],
            edgeId = edge.join('-');

        return {
            id: edgeId,
            value: 10,
            source: nodes[edge[0]],
            target: nodes[edge[1]]
        };
    });
}

function id(d) {
    return d.id;
}

function degree(d) {
    return d.adjacency ? d.adjacency.length : 1;
}

function getChildren(node, allNodes) {
    return _.map(node.adjacency, function(nodeId) {
        return allNodes[nodeId];
    });
}

function dblclick(d) {
    d.fixed = false;
}

function nodeExplorer() {

    var color = d3.scale.category20();

    var pie = d3.layout.pie()
        .value(degree);

    var radius = d3.scale.sqrt()
        .range([1, 8]);

    var drag = d3.behavior.drag();

    var dispatcher = new EventEmitter();

    var nodeLocations = {};

    var width, height;

    function circleRadius() {
        return width / 4;
    }

    function radialLayout(centerNode, nodes, radius) {
         var slices = pie(_.sortBy(_.filter(nodes, function(node) {
            return !node.layedout && _.contains(centerNode.adjacency, node.id);
         }), 'id'));

        _.each(slices, function(slice) {
            var previousXY = nodeLocations[slice.data.id];

            slice.data.x = centerNode.x + circleRadius() * Math.sin((slice.startAngle + slice.endAngle) / 2);
            slice.data.y = centerNode.y + circleRadius() * Math.cos((slice.startAngle + slice.endAngle) / 2);
            slice.data.layedout = true;
            // nodeLocations[slice.data.id] = [slice.data.forceX, slice.data.forceY];
        });
    }


    function chart(selection) {
        selection.each(function(data) {
            var allNodes = data.nodes;
            var root = allNodes[data.root];
            var expandedNodeIds = data.expandedNodes;

            var centerX = width / 2;
            var centerY = height / 2;

            var nodes = getAdjacentNodes(allNodes, expandedNodeIds);

            if (root) {
                var previousXY = nodeLocations[root.id];
                root.x = previousXY ? previousXY[0] : centerX;
                root.y = previousXY ? previousXY[1] : centerY;
                root.layedout = true;

                _.each(nodes, function(node) {
                    if (node.id !== root.id) {
                        node.layedout = false;
                    }
                });

                _.each(expandedNodeIds, function(nodeId, i) {
                    var centerNode = allNodes[nodeId];
                    if (centerNode) {
                        radialLayout(centerNode, nodes, i ? circleRadius() / 4 : circleRadius());
                    }
                });

                if (nodes.length == 0) {
                    nodes.push(root);
                }
            }

            var edges = determineAllEdges(nodes, allNodes);

            // Select the svg element, if it exists.
            var svg = d3.select(this).selectAll("svg").data([data]);

            // Otherwise, create the skeletal chart.
            var gEnter = svg.enter().append("svg")
		.attr('width', "100%")
		.attr('height', "100%");

            gEnter.append('g')
                .classed('links', true);
            gEnter.append('g')
                .classed('nodes', true);

            var link = svg.select('.links').selectAll(".link")
                    .data(edges, id);
                link.exit()
                    .remove()
                link.enter().append("line")
                    .attr("class", "link")
                    .style("stroke-width", function(d) { return Math.sqrt(d.value); });

            var node = svg.select('.nodes').selectAll(".node")
                    .data(nodes, id);

            node.exit()
                .remove();

            gEnter = node.enter().append("g")
                .attr("class", "node")
                .on("dblclick", dblclick)
                .on('click', function(d) {
                    if (d3.event.defaultPrevented) {
                        return; // click suppressed
                    }
                    dispatcher.emit('node.click', d);
                })
                .call(drag);

            gEnter.append("circle")
                .attr('class', 'outer')
                .style("fill", function(d) {
                    return color(d.label_major);
                })
                .attr("r", function(d) {
                    return radius(degree(d)) + 4;
                });

            gEnter.append("circle")
                .style("fill", function(d) {
                    return color(d.label_major);
                })
                .attr("r", function(d) {
                    return radius(degree(d));
                });

            gEnter.append("text")
                .attr('class', 'label-major')
                .attr("dy", "-.25em")
                .text(function(d) { return d.label_major; });

            gEnter.append("text")
                .attr('class', 'label-minor')
                .attr("dy", ".75em")
                .text(function(d) { return d.label_minor; });

            if (!root) {
                return;
            }

            node
                .classed('node-root', function(d) {
                    return d.id === root.id;
                })
                .classed('node-expanded', function(d) {
                    return _.contains(expandedNodeIds, d.id);
                })
                .classed('node-leaf', function(d) {
                    return _.size(d.adjacency) === 0;
                });

            function updatePositions() {
                link.attr("x1", function(d) { return d.source.x; })
                    .attr("y1", function(d) { return d.source.y; })
                    .attr("x2", function(d) { return d.target.x; })
                    .attr("y2", function(d) { return d.target.y; });

                node.attr("transform", function(d) {
                    return "translate(" + d.x + "," + d.y + ")";
                });

                node.selectAll('text')
                    .attr("dx", function(d) { return d.x > centerX ? radius(degree(d)) + 6 : - radius(degree(d)) - 6})
                    .attr("text-anchor", function(d) { return d.x > centerX ? "start" : "end"; });
            }

            var density = Math.sqrt(nodes.length / (width * height));

            drag.on('drag', function(d) {
                d.x = d3.event.x;
                d.y = d3.event.y;
                updatePositions();
            });

            updatePositions();
        });
    }


    chart.create = function(el, state) {
        d3.select(el)
            .datum(state)
            .call(chart);

        return chart;
    };

    chart.update = _.throttle(function(el, state) {
        d3.select(el)
            .datum(state)
            .call(chart);

        return chart;
    }, 500);

    chart.on = function(event, callback) {
        dispatcher.on(event, callback);
    };

    chart.teardown = function() {
        force.on('tick', null);
    };

    chart.width = function(_) {
        if (!arguments.length) return width;
        width = _;
        return chart;
    };

    chart.height = function(_) {
        if (!arguments.length) return height;
        height = _;
        return chart;
    };

    return chart;
}

module.exports = nodeExplorer;
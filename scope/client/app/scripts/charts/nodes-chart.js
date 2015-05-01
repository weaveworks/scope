var d3 = require('d3');
var dagre = require('dagre');
var _ = require('lodash');
var EventEmitter = require('events').EventEmitter;
var React = require('react');
var Node = require('./node');

var MAX_NODES = 100;

var LayoutEngines = {
    circle: {
        layout: function(nodes, edges, width, height, scale) {
            var centerX = width / 2;
            var centerY = height / 2;
            var radius = Math.min(width, height) / 3;
            var identityPie = d3.layout.pie().value(_.partial(_.identity, 1));
            var line = d3.svg.line()
                .x(function(d) { return d.x; })
                .y(function(d) { return d.y; })
                .interpolate("bundle")
                .tension(0.4);

            var slices = identityPie(_.keys(nodes).sort());

            _.each(slices, function(slice) {
                var node = nodes[slice.data];
                node.layout = 'circle';
                node.x = centerX + radius * Math.sin(slice.startAngle);
                node.y = centerY + radius * Math.cos(slice.startAngle);
                node.angle = (slice.startAngle + slice.endAngle) / 2;
                if (node.angle <= Math.PI) {
                    node.textAnchor = "start";
                } else {
                    node.textAnchor = "end";
                }
            });

            _.each(edges, function(edge) {
                var points = [{
                        x: edge.source.x,
                        y: edge.source.y
                    }, {
                        x: centerX,
                        y: centerY
                    }, {
                        x: edge.target.x,
                        y: edge.target.y
                    }];

                edge.path = line(points);
            });
        }
    },
    force: {
        layout: function(nodes, edges, width, height, scale) {
            var centerX = width / 2;
            var centerY = height / 2;
            var iterations = 0;
            var padding = 5;
            var radius = scale(1);
                
            var line = d3.svg.line()
                .x(function(d) { return d.x; })
                .y(function(d) { return d.y; });

            var force = d3.layout.force()
                .charge(-scale(20))
                .linkDistance(scale(3))
                .size([width, height])
                .nodes(_.values(nodes))
                .links(edges)
                .on('tick', function() {
                    _.each(nodes, collide(0.5));
                });

            function collide(alpha) {
                var quadtree = d3.geom.quadtree(nodes);

                return function(d) {
                    var rb = 2 * radius + padding,
                        nx1 = d.x - rb,
                        nx2 = d.x + rb,
                        ny1 = d.y - rb,
                        ny2 = d.y + rb;

                    quadtree.visit(function(quad, x1, y1, x2, y2) {
                        if (quad.point && (quad.point !== d)) {
                            var x = d.x - quad.point.x,
                                y = d.y - quad.point.y,
                                l = Math.sqrt(x * x + y * y);
                            if (l < rb) {
                                l = (l - rb) / l * alpha;
                                d.x -= x *= l;
                                d.y -= y *= l;
                                quad.point.x += x;
                                quad.point.y += y;
                            }
                        }
                        return x1 > nx2 || x2 < nx1 || y1 > ny2 || y2 < ny1;
                    });
                };
            }

            force.start();

            if (_.some(nodes, {layout: 'force'})) {
                force.alpha(0.025);
            }

            while(force.alpha() > 0.01) {
                force.tick();
                if(iterations++ > 200) {
                    break;// Avoids infinite looping
                }
            }

            _.each(nodes, function(node) {
                node.layout = 'force';
                delete node.angle;
            });

            _.each(edges, function(edge) {
                var points = [{
                        x: edge.source.x,
                        y: edge.source.y
                    }, {
                        x: edge.target.x,
                        y: edge.target.y
                    }];

                edge.path = line(points);
            });
        }
    },
    layered: {
        layout: function(nodes, edges, width, height, scale) {
            var offsetX = 0;
            var offsetY = 0;
            var g = new dagre.graphlib.Graph();

            var line = d3.svg.line()
                .interpolate("cardinal")
                .x(function(d) { return d.x; })
                .y(function(d) { return d.y; });

            g.setGraph({});

            _.each(nodes, function(node) {
                node.layout = 'layered';
                g.setNode(node.id, {id: node.id, width: scale(0.75), height: scale(0.5)});
                delete node.angle;
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
        }
    },
    solar: {
        layout: function(nodes, edges, width, height, scale, rootNodes) {
            var centerX = width / 2;
            var centerY = height / 2;
            var radius = Math.min(width, height) / 3;
            var identityPie = d3.layout.pie().value(_.partial(_.identity, 1));
            var line = d3.svg.line()
                .x(function(d) { return d.x; })
                .y(function(d) { return d.y; })
                .interpolate("bundle")
                .tension(0.4);

            var scaleRange = scale.range();
            scaleRange[1] = scaleRange[1] / 2;
            scale.range(scaleRange);

            var centerId = rootNodes[0];
            var centerNode = nodes[centerId];
            centerNode.x = centerX;
            centerNode.y = centerY;

            var circleNodes = _.without(_.keys(nodes), centerId).sort();
            var slices = identityPie(circleNodes);

            _.each(slices, function(slice) {
                var node = nodes[slice.data];
                node.layout = 'solar';
                node.x = centerX + radius * Math.sin(slice.startAngle);
                node.y = centerY + radius * Math.cos(slice.startAngle);
                node.angle = (slice.startAngle + slice.endAngle) / 2;
                if (node.angle <= Math.PI) {
                    node.textAnchor = "start";
                } else {
                    node.textAnchor = "end";
                }
            });

            _.each(edges, function(edge) {
                var points = [{
                        x: edge.source.x,
                        y: edge.source.y
                    }, {
                        x: centerX,
                        y: centerY
                    }, {
                        x: edge.target.x,
                        y: edge.target.y
                    }];

                edge.path = line(points);
            });
        }
    },
    square: {
        layout: function(nodes, edges, width, height, scale) {
            var centerX = width / 2;
            var centerY = height / 2;
            var expanse = Math.min(width, height) - 2 * scale(1);
            var n = _.size(nodes);
            var m = Math.ceil(Math.sqrt(n));
            var distance = expanse / m;
            var offsetX = (width - expanse) / 2;
            var offsetY = (height - expanse) / 2 + scale(0.5);

            var line = d3.svg.line()
                .x(function(d) { return d.x; })
                .y(function(d) { return d.y; })
                .interpolate("bundle")
                .tension(0.1);

            _.each(_.values(nodes), function(node, i) {
                var row = Math.floor(i / m);
                var col = i % m;
                node.x = offsetX + row * distance;
                node.y = offsetY + col * distance;

                node.layout = 'square';
                delete node.angle;
            });

            _.each(edges, function(edge) {
                var points = [{
                        x: edge.source.x,
                        y: edge.source.y
                    }, {
                        x: centerX,
                        y: centerY
                    }, {
                        x: edge.target.x,
                        y: edge.target.y
                    }];

                edge.path = line(points);
            });
        }
    }
};

var NodesChart = React.createClass({

    getInitialState: function() {
        return {
            nodes: {},
            edges: {}
        };
    },

    initNodes: function(topology, prevNodes) {
        var centerX = this.props.width / 2;
        var centerY = this.props.height / 2;
        var nodes = {};

        _.each(topology, function(node, id) {
            nodes[id] = prevNodes[id] || {};
            _.defaults(nodes[id], {
                x: centerX,
                y: centerY,
                textAnchor: 'start'
            });
            _.assign(nodes[id], {
                id: id,
                label: node.label_major,
                degree: _.size(node.adjacency)
            });
        }, this);

        return nodes;
    },

    initEdges: function(topology, nodes) {
        var edges = {};

        _.each(topology, function(node) {
            _.each(node.adjacency, function(adjacent) {
                var edge = [node.id, adjacent],
                    edgeId = edge.join('-');

                if (!edges[edgeId]) {
                    var source = nodes[edge[0]];
                    var target = nodes[edge[1]];

                    if(!source || !target) {
                        console.error("Missing edge node", edge[0], source, edge[1], target);
                    }

                    edges[edgeId] = {
                        id: edgeId,
                        value: 1,
                        source: source,
                        target: target
                    };
                }
            });
        }, this);

        return edges;
    },

    getNodes: function(nodes, scale) {
        return _.map(nodes, function (node) {
            var highlighted = _.includes(this.props.highlightedNodes, node.id);
            return (
                <Node
                    highlighted={highlighted}
                    onClick={this.props.onNodeClick}
                    key={node.id}
                    id={node.id}
                    angle={node.angle}
                    textAnchor={node.textAnchor}
                    label={node.label}
                    scale={scale}
                    dx={node.x}
                    dy={node.y}
                />
            );
        }, this);
    },

    getEdges: function(edges, scale) {
        return _.map(edges, function(edge) {
            return (
                <path className="link" d={edge.path} key={edge.id} />
            );
        });
    },

    getTopologyFingerprint: function(topology) {
        var nodes = _.keys(topology).sort();
        var fingerprint = [];

        _.each(topology, function(node) {
            fingerprint.push(node.id);
            if (node.adjacency) {
                fingerprint.push(node.adjacency.join(','));
            }
        });
        return fingerprint.join(';');
    },

    updateGraphState: function(props) {
        var nodes = this.initNodes(props.nodes, this.state.nodes);
        var edges = this.initEdges(props.nodes, nodes);

        this.setState({
            nodes: nodes,
            edges: edges 
        });
    },

    componentWillMount: function() {
        this.updateGraphState(this.props);
    },

    componentWillReceiveProps: function(nextProps) {
        if (this.props.layout !== nextProps.layout
            || this.getTopologyFingerprint(nextProps.nodes) !== this.getTopologyFingerprint(this.props.nodes)) {
            this.setState({
                nodes: {},
                edges: {}
            });
        }

        this.updateGraphState(nextProps);
    },

    shouldComponentUpdate: function(nextProps, nextState) {
        return !_.isUndefined(LayoutEngines[nextProps.layout]);
    },

    render: function() {
        var expanse = Math.min(this.props.height, this.props.width);
        var nodeSize = expanse / 2;
        var n = _.size(this.props.nodes);
        var scale = d3.scale.linear().range([0, nodeSize/Math.pow(n, 0.7)]);

        var engine = LayoutEngines[this.props.layout];
        if (!engine) {
            console.error('No layout engine found.');
            return (<div />);
        }

        var layoutId = [this.props.layout, this.props.context].join('-');
        console.time(layoutId);
        engine.layout(
            this.state.nodes,
            this.state.edges,
            this.props.width,
            this.props.height,
            scale,
            this.props.highlightedNodes
        );
        console.timeEnd(layoutId);

        var nodeElements = this.getNodes(this.state.nodes, scale);
        var edgeElements = this.getEdges(this.state.edges, scale);

        return (
            <svg width={this.props.width} height={this.props.height}>
                <g className="canvas">
                    <g className="edges">
                        {edgeElements}
                    </g>
                    <g className="nodes">
                        {nodeElements}
                    </g>
                </g>
            </svg>
        );
    }

});

module.exports = NodesChart;

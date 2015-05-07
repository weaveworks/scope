var _ = require('lodash');
var React = require('react');

var NodesLayout = require('./nodes-layout');
var Node = require('./node');

var MAX_NODES = 100;

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
        if (this.getTopologyFingerprint(nextProps.nodes) !== this.getTopologyFingerprint(this.props.nodes)) {
            this.setState({
                nodes: {},
                edges: {}
            });
        }

        this.updateGraphState(nextProps);
    },

    render: function() {
        var expanse = Math.min(this.props.height, this.props.width);
        var nodeSize = expanse / 2;
        var n = _.size(this.props.nodes);
        var scale = d3.scale.linear().range([0, nodeSize/Math.pow(n, 0.7)]);

        var layoutId = 'layered node chart';
        console.time(layoutId);
        NodesLayout.doLayout(
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
            <svg width="100%" height="100%">
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

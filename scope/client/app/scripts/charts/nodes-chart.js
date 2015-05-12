var _ = require('lodash');
var d3 = require('d3');
var React = require('react');

var NodesLayout = require('./nodes-layout');
var Node = require('./node');

var MAX_NODES = 100;
var MARGINS = {
    top: 120,
    left: 40,
    right: 40,
    bottom: 0
};

var line = d3.svg.line()
    .interpolate("cardinal")
    .x(function(d) { return d.x; })
    .y(function(d) { return d.y; });

var NodesChart = React.createClass({

    getInitialState: function() {
        return {
            nodes: {},
            edges: {},
            nodeScale: 1,
            translate: "0,0",
            scale: 1,
            hasZoomed: false
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
                subLabel: node.label_minor,
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
                    label={node.label}
                    subLabel={node.subLabel}
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
                <path className="link" d={line(edge.points)} key={edge.id} />
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

        var expanse = Math.min(props.height, props.width);
        var nodeSize = expanse / 2;
        var n = _.size(props.nodes);
        var nodeScale = d3.scale.linear().range([0, nodeSize/Math.pow(n, 0.7)]);

        var layoutId = 'layered node chart';
        console.time(layoutId);
        var graph = NodesLayout.doLayout(
            nodes,
            edges,
            props.width,
            props.height,
            nodeScale,
            MARGINS
        );
        console.timeEnd(layoutId);

        // adjust layout based on viewport

        var xFactor = (props.width - MARGINS.left - MARGINS.right) / graph.width;
        var yFactor = props.height / graph.height;
        var zoomFactor = Math.min(xFactor, yFactor);
        var zoomScale = this.state.scale;

        if(this.zoom && !this.state.hasZoomed && zoomFactor < 1) {
            zoomScale = zoomFactor;
            // saving in d3's behavior cache
            this.zoom.scale(zoomFactor);
        }

        this.setState({
            nodes: nodes,
            edges: edges,
            nodeScale: nodeScale,
            scale: zoomScale
        });
    },

    componentWillMount: function() {
        this.updateGraphState(this.props);
    },

    componentDidMount: function() {
        this.zoom = d3.behavior.zoom()
            .scaleExtent([0.1, 2])
            .on('zoom', this.zoomed);

        d3.select('.nodes-chart')
            .call(this.zoom);
    },

    componentWillUnmount: function() {

        // undoing .call(zoom)

        d3.select('.nodes-chart')
            .on("mousedown.zoom", null)
            .on("onwheel", null)
            .on("onmousewheel", null)
            .on("dblclick.zoom", null)
            .on("touchstart.zoom", null);
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

    zoomed: function() {
        this.setState({
            hasZoomed: true,
            translate: d3.event.translate,
            scale: d3.event.scale
        });
    },

    render: function() {
        var nodeElements = this.getNodes(this.state.nodes, this.state.nodeScale);
        var edgeElements = this.getEdges(this.state.edges, this.state.nodeScale);
        var transform = 'translate(' + this.state.translate + ')' +
            ' scale(' + this.state.scale + ')';

        return (
            <svg width="100%" height="100%" className="nodes-chart">
                <g className="canvas" transform={transform}>
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

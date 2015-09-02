const _ = require('lodash');
const d3 = require('d3');
const debug = require('debug')('scope:nodes-chart');
const React = require('react');
const timely = require('timely');

const Edge = require('./edge');
const Naming = require('../constants/naming');
const NodesLayout = require('./nodes-layout');
const Node = require('./node');

const MARGINS = {
  top: 130,
  left: 40,
  right: 40,
  bottom: 0
};

const NodesChart = React.createClass({

  getInitialState: function() {
    return {
      nodes: {},
      edges: {},
      nodeScale: 1,
      translate: '0,0',
      scale: 1,
      hasZoomed: false
    };
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

  componentWillReceiveProps: function(nextProps) {
    if (nextProps.nodes !== this.props.nodes) {
      this.setState({
        nodes: {},
        edges: {}
      });
      this.updateGraphState(nextProps);
    }
  },

  componentWillUnmount: function() {
    // undoing .call(zoom)

    d3.select('.nodes-chart')
      .on('mousedown.zoom', null)
      .on('onwheel', null)
      .on('onmousewheel', null)
      .on('dblclick.zoom', null)
      .on('touchstart.zoom', null);
  },

  getTopologyFingerprint: function(topology) {
    const fingerprint = [];

    _.each(topology, function(node) {
      fingerprint.push(node.id);
      if (node.adjacency) {
        fingerprint.push(node.adjacency.join(','));
      }
    });
    return fingerprint.join(';');
  },

  renderGraphNodes: function(nodes, scale) {
    return _.map(nodes, function(node) {
      const highlighted = _.includes(this.props.highlightedNodeIds, node.id);
      return (
        <Node
          highlighted={highlighted}
          onClick={this.props.onNodeClick}
          key={node.id}
          id={node.id}
          label={node.label}
          pseudo={node.pseudo}
          subLabel={node.subLabel}
          scale={scale}
          dx={node.x}
          dy={node.y}
        />
      );
    }, this);
  },

  renderGraphEdges: function(edges) {
    return _.map(edges, function(edge) {
      const highlighted = _.includes(this.props.highlightedEdgeIds, edge.id);
      return (
        <Edge key={edge.id} id={edge.id} points={edge.points} highlighted={highlighted} />
      );
    }, this);
  },

  render: function() {
    const nodeElements = this.renderGraphNodes(this.state.nodes, this.state.nodeScale);
    const edgeElements = this.renderGraphEdges(this.state.edges, this.state.nodeScale);
    const transform = 'translate(' + this.state.translate + ')' +
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
  },

  initNodes: function(topology, prevNodes) {
    const centerX = this.props.width / 2;
    const centerY = this.props.height / 2;
    const nodes = {};

    topology.forEach(function(node, id) {
      nodes[id] = prevNodes[id] || {};

      // use cached positions if available
      _.defaults(nodes[id], {
        x: centerX,
        y: centerY
      });

      // copy relevant fields to state nodes
      _.assign(nodes[id], {
        id: id,
        label: node.get('label_major'),
        pseudo: node.get('pseudo'),
        subLabel: node.get('label_minor'),
        rank: node.get('rank')
      });
    });

    return nodes;
  },

  initEdges: function(topology, nodes) {
    const edges = {};

    topology.forEach(function(node, nodeId) {
      const adjacency = node.get('adjacency');
      if (adjacency) {
        adjacency.forEach(function(adjacent) {
          const edge = [nodeId, adjacent];
          const edgeId = edge.join(Naming.EDGE_ID_SEPARATOR);

          if (!edges[edgeId]) {
            const source = nodes[edge[0]];
            const target = nodes[edge[1]];

            if (!source || !target) {
              debug('Missing edge node', edge[0], source, edge[1], target);
            }

            edges[edgeId] = {
              id: edgeId,
              value: 1,
              source: source,
              target: target
            };
          }
        });
      }
    });

    return edges;
  },

  updateGraphState: function(props) {
    const n = props.nodes.size;

    if (n === 0) {
      return;
    }

    const nodes = this.initNodes(props.nodes, this.state.nodes);
    const edges = this.initEdges(props.nodes, nodes);

    const expanse = Math.min(props.height, props.width);
    const nodeSize = expanse / 2;
    const nodeScale = d3.scale.linear().range([0, nodeSize / Math.pow(n, 0.7)]);

    const timedLayouter = timely(NodesLayout.doLayout);
    const graph = timedLayouter(
      nodes,
      edges,
      props.width,
      props.height,
      nodeScale,
      MARGINS,
      this.props.topologyId
    );

    debug('graph layout took ' + timedLayouter.time + 'ms');

    // adjust layout based on viewport

    const xFactor = (props.width - MARGINS.left - MARGINS.right) / graph.width;
    const yFactor = props.height / graph.height;
    const zoomFactor = Math.min(xFactor, yFactor);
    let zoomScale = this.state.scale;

    if (this.zoom && !this.state.hasZoomed && zoomFactor > 0 && zoomFactor < 1) {
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

  zoomed: function() {
    this.setState({
      hasZoomed: true,
      translate: d3.event.translate,
      scale: d3.event.scale
    });
  }

});

module.exports = NodesChart;

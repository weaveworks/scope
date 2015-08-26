const _ = require('lodash');
const d3 = require('d3');
const debug = require('debug')('scope:nodes-chart');
const React = require('react');
const Spring = require('react-motion').Spring;

const AppActions = require('../actions/app-actions');
const AppStore = require('../stores/app-store');
const Edge = require('./edge');
const Naming = require('../constants/naming');
const NodesLayout = require('./nodes-layout');
const Node = require('./node');

const MARGINS = {
  top: 100,
  left: 40,
  right: 40,
  bottom: 20
};

// make sure circular layouts a bit denser with 3-6 nodes
const radiusDensity = d3.scale.threshold()
  .domain([3, 6]).range([2.5, 3.5, 3]);

const NodesChart = React.createClass({

  getInitialState: function() {
    return {
      nodes: [],
      edges: [],
      nodeScale: d3.scale.linear(),
      shiftTranslate: [MARGINS.left, MARGINS.top],
      panTranslate: [MARGINS.left, MARGINS.top],
      scale: 1,
      hasZoomed: false,
      autoShifted: false,
      maxNodesExceeded: false
    };
  },

  componentWillMount: function() {
    const state = this.updateGraphState(this.props, this.state);
    this.setState(state);
  },

  componentDidMount: function() {
    this.zoom = d3.behavior.zoom()
      .scaleExtent([0.1, 2])
      .on('zoom', this.zoomed);

    d3.select('.nodes-chart svg')
      .call(this.zoom);

    this.updateGraphState(this.props);
  },

  componentWillReceiveProps: function(nextProps) {
    // gather state, setState should be called only once here
    const state = _.assign({}, this.state);

    // wipe node states when showing different topology
    if (nextProps.topologyId !== this.props.topologyId) {
      _.assign(state, {
        autoShifted: false,
        nodes: [],
        edges: [],
        hasZoomed: false,
        initialLayout: true
      });
      this.updateGraphState(nextProps);
    }
    // FIXME add PureRenderMixin, Immutables, and move the following functions to render()
    if (nextProps.nodes !== this.props.nodes) {
      _.assign(state, this.updateGraphState(nextProps, state));
    }
    if (this.props.selectedNodeId !== nextProps.selectedNodeId) {
      _.assign(state, this.restoreLayout(state));
    }
    if (nextProps.selectedNodeId) {
      _.assign(state, this.centerSelectedNode(nextProps, state));
    }

    this.setState(state);
  },

  componentWillUnmount: function() {
    // undoing .call(zoom)

    d3.select('.nodes-chart svg')
      .on('mousedown.zoom', null)
      .on('onwheel', null)
      .on('onmousewheel', null)
      .on('dblclick.zoom', null)
      .on('touchstart.zoom', null);
  },

  renderGraphNodes: function(nodes, scale) {
    const hasSelectedNode = this.props.selectedNodeId && this.props.nodes.has(this.props.selectedNodeId);
    const adjacency = hasSelectedNode ? AppStore.getAdjacentNodes(this.props.selectedNodeId) : null;
    const onNodeClick = this.props.onNodeClick;

    _.each(nodes, function(node) {
      node.highlighted = _.includes(this.props.highlightedNodeIds, node.id)
        || this.props.selectedNodeId === node.id;
      node.focused = hasSelectedNode
        && (this.props.selectedNodeId === node.id || adjacency.includes(node.id));
      node.blurred = hasSelectedNode && !node.focused;
    }, this);

    return _.chain(nodes)
      .sortBy(function(node) {
        if (node.blurred) {
          return 0;
        }
        if (node.highlighted) {
          return 2;
        }
        return 1;
      })
      .map(function(node) {
        return (
          <Node
            blurred={node.blurred}
            focused={node.focused}
            highlighted={node.highlighted}
            onClick={onNodeClick}
            key={node.id}
            id={node.id}
            label={node.label}
            pseudo={node.pseudo}
            subLabel={node.subLabel}
            rank={node.rank}
            scale={scale}
            dx={node.x}
            dy={node.y}
          />
        );
      })
      .value();
  },

  renderGraphEdges: function(edges) {
    const selectedNodeId = this.props.selectedNodeId;
    const hasSelectedNode = selectedNodeId && this.props.nodes.has(selectedNodeId);

    return _.map(edges, function(edge) {
      const highlighted = _.includes(this.props.highlightedEdgeIds, edge.id);
      const blurred = hasSelectedNode
        && edge.source.id !== selectedNodeId
        && edge.target.id !== selectedNodeId;
      const points = [{
        x: edge.source.x,
        y: edge.source.y
      }, {
        x: edge.target.x,
        y: edge.target.y
      }];
      return (
        <Edge key={edge.id} id={edge.id} points={points} blurred={blurred}
          highlighted={highlighted} />
      );
    }, this);
  },

  render: function() {
    const nodeElements = this.renderGraphNodes(this.state.nodes, this.state.nodeScale);
    const edgeElements = this.renderGraphEdges(this.state.edges, this.state.nodeScale);
    let scale = this.state.scale;

    // only animate shift behavior, not panning
    const panTranslate = this.state.panTranslate;
    const shiftTranslate = this.state.shiftTranslate;
    let translate = panTranslate;
    let wasShifted = false;
    if (shiftTranslate[0] !== panTranslate[0] || shiftTranslate[1] !== panTranslate[1]) {
      translate = shiftTranslate;
      wasShifted = true;
    }
    const errorClassNames = this.state.maxNodesExceeded ? 'nodes-chart-error' : 'nodes-chart-error hide';
    const svgClassNames = this.state.maxNodesExceeded || _.size(nodeElements) === 0 ? 'hide' : '';

    return (
      <div className="nodes-chart">
        <div className={errorClassNames}>
          <span className="nodes-chart-error-icon fa fa-ban" />
          <div>Too many nodes to show in the browser.<br />We're working on it, but for now, try a different view?</div>
        </div>
        <svg width="100%" height="100%" className={svgClassNames} onMouseUp={this.handleMouseUp}>
          <Spring endValue={{val: translate, config: [80, 20]}}>
            {function(interpolated) {
              let interpolatedTranslate = wasShifted ? interpolated.val : panTranslate;
              const transform = 'translate(' + interpolatedTranslate + ')' +
                ' scale(' + scale + ')';
              return (
                <g className="canvas" transform={transform}>
                  <g className="edges">
                    {edgeElements}
                  </g>
                  <g className="nodes">
                    {nodeElements}
                  </g>
                </g>
              );
            }}
          </Spring>
        </svg>
      </div>
    );
  },

  initNodes: function(topology) {
    const centerX = this.props.width / 2;
    const centerY = this.props.height / 2;
    const nodes = {};

    topology.forEach(function(node, id) {
      nodes[id] = {};

      // use cached positions if available
      _.defaults(nodes[id], prevNodes[id], {
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

  centerSelectedNode: function(props, state) {
    const layoutNodes = state.nodes;
    const layoutEdges = state.edges;
    const selectedLayoutNode = layoutNodes[props.selectedNodeId];

    if (!selectedLayoutNode) {
      return {};
    }

    const adjacency = AppStore.getAdjacentNodes(props.selectedNodeId);
    const adjacentLayoutNodes = [];

    adjacency.forEach(function(adjacentId) {
      // filter loopback
      if (adjacentId !== props.selectedNodeId) {
        adjacentLayoutNodes.push(layoutNodes[adjacentId]);
      }
    });

    // shift center node a bit
    const nodeScale = state.nodeScale;
    selectedLayoutNode.x = selectedLayoutNode.px + nodeScale(1);
    selectedLayoutNode.y = selectedLayoutNode.py + nodeScale(1);

    // circle layout for adjacent nodes
    const centerX = selectedLayoutNode.x;
    const centerY = selectedLayoutNode.y;
    const adjacentCount = adjacentLayoutNodes.length;
    const density = radiusDensity(adjacentCount);
    const radius = Math.min(props.width, props.height) / density;
    const offsetAngle = Math.PI / 4;

    _.each(adjacentLayoutNodes, function(node, i) {
      const angle = offsetAngle + Math.PI * 2 * i / adjacentCount;
      node.x = centerX + radius * Math.sin(angle);
      node.y = centerY + radius * Math.cos(angle);
    });

    // fix all edges for circular nodes

    _.each(layoutEdges, function(edge) {
      if (edge.source === selectedLayoutNode
        || edge.target === selectedLayoutNode
        || _.includes(adjacentLayoutNodes, edge.source)
        || _.includes(adjacentLayoutNodes, edge.target)) {
        edge.points = [
          {x: edge.source.x, y: edge.source.y},
          {x: edge.target.x, y: edge.target.y}
        ];
      }
    });

    // shift canvas selected node out of view if it has not been shifted already
    let autoShifted = this.state.autoShifted;
    const shiftTranslate = state.shiftTranslate;

    if (!autoShifted) {
      const visibleWidth = Math.max(props.width - props.detailsWidth, 0);
      const offsetX = shiftTranslate[0];
      // normalize graph coordinates by zoomScale
      const zoomScale = state.scale;
      const outerRadius = radius + this.state.nodeScale(1.5);
      if (2 * outerRadius * zoomScale > props.width) {
        // radius too big, centering center node on canvas
        shiftTranslate[0] = -(centerX * zoomScale - (props.width + MARGINS.left) / 2);
      } else if (offsetX + (centerX + outerRadius) * zoomScale > visibleWidth) {
        // shift left if blocked by details
        const shift = (centerX + outerRadius) * zoomScale - visibleWidth;
        shiftTranslate[0] = -shift;
      } else if (offsetX + (centerX - outerRadius) * zoomScale < 0) {
        // shift right if off canvas
        const shift = offsetX - offsetX + (centerX - outerRadius) * zoomScale;
        shiftTranslate[0] = -shift;
      }
      const offsetY = shiftTranslate[1];
      if (2 * outerRadius * zoomScale > props.height) {
        // radius too big, centering center node on canvas
        shiftTranslate[1] = -(centerY * zoomScale - (props.height + MARGINS.top) / 2);
      } else if (offsetY + (centerY + outerRadius) * zoomScale > props.height) {
        // shift up if past bottom
        const shift = (centerY + outerRadius) * zoomScale - props.height;
        shiftTranslate[1] = -shift;
      } else if (offsetY + (centerY - outerRadius) * zoomScale - props.topMargin < 0) {
        // shift down if off canvas
        const shift = offsetY - offsetY + (centerY - outerRadius) * zoomScale - props.topMargin;
        shiftTranslate[1] = -shift;
      }
      // debug('shift', centerX, centerY, outerRadius, shiftTranslate);

      // saving translate in d3's panning cache
      this.zoom.translate(shiftTranslate);
      autoShifted = true;
    }

    return {
      autoShifted: autoShifted,
      edges: layoutEdges,
      nodes: layoutNodes,
      shiftTranslate: shiftTranslate
    };
  },

  isZooming: false, // distinguish pan/zoom from click

  handleMouseUp: function() {
    if (!this.isZooming) {
      AppActions.clickCloseDetails();
      // allow shifts again
      this.setState({
        autoShifted: false
      });
    } else {
      this.isZooming = false;
    }
  },

  restoreLayout: function(state) {
    const edges = state.edges;
    const nodes = state.nodes;

    _.each(nodes, function(node) {
      node.x = node.px;
      node.y = node.py;
    });

    _.each(edges, function(edge) {
      if (edge.ppoints) {
        edge.points = edge.ppoints;
      }
    });

    return {edges: edges, nodes: nodes};
  },

  updateGraphState: function(props, state) {
    const n = props.nodes.size;

    if (n === 0) {
      return {};
    }

    const nodes = this.initNodes(props.nodes, state.nodes);
    const edges = this.initEdges(props.nodes, nodes);
    const width = props.width - MARGINS.left - MARGINS.right;
    const height = props.height - MARGINS.top - MARGINS.bottom;
    const expanse = Math.min(props.height, props.width);
    const nodeSize = expanse / 3; // single node should fill a third of the screen
    const normalizedNodeSize = nodeSize / Math.sqrt(n); // assuming rectangular layout
    const nodeScale = this.state.nodeScale.range([0, normalizedNodeSize]);

    debug('fingerprint', md5(JSON.stringify(_.keys(nodes)) + JSON.stringify(_.keys(edges))));

    let graph = NodesLayout.doLayout(nodes, edges, width, height, nodeScale);
    if (this.state.initialLayout && graph.width > 0) {
      debug('running layout twice to reduce jitter on initial layout');
      graph = NodesLayout.doLayout(nodes, edges, width, height, nodeScale);
    }

    // layout was aborted
    if (!graph) {
      return {maxNodesExceeded: true};
    }

    // save coordinates for restore
    _.each(nodes, function(node) {
      node.px = node.x;
      node.py = node.y;
    });
    _.each(edges, function(edge) {
      edge.ppoints = edge.points;
    });

    // adjust layout based on viewport

    const empty = graph.width === 0;
    const xFactor = width / graph.width;
    const yFactor = height / graph.height;
    const xOffset = graph.left;
    const yOffset = graph.top;
    // only adjust zooming in
    const zoomFactor = Math.min(1, Math.min(xFactor, yFactor));
    let zoomScale = this.state.scale;
    let translate = this.state.translate;

    if (this.zoom && !this.state.hasZoomed) {
      if (zoomFactor > 0 && zoomFactor !== zoomScale) {
        zoomScale = zoomFactor;
        // saving in d3's behavior cache
        this.zoom.scale(zoomFactor);
      }

      if (xOffset < 0) {
        translate[0] = xOffset * -1 * zoomScale + MARGINS.left;
      }

      if (yOffset < 0) {
        translate[1] = yOffset * -1 * zoomScale + MARGINS.top;
      }

      // saving in d3's behavior cache
      this.zoom.translate(translate);
    }

    return {
      nodes: nodes,
      edges: edges,
      nodeScale: nodeScale,
      scale: zoomScale,
      maxNodesExceeded: false,
      translate: translate,
      initialLayout: empty
    };
  },

  zoomed: function() {
    // debug('zoomed', d3.event.scale, d3.event.translate);
    this.isZooming = true;
    this.setState({
      autoShifted: false,
      hasZoomed: true,
      panTranslate: d3.event.translate.slice(),
      shiftTranslate: d3.event.translate.slice(),
      scale: d3.event.scale
    });
  }

});

module.exports = NodesChart;

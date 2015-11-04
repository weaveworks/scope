const _ = require('lodash');
const d3 = require('d3');
const debug = require('debug')('scope:nodes-chart');
const React = require('react');
const makeMap = require('immutable').Map;
const timely = require('timely');
const Spring = require('react-motion').Spring;

const AppActions = require('../actions/app-actions');
const AppStore = require('../stores/app-store');
const Edge = require('./edge');
const Naming = require('../constants/naming');
const NodesLayout = require('./nodes-layout');
const Node = require('./node');
const NodesError = require('./nodes-error');

const MARGINS = {
  top: 130,
  left: 40,
  right: 40,
  bottom: 0
};

// make sure circular layouts a bit denser with 3-6 nodes
const radiusDensity = d3.scale.threshold()
  .domain([3, 6]).range([2.5, 3.5, 3]);

const NodesChart = React.createClass({

  getInitialState: function() {
    return {
      nodes: makeMap(),
      edges: makeMap(),
      nodeScale: d3.scale.linear(),
      shiftTranslate: [0, 0],
      panTranslate: [0, 0],
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
  },

  componentWillReceiveProps: function(nextProps) {
    // gather state, setState should be called only once here
    const state = _.assign({}, this.state);

    // wipe node states when showing different topology
    if (nextProps.topologyId !== this.props.topologyId) {
      _.assign(state, {
        autoShifted: false,
        nodes: makeMap(),
        edges: makeMap()
      });
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

    // highlighter functions
    const setHighlighted = node => {
      const highlighted = _.includes(this.props.highlightedNodeIds, node.get('id'))
        || this.props.selectedNodeId === node.get('id');
      return node.set('highlighted', highlighted);
    };
    const setFocused = node => {
      const focused = hasSelectedNode
        && (this.props.selectedNodeId === node.get('id') || adjacency.includes(node.get('id')));
      return node.set('focused', focused);
    };
    const setBlurred = node => {
      return node.set('blurred', hasSelectedNode && !node.get('focused'));
    };

    // make sure blurred nodes are in the background
    const sortNodes = node => {
      if (node.get('blurred')) {
        return 0;
      }
      if (node.get('highlighted')) {
        return 2;
      }
      return 1;
    };

    return nodes
      .toIndexedSeq()
      .map(setHighlighted)
      .map(setFocused)
      .map(setBlurred)
      .sortBy(sortNodes)
      .map(node => {
        return (<Node
            blurred={node.get('blurred')}
            focused={node.get('focused')}
            highlighted={node.get('highlighted')}
            onClick={onNodeClick}
            key={node.get('id')}
            id={node.get('id')}
            label={node.get('label')}
            pseudo={node.get('pseudo')}
            subLabel={node.get('subLabel')}
            rank={node.get('rank')}
            scale={scale}
            dx={node.get('x')}
            dy={node.get('y')}
          />
        );
      });
  },

  renderGraphEdges: function(edges) {
    const selectedNodeId = this.props.selectedNodeId;
    const hasSelectedNode = selectedNodeId && this.props.nodes.has(selectedNodeId);

    const setHighlighted = edge => {
      return edge.set('highlighted', _.includes(this.props.highlightedEdgeIds, edge.get('id')));
    };
    const setBlurred = edge => {
      return (edge.set('blurred', hasSelectedNode
        && edge.get('source') !== selectedNodeId
        && edge.get('target') !== selectedNodeId));
    };

    return edges
      .toIndexedSeq()
      .map(setHighlighted)
      .map(setBlurred)
      .map(edge => {
        return (
          <Edge key={edge.get('id')} id={edge.get('id')} points={edge.get('points')}
            blurred={edge.get('blurred')} highlighted={edge.get('highlighted')} />
        );
      });
  },

  renderMaxNodesError: function(show) {
    const errorHint = 'We\u0027re working on it, but for now, try a different view?';
    return (
      <NodesError faIconClass="fa-ban" hidden={!show}>
        <div className="centered">Too many nodes to show in the browser.<br />{errorHint}</div>
      </NodesError>
    );
  },

  renderEmptyTopologyError: function(show) {
    return (
      <NodesError faIconClass="fa-circle-thin" hidden={!show}>
        <div className="heading">Nothing to show. This can have any of these reasons:</div>
        <ul>
          <li>We haven't received any reports from probes recently. Are the probes properly configured?</li>
          <li>There are nodes, but they're currently hidden. Check the view options in the bottom-left if they allow for showing hidden nodes.</li>
          <li>Containers view only: you're not running Docker, or you don't have any containers.</li>
        </ul>
      </NodesError>
    );
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
    const svgClassNames = this.state.maxNodesExceeded || nodeElements.size === 0 ? 'hide' : '';
    const errorEmpty = this.renderEmptyTopologyError(AppStore.isTopologyEmpty());
    const errorMaxNodesExceeded = this.renderMaxNodesError(this.state.maxNodesExceeded);

    return (
      <div className="nodes-chart">
        {errorEmpty}
        {errorMaxNodesExceeded}
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
    return topology.map((node, id) => {
      // copy relevant fields to state nodes
      return makeMap({
        id: id,
        label: node.get('label_major'),
        pseudo: node.get('pseudo'),
        subLabel: node.get('label_minor'),
        rank: node.get('rank'),
        x: 0,
        y: 0
      });
    });
  },

  initEdges: function(topology, stateNodes) {
    let edges = makeMap();

    topology.forEach(function(node, nodeId) {
      const adjacency = node.get('adjacency');
      if (adjacency) {
        adjacency.forEach(function(adjacent) {
          const edge = [nodeId, adjacent];
          const edgeId = edge.join(Naming.EDGE_ID_SEPARATOR);

          if (!edges.has(edgeId)) {
            const source = edge[0];
            const target = edge[1];

            if (!stateNodes.has(source) || !stateNodes.has(target)) {
              debug('Missing edge node', edge[0], edge[1]);
            }

            edges = edges.set(edgeId, makeMap({
              id: edgeId,
              value: 1,
              source: source,
              target: target
            }));
          }
        });
      }
    });

    return edges;
  },

  centerSelectedNode: function(props, state) {
    let stateNodes = state.nodes;
    let stateEdges = state.edges;
    let selectedLayoutNode = stateNodes.get(props.selectedNodeId);

    if (!selectedLayoutNode) {
      return {};
    }

    const adjacency = AppStore.getAdjacentNodes(props.selectedNodeId);
    let adjacentLayoutNodeIds = [];

    adjacency.forEach(function(adjacentId) {
      // filter loopback
      if (adjacentId !== props.selectedNodeId) {
        adjacentLayoutNodeIds.push(adjacentId);
      }
    });

    // shift center node a bit
    const nodeScale = state.nodeScale;
    const centerX = selectedLayoutNode.get('px') + nodeScale(1);
    const centerY = selectedLayoutNode.get('py') + nodeScale(1);
    stateNodes = stateNodes.mergeIn([props.selectedNodeId], {
      x: centerX,
      y: centerY
    });

    // circle layout for adjacent nodes
    const adjacentCount = adjacentLayoutNodeIds.length;
    const density = radiusDensity(adjacentCount);
    const radius = Math.min(props.width, props.height) / density;
    const offsetAngle = Math.PI / 4;

    stateNodes = stateNodes.map((node) => {
      const index = adjacentLayoutNodeIds.indexOf(node.get('id'));
      if (index > -1) {
        const angle = offsetAngle + Math.PI * 2 * index / adjacentCount;
        return node.merge({
          x: centerX + radius * Math.sin(angle),
          y: centerY + radius * Math.cos(angle)
        });
      }
      return node;
    });

    // fix all edges for circular nodes
    stateEdges = stateEdges.map(edge => {
      if (edge.get('source') === selectedLayoutNode.get('id')
        || edge.get('target') === selectedLayoutNode.get('id')
        || _.includes(adjacentLayoutNodeIds, edge.get('source'))
        || _.includes(adjacentLayoutNodeIds, edge.get('target'))) {
        const source = stateNodes.get(edge.get('source'));
        const target = stateNodes.get(edge.get('target'));
        return edge.set('points', [
          {x: source.get('x'), y: source.get('y')},
          {x: target.get('x'), y: target.get('y')}
        ]);
      }
      return edge;
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
      edges: stateEdges,
      nodes: stateNodes,
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
    const nodes = state.nodes.map(node => {
      return node.merge({
        x: node.get('px'),
        y: node.get('py')
      });
    });

    const edges = state.edges.map(edge => {
      if (edge.has('ppoints')) {
        return edge.set('points', edge.get('ppoints'));
      }
      return edge;
    });

    return {edges, nodes};
  },

  updateGraphState: function(props, state) {
    const n = props.nodes.size;

    if (n === 0) {
      return {
        nodes: makeMap(),
        edges: makeMap()
      };
    }

    let stateNodes = this.initNodes(props.nodes, state.nodes);
    let stateEdges = this.initEdges(props.nodes, stateNodes);

    const expanse = Math.min(props.height, props.width);
    const nodeSize = expanse / 3; // single node should fill a third of the screen
    const normalizedNodeSize = nodeSize / Math.sqrt(n); // assuming rectangular layout
    const nodeScale = this.state.nodeScale.range([0, normalizedNodeSize]);
    const options = {
      width: props.width,
      height: props.height,
      scale: nodeScale,
      margins: MARGINS,
      topologyId: this.props.topologyId
    };

    const timedLayouter = timely(NodesLayout.doLayout);
    const graph = timedLayouter(stateNodes, stateEdges, options);

    debug('graph layout took ' + timedLayouter.time + 'ms');

    // layout was aborted
    if (!graph) {
      return {maxNodesExceeded: true};
    }
    stateNodes = graph.nodes;
    stateEdges = graph.edges;

    // save coordinates for restore
    stateNodes = stateNodes.map(node => {
      return node.merge({
        px: node.get('x'),
        py: node.get('y')
      });
    });
    stateEdges = stateEdges.map(edge => {
      return edge.set('ppoints', edge.get('points'));
    });

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

    return {
      nodes: stateNodes,
      edges: stateEdges,
      nodeScale: nodeScale,
      scale: zoomScale,
      maxNodesExceeded: false
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

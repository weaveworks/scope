import _ from 'lodash';
import d3 from 'd3';
import debug from 'debug';
import React from 'react';
import { Map as makeMap } from 'immutable';
import timely from 'timely';

import { DETAILS_PANEL_WIDTH } from '../constants/styles';
import { clickBackground } from '../actions/app-actions';
import AppStore from '../stores/app-store';
import Edge from './edge';
import { EDGE_ID_SEPARATOR } from '../constants/naming';
import { doLayout } from './nodes-layout';
import Node from './node';
import NodesError from './nodes-error';
import Logo from '../components/logo';

const log = debug('scope:nodes-chart');

const MARGINS = {
  top: 130,
  left: 40,
  right: 40,
  bottom: 0
};

// make sure circular layouts a bit denser with 3-6 nodes
const radiusDensity = d3.scale.threshold()
  .domain([3, 6]).range([2.5, 3.5, 3]);

export default class NodesChart extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.handleMouseClick = this.handleMouseClick.bind(this);
    this.zoomed = this.zoomed.bind(this);

    this.state = {
      nodes: makeMap(),
      edges: makeMap(),
      panTranslate: [0, 0],
      scale: 1,
      nodeScale: d3.scale.linear(),
      selectedNodeScale: d3.scale.linear(),
      hasZoomed: false,
      maxNodesExceeded: false
    };
  }

  componentWillMount() {
    const state = this.updateGraphState(this.props, this.state);
    this.setState(state);
  }

  componentDidMount() {
    // distinguish pan/zoom from click
    this.isZooming = false;

    this.zoom = d3.behavior.zoom()
      .scaleExtent([0.1, 2])
      .on('zoom', this.zoomed);

    d3.select('.nodes-chart svg')
      .call(this.zoom);
  }

  componentWillReceiveProps(nextProps) {
    // gather state, setState should be called only once here
    const state = _.assign({}, this.state);

    // wipe node states when showing different topology
    if (nextProps.topologyId !== this.props.topologyId) {
      _.assign(state, {
        nodes: makeMap(),
        edges: makeMap()
      });
    }
    // FIXME add PureRenderMixin, Immutables, and move the following functions to render()
    if (nextProps.forceRelayout || nextProps.nodes !== this.props.nodes) {
      _.assign(state, this.updateGraphState(nextProps, state));
    }
    if (this.props.selectedNodeId !== nextProps.selectedNodeId) {
      _.assign(state, this.restoreLayout(state));
    }
    if (nextProps.selectedNodeId) {
      _.assign(state, this.centerSelectedNode(nextProps, state));
    }

    this.setState(state);
  }

  componentWillUnmount() {
    // undoing .call(zoom)

    d3.select('.nodes-chart svg')
      .on('mousedown.zoom', null)
      .on('onwheel', null)
      .on('onmousewheel', null)
      .on('dblclick.zoom', null)
      .on('touchstart.zoom', null);
  }

  renderGraphNodes(nodes, nodeScale) {
    const hasSelectedNode = this.props.selectedNodeId
      && this.props.nodes.has(this.props.selectedNodeId);
    const adjacency = hasSelectedNode ? AppStore.getAdjacentNodes(this.props.selectedNodeId) : null;
    const onNodeClick = this.props.onNodeClick;
    const zoomScale = this.state.scale;
    const selectedNodeScale = this.state.selectedNodeScale;

    // highlighter functions
    const setHighlighted = node => {
      const highlighted = this.props.highlightedNodeIds.has(node.get('id'))
        || this.props.selectedNodeId === node.get('id');
      return node.set('highlighted', highlighted);
    };
    const setFocused = node => {
      const focused = hasSelectedNode
        && (this.props.selectedNodeId === node.get('id') || adjacency.includes(node.get('id')));
      return node.set('focused', focused);
    };
    const setBlurred = node => node.set('blurred', hasSelectedNode && !node.get('focused'));

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
      .map(node => <Node
        blurred={node.get('blurred')}
        focused={node.get('focused')}
        highlighted={node.get('highlighted')}
        topologyId={this.props.topologyId}
        shape={node.get('shape')}
        stack={node.get('stack')}
        onClick={onNodeClick}
        key={node.get('id')}
        id={node.get('id')}
        label={node.get('label')}
        pseudo={node.get('pseudo')}
        nodeCount={node.get('nodeCount')}
        subLabel={node.get('subLabel')}
        rank={node.get('rank')}
        selectedNodeScale={selectedNodeScale}
        nodeScale={nodeScale}
        zoomScale={zoomScale}
        dx={node.get('x')}
        dy={node.get('y')}
      />);
  }

  renderGraphEdges(edges) {
    const selectedNodeId = this.props.selectedNodeId;
    const hasSelectedNode = selectedNodeId && this.props.nodes.has(selectedNodeId);

    const setHighlighted = edge => edge.set('highlighted', this.props.highlightedEdgeIds.has(
      edge.get('id')));

    const setBlurred = edge => edge.set('blurred', hasSelectedNode
      && edge.get('source') !== selectedNodeId
      && edge.get('target') !== selectedNodeId);

    return edges
      .toIndexedSeq()
      .map(setHighlighted)
      .map(setBlurred)
      .map(edge => <Edge key={edge.get('id')} id={edge.get('id')}
        points={edge.get('points')}
        blurred={edge.get('blurred')} highlighted={edge.get('highlighted')}
      />
    );
  }

  renderMaxNodesError(show) {
    const errorHint = 'We\u0027re working on it, but for now, try a different view?';
    return (
      <NodesError faIconClass="fa-ban" hidden={!show}>
        <div className="centered">Too many nodes to show in the browser.<br />{errorHint}</div>
      </NodesError>
    );
  }

  renderEmptyTopologyError(show) {
    return (
      <NodesError faIconClass="fa-circle-thin" hidden={!show}>
        <div className="heading">Nothing to show. This can have any of these reasons:</div>
        <ul>
          <li>We haven't received any reports from probes recently.
           Are the probes properly configured?</li>
          <li>There are nodes, but they're currently hidden. Check the view options
           in the bottom-left if they allow for showing hidden nodes.</li>
          <li>Containers view only: you're not running Docker,
           or you don't have any containers.</li>
        </ul>
      </NodesError>
    );
  }

  render() {
    const nodeElements = this.renderGraphNodes(this.state.nodes, this.state.nodeScale);
    const edgeElements = this.renderGraphEdges(this.state.edges, this.state.nodeScale);
    const scale = this.state.scale;

    const translate = this.state.panTranslate;
    const transform = `translate(${translate}) scale(${scale})`;
    const svgClassNames = this.state.maxNodesExceeded || nodeElements.size === 0 ? 'hide' : '';
    const errorEmpty = this.renderEmptyTopologyError(AppStore.isTopologyEmpty());
    const errorMaxNodesExceeded = this.renderMaxNodesError(this.state.maxNodesExceeded);

    return (
      <div className="nodes-chart">
        {errorEmpty}
        {errorMaxNodesExceeded}
        <svg width="100%" height="100%" id="nodes-chart-canvas"
          className={svgClassNames} onClick={this.handleMouseClick}>
          <g transform="translate(24,24) scale(0.25)">
            <Logo />
          </g>
          <g className="canvas" transform={transform}>
            <g className="edges">
              {edgeElements}
            </g>
            <g className="nodes">
              {nodeElements}
            </g>
          </g>
        </svg>
      </div>
    );
  }

  initNodes(topology) {
    // copy relevant fields to state nodes
    return topology.map((node, id) => makeMap({
      id,
      label: node.get('label'),
      pseudo: node.get('pseudo'),
      subLabel: node.get('label_minor'),
      nodeCount: node.get('node_count'),
      rank: node.get('rank'),
      shape: node.get('shape'),
      stack: node.get('stack'),
      x: 0,
      y: 0
    }));
  }

  initEdges(topology, stateNodes) {
    let edges = makeMap();

    topology.forEach((node, nodeId) => {
      const adjacency = node.get('adjacency');
      if (adjacency) {
        adjacency.forEach(adjacent => {
          const edge = [nodeId, adjacent];
          const edgeId = edge.join(EDGE_ID_SEPARATOR);

          if (!edges.has(edgeId)) {
            const source = edge[0];
            const target = edge[1];

            if (!stateNodes.has(source) || !stateNodes.has(target)) {
              log('Missing edge node', edge[0], edge[1]);
            }

            edges = edges.set(edgeId, makeMap({
              id: edgeId,
              value: 1,
              source,
              target
            }));
          }
        });
      }
    });

    return edges;
  }

  centerSelectedNode(props, state) {
    let stateNodes = state.nodes;
    let stateEdges = state.edges;
    const selectedLayoutNode = stateNodes.get(props.selectedNodeId);

    if (!selectedLayoutNode) {
      return {};
    }

    const adjacency = AppStore.getAdjacentNodes(props.selectedNodeId);
    const adjacentLayoutNodeIds = [];

    adjacency.forEach(adjacentId => {
      // filter loopback
      if (adjacentId !== props.selectedNodeId) {
        adjacentLayoutNodeIds.push(adjacentId);
      }
    });

    // move origin node to center of viewport
    const zoomScale = state.scale;
    const translate = state.panTranslate;
    const centerX = (-translate[0] + (props.width + MARGINS.left
      - DETAILS_PANEL_WIDTH) / 2) / zoomScale;
    const centerY = (-translate[1] + (props.height + MARGINS.top) / 2) / zoomScale;
    stateNodes = stateNodes.mergeIn([props.selectedNodeId], {
      x: centerX,
      y: centerY
    });

    // circle layout for adjacent nodes
    const adjacentCount = adjacentLayoutNodeIds.length;
    const density = radiusDensity(adjacentCount);
    const radius = Math.min(props.width, props.height) / density / zoomScale;
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

    // auto-scale node size for selected nodes
    const selectedNodeScale = this.getNodeScale(props);

    return {
      selectedNodeScale,
      edges: stateEdges,
      nodes: stateNodes
    };
  }

  handleMouseClick() {
    if (!this.isZooming) {
      clickBackground();
    } else {
      this.isZooming = false;
    }
  }

  restoreLayout(state) {
    // undo any pan/zooming that might have happened
    this.zoom.scale(state.scale);
    this.zoom.translate(state.panTranslate);

    const nodes = state.nodes.map(node => node.merge({
      x: node.get('px'),
      y: node.get('py')
    }));

    const edges = state.edges.map(edge => {
      if (edge.has('ppoints')) {
        return edge.set('points', edge.get('ppoints'));
      }
      return edge;
    });

    return { edges, nodes};
  }

  updateGraphState(props, state) {
    const n = props.nodes.size;

    if (n === 0) {
      return {
        nodes: makeMap(),
        edges: makeMap()
      };
    }

    let stateNodes = this.initNodes(props.nodes, state.nodes);
    let stateEdges = this.initEdges(props.nodes, stateNodes);
    const nodeScale = this.getNodeScale(props);

    const options = {
      width: props.width,
      height: props.height,
      scale: nodeScale,
      margins: MARGINS,
      forceRelayout: props.forceRelayout,
      topologyId: this.props.topologyId
    };

    const timedLayouter = timely(doLayout);
    const graph = timedLayouter(stateNodes, stateEdges, options);

    log(`graph layout took ${timedLayouter.time}ms`);

    // layout was aborted
    if (!graph) {
      return {maxNodesExceeded: true};
    }
    stateNodes = graph.nodes;
    stateEdges = graph.edges;

    // save coordinates for restore
    stateNodes = stateNodes.map(node => node.merge({
      px: node.get('x'),
      py: node.get('y')
    }));
    stateEdges = stateEdges.map(edge => edge.set('ppoints', edge.get('points')));

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
      scale: zoomScale,
      nodeScale,
      maxNodesExceeded: false
    };
  }

  getNodeScale(props) {
    const expanse = Math.min(props.height, props.width);
    const nodeSize = expanse / 3; // single node should fill a third of the screen
    const maxNodeSize = expanse / 10;
    const normalizedNodeSize = Math.min(nodeSize / Math.sqrt(props.nodes.size), maxNodeSize);
    return this.state.nodeScale.copy().range([0, normalizedNodeSize]);
  }

  zoomed() {
    // debug('zoomed', d3.event.scale, d3.event.translate);
    this.isZooming = true;
    // dont pan while node is selected
    if (!this.props.selectedNodeId) {
      this.setState({
        hasZoomed: true,
        panTranslate: d3.event.translate.slice(),
        scale: d3.event.scale
      });
    }
  }
}

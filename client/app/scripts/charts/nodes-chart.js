import _ from 'lodash';
import d3 from 'd3';
import debug from 'debug';
import React from 'react';
import { connect } from 'react-redux';
import { Map as makeMap, fromJS, is as isDeepEqual } from 'immutable';
import timely from 'timely';
import { Set as makeSet } from 'immutable';

import { clickBackground } from '../actions/app-actions';
import { EDGE_ID_SEPARATOR } from '../constants/naming';
import { DETAILS_PANEL_WIDTH } from '../constants/styles';
import Logo from '../components/logo';
import { doLayout } from './nodes-layout';
import NodesChartElements from './nodes-chart-elements';
import { getActiveTopologyOptions, getAdjacentNodes,
  isSameTopology } from '../utils/topology-utils';

const log = debug('scope:nodes-chart');

const ZOOM_CACHE_FIELDS = ['scale', 'panTranslateX', 'panTranslateY'];

// make sure circular layouts a bit denser with 3-6 nodes
const radiusDensity = d3.scale.threshold()
  .domain([3, 6])
  .range([2.5, 3.5, 3]);

class NodesChart extends React.Component {

  constructor(props, context) {
    super(props, context);

    this.handleMouseClick = this.handleMouseClick.bind(this);
    this.zoomed = this.zoomed.bind(this);

    this.state = {
      edges: makeMap(),
      nodes: makeMap(),
      nodeScale: d3.scale.linear(),
      panTranslateX: 0,
      panTranslateY: 0,
      scale: 1,
      selectedNodeScale: d3.scale.linear(),
      hasZoomed: false,
      height: props.height || 0,
      width: props.width || 0,
      zoomCache: {}
    };
  }

  componentWillMount() {
    const state = this.updateGraphState(this.props, this.state);
    this.setState(state);
  }

  componentWillReceiveProps(nextProps) {
    // gather state, setState should be called only once here
    const state = _.assign({}, this.state);

    // wipe node states when showing different topology
    if (nextProps.topologyId !== this.props.topologyId) {
      // re-apply cached canvas zoom/pan to d3 behavior (or set defaul values)
      const defaultZoom = { scale: 1, panTranslateX: 0, panTranslateY: 0, hasZoomed: false };
      const nextZoom = this.state.zoomCache[nextProps.topologyId] || defaultZoom;
      if (nextZoom && this.zoom) {
        this.zoom.scale(nextZoom.scale);
        this.zoom.translate([nextZoom.panTranslateX, nextZoom.panTranslateY]);
      }

      // saving previous zoom state
      const prevZoom = _.pick(this.state, ZOOM_CACHE_FIELDS);
      const zoomCache = _.assign({}, this.state.zoomCache);
      zoomCache[this.props.topologyId] = prevZoom;

      // clear canvas and apply zoom state
      _.assign(state, nextZoom, { zoomCache }, {
        nodes: makeMap(),
        edges: makeMap()
      });
    }

    // reset layout dimensions only when forced
    state.height = nextProps.height;
    state.width = nextProps.width;

    _.assign(state, this.updateGraphState(nextProps, state));
    // if (nextProps.forceRelayout || nextProps.nodes !== this.props.nodes) {
      // _.assign(state, this.updateGraphState(nextProps, state));
    // }

    if (this.props.selectedNodeId !== nextProps.selectedNodeId) {
      _.assign(state, this.restoreLayout(state));
    }
    if (nextProps.selectedNodeId) {
      _.assign(state, this.centerSelectedNode(nextProps, state));
    }

    this.setState(state);
  }

  componentDidMount() {
    // distinguish pan/zoom from click
    if (this.props.noZoom) {
      return;
    }

    this.isZooming = false;

    this.zoom = d3.behavior.zoom()
      .scaleExtent([0.1, 2])
      .on('zoom', this.zoomed);

    d3.select('.nodes-chart svg')
      .call(this.zoom);
  }

  componentWillUnmount() {
    if (this.props.noZoom) {
      return;
    }

    // undoing .call(zoom)
    d3.select('.nodes-chart svg')
      .on('mousedown.zoom', null)
      .on('onwheel', null)
      .on('onmousewheel', null)
      .on('dblclick.zoom', null)
      .on('touchstart.zoom', null);
  }

  render() {
    const { edges, nodes, panTranslateX, panTranslateY, scale } = this.state;

    // not passing translates into child components for perf reasons, use getTranslate instead
    const translate = [panTranslateX, panTranslateY];
    const transform = `translate(${translate}) scale(${scale})`;
    const svgClassNames = this.props.isEmpty ? 'hide' : '';

    return (
      <div className="nodes-chart">
        <svg width="100%" height="100%" id="nodes-chart-canvas"
          className={svgClassNames} onClick={this.handleMouseClick}>
          <g transform="translate(24,24) scale(0.25)">
            <Logo />
          </g>
          <NodesChartElements layoutNodes={nodes} layoutEdges={edges}
            nodeScale={this.state.nodeScale} scale={scale} transform={transform}
            selectedNodeScale={this.state.selectedNodeScale}
            layoutPrecision={this.props.layoutPrecision} />
        </svg>
      </div>
    );
  }

  handleMouseClick() {
    if (!this.isZooming) {
      this.props.clickBackground();
    } else {
      this.isZooming = false;
    }
  }

  initNodes(topology, stateNodes) {
    let nextStateNodes = stateNodes;

    // remove nodes that have disappeared
    stateNodes.forEach((node, id) => {
      if (!topology.has(id)) {
        nextStateNodes = nextStateNodes.delete(id);
      }
    });

    // copy relevant fields to state nodes
    topology.forEach((node, id) => {
      nextStateNodes = nextStateNodes.mergeIn([id], makeMap({
        id,
        label: node.get('label'),
        pseudo: node.get('pseudo'),
        subLabel: node.get('label_minor'),
        nodeCount: node.get('node_count'),
        metrics: node.get('metrics'),
        rank: node.get('rank'),
        shape: node.get('shape'),
        stack: node.get('stack'),
        networks: node.get('networks'),
      }));
    });

    return nextStateNodes;
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
            if (stateNodes.has(source) && stateNodes.has(target)) {
              edges = edges.set(edgeId, makeMap({
                id: edgeId,
                value: 1,
                source,
                target
              }));
            }
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

    const adjacentNodes = props.adjacentNodes;
    const adjacentLayoutNodeIds = [];

    adjacentNodes.forEach(adjacentId => {
      // filter loopback
      if (adjacentId !== props.selectedNodeId) {
        adjacentLayoutNodeIds.push(adjacentId);
      }
    });

    // move origin node to center of viewport
    const zoomScale = state.scale;
    const translate = [state.panTranslateX, state.panTranslateY];
    const centerX = (-translate[0] + (state.width + props.margins.left
      - DETAILS_PANEL_WIDTH) / 2) / zoomScale;
    const centerY = (-translate[1] + (state.height + props.margins.top) / 2) / zoomScale;
    stateNodes = stateNodes.mergeIn([props.selectedNodeId], {
      x: centerX,
      y: centerY
    });

    // circle layout for adjacent nodes
    const adjacentCount = adjacentLayoutNodeIds.length;
    const density = radiusDensity(adjacentCount);
    const radius = Math.min(state.width, state.height) / density / zoomScale;
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
        return edge.set('points', fromJS([
          {x: source.get('x'), y: source.get('y')},
          {x: target.get('x'), y: target.get('y')}
        ]));
      }
      return edge;
    });

    // auto-scale node size for selected nodes
    const selectedNodeScale = this.getNodeScale(adjacentNodes, state.width, state.height);

    return {
      selectedNodeScale,
      edges: stateEdges,
      nodes: stateNodes
    };
  }

  restoreLayout(state) {
    // undo any pan/zooming that might have happened
    if (this.zoom) {
      this.zoom.scale(state.scale);
      this.zoom.translate([state.panTranslateX, state.panTranslateY]);
    }

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

    return { edges, nodes };
  }

  updateGraphState(props, state) {
    if (props.nodes.size === 0) {
      return {
        nodes: makeMap(),
        edges: makeMap()
      };
    }

    const stateNodes = this.initNodes(props.nodes, state.nodes);
    const stateEdges = this.initEdges(props.nodes, stateNodes);
    const nodeScale = this.getNodeScale(props.nodes, state.width, state.height);
    const nextState = { nodeScale };
    console.log(props.nodeOrder);
    const nodeOrder = props.nodeOrder || makeMap(stateNodes
      .toList()
      .sortBy(n => n.get('label'))
      .map((n, i) => [n.get('id'), i]));

    const options = {
      width: state.width,
      height: state.height,
      scale: nodeScale,
      margins: props.margins,
      forceRelayout: props.forceRelayout,
      topologyId: this.props.topologyId,
      topologyOptions: this.props.topologyOptions,
      nodeOrder
    };

    console.log('nodes-chart', state.height);
    const timedLayouter = timely(doLayout);
    const graph = timedLayouter(stateNodes, stateEdges, options);

    log(`graph layout took ${timedLayouter.time}ms`);

    // extract coords and save for restore
    const graphNodes = graph.nodes.map(node => makeMap({
      x: node.get('x'),
      px: node.get('x'),
      y: node.get('y'),
      py: node.get('y')
    }));

    const layoutNodes = stateNodes.mergeDeep(graphNodes);

    const layoutEdges = graph.edges
      .map(edge => edge.set('ppoints', edge.get('points')));

    // adjust layout based on viewport
    const xFactor = (state.width - props.margins.left - props.margins.right) / graph.width;
    const yFactor = state.height / graph.height;
    const zoomFactor = Math.min(xFactor, yFactor);
    let zoomScale = this.state.scale;

    if (!this.props.noZoom && !state.hasZoomed && zoomFactor > 0 && zoomFactor < 1) {
      zoomScale = zoomFactor;
      // saving in d3's behavior cache
      if (this.zoom) {
        this.zoom.scale(zoomFactor);
      }
    }

    nextState.scale = zoomScale;
    if (!isDeepEqual(layoutNodes, state.nodes)) {
      nextState.nodes = layoutNodes;
    }
    if (!isDeepEqual(layoutEdges, state.edges)) {
      nextState.edges = layoutEdges;
    }

    return nextState;
  }

  getNodeScale(nodes, width, height) {
    const expanse = Math.min(height, width);
    const nodeSize = expanse / 3; // single node should fill a third of the screen
    const maxNodeSize = expanse / 10;
    const normalizedNodeSize = Math.min(nodeSize / Math.sqrt(nodes.size), maxNodeSize);
    return this.state.nodeScale.copy().range([0, this.props.nodeSize || normalizedNodeSize]);
  }

  zoomed() {
    // debug('zoomed', d3.event.scale, d3.event.translate);
    this.isZooming = true;
    // dont pan while node is selected
    if (!this.props.selectedNodeId) {
      this.setState({
        hasZoomed: true,
        panTranslateX: d3.event.translate[0],
        panTranslateY: d3.event.translate[1],
        scale: d3.event.scale
      });
    }
  }
}

function mapStateToProps(state) {
  return {
    adjacentNodes: getAdjacentNodes(state),
    forceRelayout: state.get('forceRelayout'),
    nodes: state.get('nodes').filter(node => !node.get('filtered')),
    selectedNodeId: state.get('selectedNodeId'),
    topologyId: state.get('currentTopologyId'),
    topologyOptions: getActiveTopologyOptions(state)
  };
}

export default connect(
  mapStateToProps,
  { clickBackground }
)(NodesChart);

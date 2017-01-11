import debug from 'debug';
import React from 'react';
import { connect } from 'react-redux';
import { assign, pick, includes } from 'lodash';
import { Map as makeMap, fromJS } from 'immutable';
import timely from 'timely';

import { scaleThreshold } from 'd3-scale';
import { event as d3Event, select } from 'd3-selection';
import { zoom, zoomIdentity } from 'd3-zoom';

import { nodeAdjacenciesSelector, adjacentNodesSelector } from '../selectors/chartSelectors';
import { clickBackground } from '../actions/app-actions';
import { EDGE_ID_SEPARATOR } from '../constants/naming';
import { DETAILS_PANEL_WIDTH, NODE_BASE_SIZE } from '../constants/styles';
import Logo from '../components/logo';
import { doLayout } from './nodes-layout';
import NodesChartElements from './nodes-chart-elements';
import { getActiveTopologyOptions } from '../utils/topology-utils';

const log = debug('scope:nodes-chart');

const ZOOM_CACHE_FIELDS = ['scale', 'panTranslateX', 'panTranslateY', 'minScale', 'maxScale'];

// make sure circular layouts a bit denser with 3-6 nodes
const radiusDensity = scaleThreshold()
  .domain([3, 6])
  .range([2.5, 3.5, 3]);

const emptyLayoutState = {
  nodes: makeMap(),
  edges: makeMap(),
};


// EDGES
function initEdges(nodes) {
  let edges = makeMap();

  nodes.forEach((node, nodeId) => {
    const adjacency = node.get('adjacency');
    if (adjacency) {
      adjacency.forEach((adjacent) => {
        const edge = [nodeId, adjacent];
        const edgeId = edge.join(EDGE_ID_SEPARATOR);

        if (!edges.has(edgeId)) {
          const source = edge[0];
          const target = edge[1];
          if (nodes.has(source) && nodes.has(target)) {
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


// ZOOM STATE
function getLayoutDefaultZoom(layoutNodes, width, height) {
  const xMin = layoutNodes.minBy(n => n.get('x')).get('x');
  const xMax = layoutNodes.maxBy(n => n.get('x')).get('x');
  const yMin = layoutNodes.minBy(n => n.get('y')).get('y');
  const yMax = layoutNodes.maxBy(n => n.get('y')).get('y');

  const xFactor = width / (xMax - xMin);
  const yFactor = height / (yMax - yMin);
  const scale = Math.min(xFactor, yFactor);

  return {
    translateX: (width - ((xMax + xMin) * scale)) / 2,
    translateY: (height - ((yMax + yMin) * scale)) / 2,
    scale,
  };
}

function defaultZoomState(props, state) {
  // adjust layout based on viewport
  const width = state.width - props.margins.left - props.margins.right;
  const height = state.height - props.margins.top - props.margins.bottom;

  const { translateX, translateY, scale } = getLayoutDefaultZoom(state.nodes, width, height);

  return {
    scale,
    minScale: scale / 5,
    maxScale: Math.min(width, height) / NODE_BASE_SIZE / 3,
    panTranslateX: translateX + props.margins.left,
    panTranslateY: translateY + props.margins.top,
  };
}


// LAYOUT STATE
function updateLayout(width, height, nodes, baseOptions) {
  const edges = initEdges(nodes);
  const options = Object.assign({}, baseOptions);

  const timedLayouter = timely(doLayout);
  const graph = timedLayouter(nodes, edges, options);

  log(`graph layout took ${timedLayouter.time}ms`);

  const layoutNodes = graph.nodes.map(node => makeMap({
    x: node.get('x'),
    y: node.get('y'),
    // extract coords and save for restore
    px: node.get('x'),
    py: node.get('y')
  }));

  const layoutEdges = graph.edges.map(edge => edge.set('ppoints', edge.get('points')));

  return { layoutNodes, layoutEdges };
}

function updatedGraphState(props, state) {
  if (props.nodes.size === 0) {
    return emptyLayoutState;
  }

  const options = {
    width: state.width,
    height: state.height,
    margins: props.margins,
    forceRelayout: props.forceRelayout,
    topologyId: props.topologyId,
    topologyOptions: props.topologyOptions,
  };

  const { layoutNodes, layoutEdges } =
    updateLayout(state.width, state.height, props.nodes, options);

  return {
    nodes: layoutNodes,
    edges: layoutEdges,
  };
}

function restoredLayout(state) {
  const restoredNodes = state.nodes.map(node => node.merge({
    x: node.get('px'),
    y: node.get('py')
  }));

  const restoredEdges = state.edges.map(edge => (
    edge.has('ppoints') ? edge.set('points', edge.get('ppoints')) : edge
  ));

  return {
    nodes: restoredNodes,
    edges: restoredEdges,
  };
}


// SELECTED NODE
function centerSelectedNode(props, state) {
  let stateNodes = state.nodes;
  let stateEdges = state.edges;
  if (!stateNodes.has(props.selectedNodeId)) {
    return {};
  }

  const adjacentNodes = props.adjacentNodes;
  const adjacentLayoutNodeIds = [];

  adjacentNodes.forEach((adjacentId) => {
    // filter loopback
    if (adjacentId !== props.selectedNodeId) {
      adjacentLayoutNodeIds.push(adjacentId);
    }
  });

  // move origin node to center of viewport
  const zoomScale = state.scale;
  const translate = [state.panTranslateX, state.panTranslateY];
  const viewportHalfWidth = ((state.width + props.margins.left) - DETAILS_PANEL_WIDTH) / 2;
  const viewportHalfHeight = (state.height + props.margins.top) / 2;
  const centerX = (-translate[0] + viewportHalfWidth) / zoomScale;
  const centerY = (-translate[1] + viewportHalfHeight) / zoomScale;
  stateNodes = stateNodes.mergeIn([props.selectedNodeId], {
    x: centerX,
    y: centerY
  });

  // circle layout for adjacent nodes
  const adjacentCount = adjacentLayoutNodeIds.length;
  const density = radiusDensity(adjacentCount);
  const radius = Math.min(state.width, state.height) / density / zoomScale;
  const offsetAngle = Math.PI / 4;

  stateNodes = stateNodes.map((node, nodeId) => {
    const index = adjacentLayoutNodeIds.indexOf(nodeId);
    if (index > -1) {
      const angle = offsetAngle + ((Math.PI * 2 * index) / adjacentCount);
      return node.merge({
        x: centerX + (radius * Math.sin(angle)),
        y: centerY + (radius * Math.cos(angle))
      });
    }
    return node;
  });

  // fix all edges for circular nodes
  stateEdges = stateEdges.map((edge) => {
    if (edge.get('source') === props.selectedNodeId
      || edge.get('target') === props.selectedNodeId
      || includes(adjacentLayoutNodeIds, edge.get('source'))
      || includes(adjacentLayoutNodeIds, edge.get('target'))) {
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
  // const selectedNodeScale = getNodeScale(adjacentNodes.size, state.width, state.height);

  return {
    selectedScale: 1,
    edges: stateEdges,
    nodes: stateNodes
  };
}


class NodesChart extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.state = Object.assign({
      scale: 1,
      minScale: 1,
      maxScale: 1,
      panTranslateX: 0,
      panTranslateY: 0,
      selectedScale: 1,
      height: props.height || 0,
      width: props.width || 0,
      zoomCache: {},
    }, emptyLayoutState);

    this.handleMouseClick = this.handleMouseClick.bind(this);
    this.zoomed = this.zoomed.bind(this);
  }

  componentWillMount() {
    const state = updatedGraphState(this.props, this.state);
    // debugger;
    // assign(state, this.restoreZoomState(this.props, Object.assign(this.state, state)));
    this.setState(state);
  }

  componentWillReceiveProps(nextProps) {
    // gather state, setState should be called only once here
    const state = assign({}, this.state);

    const topologyChanged = nextProps.topologyId !== this.props.topologyId;

    // wipe node states when showing different topology
    if (topologyChanged) {
      assign(state, emptyLayoutState);
    }

    // reset layout dimensions only when forced
    state.height = nextProps.forceRelayout ? nextProps.height : (state.height || nextProps.height);
    state.width = nextProps.forceRelayout ? nextProps.width : (state.width || nextProps.width);

    if (nextProps.forceRelayout || nextProps.nodes !== this.props.nodes) {
      assign(state, updatedGraphState(nextProps, state));
    }

    console.log(`Prepare ${nextProps.nodes.size}`);
    if (nextProps.nodes.size > 0) {
      console.log(state.zoomCache);
      assign(state, this.restoreZoomState(nextProps, state));
    }

    // if (this.props.selectedNodeId !== nextProps.selectedNodeId) {
    //   // undo any pan/zooming that might have happened
    //   this.setZoom(state);
    //   assign(state, restoredLayout(state));
    // }
    //
    // if (nextProps.selectedNodeId) {
    //   assign(state, centerSelectedNode(nextProps, state));
    // }

    if (topologyChanged) {
      // saving previous zoom state
      const prevZoom = pick(this.state, ZOOM_CACHE_FIELDS);
      const zoomCache = assign({}, this.state.zoomCache);
      zoomCache[this.props.topologyId] = prevZoom;
      assign(state, { zoomCache });
    }

    // console.log(topologyChanged);
    // console.log(state);
    this.setState(state);
  }

  componentDidMount() {
    // distinguish pan/zoom from click
    this.isZooming = false;
    // debugger;

    this.zoom = zoom()
      .scaleExtent([this.state.minScale, this.state.maxScale])
      .on('zoom', this.zoomed);

    this.svg = select('.nodes-chart svg');
    this.svg.call(this.zoom);
    // this.setZoom(this.state);
  }

  componentWillUnmount() {
    // undoing .call(zoom)
    this.svg
      .on('mousedown.zoom', null)
      .on('onwheel', null)
      .on('onmousewheel', null)
      .on('dblclick.zoom', null)
      .on('touchstart.zoom', null);
  }

  isSmallTopology() {
    return this.state.nodes.size < 100;
  }

  render() {
    const { edges, nodes, panTranslateX, panTranslateY, scale } = this.state;
    console.log(`Render ${nodes.size}`);

    // not passing translates into child components for perf reasons, use getTranslate instead
    const translate = [panTranslateX, panTranslateY];
    const transform = `translate(${translate}) scale(${scale})`;
    const svgClassNames = this.props.isEmpty ? 'hide' : '';

    return (
      <div className="nodes-chart">
        <svg
          width="100%" height="100%" id="nodes-chart-canvas"
          className={svgClassNames} onClick={this.handleMouseClick}>
          <g transform="translate(24,24) scale(0.25)">
            <Logo />
          </g>
          <NodesChartElements
            layoutNodes={nodes}
            layoutEdges={edges}
            scale={scale}
            selectedScale={this.state.selectedScale}
            transform={transform}
            isAnimated={this.isSmallTopology()} />
        </svg>
      </div>
    );
  }

  handleMouseClick() {
    if (!this.isZooming || this.props.selectedNodeId) {
      this.props.clickBackground();
    } else {
      this.isZooming = false;
    }
  }

  restoreZoomState(props, state) {
    // re-apply cached canvas zoom/pan to d3 behavior (or set the default values)
    const nextZoom = state.zoomCache[props.topologyId] || defaultZoomState(props, state);
    if (this.zoom) {
      this.zoom = this.zoom.scaleExtent([nextZoom.minScale, nextZoom.maxScale]);
      this.setZoom(nextZoom);
    }

    return nextZoom;
  }

  zoomed() {
    this.isZooming = true;
    // don't pan while node is selected
    if (!this.props.selectedNodeId) {
      this.setState({
        panTranslateX: d3Event.transform.x,
        panTranslateY: d3Event.transform.y,
        scale: d3Event.transform.k
      });
    }
    // console.log(d3Event.transform);
  }

  setZoom(newZoom) {
    this.svg.call(this.zoom.transform, zoomIdentity
      .translate(newZoom.panTranslateX, newZoom.panTranslateY)
      .scale(newZoom.scale));
  }
}


function mapStateToProps(state) {
  return {
    nodes: nodeAdjacenciesSelector(state),
    adjacentNodes: adjacentNodesSelector(state),
    forceRelayout: state.get('forceRelayout'),
    selectedNodeId: state.get('selectedNodeId'),
    topologyId: state.get('currentTopologyId'),
    topologyOptions: getActiveTopologyOptions(state)
  };
}


export default connect(
  mapStateToProps,
  { clickBackground }
)(NodesChart);

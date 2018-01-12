import _ from 'lodash';
import d3 from 'd3';
import debug from 'debug';
import React from 'react';
import { connect } from 'react-redux';
import { Map as makeMap, fromJS } from 'immutable';
import timely from 'timely';

import { nodeAdjacenciesSelector, adjacentNodesSelector } from '../selectors/chartSelectors';
import { clickBackground } from '../actions/app-actions';
import { EDGE_ID_SEPARATOR } from '../constants/naming';
import { MIN_NODE_SIZE, DETAILS_PANEL_WIDTH, MAX_NODE_SIZE } from '../constants/styles';
import Logo from '../components/logo';
import { doLayout } from './nodes-layout';
import NodesChartElements from './nodes-chart-elements';
import { getActiveTopologyOptions } from '../utils/topology-utils';

const log = debug('scope:nodes-chart');

const ZOOM_CACHE_FIELDS = ['scale', 'panTranslateX', 'panTranslateY'];

// make sure circular layouts a bit denser with 3-6 nodes
const radiusDensity = d3.scale.threshold()
  .domain([3, 6])
  .range([2.5, 3.5, 3]);

/**
 * dynamic coords precision based on topology size
 */
function getLayoutPrecision(nodesCount) {
  let precision;
  if (nodesCount >= 50) {
    precision = 0;
  } else if (nodesCount > 20) {
    precision = 1;
  } else if (nodesCount > 10) {
    precision = 2;
  } else {
    precision = 3;
  }

  return precision;
}


function initEdges(nodes) {
  let edges = makeMap();

  nodes.forEach((node, nodeId) => {
    const adjacency = node.get('adjacency');
    if (adjacency) {
      adjacency.forEach(adjacent => {
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


function getNodeScale(nodesCount, width, height) {
  const expanse = Math.min(height, width);
  const nodeSize = expanse / 3; // single node should fill a third of the screen
  const maxNodeSize = Math.min(MAX_NODE_SIZE, expanse / 10);
  const normalizedNodeSize = Math.max(MIN_NODE_SIZE,
    Math.min(nodeSize / Math.sqrt(nodesCount), maxNodeSize));

  return d3.scale.linear().range([0, normalizedNodeSize]);
}


function updateLayout(width, height, nodes, baseOptions) {
  const nodeScale = getNodeScale(nodes.size, width, height);
  const edges = initEdges(nodes);

  const options = Object.assign({}, baseOptions, {
    scale: nodeScale,
  });

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

  const layoutEdges = graph.edges
    .map(edge => edge.set('ppoints', edge.get('points')));

  return { layoutNodes, layoutEdges, layoutWidth: graph.width, layoutHeight: graph.height };
}


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
      zoomCache: {},
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
      if (nextZoom) {
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
    state.height = nextProps.forceRelayout ? nextProps.height : (state.height || nextProps.height);
    state.width = nextProps.forceRelayout ? nextProps.width : (state.width || nextProps.width);

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

  componentDidMount() {
    // distinguish pan/zoom from click
    this.isZooming = false;

    this.zoom = d3.behavior.zoom()
      .scaleExtent([0.1, 2])
      .on('zoom', this.zoomed);

    d3.select('.nodes-chart svg')
      .call(this.zoom);
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

  parseHash() {
    try {
      let hash = window.location.hash;
      hash = hash.substring(hash.lastIndexOf('/') + 1);
      const end = hash.indexOf('?');
      if (end !== -1) {
        hash = hash.substring(0, end);
      }
      hash = JSON.parse(hash);
      return hash;
    } catch (e) {
      return {};
    }
  }

  replaceHash(obj) {
    let hash = window.location.hash;
    const start = hash.substring(0, hash.lastIndexOf('/') + 1);
    let end = hash.indexOf('?');
    if (end !== -1) {
      end = hash.substring(end);
    } else {
      end = '';
    }
    hash = start + JSON.stringify(obj) + end;
    window.location.hash = hash;
    window.location.reload();
  }
  render() {
    const { edges, nodes, panTranslateX, panTranslateY, scale } = this.state;
    // 获取详情
    const { nodeDatas, nodeDetails } = this.props;
    const hash = this.parseHash();
    let noNodeDetail = false;
    if (!nodeDetails || nodeDetails.size === 0) {
      noNodeDetail = true;
    }
    if (noNodeDetail && nodeDatas && nodeDatas.size > 0 && hash.selectedLabel) {
      nodeDatas.map((node) => {
        // 不同版本的mesos生成的容器名称不一样
        let shortLable = hash.selectedLabel;
        if (hash.selectedLabel.indexOf('.') !== -1) {
          shortLable = shortLable.substring(0, hash.selectedLabel.indexOf('-') + 1) +
            shortLable.substring(hash.selectedLabel.lastIndexOf('.') + 1);
        }
        if (node.get('label') === hash.selectedLabel || node.get('label') === shortLable) {
          hash.nodeDetails = [{id: node.get('id'), label: node.get('label'),
            topologyId: 'containers'}];
          delete hash.selectedLabel;
          hash.selectedNodeId = node.get('id');
          this.replaceHash(hash);
        }
        return null;
      });
    }
    const ns = window.isNormal() ? nodes : makeMap();

    // not passing translates into child components for perf reasons, use getTranslate instead
    const translate = [panTranslateX, panTranslateY];
    const transform = `translate(${translate}) scale(${scale})`;
    const svgClassNames = this.props.isEmpty ? 'hide' : '';

    const layoutPrecision = getLayoutPrecision(ns.size);
    return (
      <div className="nodes-chart">
        <svg width="100%" height="100%" id="nodes-chart-canvas"
          className={svgClassNames} onClick={this.handleMouseClick}>
          <g transform="translate(24,24) scale(0.25)">
            <Logo />
          </g>
          <NodesChartElements
            layoutNodes={ns}
            layoutEdges={edges}
            nodeScale={this.state.nodeScale}
            scale={scale}
            transform={transform}
            selectedNodeScale={this.state.selectedNodeScale}
            layoutPrecision={layoutPrecision} />
        </svg>
      </div>
    );
  }

  handleMouseClick() {
    if (!window.isNormal()) {
      return;
    }
    if (!this.isZooming) {
      this.props.clickBackground();
    } else {
      this.isZooming = false;
    }
  }

  centerSelectedNode(props, state) {
    let stateNodes = state.nodes;
    let stateEdges = state.edges;
    if (!stateNodes.has(props.selectedNodeId)) {
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

    stateNodes = stateNodes.map((node, nodeId) => {
      const index = adjacentLayoutNodeIds.indexOf(nodeId);
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
      if (edge.get('source') === props.selectedNodeId
        || edge.get('target') === props.selectedNodeId
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
    const selectedNodeScale = getNodeScale(adjacentNodes.size, state.width, state.height);

    return {
      selectedNodeScale,
      edges: stateEdges,
      nodes: stateNodes
    };
  }

  restoreLayout(state) {
    // undo any pan/zooming that might have happened
    this.zoom.scale(state.scale);
    this.zoom.translate([state.panTranslateX, state.panTranslateY]);

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

    const options = {
      width: state.width,
      height: state.height,
      margins: props.margins,
      forceRelayout: props.forceRelayout,
      topologyId: props.topologyId,
      topologyOptions: props.topologyOptions,
    };

    const { layoutNodes, layoutEdges, layoutWidth, layoutHeight } = updateLayout(
      state.width, state.height, props.nodes, options);
    //
    // adjust layout based on viewport
    const xFactor = (state.width - props.margins.left - props.margins.right) / layoutWidth;
    const yFactor = state.height / layoutHeight;
    const zoomFactor = Math.min(xFactor, yFactor);
    let zoomScale = state.scale;

    if (this.zoom && !state.hasZoomed && zoomFactor > 0 && zoomFactor < 1) {
      zoomScale = zoomFactor;
      // saving in d3's behavior cache
      this.zoom.scale(zoomFactor);
    }

    return {
      scale: zoomScale,
      nodes: layoutNodes,
      edges: layoutEdges,
      nodeScale: getNodeScale(props.nodes.size, state.width, state.height),
    };
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
    nodes: nodeAdjacenciesSelector(state),
    adjacentNodes: adjacentNodesSelector(state),
    forceRelayout: state.get('forceRelayout'),
    selectedNodeId: state.get('selectedNodeId'),
    topologyId: state.get('currentTopologyId'),
    topologyOptions: getActiveTopologyOptions(state),
    nodeDetails: state.get('nodeDetails'),
    nodeDatas: state.get('nodes')
  };
}


export default connect(
  mapStateToProps,
  { clickBackground }
)(NodesChart);

import _ from 'lodash';
import debug from 'debug';
import { fromJS, is as isDeepEqual, List, Map, OrderedMap, Set } from 'immutable';
import { Store } from 'flux/utils';

import AppDispatcher from '../dispatcher/app-dispatcher';
import ActionTypes from '../constants/action-types';
import { EDGE_ID_SEPARATOR } from '../constants/naming';
import { findTopologyById, setTopologyUrlsById, updateTopologyIds,
  filterHiddenTopologies } from '../utils/topology-utils';

const makeList = List;
const makeMap = Map;
const makeOrderedMap = OrderedMap;
const makeSet = Set;
const log = debug('scope:app-store');

const error = debug('scope:error');

// Helpers

function makeNode(node) {
  return {
    id: node.id,
    label: node.label,
    label_minor: node.label_minor,
    node_count: node.node_count,
    rank: node.rank,
    pseudo: node.pseudo,
    stack: node.stack,
    shape: node.shape,
    adjacency: node.adjacency,
    metrics: node.metrics
  };
}

// Initial values

let topologyOptions = makeOrderedMap(); // topologyId -> options
let controlStatus = makeMap();
let currentTopology = null;
let currentTopologyId = 'containers';
let errorUrl = null;
let forceRelayout = false;
let highlightedEdgeIds = makeSet();
let highlightedNodeIds = makeSet();
let hostname = '...';
let version = '...';
let versionUpdate = null;
let plugins = [];
let mouseOverEdgeId = null;
let mouseOverNodeId = null;
let nodeDetails = makeOrderedMap(); // nodeId -> details
let nodes = makeOrderedMap(); // nodeId -> node
let selectedNodeId = null;
let topologies = makeList();
let topologiesLoaded = false;
let topologyUrlsById = makeOrderedMap(); // topologyId -> topologyUrl
let routeSet = false;
let controlPipes = makeOrderedMap(); // pipeId -> controlPipe
let updatePausedAt = null; // Date
let websocketClosed = true;
let showingHelp = false;
let tableSortOrder = null;

let selectedMetric = null;
let pinnedMetric = selectedMetric;
// class of metric, e.g. 'cpu', rather than 'host_cpu' or 'process_cpu'.
// allows us to keep the same metric "type" selected when the topology changes.
let pinnedMetricType = null;
let availableCanvasMetrics = makeList();


const topologySorter = topology => topology.get('rank');

// adds ID field to topology (based on last part of URL path) and save urls in
// map for easy lookup
function processTopologies(nextTopologies) {
  // filter out hidden topos
  const visibleTopologies = filterHiddenTopologies(nextTopologies);

  // add IDs to topology objects in-place
  const topologiesWithId = updateTopologyIds(visibleTopologies);

  // cache URLs by ID
  topologyUrlsById = setTopologyUrlsById(topologyUrlsById, topologiesWithId);

  const immNextTopologies = fromJS(topologiesWithId).sortBy(topologySorter);
  topologies = topologies.mergeDeep(immNextTopologies);
}

function setTopology(topologyId) {
  currentTopology = findTopologyById(topologies, topologyId);
  currentTopologyId = topologyId;
}

function setDefaultTopologyOptions(topologyList) {
  topologyList.forEach(topology => {
    let defaultOptions = makeOrderedMap();
    if (topology.has('options') && topology.get('options')) {
      topology.get('options').forEach((option) => {
        const optionId = option.get('id');
        const defaultValue = option.get('defaultValue');
        defaultOptions = defaultOptions.set(optionId, defaultValue);
      });
    }

    if (defaultOptions.size) {
      topologyOptions = topologyOptions.set(
        topology.get('id'),
        defaultOptions
      );
    }
  });
}

function closeNodeDetails(nodeId) {
  if (nodeDetails.size > 0) {
    const popNodeId = nodeId || nodeDetails.keySeq().last();
    // remove pipe if it belongs to the node being closed
    controlPipes = controlPipes.filter(pipe => pipe.get('nodeId') !== popNodeId);
    nodeDetails = nodeDetails.delete(popNodeId);
  }
  if (nodeDetails.size === 0 || selectedNodeId === nodeId) {
    selectedNodeId = null;
  }
}

function closeAllNodeDetails() {
  while (nodeDetails.size) {
    closeNodeDetails();
  }
}

function resumeUpdate() {
  updatePausedAt = null;
}

// Store API

export class AppStore extends Store {

  // keep at the top
  getAppState() {
    const cp = this.getControlPipe();
    return {
      controlPipe: cp ? cp.toJS() : null,
      nodeDetails: this.getNodeDetailsState().toJS(),
      selectedNodeId,
      pinnedMetricType,
      topologyId: currentTopologyId,
      topologyOptions: topologyOptions.toJS() // all options
    };
  }

  getTableSortOrder() {
    return tableSortOrder;
  }

  getShowingHelp() {
    return showingHelp;
  }

  getActiveTopologyOptions() {
    // options for current topology, sub-topologies share options with parent
    if (currentTopology && currentTopology.get('parentId')) {
      return topologyOptions.get(currentTopology.get('parentId'));
    }
    return topologyOptions.get(currentTopologyId);
  }

  getAdjacentNodes(nodeId) {
    let adjacentNodes = makeSet();

    if (nodes.has(nodeId)) {
      adjacentNodes = makeSet(nodes.getIn([nodeId, 'adjacency']));
      // fill up set with reverse edges
      nodes.forEach((node, id) => {
        if (node.get('adjacency') && node.get('adjacency').includes(nodeId)) {
          adjacentNodes = adjacentNodes.add(id);
        }
      });
    }

    return adjacentNodes;
  }

  getPinnedMetric() {
    return pinnedMetric;
  }

  getSelectedMetric() {
    return selectedMetric;
  }

  getAvailableCanvasMetrics() {
    return availableCanvasMetrics;
  }

  getAvailableCanvasMetricsTypes() {
    return makeMap(this.getAvailableCanvasMetrics().map(m => [m.get('id'), m.get('label')]));
  }

  getControlStatus() {
    return controlStatus;
  }

  getControlPipe() {
    return controlPipes.last();
  }

  getCurrentTopology() {
    if (!currentTopology) {
      currentTopology = setTopology(currentTopologyId);
    }
    return currentTopology;
  }

  getCurrentTopologyId() {
    return currentTopologyId;
  }

  getCurrentTopologyOptions() {
    return currentTopology && currentTopology.get('options') || makeOrderedMap();
  }

  getCurrentTopologyUrl() {
    return currentTopology && currentTopology.get('url');
  }

  getErrorUrl() {
    return errorUrl;
  }

  getHighlightedEdgeIds() {
    return highlightedEdgeIds;
  }

  getHighlightedNodeIds() {
    return highlightedNodeIds;
  }

  getHostname() {
    return hostname;
  }

  getNodeDetails() {
    return nodeDetails;
  }

  getNodeDetailsState() {
    return nodeDetails.toIndexedSeq().map(details => ({
      id: details.id, label: details.label, topologyId: details.topologyId
    }));
  }

  getTopCardNodeId() {
    return nodeDetails.last() && nodeDetails.last().id;
  }

  getNodes() {
    return nodes;
  }

  getSelectedNodeId() {
    return selectedNodeId;
  }

  getTopologies() {
    return topologies;
  }

  getTopologyUrlsById() {
    return topologyUrlsById;
  }

  getUpdatePausedAt() {
    return updatePausedAt;
  }

  getVersion() {
    return version;
  }

  getVersionUpdate() {
    return versionUpdate;
  }

  getPlugins() {
    return plugins;
  }

  isForceRelayout() {
    return forceRelayout;
  }

  isRouteSet() {
    return routeSet;
  }

  isTopologiesLoaded() {
    return topologiesLoaded;
  }

  isTopologyEmpty() {
    return currentTopology && currentTopology.get('stats')
      && currentTopology.get('stats').get('node_count') === 0 && nodes.size === 0;
  }

  isUpdatePaused() {
    return updatePausedAt !== null;
  }

  isWebsocketClosed() {
    return websocketClosed;
  }

  __onDispatch(payload) {
    if (!payload.type) {
      error('Payload missing a type!', payload);
    }

    switch (payload.type) {
      case ActionTypes.CHANGE_TOPOLOGY_OPTION: {
        resumeUpdate();
        // set option on parent topology
        const topology = findTopologyById(topologies, payload.topologyId);
        if (topology) {
          const topologyId = topology.get('parentId') || topology.get('id');
          if (topologyOptions.getIn([topologyId, payload.option]) !== payload.value) {
            nodes = nodes.clear();
          }
          topologyOptions = topologyOptions.setIn(
            [topologyId, payload.option],
            payload.value
          );
          this.__emitChange();
        }
        break;
      }
      case ActionTypes.CLEAR_CONTROL_ERROR: {
        controlStatus = controlStatus.removeIn([payload.nodeId, 'error']);
        this.__emitChange();
        break;
      }
      case ActionTypes.CLICK_BACKGROUND: {
        closeAllNodeDetails();
        this.__emitChange();
        break;
      }
      case ActionTypes.CLICK_CLOSE_DETAILS: {
        closeNodeDetails(payload.nodeId);
        this.__emitChange();
        break;
      }
      case ActionTypes.SORT_ORDER_CHANGED: {
        tableSortOrder = makeMap((payload.newOrder || []).map((n, i) => [n.id, i]));
        this.__emitChange();
        break;
      }
      case ActionTypes.CLICK_CLOSE_TERMINAL: {
        controlPipes = controlPipes.clear();
        this.__emitChange();
        break;
      }
      case ActionTypes.CLICK_FORCE_RELAYOUT: {
        forceRelayout = true;
        // fire only once, reset after emitChange
        setTimeout(() => {
          forceRelayout = false;
        }, 0);
        this.__emitChange();
        break;
      }
      case ActionTypes.CLICK_NODE: {
        const prevSelectedNodeId = selectedNodeId;
        const prevDetailsStackSize = nodeDetails.size;
        // click on sibling closes all
        closeAllNodeDetails();
        // select new node if it's not the same (in that case just delesect)
        if (prevDetailsStackSize > 1 || prevSelectedNodeId !== payload.nodeId) {
          // dont set origin if a node was already selected, suppresses animation
          const origin = prevSelectedNodeId === null ? payload.origin : null;
          nodeDetails = nodeDetails.set(
            payload.nodeId,
            {
              id: payload.nodeId,
              label: payload.label,
              origin,
              topologyId: currentTopologyId
            }
          );
          selectedNodeId = payload.nodeId;
        }
        this.__emitChange();
        break;
      }
      case ActionTypes.CLICK_PAUSE_UPDATE: {
        updatePausedAt = new Date;
        this.__emitChange();
        break;
      }
      case ActionTypes.CLICK_RELATIVE: {
        if (nodeDetails.has(payload.nodeId)) {
          // bring to front
          const details = nodeDetails.get(payload.nodeId);
          nodeDetails = nodeDetails.delete(payload.nodeId);
          nodeDetails = nodeDetails.set(payload.nodeId, details);
        } else {
          nodeDetails = nodeDetails.set(
            payload.nodeId,
            {
              id: payload.nodeId,
              label: payload.label,
              origin: payload.origin,
              topologyId: payload.topologyId
            }
          );
        }
        this.__emitChange();
        break;
      }
      case ActionTypes.CLICK_RESUME_UPDATE: {
        resumeUpdate();
        this.__emitChange();
        break;
      }
      case ActionTypes.CLICK_SHOW_TOPOLOGY_FOR_NODE: {
        resumeUpdate();
        nodeDetails = nodeDetails.filter((v, k) => k === payload.nodeId);
        controlPipes = controlPipes.clear();
        selectedNodeId = payload.nodeId;
        if (payload.topologyId !== currentTopologyId) {
          setTopology(payload.topologyId);
          nodes = nodes.clear();
        }
        availableCanvasMetrics = makeList();
        tableSortOrder = null;
        this.__emitChange();
        break;
      }
      case ActionTypes.CLICK_TOPOLOGY: {
        resumeUpdate();
        closeAllNodeDetails();
        if (payload.topologyId !== currentTopologyId) {
          setTopology(payload.topologyId);
          nodes = nodes.clear();
        }
        availableCanvasMetrics = makeList();
        tableSortOrder = null;

        this.__emitChange();
        break;
      }
      case ActionTypes.CLOSE_WEBSOCKET: {
        if (!websocketClosed) {
          websocketClosed = true;
          this.__emitChange();
        }
        break;
      }
      case ActionTypes.SELECT_METRIC: {
        selectedMetric = payload.metricId;
        this.__emitChange();
        break;
      }
      case ActionTypes.PIN_METRIC: {
        pinnedMetric = payload.metricId;
        pinnedMetricType = this.getAvailableCanvasMetricsTypes().get(payload.metricId);
        selectedMetric = payload.metricId;
        this.__emitChange();
        break;
      }
      case ActionTypes.UNPIN_METRIC: {
        pinnedMetric = null;
        pinnedMetricType = null;
        this.__emitChange();
        break;
      }
      case ActionTypes.SHOW_HELP: {
        showingHelp = true;
        this.__emitChange();
        break;
      }
      case ActionTypes.HIDE_HELP: {
        showingHelp = false;
        this.__emitChange();
        break;
      }
      case ActionTypes.DESELECT_NODE: {
        closeNodeDetails();
        this.__emitChange();
        break;
      }
      case ActionTypes.DO_CONTROL: {
        controlStatus = controlStatus.set(payload.nodeId, makeMap({
          pending: true,
          error: null
        }));
        this.__emitChange();
        break;
      }
      case ActionTypes.ENTER_EDGE: {
        // clear old highlights
        highlightedNodeIds = highlightedNodeIds.clear();
        highlightedEdgeIds = highlightedEdgeIds.clear();

        // highlight edge
        highlightedEdgeIds = highlightedEdgeIds.add(payload.edgeId);

        // highlight adjacent nodes
        highlightedNodeIds = highlightedNodeIds.union(payload.edgeId.split(EDGE_ID_SEPARATOR));

        this.__emitChange();
        break;
      }
      case ActionTypes.ENTER_NODE: {
        const nodeId = payload.nodeId;
        const adjacentNodes = this.getAdjacentNodes(nodeId);

        // clear old highlights
        highlightedNodeIds = highlightedNodeIds.clear();
        highlightedEdgeIds = highlightedEdgeIds.clear();

        // highlight nodes
        highlightedNodeIds = highlightedNodeIds.add(nodeId);
        highlightedNodeIds = highlightedNodeIds.union(adjacentNodes);

        // highlight edges
        if (adjacentNodes.size > 0) {
          // all neighbour combinations because we dont know which direction exists
          highlightedEdgeIds = highlightedEdgeIds.union(adjacentNodes.flatMap((adjacentId) => [
            [adjacentId, nodeId].join(EDGE_ID_SEPARATOR),
            [nodeId, adjacentId].join(EDGE_ID_SEPARATOR)
          ]));
        }

        this.__emitChange();
        break;
      }
      case ActionTypes.LEAVE_EDGE: {
        highlightedEdgeIds = highlightedEdgeIds.clear();
        highlightedNodeIds = highlightedNodeIds.clear();
        this.__emitChange();
        break;
      }
      case ActionTypes.LEAVE_NODE: {
        highlightedEdgeIds = highlightedEdgeIds.clear();
        highlightedNodeIds = highlightedNodeIds.clear();
        this.__emitChange();
        break;
      }
      case ActionTypes.OPEN_WEBSOCKET: {
        // flush nodes cache after re-connect
        nodes = nodes.clear();
        websocketClosed = false;

        this.__emitChange();
        break;
      }
      case ActionTypes.DO_CONTROL_ERROR: {
        controlStatus = controlStatus.set(payload.nodeId, makeMap({
          pending: false,
          error: payload.error
        }));
        this.__emitChange();
        break;
      }
      case ActionTypes.DO_CONTROL_SUCCESS: {
        controlStatus = controlStatus.set(payload.nodeId, makeMap({
          pending: false,
          error: null
        }));
        this.__emitChange();
        break;
      }
      case ActionTypes.RECEIVE_CONTROL_PIPE: {
        controlPipes = controlPipes.set(payload.pipeId, makeOrderedMap({
          id: payload.pipeId,
          nodeId: payload.nodeId,
          raw: payload.rawTty
        }));
        this.__emitChange();
        break;
      }
      case ActionTypes.RECEIVE_CONTROL_PIPE_STATUS: {
        if (controlPipes.has(payload.pipeId)) {
          controlPipes = controlPipes.setIn([payload.pipeId, 'status'], payload.status);
          this.__emitChange();
        }
        break;
      }
      case ActionTypes.RECEIVE_ERROR: {
        if (errorUrl !== null) {
          errorUrl = payload.errorUrl;
          this.__emitChange();
        }
        break;
      }
      case ActionTypes.RECEIVE_NODE_DETAILS: {
        errorUrl = null;

        // disregard if node is not selected anymore
        if (nodeDetails.has(payload.details.id)) {
          nodeDetails = nodeDetails.update(payload.details.id, obj => {
            const result = Object.assign({}, obj);
            result.notFound = false;
            result.details = payload.details;
            return result;
          });
        }
        this.__emitChange();
        break;
      }
      case ActionTypes.RECEIVE_NODES_DELTA: {
        const emptyMessage = !payload.delta.add && !payload.delta.remove
          && !payload.delta.update;
        // this action is called frequently, good to check if something changed
        const emitChange = !emptyMessage || errorUrl !== null;

        if (!emptyMessage) {
          log('RECEIVE_NODES_DELTA',
            'remove', _.size(payload.delta.remove),
            'update', _.size(payload.delta.update),
            'add', _.size(payload.delta.add));
        }

        errorUrl = null;

        // nodes that no longer exist
        _.each(payload.delta.remove, (nodeId) => {
          // in case node disappears before mouseleave event
          if (mouseOverNodeId === nodeId) {
            mouseOverNodeId = null;
          }
          if (nodes.has(nodeId) && _.includes(mouseOverEdgeId, nodeId)) {
            mouseOverEdgeId = null;
          }
          nodes = nodes.delete(nodeId);
        });

        // update existing nodes
        _.each(payload.delta.update, (node) => {
          if (nodes.has(node.id)) {
            nodes = nodes.set(node.id, nodes.get(node.id).merge(fromJS(node)));
          }
        });

        // add new nodes
        _.each(payload.delta.add, (node) => {
          nodes = nodes.set(node.id, fromJS(makeNode(node)));
        });

        availableCanvasMetrics = nodes
          .valueSeq()
          .flatMap(n => (n.get('metrics') || makeList()).map(m => (
            makeMap({id: m.get('id'), label: m.get('label')})
          )))
          .toSet()
          .toList()
          .sortBy(m => m.get('label'));

        const similarTypeMetric = availableCanvasMetrics
          .find(m => m.get('label') === pinnedMetricType);
        pinnedMetric = similarTypeMetric && similarTypeMetric.get('id');
        // if something in the current topo is not already selected, select it.
        if (!availableCanvasMetrics.map(m => m.get('id')).toSet().has(selectedMetric)) {
          selectedMetric = pinnedMetric;
        }

        if (emitChange) {
          this.__emitChange();
        }
        break;
      }
      case ActionTypes.RECEIVE_NOT_FOUND: {
        if (nodeDetails.has(payload.nodeId)) {
          nodeDetails = nodeDetails.update(payload.nodeId, obj => {
            const result = Object.assign({}, obj);
            result.notFound = true;
            return result;
          });
          this.__emitChange();
        }
        break;
      }
      case ActionTypes.RECEIVE_TOPOLOGIES: {
        errorUrl = null;
        topologyUrlsById = topologyUrlsById.clear();
        processTopologies(payload.topologies);
        setTopology(currentTopologyId);
        // only set on first load, if options are not already set via route
        if (!topologiesLoaded && topologyOptions.size === 0) {
          setDefaultTopologyOptions(topologies);
        }
        topologiesLoaded = true;

        this.__emitChange();
        break;
      }
      case ActionTypes.RECEIVE_API_DETAILS: {
        errorUrl = null;
        hostname = payload.hostname;
        version = payload.version;
        plugins = payload.plugins;
        versionUpdate = payload.newVersion;
        this.__emitChange();
        break;
      }
      case ActionTypes.ROUTE_TOPOLOGY: {
        routeSet = true;
        if (currentTopologyId !== payload.state.topologyId) {
          nodes = nodes.clear();
        }
        setTopology(payload.state.topologyId);
        setDefaultTopologyOptions(topologies);
        selectedNodeId = payload.state.selectedNodeId;
        pinnedMetricType = payload.state.pinnedMetricType;
        if (payload.state.controlPipe) {
          controlPipes = makeOrderedMap({
            [payload.state.controlPipe.id]:
              makeOrderedMap(payload.state.controlPipe)
          });
        } else {
          controlPipes = controlPipes.clear();
        }
        if (payload.state.nodeDetails) {
          const payloadNodeDetails = makeOrderedMap(
            payload.state.nodeDetails.map(obj => [obj.id, obj]));
          // check if detail IDs have changed
          if (!isDeepEqual(nodeDetails.keySeq(), payloadNodeDetails.keySeq())) {
            nodeDetails = payloadNodeDetails;
          }
        } else {
          nodeDetails = nodeDetails.clear();
        }
        topologyOptions = fromJS(payload.state.topologyOptions)
          || topologyOptions;
        this.__emitChange();
        break;
      }
      default: {
        break;
      }
    }
  }
}

export default new AppStore(AppDispatcher);

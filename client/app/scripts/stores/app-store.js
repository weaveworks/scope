import _ from 'lodash';
import debug from 'debug';
import { fromJS, List, Map, OrderedMap, Set } from 'immutable';
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
let selectedMetric = 'process_cpu_usage_percent';

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
  currentTopologyId = topologyId;
  currentTopology = findTopologyById(topologies, topologyId);
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

    if (topology.has('sub_topologies')) {
      setDefaultTopologyOptions(topology.get('sub_topologies'));
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
    return {
      controlPipe: this.getControlPipe(),
      nodeDetails: this.getNodeDetailsState(),
      selectedNodeId,
      topologyId: currentTopologyId,
      topologyOptions: topologyOptions.toJS() // all options
    };
  }

  getActiveTopologyOptions() {
    // options for current topology
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

  getSelectedMetric() {
    return selectedMetric;
  }

  getControlStatus() {
    return controlStatus.toJS();
  }

  getControlPipe() {
    const cp = controlPipes.last();
    return cp && cp.toJS();
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
    })).toJS();
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
        if (topologyOptions.getIn([payload.topologyId, payload.option])
          !== payload.value) {
          nodes = nodes.clear();
        }
        topologyOptions = topologyOptions.setIn(
          [payload.topologyId, payload.option],
          payload.value
        );
        this.__emitChange();
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
        _.each(payload.delta.update, function(node) {
          if (nodes.has(node.id)) {
            nodes = nodes.set(node.id, nodes.get(node.id).merge(Immutable.fromJS(node)));
          }
        });

        // add new nodes
        _.each(payload.delta.add, (node) => {
          nodes = nodes.set(node.id, fromJS(makeNode(node)));
        });

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
        if (payload.state.controlPipe) {
          controlPipes = makeOrderedMap({
            [payload.state.controlPipe.id]:
              makeOrderedMap(payload.state.controlPipe)
          });
        } else {
          controlPipes = controlPipes.clear();
        }
        if (payload.state.nodeDetails) {
          nodeDetails = makeOrderedMap(payload.state.nodeDetails.map(obj => [obj.id, obj]));
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

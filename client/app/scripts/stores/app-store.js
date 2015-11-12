const EventEmitter = require('events').EventEmitter;
const _ = require('lodash');
const debug = require('debug')('scope:app-store');
const Immutable = require('immutable');

const AppDispatcher = require('../dispatcher/app-dispatcher');
const ActionTypes = require('../constants/action-types');
const Naming = require('../constants/naming');

const makeOrderedMap = Immutable.OrderedMap;
const makeSet = Immutable.Set;

// Helpers

function findCurrentTopology(subTree, topologyId) {
  let foundTopology;

  _.each(subTree, function(topology) {
    if (_.endsWith(topology.url, topologyId)) {
      foundTopology = topology;
    }
    if (!foundTopology) {
      foundTopology = findCurrentTopology(topology.sub_topologies, topologyId);
    }
    if (foundTopology) {
      return false;
    }
  });

  return foundTopology;
}

function makeNode(node) {
  return {
    id: node.id,
    label_major: node.label_major,
    label_minor: node.label_minor,
    rank: node.rank,
    pseudo: node.pseudo,
    adjacency: node.adjacency
  };
}

// Initial values

let topologyOptions = makeOrderedMap();
let adjacentNodes = makeSet();
let controlError = null;
let controlPending = false;
let currentTopology = null;
let currentTopologyId = 'containers';
let errorUrl = null;
let version = '';
let mouseOverEdgeId = null;
let mouseOverNodeId = null;
let nodes = makeOrderedMap();
let nodeDetails = null;
let selectedNodeId = null;
let topologies = [];
let topologiesLoaded = false;
let routeSet = false;
let websocketClosed = true;

function processTopologies(topologyList) {
  // adds ID field to topology, based on last part of URL path
  _.each(topologyList, function(topology) {
    topology.id = topology.url.split('/').pop();
    processTopologies(topology.sub_topologies);
  });
  return topologyList;
}

function setTopology(topologyId) {
  currentTopologyId = topologyId;
  currentTopology = findCurrentTopology(topologies, topologyId);
}

function setDefaultTopologyOptions(topologyList) {
  _.each(topologyList, function(topology) {
    let defaultOptions = makeOrderedMap();
    _.each(topology.options, function(items, option) {
      _.each(items, function(item) {
        if (item.default === true) {
          defaultOptions = defaultOptions.set(option, item.value);
        }
      });
    });

    if (defaultOptions.size) {
      topologyOptions = topologyOptions.set(
        topology.id,
        defaultOptions
      );
    }

    setDefaultTopologyOptions(topology.sub_topologies);
  });
}

// Store API

const AppStore = Object.assign({}, EventEmitter.prototype, {

  CHANGE_EVENT: 'change',

  // keep at the top
  getAppState: function() {
    return {
      topologyId: currentTopologyId,
      selectedNodeId: this.getSelectedNodeId(),
      topologyOptions: topologyOptions.toJS() // all options
    };
  },

  getActiveTopologyOptions: function() {
    // options for current topology
    return topologyOptions.get(currentTopologyId);
  },

  getAdjacentNodes: function(nodeId) {
    adjacentNodes = adjacentNodes.clear();

    if (nodes.has(nodeId)) {
      adjacentNodes = makeSet(nodes.get(nodeId).get('adjacency'));
      // fill up set with reverse edges
      nodes.forEach(function(node, id) {
        if (node.get('adjacency') && node.get('adjacency').includes(nodeId)) {
          adjacentNodes = adjacentNodes.add(id);
        }
      });
    }

    return adjacentNodes;
  },

  getControlError: function() {
    return controlError;
  },

  getCurrentTopology: function() {
    if (!currentTopology) {
      currentTopology = setTopology(currentTopologyId);
    }
    return currentTopology;
  },

  getCurrentTopologyId: function() {
    return currentTopologyId;
  },

  getCurrentTopologyOptions: function() {
    return currentTopology && currentTopology.options;
  },

  getCurrentTopologyUrl: function() {
    return currentTopology && currentTopology.url;
  },

  getErrorUrl: function() {
    return errorUrl;
  },

  getHighlightedEdgeIds: function() {
    if (mouseOverNodeId && nodes.has(mouseOverNodeId)) {
      // all neighbour combinations because we dont know which direction exists
      const adjacency = nodes.get(mouseOverNodeId).get('adjacency');
      if (adjacency) {
        return _.flatten(
          adjacency.forEach(function(nodeId) {
            return [
              [nodeId, mouseOverNodeId].join(Naming.EDGE_ID_SEPARATOR),
              [mouseOverNodeId, nodeId].join(Naming.EDGE_ID_SEPARATOR)
            ];
          })
        );
      }
    }
    if (mouseOverEdgeId) {
      return mouseOverEdgeId;
    }
    return null;
  },

  getHighlightedNodeIds: function() {
    if (mouseOverNodeId) {
      const adjacency = this.getAdjacentNodes(mouseOverNodeId);
      if (adjacency.size) {
        return _.union(adjacency.toJS(), [mouseOverNodeId]);
      }
    }
    if (mouseOverEdgeId) {
      return mouseOverEdgeId.split(Naming.EDGE_ID_SEPARATOR);
    }
    return null;
  },

  getNodeDetails: function() {
    return nodeDetails;
  },

  getNodes: function() {
    return nodes;
  },

  getSelectedNodeId: function() {
    return selectedNodeId;
  },

  getTopologies: function() {
    return topologies;
  },

  getVersion: function() {
    return version;
  },

  isControlPending: function() {
    return controlPending;
  },

  isRouteSet: function() {
    return routeSet;
  },

  isTopologiesLoaded: function() {
    return topologiesLoaded;
  },

  isTopologyEmpty: function() {
    return currentTopology && currentTopology.stats && currentTopology.stats.node_count === 0 && nodes.size === 0;
  },

  isWebsocketClosed: function() {
    return websocketClosed;
  }

});

// Store Dispatch Hooks

AppStore.registeredCallback = function(payload) {
  switch (payload.type) {

  case ActionTypes.CHANGE_TOPOLOGY_OPTION:
    if (topologyOptions.getIn([payload.topologyId, payload.option])
      !== payload.value) {
      nodes = nodes.clear();
    }
    topologyOptions = topologyOptions.setIn(
      [payload.topologyId, payload.option],
      payload.value
    );
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.CLEAR_CONTROL_ERROR:
    controlError = null;
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.CLICK_CLOSE_DETAILS:
    selectedNodeId = null;
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.CLICK_NODE:
    if (payload.nodeId === selectedNodeId) {
      // clicking same node twice unsets the selection
      selectedNodeId = null;
    } else {
      selectedNodeId = payload.nodeId;
    }
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.CLICK_TOPOLOGY:
    selectedNodeId = null;
    if (payload.topologyId !== currentTopologyId) {
      setTopology(payload.topologyId);
      nodes = nodes.clear();
    }
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.CLOSE_WEBSOCKET:
    websocketClosed = true;
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.DO_CONTROL:
    controlPending = true;
    controlError = null;
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.ENTER_EDGE:
    mouseOverEdgeId = payload.edgeId;
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.ENTER_NODE:
    mouseOverNodeId = payload.nodeId;
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.HIT_ESC_KEY:
    nodeDetails = null;
    selectedNodeId = null;
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.LEAVE_EDGE:
    mouseOverEdgeId = null;
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.LEAVE_NODE:
    mouseOverNodeId = null;
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.OPEN_WEBSOCKET:
    // flush nodes cache after re-connect
    nodes = nodes.clear();
    websocketClosed = false;

    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.DO_CONTROL_ERROR:
    controlPending = false;
    controlError = payload.error;
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.DO_CONTROL_SUCCESS:
    controlPending = false;
    controlError = null;
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.RECEIVE_ERROR:
    errorUrl = payload.errorUrl;
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.RECEIVE_NODE_DETAILS:
    errorUrl = null;
    nodeDetails = payload.details;
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.RECEIVE_NODES_DELTA:
    const emptyMessage = !payload.delta.add && !payload.delta.remove
      && payload.delta.update;

    if (!emptyMessage) {
      debug('RECEIVE_NODES_DELTA',
        'remove', _.size(payload.delta.remove),
        'update', _.size(payload.delta.update),
        'add', _.size(payload.delta.add));
    }

    errorUrl = null;

    // nodes that no longer exist
    _.each(payload.delta.remove, function(nodeId) {
      // in case node disappears before mouseleave event
      if (mouseOverNodeId === nodeId) {
        mouseOverNodeId = null;
      }
      if (nodes.has(nodeId) && _.contains(mouseOverEdgeId, nodeId)) {
        mouseOverEdgeId = null;
      }
      nodes = nodes.delete(nodeId);
    });

    // update existing nodes
    _.each(payload.delta.update, function(node) {
      nodes = nodes.set(node.id, nodes.get(node.id).merge(makeNode(node)));
    });

    // add new nodes
    _.each(payload.delta.add, function(node) {
      nodes = nodes.set(node.id, Immutable.fromJS(makeNode(node)));
    });

    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.RECEIVE_TOPOLOGIES:
    errorUrl = null;
    topologies = processTopologies(payload.topologies);
    setTopology(currentTopologyId);
    // only set on first load, if options are not already set via route
    if (!topologiesLoaded && topologyOptions.size === 0) {
      setDefaultTopologyOptions(topologies);
    }
    topologiesLoaded = true;
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.RECEIVE_API_DETAILS:
    errorUrl = null;
    version = payload.version;
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  case ActionTypes.ROUTE_TOPOLOGY:
    routeSet = true;
    if (currentTopologyId !== payload.state.topologyId) {
      nodes = nodes.clear();
    }
    setTopology(payload.state.topologyId);
    setDefaultTopologyOptions(topologies);
    selectedNodeId = payload.state.selectedNodeId;
    topologyOptions = Immutable.fromJS(payload.state.topologyOptions)
      || topologyOptions;
    AppStore.emit(AppStore.CHANGE_EVENT);
    break;

  default:
    break;

  }
};

AppStore.dispatchToken = AppDispatcher.register(AppStore.registeredCallback);

module.exports = AppStore;

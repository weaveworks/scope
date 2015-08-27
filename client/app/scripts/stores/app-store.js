const EventEmitter = require('events').EventEmitter;
const _ = require('lodash');
const assign = require('object-assign');
const debug = require('debug')('scope:app-store');
const Immutable = require('immutable');

const AppDispatcher = require('../dispatcher/app-dispatcher');
const ActionTypes = require('../constants/action-types');
const Naming = require('../constants/naming');

const makeOrderedMap = Immutable.OrderedMap;

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

let currentTopologyId = 'containers';
let errorUrl = null;
let version = '';
let mouseOverEdgeId = null;
let mouseOverNodeId = null;
let nodes = makeOrderedMap();
let nodeDetails = null;
let selectedNodeId = null;
let topologies = [];
let websocketClosed = true;

// Store API

const AppStore = assign({}, EventEmitter.prototype, {

  CHANGE_EVENT: 'change',

  getAppState: function() {
    return {
      topologyId: currentTopologyId,
      selectedNodeId: this.getSelectedNodeId()
    };
  },

  getCurrentTopology: function() {
    return findCurrentTopology(topologies, currentTopologyId);
  },

  getCurrentTopologyId: function() {
    return currentTopologyId;
  },

  getCurrentTopologyUrl: function() {
    const topology = this.getCurrentTopology();

    if (topology) {
      return topology.url;
    }
  },

  getErrorUrl: function() {
    return errorUrl;
  },

  getHighlightedEdgeIds: function() {
    if (mouseOverNodeId) {
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
      const adjacency = nodes.get(mouseOverNodeId).get('adjacency');
      if (adjacency) {
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

  getTopologyIdForUrl: function(url) {
    return url.split('/').pop();
  },

  getVersion: function() {
    return version;
  },

  isWebsocketClosed: function() {
    return websocketClosed;
  }

});

// Store Dispatch Hooks

AppStore.registeredCallback = function(payload) {
  switch (payload.type) {

    case ActionTypes.CLICK_CLOSE_DETAILS:
      selectedNodeId = null;
      AppStore.emit(AppStore.CHANGE_EVENT);
      break;

    case ActionTypes.CLICK_NODE:
      selectedNodeId = payload.nodeId;
      AppStore.emit(AppStore.CHANGE_EVENT);
      break;

    case ActionTypes.CLICK_TOPOLOGY:
      selectedNodeId = null;
      if (payload.topologyId !== currentTopologyId) {
        currentTopologyId = payload.topologyId;
        nodes = nodes.clear();
      }
      AppStore.emit(AppStore.CHANGE_EVENT);
      break;

    case ActionTypes.CLOSE_WEBSOCKET:
      nodes = nodes.clear();
      websocketClosed = true;
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
      nodes = {};
      websocketClosed = false;

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
      debug('RECEIVE_NODES_DELTA',
        'remove', _.size(payload.delta.remove),
        'update', _.size(payload.delta.update),
        'add', _.size(payload.delta.add));

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
        nodes = nodes.set(node.id, nodes.get(node.id).mergeDeep(makeNode(node)));
      });

      // add new nodes
      _.each(payload.delta.add, function(node) {
        nodes = nodes.set(node.id, Immutable.fromJS(makeNode(node)));
      });

      AppStore.emit(AppStore.CHANGE_EVENT);
      break;

    case ActionTypes.RECEIVE_TOPOLOGIES:
      errorUrl = null;
      topologies = payload.topologies;
      AppStore.emit(AppStore.CHANGE_EVENT);
      break;

    case ActionTypes.RECEIVE_API_DETAILS:
      errorUrl = null;
      version = payload.version;
      AppStore.emit(AppStore.CHANGE_EVENT);
      break;

    case ActionTypes.ROUTE_TOPOLOGY:
      if (currentTopologyId !== payload.state.topologyId) {
        nodes = nodes.clear();
      }
      currentTopologyId = payload.state.topologyId;
      selectedNodeId = payload.state.selectedNodeId;
      AppStore.emit(AppStore.CHANGE_EVENT);
      break;

    default:
      break;

  }
};

AppStore.dispatchToken = AppDispatcher.register(AppStore.registeredCallback);

module.exports = AppStore;

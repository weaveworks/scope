const EventEmitter = require('events').EventEmitter;
const _ = require('lodash');
const assign = require('object-assign');

const AppDispatcher = require('../dispatcher/app-dispatcher');
const ActionTypes = require('../constants/action-types');

// Helpers

function isUrlForTopologyId(url, topologyId) {
  return _.endsWith(url, topologyId);
}

// Initial values

let connectionState = 'disconnected';
let currentGrouping = 'none';
let currentTopologyId = 'applications';
let mouseOverNode = null;
let nodes = {};
let nodeDetails = null;
let selectedNodeId = null;
let topologies = [];

// Store API

const AppStore = assign({}, EventEmitter.prototype, {

  CHANGE_EVENT: 'change',

  getAppState: function() {
    return {
      topologyId: currentTopologyId,
      grouping: this.getCurrentGrouping(),
      selectedNodeId: this.getSelectedNodeId()
    };
  },

  getConnectionState: function() {
    return connectionState;
  },

  getCurrentTopology: function() {
    return _.find(topologies, function(topology) {
      return isUrlForTopologyId(topology.url, currentTopologyId);
    });
  },

  getCurrentTopologyUrl: function() {
    const topology = this.getCurrentTopology();

    if (topology) {
      return topology.grouped_url && currentGrouping === 'grouped' ? topology.grouped_url : topology.url;
    }
  },

  getCurrentGrouping: function() {
    return currentGrouping;
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
  }
});

// Store Dispatch Hooks

AppStore.registeredCallback = function(payload) {
  switch (payload.type) {

    case ActionTypes.CLICK_CLOSE_DETAILS:
      selectedNodeId = null;
      AppStore.emit(AppStore.CHANGE_EVENT);
      break;

    case ActionTypes.CLICK_GROUPING:
      if (payload.grouping !== currentGrouping) {
        currentGrouping = payload.grouping;
        nodes = {};
        AppStore.emit(AppStore.CHANGE_EVENT);
      }
      break;

    case ActionTypes.CLICK_NODE:
      selectedNodeId = payload.nodeId;
      AppStore.emit(AppStore.CHANGE_EVENT);
      break;

    case ActionTypes.CLICK_TOPOLOGY:
      if (payload.topologyId !== currentTopologyId) {
        currentTopologyId = payload.topologyId;
        nodes = {};
      }
      AppStore.emit(AppStore.CHANGE_EVENT);
      break;

    case ActionTypes.ENTER_NODE:
      mouseOverNode = payload.nodeId;
      AppStore.emit(AppStore.CHANGE_EVENT);
      break;

    case ActionTypes.HIT_ESC_KEY:
      nodeDetails = null;
      selectedNodeId = null;
      AppStore.emit(AppStore.CHANGE_EVENT);
      break;

    case ActionTypes.LEAVE_NODE:
      mouseOverNode = null;
      AppStore.emit(AppStore.CHANGE_EVENT);
      break;

    case ActionTypes.RECEIVE_NODE_DETAILS:
      nodeDetails = payload.details;
      AppStore.emit(AppStore.CHANGE_EVENT);
      break;

    case ActionTypes.RECEIVE_NODES_DELTA:
      console.log('RECEIVE_NODES_DELTA',
        'remove', _.size(payload.delta.remove),
        'update', _.size(payload.delta.update),
        'add', _.size(payload.delta.add));

      connectionState = 'connected';

      // nodes that no longer exist
      _.each(payload.delta.remove, function(nodeId) {
        // in case node disappears before mouseleave event
        if (mouseOverNode === nodeId) {
          mouseOverNode = null;
        }
        delete nodes[nodeId];
      });

      // update existing nodes
      _.each(payload.delta.update, function(node) {
        nodes[node.id] = node;
      });

      // add new nodes
      _.each(payload.delta.add, function(node) {
        nodes[node.id] = node;
      });

      AppStore.emit(AppStore.CHANGE_EVENT);
      break;

    case ActionTypes.RECEIVE_TOPOLOGIES:
      topologies = payload.topologies;
      AppStore.emit(AppStore.CHANGE_EVENT);
      break;

    case ActionTypes.ROUTE_TOPOLOGY:
      nodes = {};
      currentTopologyId = payload.state.topologyId;
      currentGrouping = payload.state.grouping;
      selectedNodeId = payload.state.selectedNodeId;
      AppStore.emit(AppStore.CHANGE_EVENT);
      break;

    default:
      break;

  }
};

AppStore.dispatchToken = AppDispatcher.register(AppStore.registeredCallback);

module.exports = AppStore;

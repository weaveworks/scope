let ActionTypes;
let AppDispatcher;
let AppStore;
let RouterUtils;
let WebapiUtils;

module.exports = {
  clickCloseDetails: function() {
    AppDispatcher.dispatch({
      type: ActionTypes.CLICK_CLOSE_DETAILS
    });
    RouterUtils.updateRoute();
  },

  clickNode: function(nodeId) {
    AppDispatcher.dispatch({
      type: ActionTypes.CLICK_NODE,
      nodeId: nodeId
    });
    RouterUtils.updateRoute();
    WebapiUtils.getNodeDetails(AppStore.getCurrentTopologyUrl(), AppStore.getSelectedNodeId());
  },

  clickTopology: function(topologyId) {
    AppDispatcher.dispatch({
      type: ActionTypes.CLICK_TOPOLOGY,
      topologyId: topologyId
    });
    RouterUtils.updateRoute();
    WebapiUtils.getNodesDelta(AppStore.getCurrentTopologyUrl());
  },

  closeWebsocket: function() {
    AppDispatcher.dispatch({
      type: ActionTypes.CLOSE_WEBSOCKET
    });
  },

  enterEdge: function(edgeId) {
    AppDispatcher.dispatch({
      type: ActionTypes.ENTER_EDGE,
      edgeId: edgeId
    });
  },

  enterNode: function(nodeId) {
    AppDispatcher.dispatch({
      type: ActionTypes.ENTER_NODE,
      nodeId: nodeId
    });
  },

  hitEsc: function() {
    AppDispatcher.dispatch({
      type: ActionTypes.HIT_ESC_KEY
    });
    RouterUtils.updateRoute();
  },

  leaveEdge: function(edgeId) {
    AppDispatcher.dispatch({
      type: ActionTypes.LEAVE_EDGE,
      edgeId: edgeId
    });
  },

  leaveNode: function(nodeId) {
    AppDispatcher.dispatch({
      type: ActionTypes.LEAVE_NODE,
      nodeId: nodeId
    });
  },

  receiveNodeDetails: function(details) {
    AppDispatcher.dispatch({
      type: ActionTypes.RECEIVE_NODE_DETAILS,
      details: details
    });
  },

  receiveNodesDelta: function(delta) {
    AppDispatcher.dispatch({
      type: ActionTypes.RECEIVE_NODES_DELTA,
      delta: delta
    });
  },

  receiveTopologies: function(topologies) {
    AppDispatcher.dispatch({
      type: ActionTypes.RECEIVE_TOPOLOGIES,
      topologies: topologies
    });
    WebapiUtils.getNodesDelta(AppStore.getCurrentTopologyUrl());
    WebapiUtils.getNodeDetails(AppStore.getCurrentTopologyUrl(), AppStore.getSelectedNodeId());
  },

  receiveApiDetails: function(apiDetails) {
    AppDispatcher.dispatch({
        type: ActionTypes.RECEIVE_API_DETAILS,
        version: apiDetails.version
    });
  },

  receiveError: function(errorUrl) {
    AppDispatcher.dispatch({
        errorUrl: errorUrl,
        type: ActionTypes.RECEIVE_ERROR
    });
  },

  route: function(state) {
    AppDispatcher.dispatch({
      state: state,
      type: ActionTypes.ROUTE_TOPOLOGY
    });
    WebapiUtils.getNodesDelta(AppStore.getCurrentTopologyUrl());
    WebapiUtils.getNodeDetails(AppStore.getCurrentTopologyUrl(), AppStore.getSelectedNodeId());
  }
};

// require below export to break circular dep

AppDispatcher = require('../dispatcher/app-dispatcher');
ActionTypes = require('../constants/action-types');

RouterUtils = require('../utils/router-utils');
WebapiUtils = require('../utils/web-api-utils');
AppStore = require('../stores/app-store');

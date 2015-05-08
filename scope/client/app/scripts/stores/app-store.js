
var EventEmitter = require('events').EventEmitter;
var _ = require('lodash');
var assign = require('object-assign');

var AppDispatcher = require('../dispatcher/app-dispatcher');
var ActionTypes = require('../constants/action-types');
var TopologyStore = require('./topology-store');
// var topologies = require('../constants/topologies');


// Initial values

var currentTopology = 'processname';
var currentTopologyMode = 'individual';
var nodeDetails = null;
var selectedNodeId = null;
var topologies = [];

// Store API

var AppStore = assign({}, EventEmitter.prototype, {

	CHANGE_EVENT: 'change',

	getAppState: function() {
		return {
			currentTopology: this.getCurrentTopology(),
			currentTopologyMode: this.getCurrentTopologyMode(),
			selectedNodeId: this.getSelectedNodeId()
		};
	},

	getCurrentTopology: function() {
		return currentTopology;
	},

	getCurrentTopologyMode: function() {
		return currentTopologyMode;
	},

	getNodeDetails: function() {
		return nodeDetails;
	},

	getSelectedNodeId: function() {
		return selectedNodeId;
	},

	getTopologies: function() {
		return topologies;
	},

	getTopologyForUrl: function(url) {
		return url.split('/').pop();
	},

	getUrlForTopology: function(topologyId) {
		var topology =  _.find(topologies, function(topology) {
			return this.isUrlForTopology(topology.url, topologyId);
		}, this);

		if (topology) {
			return topology.grouped_url && currentTopologyMode == 'class' ? topology.grouped_url : topology.url;
		}
	},

	isUrlForTopology: function(url, topologyId) {
		return _.endsWith(url, topologyId);
	}

});


// Store Dispatch Hooks

AppStore.dispatchToken = AppDispatcher.register(function(payload) {
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
			currentTopology = payload.topologyId;
			AppDispatcher.waitFor([TopologyStore.dispatchToken]);
			AppStore.emit(AppStore.CHANGE_EVENT);
			break;

		case ActionTypes.CLICK_TOPOLOGY_MODE:
			currentTopologyMode = payload.mode;
			AppDispatcher.waitFor([TopologyStore.dispatchToken]);
			AppStore.emit(AppStore.CHANGE_EVENT);
			break;

		case ActionTypes.HIT_ESC_KEY:
			nodeDetails = null;
			selectedNodeId = null;
			AppStore.emit(AppStore.CHANGE_EVENT);
			break;

		case ActionTypes.RECEIVE_NODE_DETAILS:
			nodeDetails = payload.details;
			AppStore.emit(AppStore.CHANGE_EVENT);
			break;

		case ActionTypes.RECEIVE_TOPOLOGIES:
			topologies = payload.topologies;
			AppStore.emit(AppStore.CHANGE_EVENT);
			break;

		case ActionTypes.ROUTE_TOPOLOGY:
			currentTopology = payload.state.currentTopology;
			currentTopologyMode = payload.state.currentTopologyMode;
			selectedNodeId = payload.state.selectedNodeId;
			AppDispatcher.waitFor([TopologyStore.dispatchToken]);
			AppStore.emit(AppStore.CHANGE_EVENT);
			break;

		default:
			break;

	}
});

module.exports = AppStore;
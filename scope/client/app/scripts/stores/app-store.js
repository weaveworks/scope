
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
var detailsView = false;
var explorerExpandedNodes = [];
var nodeDetails = null;
var topologies = [];

// Store API

var AppStore = assign({}, EventEmitter.prototype, {

	CHANGE_EVENT: 'change',

	getAppState: function() {
		return {
			currentTopology: this.getCurrentTopology(),
			currentTopologyMode: this.getCurrentTopologyMode(),
			detailsView: this.getDetailsView(),
			explorerExpandedNodes: this.getExplorerExpandedNodes()
		};
	},

	getCurrentTopology: function() {
		return currentTopology;
	},

	getCurrentTopologyMode: function() {
		return currentTopologyMode;
	},

	getDetailsView: function() {
		return detailsView;
	},

	getExplorerExpandedNodes: function() {
		return explorerExpandedNodes;
	},

	getNodeDetails: function() {
		return nodeDetails;
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
			detailsView = false;
			explorerExpandedNodes = [];
			AppStore.emit(AppStore.CHANGE_EVENT);
			break;

		case ActionTypes.CLICK_NODE:
			detailsView = 'explorer';
			if (!_.contains(explorerExpandedNodes, payload.nodeId)) {
				explorerExpandedNodes.push(payload.nodeId);
			} else {
				explorerExpandedNodes = _.without(explorerExpandedNodes, payload.nodeId);
			}
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
			detailsView = false;
			nodeDetails = null;
			explorerExpandedNodes = [];
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
			detailsView = payload.state.detailsView;
			explorerExpandedNodes = payload.state.explorerExpandedNodes;
			currentView = payload.state.currentView;
			AppDispatcher.waitFor([TopologyStore.dispatchToken]);
			AppStore.emit(AppStore.CHANGE_EVENT);
			break;

		default:
			break;

	}
});

module.exports = AppStore;
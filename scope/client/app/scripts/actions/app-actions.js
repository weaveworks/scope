var AppDispatcher = require('../dispatcher/app-dispatcher');
var ActionTypes = require('../constants/action-types');

module.exports = {
	clickNode: function(nodeId) {
		AppDispatcher.dispatch({
			type: ActionTypes.CLICK_NODE,
			nodeId: nodeId
		});
		RouterUtils.updateRoute();
	},

	clickTopology: function(topologyId) {
		AppDispatcher.dispatch({
			type: ActionTypes.CLICK_TOPOLOGY,
			topologyId: topologyId
		});
		RouterUtils.updateRoute();
		WebapiUtils.getNodesDelta(AppStore.getUrlForTopology(AppStore.getCurrentTopology()));
	},

	clickTopologyMode: function(mode) {
		AppDispatcher.dispatch({
			type: ActionTypes.CLICK_TOPOLOGY_MODE,
			mode: mode
		});
		RouterUtils.updateRoute();
		WebapiUtils.getNodesDelta(AppStore.getUrlForTopology(AppStore.getCurrentTopology()));
	},

	clickView: function(viewId) {
		AppDispatcher.dispatch({
			type: ActionTypes.CLICK_VIEW,
			viewId: viewId
		});
		RouterUtils.updateRoute();
	},

	hitEsc: function() {
		AppDispatcher.dispatch({
			type: ActionTypes.HIT_ESC_KEY
		});
		RouterUtils.updateRoute();
	},

	receiveNodeDetails: function(details) {
		AppDispatcher.dispatch({
			type: ActionTypes.RECEIVE_NODE_DETAILS,
			details: details
		});
	},

	receiveTopologies: function(topologies) {
		AppDispatcher.dispatch({
			type: ActionTypes.RECEIVE_TOPOLOGIES,
			topologies: topologies
		});
		WebapiUtils.getNodesDelta(AppStore.getUrlForTopology(AppStore.getCurrentTopology()));
	},

	route: function(state) {
		AppDispatcher.dispatch({
			state: state,
			type: ActionTypes.ROUTE_TOPOLOGY
		});
		WebapiUtils.getNodesDelta(AppStore.getUrlForTopology(AppStore.getCurrentTopology()));
	}
};

// breaking circular deps

var RouterUtils = require('../utils/router-utils');
var WebapiUtils = require('../utils/web-api-utils');
var AppStore = require('../stores/app-store');
var AppDispatcher = require('../dispatcher/app-dispatcher');
var ActionTypes = require('../constants/action-types');

module.exports = {
	clickCloseDetails: function() {
		AppDispatcher.dispatch({
			type: ActionTypes.CLICK_CLOSE_DETAILS
		});
		RouterUtils.updateRoute();
	},

	clickGrouping: function(grouping) {
		AppDispatcher.dispatch({
			type: ActionTypes.CLICK_GROUPING,
			grouping: grouping
		});
		RouterUtils.updateRoute();
		WebapiUtils.getNodesDelta(AppStore.getUrlForTopology(AppStore.getCurrentTopology()));
	},

	clickNode: function(nodeId) {
		AppDispatcher.dispatch({
			type: ActionTypes.CLICK_NODE,
			nodeId: nodeId
		});
		RouterUtils.updateRoute();
		WebapiUtils.getNodeDetails(AppStore.getUrlForTopology(AppStore.getCurrentTopology()), AppStore.getSelectedNodeId());
	},

	clickTopology: function(topologyId) {
		AppDispatcher.dispatch({
			type: ActionTypes.CLICK_TOPOLOGY,
			topologyId: topologyId
		});
		RouterUtils.updateRoute();
		WebapiUtils.getNodesDelta(AppStore.getUrlForTopology(AppStore.getCurrentTopology()));
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
		WebapiUtils.getNodeDetails(AppStore.getUrlForTopology(AppStore.getCurrentTopology()), AppStore.getSelectedNodeId());
	},

	route: function(state) {
		AppDispatcher.dispatch({
			state: state,
			type: ActionTypes.ROUTE_TOPOLOGY
		});
		WebapiUtils.getNodesDelta(AppStore.getUrlForTopology(AppStore.getCurrentTopology()));
		WebapiUtils.getNodeDetails(AppStore.getUrlForTopology(AppStore.getCurrentTopology()), AppStore.getSelectedNodeId());
	}
};

// breaking circular deps

var RouterUtils = require('../utils/router-utils');
var WebapiUtils = require('../utils/web-api-utils');
var AppStore = require('../stores/app-store');
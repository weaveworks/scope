var EventEmitter = require('events').EventEmitter;
var _ = require('lodash');
var assign = require('object-assign');

var AppDispatcher = require('../dispatcher/app-dispatcher');
var ActionTypes = require('../constants/action-types');



// Initial values

var nodes = {};
var mouseOverNode = null;

// Store API

var TopologyStore = assign({}, EventEmitter.prototype, {

	CHANGE_EVENT: 'change',

	getNodes: function() {
		return nodes;
	}

});


// Store Dispatch Hooks

TopologyStore.dispatchToken = AppDispatcher.register(function(payload) {
	switch (payload.type) {
		case ActionTypes.CLICK_GROUPING:
			nodes = {};
			TopologyStore.emit(TopologyStore.CHANGE_EVENT);
			break;

		case ActionTypes.CLICK_TOPOLOGY:
			nodes = {};
			TopologyStore.emit(TopologyStore.CHANGE_EVENT);
			break;

		case ActionTypes.ENTER_NODE:
			mouseOverNode = payload.nodeId;
			TopologyStore.emit(TopologyStore.CHANGE_EVENT);
			break;

		case ActionTypes.LEAVE_NODE:
			mouseOverNode = null;
			TopologyStore.emit(TopologyStore.CHANGE_EVENT);
			break;

		case ActionTypes.RECEIVE_NODES_DELTA:
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

			TopologyStore.emit(TopologyStore.CHANGE_EVENT);
			break;

		case ActionTypes.ROUTE_TOPOLOGY:
			nodes = {};
			TopologyStore.emit(TopologyStore.CHANGE_EVENT);
			break;

		default:
			break;

	}
});

module.exports = TopologyStore;
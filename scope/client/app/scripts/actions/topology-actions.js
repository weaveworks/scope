var AppDispatcher = require('../dispatcher/app-dispatcher');
var ActionTypes = require('../constants/action-types');

module.exports = {
	checkFilterAdjacent: function(value) {
		AppDispatcher.dispatch({
			type: ActionTypes.CHECK_FILTER_ADJACENT,
			value: value
		});
	},

	enterNode: function(nodeId) {
		AppDispatcher.dispatch({
			type: ActionTypes.ENTER_NODE,
			nodeId: nodeId
		});
	},

	inputFilterText: function(text) {
		AppDispatcher.dispatch({
			type: ActionTypes.INPUT_FILTER_TEXT,
			text: text
		});
	},

	leaveNode: function(nodeId) {
		AppDispatcher.dispatch({
			type: ActionTypes.LEAVE_NODE,
			nodeId: nodeId
		});
	},

	receiveNodesDelta: function(delta) {
		AppDispatcher.dispatch({
			type: ActionTypes.RECEIVE_NODES_DELTA,
			delta: delta
		});
	}
};
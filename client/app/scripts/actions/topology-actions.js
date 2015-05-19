var AppDispatcher = require('../dispatcher/app-dispatcher');
var ActionTypes = require('../constants/action-types');

module.exports = {
	enterNode: function(nodeId) {
		AppDispatcher.dispatch({
			type: ActionTypes.ENTER_NODE,
			nodeId: nodeId
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
var Renderers = {
	getNodeHost: function(nodeLabel) {
		return nodeLabel.split('@')[1] ? nodeLabel.split('@')[1].substr(5) : nodeLabel;
	},
	getNodeProc: function(nodeLabel) {
		return nodeLabel.split('@')[0];
	}
};

module.exports = Renderers;
var reqwest = require('reqwest');

var TopologyActions = require('../actions/topology-actions');
var AppActions = require('../actions/app-actions');
var AppStore = require('../stores/app-store');

var WS_URL = window.WS_URL || 'ws://' + location.host;


var socket;
var reconnectTimer = 0;
var currentUrl = null;
var updateFrequency = '5s';
var topologyTimer = 0;

function createWebsocket(topologyUrl) {
		if (socket) {
			socket.onclose = null;
			socket.close();
		}

		socket = new WebSocket(WS_URL + topologyUrl + '/ws?t=' + updateFrequency);

		socket.onclose = function() {
			clearTimeout(reconnectTimer);
			socket = null;

			reconnectTimer = setTimeout(function() {
				createWebsocket(topologyUrl);
			}, 5000);
		}

		socket.onmessage = function(event) {
			var msg = JSON.parse(event.data);
			if (msg.add || msg.remove || msg.update) {
				TopologyActions.receiveNodesDelta(msg);
			}
		};

		currentUrl = topologyUrl;
}

function getTopologies() {
	clearTimeout(topologyTimer);
	reqwest('/api/topology', function(res) {
		AppActions.receiveTopologies(res);
		topologyTimer = setTimeout(getTopologies, 10000);
	});
}

function getNodeDetails(topology, nodeId) {
	var url = [AppStore.getUrlForTopology(topology), nodeId].join('/');
	reqwest(url, function(res) {
		AppActions.receiveNodeDetails(res.node);
	});
}

module.exports = {
	getNodeDetails: getNodeDetails,

	getTopologies: getTopologies,

	getNodesDelta: function(topologyUrl) {
		if (topologyUrl) {
			createWebsocket(topologyUrl);
		}
	}
}

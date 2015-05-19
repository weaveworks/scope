var reqwest = require('reqwest');

var AppActions = require('../actions/app-actions');

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
				AppActions.receiveNodesDelta(msg);
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

function getNodeDetails(topologyUrl, nodeId) {
	if (topologyUrl && nodeId) {
		var url = [topologyUrl, nodeId].join('/');
		reqwest(url, function(res) {
			AppActions.receiveNodeDetails(res.node);
		});
	}
}

module.exports = {
	getNodeDetails: getNodeDetails,

	getTopologies: getTopologies,

	getNodesDelta: function(topologyUrl) {
		if (topologyUrl && topologyUrl !== currentUrl) {
			createWebsocket(topologyUrl);
		}
	}
}

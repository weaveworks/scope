const reqwest = require('reqwest');

const AppActions = require('../actions/app-actions');

const WS_URL = window.WS_URL || 'ws://' + location.host;


let socket;
let reconnectTimer = 0;
let currentUrl = null;
let updateFrequency = '5s';
let topologyTimer = 0;

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
  };

  socket.onmessage = function(event) {
    const msg = JSON.parse(event.data);
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
    const url = [topologyUrl, nodeId].join('/');
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
};


const debug = require('debug')('scope:web-api-utils');
const reqwest = require('reqwest');

const AppActions = require('../actions/app-actions');

const WS_URL = window.WS_URL || 'ws://' + location.host;


const apiTimerInterval = 10000;
const reconnectTimerInterval = 5000;
const topologyTimerInterval = apiTimerInterval;
const updateFrequency = '5s';

let socket;
let reconnectTimer = 0;
let currentUrl = null;
let topologyTimer = 0;
let apiDetailsTimer = 0;

function createWebsocket(topologyUrl) {
  if (socket) {
    socket.onclose = null;
    socket.onerror = null;
    socket.close();
  }

  socket = new WebSocket(WS_URL + topologyUrl + '/ws?t=' + updateFrequency);

  socket.onclose = function() {
    clearTimeout(reconnectTimer);
    socket = null;
    AppActions.closeWebsocket();
    debug('Closed websocket to ' + currentUrl);

    reconnectTimer = setTimeout(function() {
      createWebsocket(topologyUrl);
    }, reconnectTimerInterval);
  };

  socket.onerror = function() {
    debug('Error in websocket to ' + currentUrl);
    AppActions.receiveError(currentUrl);
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
  const url = '/api/topology';
  reqwest({
    url: url,
    success: function(res) {
      AppActions.receiveTopologies(res);
      topologyTimer = setTimeout(getTopologies, topologyTimerInterval);
    },
    error: function(err) {
      debug('Error in topology request: ' + err);
      AppActions.receiveError(url);
      topologyTimer = setTimeout(getTopologies, topologyTimerInterval / 2);
    }
  });
}

function getNodeDetails(topologyUrl, nodeId) {
  if (topologyUrl && nodeId) {
    const url = [topologyUrl, encodeURIComponent(nodeId)].join('/');
    reqwest({
      url: url,
      success: function(res) {
        AppActions.receiveNodeDetails(res.node);
      },
      error: function(err) {
        debug('Error in node details request: ' + err);
        AppActions.receiveError(topologyUrl);
      }
    });
  }
}

function getApiDetails() {
  clearTimeout(apiDetailsTimer);
  const url = '/api';
  reqwest({
    url: url,
    success: function(res) {
      AppActions.receiveApiDetails(res);
      apiDetailsTimer = setTimeout(getApiDetails, apiTimerInterval);
    },
    error: function(err) {
      debug('Error in api details request: ' + err);
      AppActions.receiveError(url);
      apiDetailsTimer = setTimeout(getApiDetails, apiTimerInterval / 2);
    }
  });
}

module.exports = {
  getNodeDetails: getNodeDetails,

  getTopologies: getTopologies,

  getApiDetails: getApiDetails,

  getNodesDelta: function(topologyUrl) {
    if (topologyUrl && topologyUrl !== currentUrl) {
      createWebsocket(topologyUrl);
    }
  }
};


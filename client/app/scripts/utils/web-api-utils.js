
const debug = require('debug')('scope:web-api-utils');
const reqwest = require('reqwest');

const AppActions = require('../actions/app-actions');

const wsProto = location.protocol === 'https:' ? 'wss' : 'ws';
const wsUrl = __WS_URL__ || wsProto + '://' + location.host + location.pathname.replace(/\/$/, '');

const apiTimerInterval = 10000;
const reconnectTimerInterval = 5000;
const topologyTimerInterval = apiTimerInterval;
const updateFrequency = '5s';

let socket;
let reconnectTimer = 0;
let currentUrl = null;
let currentOptions = null;
let topologyTimer = 0;
let apiDetailsTimer = 0;
let controlErrorTimer = 0;

function buildOptionsQuery(options) {
  if (options) {
    return options.reduce(function(query, value, param) {
      return `${query}&${param}=${value}`;
    }, '');
  }
  return '';
}

function createWebsocket(topologyUrl, optionsQuery) {
  if (socket) {
    socket.onclose = null;
    socket.onerror = null;
    socket.close();
  }

  socket = new WebSocket(wsUrl + topologyUrl
    + '/ws?t=' + updateFrequency + '&' + optionsQuery);

  socket.onopen = function() {
    AppActions.openWebsocket();
  };

  socket.onclose = function() {
    clearTimeout(reconnectTimer);
    socket = null;
    AppActions.closeWebsocket();
    debug('Closed websocket to ' + topologyUrl);

    reconnectTimer = setTimeout(function() {
      createWebsocket(topologyUrl, optionsQuery);
    }, reconnectTimerInterval);
  };

  socket.onerror = function() {
    debug('Error in websocket to ' + topologyUrl);
    AppActions.receiveError(currentUrl);
  };

  socket.onmessage = function(event) {
    const msg = JSON.parse(event.data);
    AppActions.receiveNodesDelta(msg);
  };
}

/* keep URLs relative */

function getTopologies(options) {
  clearTimeout(topologyTimer);
  const optionsQuery = buildOptionsQuery(options);
  const url = `api/topology?${optionsQuery}`;
  reqwest({
    url: url,
    success: function(res) {
      AppActions.receiveTopologies(res);
      topologyTimer = setTimeout(function() {
        getTopologies(options);
      }, topologyTimerInterval / 2);
    },
    error: function(err) {
      debug('Error in topology request: ' + err);
      AppActions.receiveError(url);
      topologyTimer = setTimeout(function() {
        getTopologies(options);
      }, topologyTimerInterval / 2);
    }
  });
}

function getTopology(topologyUrl, options) {
  const optionsQuery = buildOptionsQuery(options);

  // only recreate websocket if url changed
  if (topologyUrl && (topologyUrl !== currentUrl || currentOptions !== optionsQuery)) {
    createWebsocket(topologyUrl, optionsQuery);
    currentUrl = topologyUrl;
    currentOptions = optionsQuery;
  }
}

function getNodeDetails(topologyUrl, nodeId, options) {
  const optionsQuery = buildOptionsQuery(options);

  if (topologyUrl && nodeId) {
    const url = [topologyUrl, '/', encodeURIComponent(nodeId), '?', optionsQuery]
      .join('').substr(1);
    reqwest({
      url: url,
      success: function(res) {
        AppActions.receiveNodeDetails(res.node);
      },
      error: function(err) {
        debug('Error in node details request: ' + err.responseText);
        // dont treat missing node as error
        if (err.status !== 404) {
          AppActions.receiveError(topologyUrl);
        }
      }
    });
  }
}

function getApiDetails() {
  clearTimeout(apiDetailsTimer);
  const url = 'api';
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

function doControl(probeId, nodeId, control) {
  clearTimeout(controlErrorTimer);
  const url = `api/control/${encodeURIComponent(probeId)}/`
    + `${encodeURIComponent(nodeId)}/${control}`;
  reqwest({
    method: 'POST',
    url: url,
    success: function() {
      AppActions.receiveControlSuccess();
    },
    error: function(err) {
      AppActions.receiveControlError(err.response);
      controlErrorTimer = setTimeout(function() {
        AppActions.clearControlError();
      }, 10000);
    }
  });
}

module.exports = {
  doControl: doControl,

  getNodeDetails: getNodeDetails,

  getTopologies: getTopologies,

  getApiDetails: getApiDetails,

  getNodesDelta: getTopology
};

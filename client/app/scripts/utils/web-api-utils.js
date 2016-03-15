import debug from 'debug';
import reqwest from 'reqwest';

import { clearControlError, closeWebsocket, openWebsocket, receiveError,
  receiveApiDetails, receiveNodesDelta, receiveNodeDetails, receiveControlError,
  receiveControlPipe, receiveControlPipeStatus, receiveControlSuccess,
  receiveTopologies, receiveNotFound } from '../actions/app-actions';

import { API_INTERVAL, TOPOLOGY_INTERVAL } from '../constants/timer';

const log = debug('scope:web-api-utils');

const reconnectTimerInterval = 5000;
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

export function basePath(urlPath) {
  //
  // "/scope/terminal.html" -> "/scope"
  // "/scope/" -> "/scope"
  // "/scope" -> "/scope"
  // "/" -> ""
  //
  const parts = urlPath.split('/');
  // if the last item has a "." in it, e.g. foo.html...
  if (parts[parts.length - 1].indexOf('.') !== -1) {
    return parts.slice(0, -1).join('/');
  }
  return parts.join('/').replace(/\/$/, '');
}

export function basePathSlash(urlPath) {
  //
  // "/scope/terminal.html" -> "/scope/"
  // "/scope/" -> "/scope/"
  // "/scope" -> "/scope/"
  // "/" -> "/"
  //
  return basePath(urlPath) + '/';
}

const wsProto = location.protocol === 'https:' ? 'wss' : 'ws';
const wsUrl = wsProto + '://' + location.host + basePath(location.pathname);

function createWebsocket(topologyUrl, optionsQuery) {
  if (socket) {
    socket.onclose = null;
    socket.onerror = null;
    socket.close();
    // onclose() is not called, but that's fine since we're opening a new one
    // right away
  }

  socket = new WebSocket(wsUrl + topologyUrl
    + '/ws?t=' + updateFrequency + '&' + optionsQuery);

  socket.onopen = function() {
    openWebsocket();
  };

  socket.onclose = function() {
    clearTimeout(reconnectTimer);
    log('Closing websocket to ' + topologyUrl, socket.readyState);
    socket = null;
    closeWebsocket();

    reconnectTimer = setTimeout(function() {
      createWebsocket(topologyUrl, optionsQuery);
    }, reconnectTimerInterval);
  };

  socket.onerror = function() {
    log('Error in websocket to ' + topologyUrl);
    receiveError(currentUrl);
  };

  socket.onmessage = function(event) {
    const msg = JSON.parse(event.data);
    receiveNodesDelta(msg);
  };
}

/* keep URLs relative */

export function getTopologies(options) {
  clearTimeout(topologyTimer);
  const optionsQuery = buildOptionsQuery(options);
  const url = `api/topology?${optionsQuery}`;
  reqwest({
    url: url,
    success: function(res) {
      receiveTopologies(res);
      topologyTimer = setTimeout(function() {
        getTopologies(options);
      }, TOPOLOGY_INTERVAL);
    },
    error: function(err) {
      log('Error in topology request: ' + err.responseText);
      receiveError(url);
      topologyTimer = setTimeout(function() {
        getTopologies(options);
      }, TOPOLOGY_INTERVAL / 2);
    }
  });
}

export function getNodesDelta(topologyUrl, options) {
  const optionsQuery = buildOptionsQuery(options);

  // only recreate websocket if url changed
  if (topologyUrl && (topologyUrl !== currentUrl || currentOptions !== optionsQuery)) {
    createWebsocket(topologyUrl, optionsQuery);
    currentUrl = topologyUrl;
    currentOptions = optionsQuery;
  }
}

export function getNodeDetails(topologyUrlsById, nodeMap) {
  // get details for all opened nodes
  const obj = nodeMap.last();
  if (obj && topologyUrlsById.has(obj.topologyId)) {
    const topologyUrl = topologyUrlsById.get(obj.topologyId);
    const url = [topologyUrl, '/', encodeURIComponent(obj.id)]
      .join('').substr(1);
    reqwest({
      url: url,
      success: function(res) {
        // make sure node is still selected
        if (nodeMap.has(res.node.id)) {
          receiveNodeDetails(res.node);
        }
      },
      error: function(err) {
        log('Error in node details request: ' + err.responseText);
        // dont treat missing node as error
        if (err.status === 404) {
          receiveNotFound(obj.id);
        } else {
          receiveError(topologyUrl);
        }
      }
    });
  } else {
    log('No details or url found for ', obj);
  }
}

export function getApiDetails() {
  clearTimeout(apiDetailsTimer);
  const url = 'api';
  reqwest({
    url: url,
    success: function(res) {
      receiveApiDetails(res);
      apiDetailsTimer = setTimeout(getApiDetails, API_INTERVAL);
    },
    error: function(err) {
      log('Error in api details request: ' + err.responseText);
      receiveError(url);
      apiDetailsTimer = setTimeout(getApiDetails, API_INTERVAL / 2);
    }
  });
}

export function doControlRequest(nodeId, control) {
  clearTimeout(controlErrorTimer);
  const url = `api/control/${encodeURIComponent(control.probeId)}/`
    + `${encodeURIComponent(control.nodeId)}/${control.id}`;
  reqwest({
    method: 'POST',
    url: url,
    success: function(res) {
      receiveControlSuccess(nodeId);
      if (res && res.pipe) {
        receiveControlPipe(res.pipe, nodeId, res.raw_tty, true);
      }
    },
    error: function(err) {
      receiveControlError(nodeId, err.response);
      controlErrorTimer = setTimeout(function() {
        clearControlError(nodeId);
      }, 10000);
    }
  });
}

export function deletePipe(pipeId) {
  const url = `api/pipe/${encodeURIComponent(pipeId)}`;
  reqwest({
    method: 'DELETE',
    url: url,
    success: function() {
      log('Closed the pipe!');
    },
    error: function(err) {
      log('Error closing pipe:' + err);
      receiveError(url);
    }
  });
}

export function getPipeStatus(pipeId) {
  const url = `api/pipe/${encodeURIComponent(pipeId)}/check`;
  reqwest({
    method: 'GET',
    url: url,
    error: function(err) {
      log('ERROR: unexpected response:', err);
    },
    success: function(res) {
      const status = {
        204: 'PIPE_ALIVE',
        404: 'PIPE_DELETED'
      }[res.status];

      if (!status) {
        log('Unexpected pipe status:', res.status);
        return;
      }

      receiveControlPipeStatus(pipeId, status);
    }
  });
}

import debug from 'debug';
import reqwest from 'reqwest';
import trimStart from 'lodash/trimStart';

import { blurSearch, clearControlError, closeWebsocket, openWebsocket, receiveError,
  receiveApiDetails, receiveNodesDelta, receiveNodeDetails, receiveControlError,
  receiveControlNodeRemoved, receiveControlPipe, receiveControlPipeStatus,
  receiveControlSuccess, receiveTopologies, receiveNotFound,
  receiveNodesForTopology } from '../actions/app-actions';

import { API_INTERVAL, TOPOLOGY_INTERVAL } from '../constants/timer';

const log = debug('scope:web-api-utils');

const reconnectTimerInterval = 5000;
const updateFrequency = '5s';
const FIRST_RENDER_TOO_LONG_THRESHOLD = 100; // ms

let socket;
let reconnectTimer = 0;
let currentUrl = null;
let currentOptions = null;
let topologyTimer = 0;
let apiDetailsTimer = 0;
let controlErrorTimer = 0;
let createWebsocketAt = 0;
let firstMessageOnWebsocketAt = 0;

export function buildOptionsQuery(options) {
  if (options) {
    return options.map((value, param) => `${param}=${value}`).join('&');
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
  return `${basePath(urlPath)}/`;
}

// JJP - `apiPath` is used to get API URLs right when running as a React component.
// This needs to be refactored to just accept a URL prop on the scope component.
let apiPath;
let websocketUrl;
const isIframe = window.location !== window.parent.location;
const isStandalone = window.location.pathname === '/'
  || window.location.pathname === '/demo/'
  || window.location.pathname === '/scoped/'
  || /\/(.+).html/.test(window.location.pathname);
const wsProto = location.protocol === 'https:' ? 'wss' : 'ws';

if (isIframe || isStandalone) {
  apiPath = 'api';
  websocketUrl = `${wsProto}://${location.host}${basePath(location.pathname)}`;
} else {
  apiPath = `/api${basePath(window.location.pathname)}/api`;
  websocketUrl = `${wsProto}://${location.host}/api${basePath(window.location.pathname)}`;
}

export const wsUrl = websocketUrl;

function createWebsocket(topologyUrl, optionsQuery, dispatch) {
  if (socket) {
    socket.onclose = null;
    socket.onerror = null;
    socket.close();
    // onclose() is not called, but that's fine since we're opening a new one
    // right away
  }

  // profiling
  createWebsocketAt = new Date();
  firstMessageOnWebsocketAt = 0;

  socket = new WebSocket(`${wsUrl}${topologyUrl}/ws?t=${updateFrequency}&${optionsQuery}`);

  socket.onopen = () => {
    dispatch(openWebsocket());
  };

  socket.onclose = () => {
    clearTimeout(reconnectTimer);
    log(`Closing websocket to ${topologyUrl}`, socket.readyState);
    socket = null;
    dispatch(closeWebsocket());

    reconnectTimer = setTimeout(() => {
      createWebsocket(topologyUrl, optionsQuery, dispatch);
    }, reconnectTimerInterval);
  };

  socket.onerror = () => {
    log(`Error in websocket to ${topologyUrl}`);
    dispatch(receiveError(currentUrl));
  };

  socket.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    dispatch(receiveNodesDelta(msg));

    // profiling (receiveNodesDelta triggers synchronous render)
    if (!firstMessageOnWebsocketAt) {
      firstMessageOnWebsocketAt = new Date();
      const timeToFirstMessage = firstMessageOnWebsocketAt - createWebsocketAt;
      if (timeToFirstMessage > FIRST_RENDER_TOO_LONG_THRESHOLD) {
        log('Time (ms) to first nodes render after websocket was created',
          firstMessageOnWebsocketAt - createWebsocketAt);
      }
    }
  };
}

/* keep URLs relative */

/**
 * Gets nodes for all topologies (for search)
 */
export function getAllNodes(getState, dispatch) {
  const state = getState();
  const topologyOptions = state.get('topologyOptions');
  // fetch sequentially
  state.get('topologyUrlsById')
    .reduce((sequence, topologyUrl, topologyId) => sequence.then(() => {
      const optionsQuery = buildOptionsQuery(topologyOptions.get(topologyId));
      // Trim the leading slash from the url before requesting.
      // This ensures that scope will request from the correct route if embedded in an iframe.
      return fetch(`${trimStart(topologyUrl, '/')}?${optionsQuery}`);
    })
    .then(response => response.json())
    .then(json => dispatch(receiveNodesForTopology(json.nodes, topologyId))),
    Promise.resolve());
}

export function getTopologies(options, dispatch) {
  clearTimeout(topologyTimer);
  const optionsQuery = buildOptionsQuery(options);
  const url = `${apiPath}/topology?${optionsQuery}`;
  reqwest({
    url,
    success: (res) => {
      dispatch(receiveTopologies(res));
      topologyTimer = setTimeout(() => {
        getTopologies(options, dispatch);
      }, TOPOLOGY_INTERVAL);
    },
    error: (err) => {
      log(`Error in topology request: ${err.responseText}`);
      dispatch(receiveError(url));
      topologyTimer = setTimeout(() => {
        getTopologies(options, dispatch);
      }, TOPOLOGY_INTERVAL);
    }
  });
}

export function getNodesDelta(topologyUrl, options, dispatch) {
  const optionsQuery = buildOptionsQuery(options);

  // only recreate websocket if url changed
  if (topologyUrl && (topologyUrl !== currentUrl || currentOptions !== optionsQuery)) {
    createWebsocket(topologyUrl, optionsQuery, dispatch);
    currentUrl = topologyUrl;
    currentOptions = optionsQuery;
  }
}

export function getNodeDetails(topologyUrlsById, currentTopologyId, options, nodeMap, dispatch) {
  // get details for all opened nodes
  const obj = nodeMap.last();
  if (obj && topologyUrlsById.has(obj.topologyId)) {
    const topologyUrl = topologyUrlsById.get(obj.topologyId);
    let urlComponents = [apiPath, '/', trimStart(topologyUrl, '/api'), '/', encodeURIComponent(obj.id)];
    if (currentTopologyId === obj.topologyId) {
      // Only forward filters for nodes in the current topology
      const optionsQuery = buildOptionsQuery(options);
      urlComponents = urlComponents.concat(['?', optionsQuery]);
    }
    const url = urlComponents.join('');

    reqwest({
      url,
      success: (res) => {
        // make sure node is still selected
        if (nodeMap.has(res.node.id)) {
          dispatch(receiveNodeDetails(res.node));
        }
      },
      error: (err) => {
        log(`Error in node details request: ${err.responseText}`);
        // dont treat missing node as error
        if (err.status === 404) {
          dispatch(receiveNotFound(obj.id));
        } else {
          dispatch(receiveError(topologyUrl));
        }
      }
    });
  } else if (obj) {
    log('No details or url found for ', obj);
  }
}

export function getApiDetails(dispatch) {
  clearTimeout(apiDetailsTimer);
  const url = apiPath;
  reqwest({
    url,
    success: (res) => {
      dispatch(receiveApiDetails(res));
      apiDetailsTimer = setTimeout(() => {
        getApiDetails(dispatch);
      }, API_INTERVAL);
    },
    error: (err) => {
      log(`Error in api details request: ${err.responseText}`);
      receiveError(url);
      apiDetailsTimer = setTimeout(() => {
        getApiDetails(dispatch);
      }, API_INTERVAL / 2);
    }
  });
}

export function doControlRequest(nodeId, control, dispatch) {
  clearTimeout(controlErrorTimer);
  const url = `${apiPath}/control/${encodeURIComponent(control.probeId)}/`
    + `${encodeURIComponent(control.nodeId)}/${control.id}`;
  reqwest({
    method: 'POST',
    url,
    success: (res) => {
      dispatch(receiveControlSuccess(nodeId));
      if (res) {
        if (res.pipe) {
          dispatch(blurSearch());
          const resizeTtyControl = res.resize_tty_control &&
            {id: res.resize_tty_control, probeId: control.probeId, nodeId: control.nodeId};
          dispatch(receiveControlPipe(
            res.pipe,
            nodeId,
            res.raw_tty,
            resizeTtyControl,
            control
          ));
        }
        if (res.removedNode) {
          dispatch(receiveControlNodeRemoved(nodeId));
        }
      }
    },
    error: (err) => {
      dispatch(receiveControlError(nodeId, err.response));
      controlErrorTimer = setTimeout(() => {
        dispatch(clearControlError(nodeId));
      }, 10000);
    }
  });
}


export function doResizeTty(pipeId, control, cols, rows) {
  const url = `${apiPath}/control/${encodeURIComponent(control.probeId)}/`
    + `${encodeURIComponent(control.nodeId)}/${control.id}`;

  return reqwest({
    method: 'POST',
    url,
    data: JSON.stringify({pipeID: pipeId, width: cols.toString(), height: rows.toString()}),
  })
    .fail((err) => {
      log(`Error resizing pipe: ${err}`);
    });
}


export function deletePipe(pipeId, dispatch) {
  const url = `${apiPath}/pipe/${encodeURIComponent(pipeId)}`;
  reqwest({
    method: 'DELETE',
    url,
    success: () => {
      log('Closed the pipe!');
    },
    error: (err) => {
      log(`Error closing pipe:${err}`);
      dispatch(receiveError(url));
    }
  });
}


export function getPipeStatus(pipeId, dispatch) {
  const url = `${apiPath}/pipe/${encodeURIComponent(pipeId)}/check`;
  reqwest({
    method: 'GET',
    url,
    complete: (res) => {
      const status = {
        204: 'PIPE_ALIVE',
        404: 'PIPE_DELETED'
      }[res.status];

      if (!status) {
        log('Unexpected pipe status:', res.status);
        return;
      }

      dispatch(receiveControlPipeStatus(pipeId, status));
    }
  });
}

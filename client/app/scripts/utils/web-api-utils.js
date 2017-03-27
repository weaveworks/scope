import debug from 'debug';
import reqwest from 'reqwest';
import defaults from 'lodash/defaults';
import { Map as makeMap, List } from 'immutable';

import { blurSearch, clearControlError, closeWebsocket, openWebsocket, receiveError,
  receiveApiDetails, receiveNodesDelta, receiveNodeDetails, receiveControlError,
  receiveControlNodeRemoved, receiveControlPipe, receiveControlPipeStatus,
  receiveControlSuccess, receiveTopologies, receiveNotFound,
  receiveNodesForTopology } from '../actions/app-actions';

import { layersTopologyIdsSelector } from '../selectors/resource-view/layout';
import { API_INTERVAL, TOPOLOGY_INTERVAL } from '../constants/timer';

const log = debug('scope:web-api-utils');

const reconnectTimerInterval = 5000;
const updateFrequency = '5s';
const FIRST_RENDER_TOO_LONG_THRESHOLD = 100; // ms
const csrfToken = (() => {
  // Check for token at window level or parent level (for iframe);
  /* eslint-disable no-underscore-dangle */
  const token = typeof window !== 'undefined'
    ? window.__WEAVEWORKS_CSRF_TOKEN || parent.__WEAVEWORKS_CSRF_TOKEN
    : null;
  /* eslint-enable no-underscore-dangle */
  if (!token || token === '$__CSRF_TOKEN_PLACEHOLDER__') {
    // Authfe did not replace the token in the static html.
    return null;
  }

  return token;
})();

let socket;
let reconnectTimer = 0;
let currentUrl = null;
let currentOptions = null;
let topologyTimer = 0;
let apiDetailsTimer = 0;
let controlErrorTimer = 0;
let createWebsocketAt = 0;
let firstMessageOnWebsocketAt = 0;
let continuePolling = true;

export function buildOptionsQuery(options) {
  if (options) {
    return options.map((value, param) => {
      if (List.isList(value)) {
        value = value.join(',');
      }
      return `${param}=${value}`;
    }).join('&');
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

export function getApiPath(pathname = window.location.pathname) {
  if (process.env.SCOPE_API_PREFIX) {
    return basePath(`${process.env.SCOPE_API_PREFIX}${pathname}`);
  }

  return basePath(pathname);
}

export function getWebsocketUrl(host = window.location.host, pathname = window.location.pathname) {
  const wsProto = location.protocol === 'https:' ? 'wss' : 'ws';
  return `${wsProto}://${host}${process.env.SCOPE_API_PREFIX || ''}${basePath(pathname)}`;
}

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

  socket = new WebSocket(`${getWebsocketUrl()}${topologyUrl}/ws?t=${updateFrequency}&${optionsQuery}`);

  socket.onopen = () => {
    dispatch(openWebsocket());
  };

  socket.onclose = () => {
    clearTimeout(reconnectTimer);
    log(`Closing websocket to ${topologyUrl}`, socket.readyState);
    socket = null;
    dispatch(closeWebsocket());

    if (continuePolling) {
      reconnectTimer = setTimeout(() => {
        createWebsocket(topologyUrl, optionsQuery, dispatch);
      }, reconnectTimerInterval);
    }
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

/**
  * XHR wrapper. Applies a CSRF token (if it exists) and content-type to all requests.
  * Any opts that get passed in will override the defaults.
  */
function doRequest(opts) {
  const config = defaults(opts, {
    contentType: 'application/json',
    type: 'json'
  });
  if (csrfToken) {
    config.headers = Object.assign({}, config.headers, { 'X-CSRF-Token': csrfToken });
  }

  return reqwest(config);
}

/**
 * Does a one-time fetch of all the nodes for a custom list of topologies.
 */
function getNodesForTopologies(getState, dispatch, topologyIds, topologyOptions = makeMap()) {
  // fetch sequentially
  getState().get('topologyUrlsById')
    .filter((_, topologyId) => topologyIds.contains(topologyId))
    .reduce((sequence, topologyUrl, topologyId) => sequence.then(() => {
      const optionsQuery = buildOptionsQuery(topologyOptions.get(topologyId));
      return doRequest({ url: `${getApiPath()}${topologyUrl}?${optionsQuery}` });
    })
    .then(json => dispatch(receiveNodesForTopology(json.nodes, topologyId))),
    Promise.resolve());
}

/**
 * Gets nodes for all topologies (for search).
 */
export function getAllNodes(getState, dispatch) {
  const state = getState();
  const topologyOptions = state.get('topologyOptions');
  const topologyIds = state.get('topologyUrlsById').keySeq();
  getNodesForTopologies(getState, dispatch, topologyIds, topologyOptions);
}

/**
 * One-time update of all the nodes of topologies that appear in the current resource view.
 * TODO: Replace the one-time snapshot with periodic polling.
 */
export function getResourceViewNodesSnapshot(getState, dispatch) {
  const topologyIds = layersTopologyIdsSelector(getState());
  getNodesForTopologies(getState, dispatch, topologyIds);
}

export function getTopologies(options, dispatch, initialPoll) {
  // Used to resume polling when navigating between pages in Weave Cloud.
  continuePolling = initialPoll === true ? true : continuePolling;
  clearTimeout(topologyTimer);
  const optionsQuery = buildOptionsQuery(options);
  const url = `${getApiPath()}/api/topology?${optionsQuery}`;
  doRequest({
    url,
    success: (res) => {
      if (continuePolling) {
        dispatch(receiveTopologies(res));
        topologyTimer = setTimeout(() => {
          getTopologies(options, dispatch);
        }, TOPOLOGY_INTERVAL);
      }
    },
    error: (req) => {
      log(`Error in topology request: ${req.responseText}`);
      dispatch(receiveError(url));
      // Only retry in stand-alone mode
      if (continuePolling) {
        topologyTimer = setTimeout(() => {
          getTopologies(options, dispatch);
        }, TOPOLOGY_INTERVAL);
      }
    }
  });
}

// TODO: topologyUrl and options are always used for the current topology so they as arguments
// can be replaced by the `state` and then retrieved here internally from selectors.
export function getNodesDelta(topologyUrl, options, dispatch) {
  const optionsQuery = buildOptionsQuery(options);
  // Only recreate websocket if url changed or if forced (weave cloud instance reload);
  // Check for truthy options and that options have changed.
  const isNewOptions = currentOptions && currentOptions !== optionsQuery;
  const isNewUrl = topologyUrl !== currentUrl || isNewOptions;
  // `topologyUrl` can be undefined initially, so only create a socket if it is truthy
  // and no socket exists, or if we get a new url.
  if ((topologyUrl && !socket) || (topologyUrl && isNewUrl)) {
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
    let urlComponents = [getApiPath(), topologyUrl, '/', encodeURIComponent(obj.id)];
    if (currentTopologyId === obj.topologyId) {
      // Only forward filters for nodes in the current topology
      const optionsQuery = buildOptionsQuery(options);
      urlComponents = urlComponents.concat(['?', optionsQuery]);
    }
    const url = urlComponents.join('');

    doRequest({
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
  const url = `${getApiPath()}/api`;
  doRequest({
    url,
    success: (res) => {
      dispatch(receiveApiDetails(res));
      if (continuePolling) {
        apiDetailsTimer = setTimeout(() => {
          getApiDetails(dispatch);
        }, API_INTERVAL);
      }
    },
    error: (req) => {
      log(`Error in api details request: ${req.responseText}`);
      receiveError(url);
      if (continuePolling) {
        apiDetailsTimer = setTimeout(() => {
          getApiDetails(dispatch);
        }, API_INTERVAL / 2);
      }
    }
  });
}

export function doControlRequest(nodeId, control, dispatch) {
  clearTimeout(controlErrorTimer);
  const url = `${getApiPath()}/api/control/${encodeURIComponent(control.probeId)}/`
    + `${encodeURIComponent(control.nodeId)}/${control.id}`;
  doRequest({
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
  const url = `${getApiPath()}/api/control/${encodeURIComponent(control.probeId)}/`
    + `${encodeURIComponent(control.nodeId)}/${control.id}`;

  return doRequest({
    method: 'POST',
    url,
    data: JSON.stringify({pipeID: pipeId, width: cols.toString(), height: rows.toString()}),
  })
    .fail((err) => {
      log(`Error resizing pipe: ${err}`);
    });
}


export function deletePipe(pipeId, dispatch) {
  const url = `${getApiPath()}/api/pipe/${encodeURIComponent(pipeId)}`;
  doRequest({
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
  const url = `${getApiPath()}/api/pipe/${encodeURIComponent(pipeId)}/check`;
  doRequest({
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

export function stopPolling() {
  clearTimeout(apiDetailsTimer);
  clearTimeout(topologyTimer);
  continuePolling = false;
}

export function teardownWebsockets() {
  clearTimeout(reconnectTimer);
  if (socket) {
    socket.onerror = null;
    socket.onclose = null;
    socket.onmessage = null;
    socket.onopen = null;
    socket.close();
    socket = null;
    currentOptions = null;
  }
}

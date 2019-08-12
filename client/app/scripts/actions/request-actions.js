/*

This file consists of functions that both dispatch actions to Redux and also make API requests.

TODO: Refactor all the methods below so that the split between actions and
requests is more clear, and make user components make explicit calls to requests
and dispatch actions when handling request promises.

*/
import debug from 'debug';
import { fromJS } from 'immutable';

import ActionTypes from '../constants/action-types';
import { RESOURCE_VIEW_MODE } from '../constants/naming';
import {
  API_REFRESH_INTERVAL,
  TOPOLOGY_REFRESH_INTERVAL,
} from '../constants/timer';
import { updateRoute } from '../utils/router-utils';
import { getCurrentTopologyUrl } from '../utils/topology-utils';
import {
  doRequest,
  getApiPath,
  getAllNodes,
  getNodesOnce,
  deletePipe,
  getNodeDetails,
  getResourceViewNodesSnapshot,
  topologiesUrl,
  buildWebsocketUrl,
} from '../utils/web-api-utils';
import {
  availableMetricTypesSelector,
  pinnedMetricSelector,
} from '../selectors/node-metric';
import {
  isResourceViewModeSelector,
  resourceViewAvailableSelector,
  activeTopologyOptionsSelector,
} from '../selectors/topology';
import { isPausedSelector } from '../selectors/time-travel';

import {
  receiveControlNodeRemoved,
  receiveControlPipeStatus,
  receiveControlSuccess,
  receiveControlError,
  receiveError,
  pinMetric,
  openWebsocket,
  closeWebsocket,
  receiveNodesDelta,
  clearControlError,
  blurSearch,
} from './app-actions';


const log = debug('scope:app-actions');
const reconnectTimerInterval = 5000;
const FIRST_RENDER_TOO_LONG_THRESHOLD = 100; // ms

let socket;
let topologyTimer = 0;
let controlErrorTimer = 0;
let reconnectTimer = 0;
let apiDetailsTimer = 0;
let continuePolling = true;
let firstMessageOnWebsocketAt = null;
let createWebsocketAt = null;
let currentUrl = null;

function createWebsocket(websocketUrl, getState, dispatch) {
  if (socket) {
    socket.onclose = null;
    socket.onerror = null;
    socket.close();
    // onclose() is not called, but that's fine since we're opening a new one
    // right away
  }

  // profiling
  createWebsocketAt = new Date();
  firstMessageOnWebsocketAt = null;

  socket = new WebSocket(websocketUrl);

  socket.onopen = () => {
    log(`Opening websocket to ${websocketUrl}`);
    dispatch(openWebsocket());
  };

  socket.onclose = () => {
    clearTimeout(reconnectTimer);
    log(`Closing websocket to ${websocketUrl}`, socket.readyState);
    socket = null;
    dispatch(closeWebsocket());

    if (continuePolling && !isPausedSelector(getState())) {
      reconnectTimer = setTimeout(() => {
        createWebsocket(websocketUrl, getState, dispatch);
      }, reconnectTimerInterval);
    }
  };

  socket.onerror = () => {
    log(`Error in websocket to ${websocketUrl}`);
    dispatch(receiveError(websocketUrl));
  };

  socket.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    dispatch(receiveNodesDelta(msg));

    // profiling (receiveNodesDelta triggers synchronous render)
    if (!firstMessageOnWebsocketAt) {
      firstMessageOnWebsocketAt = new Date();
      const timeToFirstMessage = firstMessageOnWebsocketAt - createWebsocketAt;
      if (timeToFirstMessage > FIRST_RENDER_TOO_LONG_THRESHOLD) {
        log(
          'Time (ms) to first nodes render after websocket was created',
          firstMessageOnWebsocketAt - createWebsocketAt
        );
      }
    }
  };
}

function teardownWebsockets() {
  clearTimeout(reconnectTimer);
  if (socket) {
    socket.onerror = null;
    socket.onclose = null;
    socket.onmessage = null;
    socket.onopen = null;
    socket.close();
    socket = null;
    currentUrl = null;
  }
}

function updateWebsocketChannel(getState, dispatch, forceRequest) {
  const topologyUrl = getCurrentTopologyUrl(getState());
  const topologyOptions = activeTopologyOptionsSelector(getState());
  const websocketUrl = buildWebsocketUrl(topologyUrl, topologyOptions, getState());
  // Only recreate websocket if url changed or if forced (weave cloud instance reload);
  const isNewUrl = websocketUrl !== currentUrl;
  // `topologyUrl` can be undefined initially, so only create a socket if it is truthy
  // and no socket exists, or if we get a new url.
  if (topologyUrl && (!socket || isNewUrl || forceRequest)) {
    createWebsocket(websocketUrl, getState, dispatch);
    currentUrl = websocketUrl;
  }
}

function getNodes(getState, dispatch, forceRequest = false) {
  if (isPausedSelector(getState())) {
    getNodesOnce(getState, dispatch);
  } else {
    updateWebsocketChannel(getState, dispatch, forceRequest);
  }
  getNodeDetails(getState, dispatch);
}

export function pauseTimeAtNow() {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.PAUSE_TIME_AT_NOW
    });
    updateRoute(getState);
    if (!getState().get('nodesLoaded')) {
      getNodes(getState, dispatch);
      if (isResourceViewModeSelector(getState())) {
        getResourceViewNodesSnapshot(getState(), dispatch);
      }
    }
  };
}

function receiveTopologies(topologies) {
  return (dispatch, getState) => {
    const firstLoad = !getState().get('topologiesLoaded');
    dispatch({
      topologies,
      type: ActionTypes.RECEIVE_TOPOLOGIES
    });
    getNodes(getState, dispatch);
    // Populate search matches on first load
    const state = getState();
    // Fetch all the relevant nodes once on first load
    if (firstLoad && isResourceViewModeSelector(state)) {
      getResourceViewNodesSnapshot(state, dispatch);
    }
  };
}

function getTopologiesOnce(getState, dispatch) {
  const url = topologiesUrl(getState());
  doRequest({
    error: (req) => {
      log(`Error in topology request: ${req.responseText}`);
      dispatch(receiveError(url));
    },
    success: (res) => {
      dispatch(receiveTopologies(res));
    },
    url
  });
}

function pollTopologies(getState, dispatch, initialPoll = false) {
  // Used to resume polling when navigating between pages in Weave Cloud.
  continuePolling = initialPoll === true ? true : continuePolling;
  clearTimeout(topologyTimer);
  // NOTE: getState is called every time to make sure the up-to-date state is used.
  const url = topologiesUrl(getState());
  doRequest({
    error: (req) => {
      log(`Error in topology request: ${req.responseText}`);
      dispatch(receiveError(url));
      // Only retry in stand-alone mode
      if (continuePolling && !isPausedSelector(getState())) {
        topologyTimer = setTimeout(() => {
          pollTopologies(getState, dispatch);
        }, TOPOLOGY_REFRESH_INTERVAL);
      }
    },
    success: (res) => {
      if (continuePolling && !isPausedSelector(getState())) {
        dispatch(receiveTopologies(res));
        topologyTimer = setTimeout(() => {
          pollTopologies(getState, dispatch);
        }, TOPOLOGY_REFRESH_INTERVAL);
      }
    },
    url
  });
}

function getTopologies(getState, dispatch, forceRequest) {
  if (isPausedSelector(getState())) {
    getTopologiesOnce(getState, dispatch);
  } else {
    pollTopologies(getState, dispatch, forceRequest);
  }
}

export function jumpToTime(timestamp) {
  return (dispatch, getState) => {
    dispatch({
      timestamp,
      type: ActionTypes.JUMP_TO_TIME,
    });
    updateRoute(getState);
    getTopologies(getState, dispatch);
    if (!getState().get('nodesLoaded')) {
      getNodes(getState, dispatch);
      if (isResourceViewModeSelector(getState())) {
        getResourceViewNodesSnapshot(getState(), dispatch);
      }
    } else {
      // Get most recent details before freezing the state.
      getNodeDetails(getState, dispatch);
    }
  };
}

export function receiveApiDetails(apiDetails) {
  return (dispatch, getState) => {
    const isFirstTime = !getState().get('version');
    const pausedAt = getState().get('pausedAt');

    dispatch({
      capabilities: fromJS(apiDetails.capabilities || {}),
      hostname: apiDetails.hostname,
      newVersion: apiDetails.newVersion,
      plugins: apiDetails.plugins,
      type: ActionTypes.RECEIVE_API_DETAILS,
      version: apiDetails.version,
    });

    // On initial load either start time travelling at the pausedAt timestamp
    // (if it was given as URL param) if time travelling is enabled, otherwise
    // simply pause at the present time which is arguably the next best thing
    // we could do.
    // NOTE: We can't make this decision before API details are received because
    // we have no prior info on whether time travel would be available.
    if (isFirstTime && pausedAt) {
      if (apiDetails.capabilities && apiDetails.capabilities.historic_reports) {
        dispatch(jumpToTime(pausedAt));
      } else {
        dispatch(pauseTimeAtNow());
      }
    }
  };
}

export function getApiDetails(dispatch) {
  clearTimeout(apiDetailsTimer);
  const url = `${getApiPath()}/api`;
  doRequest({
    error: (req) => {
      log(`Error in api details request: ${req.responseText}`);
      receiveError(url);
      if (continuePolling) {
        apiDetailsTimer = setTimeout(() => {
          getApiDetails(dispatch);
        }, API_REFRESH_INTERVAL / 2);
      }
    },
    success: (res) => {
      dispatch(receiveApiDetails(res));
      if (continuePolling) {
        apiDetailsTimer = setTimeout(() => {
          getApiDetails(dispatch);
        }, API_REFRESH_INTERVAL);
      }
    },
    url
  });
}

function stopPolling() {
  clearTimeout(apiDetailsTimer);
  clearTimeout(topologyTimer);
  continuePolling = false;
}

export function focusSearch() {
  return (dispatch, getState) => {
    dispatch({ type: ActionTypes.FOCUS_SEARCH });
    // update nodes cache to allow search across all topologies,
    // wait a second until animation is over
    // NOTE: This will cause matching recalculation (and rerendering)
    // of all the nodes in the topology, instead applying it only on
    // the nodes delta. The solution would be to implement deeper
    // search selectors with per-node caching instead of per-topology.
    setTimeout(() => {
      getAllNodes(getState(), dispatch);
    }, 1200);
  };
}

export function getPipeStatus(pipeId, dispatch) {
  const url = `${getApiPath()}/api/pipe/${encodeURIComponent(pipeId)}/check`;
  doRequest({
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
    },
    method: 'GET',
    url
  });
}

export function receiveControlPipe(pipeId, nodeId, rawTty, resizeTtyControl, control) {
  return (dispatch, getState) => {
    const state = getState();
    if (state.get('nodeDetails').last()
      && nodeId !== state.get('nodeDetails').last().id) {
      log('Node was deselected before we could set up control!');
      deletePipe(pipeId, dispatch);
      return;
    }

    const controlPipe = state.get('controlPipes').last();
    if (controlPipe && controlPipe.get('id') !== pipeId) {
      deletePipe(controlPipe.get('id'), dispatch);
    }

    dispatch({
      control,
      nodeId,
      pipeId,
      rawTty,
      resizeTtyControl,
      type: ActionTypes.RECEIVE_CONTROL_PIPE
    });

    updateRoute(getState);
  };
}

function doControlRequest(nodeId, control, dispatch) {
  clearTimeout(controlErrorTimer);
  const url = `${getApiPath()}/api/control/${encodeURIComponent(control.probeId)}/`
    + `${encodeURIComponent(control.nodeId)}/${control.id}`;
  doRequest({
    error: (err) => {
      dispatch(receiveControlError(nodeId, err.response));
      controlErrorTimer = setTimeout(() => {
        dispatch(clearControlError(nodeId));
      }, 10000);
    },
    method: 'POST',
    success: (res) => {
      dispatch(receiveControlSuccess(nodeId));
      if (res) {
        if (res.pipe) {
          dispatch(blurSearch());
          const resizeTtyControl = res.resize_tty_control
            && { id: res.resize_tty_control, nodeId: control.nodeId, probeId: control.probeId };
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
    url
  });
}

export function doControl(nodeId, control) {
  return (dispatch) => {
    dispatch({
      control,
      nodeId,
      type: ActionTypes.DO_CONTROL
    });
    doControlRequest(nodeId, control, dispatch);
  };
}

export function shutdown() {
  return (dispatch) => {
    stopPolling();
    teardownWebsockets();
    dispatch({
      type: ActionTypes.SHUTDOWN
    });
  };
}

export function setResourceView() {
  return (dispatch, getState) => {
    if (resourceViewAvailableSelector(getState())) {
      dispatch({
        type: ActionTypes.SET_VIEW_MODE,
        viewMode: RESOURCE_VIEW_MODE,
      });
      // Pin the first metric if none of the visible ones is pinned.
      const state = getState();
      if (!pinnedMetricSelector(state)) {
        const firstAvailableMetricType = availableMetricTypesSelector(state).first();
        dispatch(pinMetric(firstAvailableMetricType));
      }
      getResourceViewNodesSnapshot(getState(), dispatch);
      updateRoute(getState);
    }
  };
}

export function changeTopologyOption(option, value, topologyId, addOrRemove) {
  return (dispatch, getState) => {
    dispatch({
      addOrRemove,
      option,
      topologyId,
      type: ActionTypes.CHANGE_TOPOLOGY_OPTION,
      value
    });
    updateRoute(getState);
    // update all request workers with new options
    getTopologies(getState, dispatch);
    getNodes(getState, dispatch);
  };
}

export function getTopologiesWithInitialPoll() {
  return (dispatch, getState) => {
    getTopologies(getState, dispatch, true);
  };
}

export function resumeTime() {
  return (dispatch, getState) => {
    if (isPausedSelector(getState())) {
      dispatch({
        type: ActionTypes.RESUME_TIME
      });
      updateRoute(getState);
      // After unpausing, all of the following calls will re-activate polling.
      getTopologies(getState, dispatch);
      getNodes(getState, dispatch, true);
      if (isResourceViewModeSelector(getState())) {
        getResourceViewNodesSnapshot(getState(), dispatch);
      }
    }
  };
}

export function route(urlState) {
  return (dispatch, getState) => {
    dispatch({
      state: urlState,
      type: ActionTypes.ROUTE_TOPOLOGY
    });
    // Handle Time Travel state update through separate actions as it's more complex.
    // This is mostly to handle switching contexts Explore <-> Monitor in WC while
    // the timestamp keeps changing - e.g. if we were Time Travelling in Scope and
    // then went live in Monitor, switching back to Explore should properly close
    // the Time Travel etc, not just update the pausedAt state directly.
    if (!urlState.pausedAt) {
      dispatch(resumeTime());
    } else {
      dispatch(jumpToTime(urlState.pausedAt));
    }
    // update all request workers with new options
    getTopologies(getState, dispatch);
    getNodes(getState, dispatch);
    // If we are landing on the resource view page, we need to fetch not only all the
    // nodes for the current topology, but also the nodes of all the topologies that make
    // the layers in the resource view.
    const state = getState();
    if (isResourceViewModeSelector(state)) {
      getResourceViewNodesSnapshot(state, dispatch);
    }
  };
}

export function clickCloseDetails(nodeId) {
  return (dispatch, getState) => {
    dispatch({
      nodeId,
      type: ActionTypes.CLICK_CLOSE_DETAILS
    });
    // Pull the most recent details for the next details panel that comes into focus.
    getNodeDetails(getState, dispatch);
    updateRoute(getState);
  };
}

export function clickNode(nodeId, label, origin, topologyId = null) {
  return (dispatch, getState) => {
    dispatch({
      label,
      nodeId,
      origin,
      topologyId,
      type: ActionTypes.CLICK_NODE,
    });
    updateRoute(getState);
    getNodeDetails(getState, dispatch);
  };
}

export function clickRelative(nodeId, topologyId, label, origin) {
  return (dispatch, getState) => {
    dispatch({
      label,
      nodeId,
      origin,
      topologyId,
      type: ActionTypes.CLICK_RELATIVE
    });
    updateRoute(getState);
    getNodeDetails(getState, dispatch);
  };
}

function updateTopology(dispatch, getState) {
  const state = getState();
  // If we're in the resource view, get the snapshot of all the relevant node topologies.
  if (isResourceViewModeSelector(state)) {
    getResourceViewNodesSnapshot(state, dispatch);
  }
  updateRoute(getState);
  // NOTE: This is currently not needed for our static resource
  // view, but we'll need it here later and it's simpler to just
  // keep it than to redo the nodes delta updating logic.
  getNodes(getState, dispatch);
}

export function clickShowTopologyForNode(topologyId, nodeId) {
  return (dispatch, getState) => {
    dispatch({
      nodeId,
      topologyId,
      type: ActionTypes.CLICK_SHOW_TOPOLOGY_FOR_NODE
    });
    updateTopology(dispatch, getState);
  };
}

export function clickTopology(topologyId) {
  return (dispatch, getState) => {
    dispatch({
      topologyId,
      type: ActionTypes.CLICK_TOPOLOGY
    });
    updateTopology(dispatch, getState);
  };
}

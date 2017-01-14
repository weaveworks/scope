import debug from 'debug';
import trimStart from 'lodash/trimStart';
import reduce from 'lodash/reduce';
import map from 'lodash/map';
import each from 'lodash/each';
import { fromJS } from 'immutable';

import ActionTypes from '../constants/action-types';
import { saveGraph } from '../utils/file-utils';
import { modulo } from '../utils/math-utils';
import { updateRoute } from '../utils/router-utils';
import { parseQuery, searchTopology } from '../utils/search-utils';
import { bufferDeltaUpdate, resumeUpdate,
  resetUpdateBuffer } from '../utils/update-buffer-utils';
import { doControlRequest, getAllNodes, getNodesDelta, getNodeDetails,
  getTopologies, deletePipe } from '../utils/web-api-utils';
import { getActiveTopologyOptions,
  getCurrentTopologyUrl } from '../utils/topology-utils';
import { storageSet } from '../utils/storage-utils';

const log = debug('scope:app-actions');

export function showHelp() {
  return {type: ActionTypes.SHOW_HELP};
}


export function hideHelp() {
  return {type: ActionTypes.HIDE_HELP};
}


export function toggleHelp() {
  return (dispatch, getState) => {
    if (getState().get('showingHelp')) {
      dispatch(hideHelp());
    } else {
      dispatch(showHelp());
    }
  };
}


export function sortOrderChanged(sortedBy, sortedDesc) {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.SORT_ORDER_CHANGED,
      sortedBy,
      sortedDesc
    });
    updateRoute(getState);
  };
}


//
// Networks
//


export function showNetworks(visible) {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.SHOW_NETWORKS,
      visible
    });

    updateRoute(getState);
  };
}


export function selectNetwork(networkId) {
  return {
    type: ActionTypes.SELECT_NETWORK,
    networkId
  };
}

export function pinNetwork(networkId) {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.PIN_NETWORK,
      networkId,
    });

    updateRoute(getState);
  };
}

export function unpinNetwork(networkId) {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.UNPIN_NETWORK,
      networkId,
    });

    updateRoute(getState);
  };
}


//
// Metrics
//


export function selectMetric(metricId) {
  return {
    type: ActionTypes.SELECT_METRIC,
    metricId
  };
}

export function pinMetric(metricId) {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.PIN_METRIC,
      metricId,
    });
    updateRoute(getState);
  };
}

export function unpinMetric() {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.UNPIN_METRIC,
    });
    updateRoute(getState);
  };
}

export function pinNextMetric(delta) {
  return (dispatch, getState) => {
    const state = getState();
    const metrics = state.get('availableCanvasMetrics').map(m => m.get('id'));
    const currentIndex = metrics.indexOf(state.get('selectedMetric'));
    const nextIndex = modulo(currentIndex + delta, metrics.count());
    const nextMetric = metrics.get(nextIndex);

    dispatch(pinMetric(nextMetric));
  };
}

export function unpinSearch(query) {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.UNPIN_SEARCH,
      query
    });
    updateRoute(getState);
  };
}

export function blurSearch() {
  return { type: ActionTypes.BLUR_SEARCH };
}

export function changeTopologyOption(option, value, topologyId) {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.CHANGE_TOPOLOGY_OPTION,
      topologyId,
      option,
      value
    });
    updateRoute(getState);
    // update all request workers with new options
    resetUpdateBuffer();
    const state = getState();
    getTopologies(getActiveTopologyOptions(state), dispatch);
    getNodesDelta(
      getCurrentTopologyUrl(state),
      getActiveTopologyOptions(state),
      dispatch
    );
    getNodeDetails(
      state.get('topologyUrlsById'),
      state.get('currentTopologyId'),
      getActiveTopologyOptions(state),
      state.get('nodeDetails'),
      dispatch
    );
  };
}

export function clickBackground() {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.CLICK_BACKGROUND
    });
    updateRoute(getState);
  };
}

export function clickCloseDetails(nodeId) {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.CLICK_CLOSE_DETAILS,
      nodeId
    });
    updateRoute(getState);
  };
}

export function clickCloseTerminal(pipeId, closePipe) {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.CLICK_CLOSE_TERMINAL,
      pipeId
    });
    if (closePipe) {
      deletePipe(pipeId, dispatch);
    }
    updateRoute(getState);
  };
}

export function clickDownloadGraph() {
  return (dispatch) => {
    dispatch({ type: ActionTypes.SET_EXPORTING_GRAPH, exporting: true });
    saveGraph();
    dispatch({ type: ActionTypes.SET_EXPORTING_GRAPH, exporting: false });
  };
}

export function clickForceRelayout() {
  return (dispatch) => {
    dispatch({
      type: ActionTypes.CLICK_FORCE_RELAYOUT,
      forceRelayout: true
    });
    // fire only once, reset after dispatch
    setTimeout(() => {
      dispatch({
        type: ActionTypes.CLICK_FORCE_RELAYOUT,
        forceRelayout: false
      });
    }, 100);
  };
}

export function toggleGridMode(enabledArgument) {
  return (dispatch, getState) => {
    const enabled = (enabledArgument === undefined) ?
      !getState().get('gridMode') :
      enabledArgument;
    dispatch({
      type: ActionTypes.SET_GRID_MODE,
      enabled
    });
    updateRoute(getState);
    if (!enabled) {
      dispatch(clickForceRelayout());
    }
  };
}

export function clickNode(nodeId, label, origin) {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.CLICK_NODE,
      origin,
      label,
      nodeId
    });
    updateRoute(getState);
    const state = getState();
    getNodeDetails(
      state.get('topologyUrlsById'),
      state.get('currentTopologyId'),
      getActiveTopologyOptions(state),
      state.get('nodeDetails'),
      dispatch
    );
  };
}

export function clickPauseUpdate() {
  return {
    type: ActionTypes.CLICK_PAUSE_UPDATE
  };
}

export function clickRelative(nodeId, topologyId, label, origin) {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.CLICK_RELATIVE,
      label,
      origin,
      nodeId,
      topologyId
    });
    updateRoute(getState);
    const state = getState();
    getNodeDetails(
      state.get('topologyUrlsById'),
      state.get('currentTopologyId'),
      getActiveTopologyOptions(state),
      state.get('nodeDetails'),
      dispatch
    );
  };
}

export function clickResumeUpdate() {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.CLICK_RESUME_UPDATE
    });
    resumeUpdate(getState);
  };
}

export function clickShowTopologyForNode(topologyId, nodeId) {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.CLICK_SHOW_TOPOLOGY_FOR_NODE,
      topologyId,
      nodeId
    });
    updateRoute(getState);
    // update all request workers with new options
    resetUpdateBuffer();
    const state = getState();
    getNodesDelta(
      getCurrentTopologyUrl(state),
      getActiveTopologyOptions(state),
      dispatch
    );
  };
}

export function clickTopology(topologyId) {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.CLICK_TOPOLOGY,
      topologyId
    });
    updateRoute(getState);
    // update all request workers with new options
    resetUpdateBuffer();
    const state = getState();
    getNodesDelta(
      getCurrentTopologyUrl(state),
      getActiveTopologyOptions(state),
      dispatch
    );
  };
}

export function openWebsocket() {
  return {
    type: ActionTypes.OPEN_WEBSOCKET
  };
}

export function clearControlError(nodeId) {
  return {
    type: ActionTypes.CLEAR_CONTROL_ERROR,
    nodeId
  };
}

export function closeWebsocket() {
  return {
    type: ActionTypes.CLOSE_WEBSOCKET
  };
}

export function doControl(nodeId, control) {
  return (dispatch) => {
    dispatch({
      type: ActionTypes.DO_CONTROL,
      nodeId,
      control
    });
    doControlRequest(nodeId, control, dispatch);
  };
}

export function doSearch(searchQuery) {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.DO_SEARCH,
      searchQuery
    });
    updateRoute(getState);
  };
}

export function enterEdge(edgeId) {
  return {
    type: ActionTypes.ENTER_EDGE,
    edgeId
  };
}

export function enterNode(nodeId) {
  return {
    type: ActionTypes.ENTER_NODE,
    nodeId
  };
}

export function focusSearch() {
  return (dispatch, getState) => {
    dispatch({ type: ActionTypes.FOCUS_SEARCH });
    // update nodes cache to allow search across all topologies,
    // wait a second until animation is over
    setTimeout(() => {
      getAllNodes(getState, dispatch);
    }, 1200);
  };
}

export function hitBackspace() {
  return (dispatch, getState) => {
    const state = getState();
    // remove last pinned query if search query is empty
    if (state.get('searchFocused') && !state.get('searchQuery')) {
      const query = state.get('pinnedSearches').last();
      if (query) {
        dispatch({
          type: ActionTypes.UNPIN_SEARCH,
          query
        });
        updateRoute(getState);
      }
    }
  };
}

export function hitEnter() {
  return (dispatch, getState) => {
    const state = getState();
    // pin query based on current search field
    if (state.get('searchFocused')) {
      const query = state.get('searchQuery');
      if (query && parseQuery(query)) {
        dispatch({
          type: ActionTypes.PIN_SEARCH,
          query
        });
        updateRoute(getState);
      }
    }
  };
}

export function hitEsc() {
  return (dispatch, getState) => {
    const state = getState();
    const controlPipe = state.get('controlPipes').last();
    if (controlPipe && controlPipe.get('status') === 'PIPE_DELETED') {
      dispatch({
        type: ActionTypes.CLICK_CLOSE_TERMINAL,
        pipeId: controlPipe.get('id')
      });
      updateRoute(getState);
      // Don't deselect node on ESC if there is a controlPipe (keep terminal open)
    } else if (state.get('searchFocused')) {
      if (state.get('searchQuery')) {
        dispatch(doSearch(''));
      } else {
        dispatch(blurSearch());
      }
    } else if (state.get('showingHelp')) {
      dispatch(hideHelp());
    } else if (state.get('nodeDetails').last() && !controlPipe) {
      dispatch({ type: ActionTypes.DESELECT_NODE });
      updateRoute(getState);
    }
  };
}

export function leaveEdge(edgeId) {
  return {
    type: ActionTypes.LEAVE_EDGE,
    edgeId
  };
}

export function leaveNode(nodeId) {
  return {
    type: ActionTypes.LEAVE_NODE,
    nodeId
  };
}

export function receiveControlError(nodeId, err) {
  return {
    type: ActionTypes.DO_CONTROL_ERROR,
    nodeId,
    error: err
  };
}

export function receiveControlSuccess(nodeId) {
  return {
    type: ActionTypes.DO_CONTROL_SUCCESS,
    nodeId
  };
}

export function receiveNodeDetails(details) {
  return {
    type: ActionTypes.RECEIVE_NODE_DETAILS,
    details
  };
}

export function receiveNodesDelta(delta) {
  return (dispatch, getState) => {
    //
    // allow css-animation to run smoothly by scheduling it to run on the
    // next tick after any potentially expensive canvas re-draws have been
    // completed.
    //
    setTimeout(() => dispatch({ type: ActionTypes.SET_RECEIVED_NODES_DELTA }), 0);

    if (delta.add || delta.update || delta.remove) {
      const state = getState();
      if (state.get('updatePausedAt') !== null) {
        bufferDeltaUpdate(delta);
      } else {
        dispatch({
          type: ActionTypes.RECEIVE_NODES_DELTA,
          delta
        });
      }
    }
  };
}

export function receiveNodesForTopology(nodes, topologyId) {
  return {
    type: ActionTypes.RECEIVE_NODES_FOR_TOPOLOGY,
    nodes,
    topologyId
  };
}

export function receiveTopologies(topologies) {
  return (dispatch, getState) => {
    const firstLoad = !getState().get('topologiesLoaded');
    dispatch({
      type: ActionTypes.RECEIVE_TOPOLOGIES,
      topologies
    });
    const state = getState();
    getNodesDelta(
      getCurrentTopologyUrl(state),
      getActiveTopologyOptions(state),
      dispatch
    );
    getNodeDetails(
      state.get('topologyUrlsById'),
      state.get('currentTopologyId'),
      getActiveTopologyOptions(state),
      state.get('nodeDetails'),
      dispatch
    );
    // populate search matches on first load
    if (firstLoad && state.get('searchQuery')) {
      dispatch(focusSearch());
    }
  };
}

export function receiveApiDetails(apiDetails) {
  return {
    type: ActionTypes.RECEIVE_API_DETAILS,
    hostname: apiDetails.hostname,
    version: apiDetails.version,
    plugins: apiDetails.plugins
  };
}

export function receiveControlNodeRemoved(nodeId) {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.RECEIVE_CONTROL_NODE_REMOVED,
      nodeId
    });
    updateRoute(getState);
  };
}

export function receiveControlPipeFromParams(pipeId, rawTty, resizeTtyControl) {
  // TODO add nodeId
  return {
    type: ActionTypes.RECEIVE_CONTROL_PIPE,
    pipeId,
    rawTty,
    resizeTtyControl
  };
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
      type: ActionTypes.RECEIVE_CONTROL_PIPE,
      nodeId,
      pipeId,
      rawTty,
      resizeTtyControl,
      control
    });

    updateRoute(getState);
  };
}

export function receiveControlPipeStatus(pipeId, status) {
  return {
    type: ActionTypes.RECEIVE_CONTROL_PIPE_STATUS,
    pipeId,
    status
  };
}

export function receiveError(errorUrl) {
  return {
    errorUrl,
    type: ActionTypes.RECEIVE_ERROR
  };
}

export function receiveNotFound(nodeId) {
  return {
    nodeId,
    type: ActionTypes.RECEIVE_NOT_FOUND
  };
}

export function route(urlState) {
  return (dispatch, getState) => {
    dispatch({
      state: urlState,
      type: ActionTypes.ROUTE_TOPOLOGY
    });
    // update all request workers with new options
    const state = getState();
    getTopologies(getActiveTopologyOptions(state), dispatch);
    getNodesDelta(
      getCurrentTopologyUrl(state),
      getActiveTopologyOptions(state),
      dispatch
    );
    getNodeDetails(
      state.get('topologyUrlsById'),
      state.get('currentTopologyId'),
      getActiveTopologyOptions(state),
      state.get('nodeDetails'),
      dispatch
    );
  };
}

export function resetLocalViewState() {
  return (dispatch) => {
    dispatch({type: ActionTypes.RESET_LOCAL_VIEW_STATE});
    storageSet('scopeViewState', '');
    window.location.href = window.location.href.split('#')[0];
  };
}

export function toggleTroubleshootingMenu(ev) {
  if (ev) { ev.preventDefault(); ev.stopPropagation(); }
  return {
    type: ActionTypes.TOGGLE_TROUBLESHOOTING_MENU
  };
}

function convertQueryString(paramString) {
  const pairs = trimStart(paramString, '?').split('&');
  return reduce(pairs, (result, pair) => {
    const [k, v] = pair.split('=');
    result[k] = v;
    return result;
  }, {});
}

function waterfall(series, target, cb) {
  function next(result) {
    const fn = series.shift();
    if (fn) {
      try {
        fn(result, next);
      } catch (e) {
        cb(e);
      }
    } else {
      cb(null, result);
    }
  }
  next(target, next);
}

export function translateUrlParamsToViewState(queryString) {
  return () => {
    const params = convertQueryString(queryString);
    if (params.node) {
      // Get the list of topologies
      fetch('api/topology').then(response => response.json())
        .then((topologies) => {
          // Queue up a list of functions to run, one after the other.
          const series = map(topologies, topo => (result, cb) => {
            // Fetch the node for each topology
            // TODO: Get this buildOptions to work
            // buildOptionsQuery(topo.options);
            fetch(`${trimStart(topo.url, '/')}`)
              .then(res => res.json())
              .then((json) => {
                // Append each node in the list to the result object.
                each(json.nodes, (node, id) => {
                  result[id] = node;
                });
                cb(result);
              })
              .catch((e) => { throw e; });
          });
          // Run the series
          waterfall(series, {}, (err, nodes) => {
            if (err) { throw err; }
            // const collection = flatten(map(r, ({nodes}) => values(nodes)));
            const result = searchTopology(fromJS(nodes), { query: params.node });
            console.log(result.toJS());
          });
        });
    }
  };
}

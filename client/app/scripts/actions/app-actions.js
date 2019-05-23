import debug from 'debug';
import { fromJS } from 'immutable';

import ActionTypes from '../constants/action-types';
import { saveGraph } from '../utils/file-utils';
import { clearStoredViewState, updateRoute } from '../utils/router-utils';
import {
  doControlRequest,
  getAllNodes,
  getResourceViewNodesSnapshot,
  getNodeDetails,
  getTopologies,
  deletePipe,
  stopPolling,
  teardownWebsockets,
  getNodes,
} from '../utils/web-api-utils';
import { isPausedSelector } from '../selectors/time-travel';
import {
  availableMetricTypesSelector,
  nextPinnedMetricTypeSelector,
  previousPinnedMetricTypeSelector,
  pinnedMetricSelector,
} from '../selectors/node-metric';
import {
  isResourceViewModeSelector,
  resourceViewAvailableSelector,
} from '../selectors/topology';

import {
  GRAPH_VIEW_MODE,
  TABLE_VIEW_MODE,
  RESOURCE_VIEW_MODE,
} from '../constants/naming';


const log = debug('scope:app-actions');


export function showHelp() {
  return { type: ActionTypes.SHOW_HELP };
}


export function hideHelp() {
  return { type: ActionTypes.HIDE_HELP };
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
      sortedBy,
      sortedDesc,
      type: ActionTypes.SORT_ORDER_CHANGED
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
    networkId,
    type: ActionTypes.SELECT_NETWORK
  };
}

export function pinNetwork(networkId) {
  return (dispatch, getState) => {
    dispatch({
      networkId,
      type: ActionTypes.PIN_NETWORK,
    });

    updateRoute(getState);
  };
}

export function unpinNetwork(networkId) {
  return (dispatch, getState) => {
    dispatch({
      networkId,
      type: ActionTypes.UNPIN_NETWORK,
    });

    updateRoute(getState);
  };
}


//
// Metrics
//

export function hoverMetric(metricType) {
  return {
    metricType,
    type: ActionTypes.HOVER_METRIC,
  };
}

export function unhoverMetric() {
  return {
    type: ActionTypes.UNHOVER_METRIC,
  };
}

export function pinMetric(metricType) {
  return (dispatch, getState) => {
    dispatch({
      metricType,
      type: ActionTypes.PIN_METRIC,
    });
    updateRoute(getState);
  };
}

export function unpinMetric() {
  return (dispatch, getState) => {
    // We always have to keep metrics pinned in the resource view.
    if (!isResourceViewModeSelector(getState())) {
      dispatch({
        type: ActionTypes.UNPIN_METRIC,
      });
      updateRoute(getState);
    }
  };
}

export function pinNextMetric() {
  return (dispatch, getState) => {
    const nextPinnedMetricType = nextPinnedMetricTypeSelector(getState());
    dispatch(pinMetric(nextPinnedMetricType));
  };
}

export function pinPreviousMetric() {
  return (dispatch, getState) => {
    const previousPinnedMetricType = previousPinnedMetricTypeSelector(getState());
    dispatch(pinMetric(previousPinnedMetricType));
  };
}

export function updateSearch(searchQuery = '', pinnedSearches = []) {
  return (dispatch, getState) => {
    dispatch({
      pinnedSearches,
      searchQuery,
      type: ActionTypes.UPDATE_SEARCH,
    });
    updateRoute(getState);
  };
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

export function blurSearch() {
  return { type: ActionTypes.BLUR_SEARCH };
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
      nodeId,
      type: ActionTypes.CLICK_CLOSE_DETAILS
    });
    // Pull the most recent details for the next details panel that comes into focus.
    getNodeDetails(getState, dispatch);
    updateRoute(getState);
  };
}

export function closeTerminal(pipeId) {
  return (dispatch, getState) => {
    dispatch({
      pipeId,
      type: ActionTypes.CLOSE_TERMINAL
    });
    updateRoute(getState);
  };
}

export function clickDownloadGraph() {
  return (dispatch) => {
    dispatch({ exporting: true, type: ActionTypes.SET_EXPORTING_GRAPH });
    saveGraph();
    dispatch({ exporting: false, type: ActionTypes.SET_EXPORTING_GRAPH });
  };
}

export function clickForceRelayout() {
  return (dispatch) => {
    dispatch({
      forceRelayout: true,
      type: ActionTypes.CLICK_FORCE_RELAYOUT
    });
    // fire only once, reset after dispatch
    setTimeout(() => {
      dispatch({
        forceRelayout: false,
        type: ActionTypes.CLICK_FORCE_RELAYOUT
      });
    }, 100);
  };
}

export function setViewportDimensions(width, height) {
  return (dispatch) => {
    dispatch({ height, type: ActionTypes.SET_VIEWPORT_DIMENSIONS, width });
  };
}

export function setGraphView() {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.SET_VIEW_MODE,
      viewMode: GRAPH_VIEW_MODE,
    });
    updateRoute(getState);
  };
}

export function setTableView() {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.SET_VIEW_MODE,
      viewMode: TABLE_VIEW_MODE,
    });
    updateRoute(getState);
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

export function cacheZoomState(zoomState) {
  return {
    type: ActionTypes.CACHE_ZOOM_STATE,
    // Make sure only proper numerical values are cached.
    zoomState: zoomState.filter(value => !window.isNaN(value)),
  };
}

export function openWebsocket() {
  return {
    type: ActionTypes.OPEN_WEBSOCKET
  };
}

export function clearControlError(nodeId) {
  return {
    nodeId,
    type: ActionTypes.CLEAR_CONTROL_ERROR
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
      control,
      nodeId,
      type: ActionTypes.DO_CONTROL
    });
    doControlRequest(nodeId, control, dispatch);
  };
}

export function enterEdge(edgeId) {
  return {
    edgeId,
    type: ActionTypes.ENTER_EDGE
  };
}

export function enterNode(nodeId) {
  return {
    nodeId,
    type: ActionTypes.ENTER_NODE
  };
}

export function hitEsc() {
  return (dispatch, getState) => {
    const state = getState();
    const controlPipe = state.get('controlPipes').last();
    if (controlPipe && controlPipe.get('status') === 'PIPE_DELETED') {
      dispatch({
        pipeId: controlPipe.get('id'),
        type: ActionTypes.CLOSE_TERMINAL
      });
      updateRoute(getState);
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
    edgeId,
    type: ActionTypes.LEAVE_EDGE
  };
}

export function leaveNode(nodeId) {
  return {
    nodeId,
    type: ActionTypes.LEAVE_NODE
  };
}

export function receiveControlError(nodeId, err) {
  return {
    error: err,
    nodeId,
    type: ActionTypes.DO_CONTROL_ERROR
  };
}

export function receiveControlSuccess(nodeId) {
  return {
    nodeId,
    type: ActionTypes.DO_CONTROL_SUCCESS
  };
}

export function receiveNodeDetails(details, requestTimestamp) {
  return {
    details,
    requestTimestamp,
    type: ActionTypes.RECEIVE_NODE_DETAILS
  };
}

export function receiveNodesDelta(delta) {
  return (dispatch, getState) => {
    if (!isPausedSelector(getState())) {
      // Allow css-animation to run smoothly by scheduling it to run on the
      // next tick after any potentially expensive canvas re-draws have been
      // completed.
      setTimeout(() => dispatch({ type: ActionTypes.SET_RECEIVED_NODES_DELTA }), 0);

      // When moving in time, we will consider the transition complete
      // only when the first batch of nodes delta has been received. We
      // do that because we want to keep the previous state blurred instead
      // of transitioning over an empty state like when switching topologies.
      if (getState().get('timeTravelTransitioning')) {
        dispatch({ type: ActionTypes.FINISH_TIME_TRAVEL_TRANSITION });
      }

      const hasChanges = delta.add || delta.update || delta.remove || delta.reset;
      if (hasChanges) {
        dispatch({
          delta,
          type: ActionTypes.RECEIVE_NODES_DELTA
        });
      }
    }
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

export function receiveNodes(nodes) {
  return {
    nodes,
    type: ActionTypes.RECEIVE_NODES,
  };
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

export function receiveNodesForTopology(nodes, topologyId) {
  return {
    nodes,
    topologyId,
    type: ActionTypes.RECEIVE_NODES_FOR_TOPOLOGY
  };
}

export function receiveTopologies(topologies) {
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

export function receiveControlNodeRemoved(nodeId) {
  return (dispatch, getState) => {
    dispatch({
      nodeId,
      type: ActionTypes.RECEIVE_CONTROL_NODE_REMOVED
    });
    updateRoute(getState);
  };
}

export function receiveControlPipeFromParams(pipeId, rawTty, resizeTtyControl) {
  // TODO add nodeId
  return {
    pipeId,
    rawTty,
    resizeTtyControl,
    type: ActionTypes.RECEIVE_CONTROL_PIPE
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

export function receiveControlPipeStatus(pipeId, status) {
  return {
    pipeId,
    status,
    type: ActionTypes.RECEIVE_CONTROL_PIPE_STATUS
  };
}

export function receiveError(errorUrl) {
  return {
    errorUrl,
    type: ActionTypes.RECEIVE_ERROR
  };
}

export function receiveNotFound(nodeId, requestTimestamp) {
  return {
    nodeId,
    requestTimestamp,
    type: ActionTypes.RECEIVE_NOT_FOUND,
  };
}

export function setContrastMode(enabled) {
  return (dispatch, getState) => {
    dispatch({
      enabled,
      type: ActionTypes.TOGGLE_CONTRAST_MODE,
    });
    updateRoute(getState);
  };
}

export function getTopologiesWithInitialPoll() {
  return (dispatch, getState) => {
    getTopologies(getState, dispatch, true);
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

export function resetLocalViewState() {
  return (dispatch) => {
    dispatch({type: ActionTypes.RESET_LOCAL_VIEW_STATE});
    clearStoredViewState();
    // eslint-disable-next-line prefer-destructuring
    window.location.href = window.location.href.split('#')[0];
  };
}

export function toggleTroubleshootingMenu(ev) {
  if (ev) { ev.preventDefault(); ev.stopPropagation(); }
  return {
    type: ActionTypes.TOGGLE_TROUBLESHOOTING_MENU
  };
}

export function changeInstance() {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.CHANGE_INSTANCE
    });
    updateRoute(getState);
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

export function setMonitorState(monitor) {
  return {
    monitor,
    type: ActionTypes.MONITOR_STATE
  };
}

export function setStoreViewState(storeViewState) {
  return {
    storeViewState,
    type: ActionTypes.SET_STORE_VIEW_STATE
  };
}

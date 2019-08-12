import ActionTypes from '../constants/action-types';
import { saveGraph } from '../utils/file-utils';
import { clearStoredViewState, updateRoute } from '../utils/router-utils';
import { isPausedSelector } from '../selectors/time-travel';
import {
  nextPinnedMetricTypeSelector,
  previousPinnedMetricTypeSelector,
} from '../selectors/node-metric';
import { isResourceViewModeSelector } from '../selectors/topology';

import {
  GRAPH_VIEW_MODE,
  TABLE_VIEW_MODE,
} from '../constants/naming';


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

export function blurSearch() {
  return { type: ActionTypes.BLUR_SEARCH };
}

export function clickBackground() {
  return (dispatch, getState) => {
    dispatch({
      type: ActionTypes.CLICK_BACKGROUND
    });
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

export function receiveNodes(nodes) {
  return {
    nodes,
    type: ActionTypes.RECEIVE_NODES,
  };
}

export function receiveNodesForTopology(nodes, topologyId) {
  return {
    nodes,
    topologyId,
    type: ActionTypes.RECEIVE_NODES_FOR_TOPOLOGY
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

export function resetLocalViewState() {
  return (dispatch) => {
    dispatch({ type: ActionTypes.RESET_LOCAL_VIEW_STATE });
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

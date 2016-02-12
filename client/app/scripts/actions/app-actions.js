import debug from 'debug';

import AppDispatcher from '../dispatcher/app-dispatcher';
import ActionTypes from '../constants/action-types';
import { saveGraph } from '../utils/file-utils';
import { modulo } from '../utils/math-utils';
import { updateRoute } from '../utils/router-utils';
import { bufferDeltaUpdate, resumeUpdate,
  resetUpdateBuffer } from '../utils/update-buffer-utils';
import { doControlRequest, getNodesDelta, getNodeDetails,
  getTopologies, deletePipe, getQueryData } from '../utils/web-api-utils';
import AppStore from '../stores/app-store';

const log = debug('scope:app-actions');

export function showHelp() {
  AppDispatcher.dispatch({type: ActionTypes.SHOW_HELP});
}

export function hideHelp() {
  AppDispatcher.dispatch({type: ActionTypes.HIDE_HELP});
}

export function toggleHelp() {
  if (AppStore.getShowingHelp()) {
    hideHelp();
  } else {
    showHelp();
  }
}

export function selectMetric(metricId) {
  AppDispatcher.dispatch({
    type: ActionTypes.SELECT_METRIC,
    metricId
  });
}

export function pinNextMetric(delta) {
  const metrics = AppStore.getAvailableCanvasMetrics().map(m => m.get('id'));
  const currentIndex = metrics.indexOf(AppStore.getSelectedMetric());
  const nextIndex = modulo(currentIndex + delta, metrics.count());
  const nextMetric = metrics.get(nextIndex);

  AppDispatcher.dispatch({
    type: ActionTypes.PIN_METRIC,
    metricId: nextMetric,
  });
  updateRoute();
}

export function pinMetric(metricId) {
  AppDispatcher.dispatch({
    type: ActionTypes.PIN_METRIC,
    metricId,
  });
  updateRoute();
}

export function unpinMetric() {
  AppDispatcher.dispatch({
    type: ActionTypes.UNPIN_METRIC,
  });
  updateRoute();
}

export function changeTopologyOption(option, value, topologyId) {
  AppDispatcher.dispatch({
    type: ActionTypes.CHANGE_TOPOLOGY_OPTION,
    topologyId,
    option,
    value
  });
  updateRoute();
  // update all request workers with new options
  resetUpdateBuffer();
  getTopologies(
    AppStore.getActiveTopologyOptions()
  );
  getNodesDelta(
    AppStore.getCurrentTopologyUrl(),
    AppStore.getActiveTopologyOptions()
  );
  getNodeDetails(
    AppStore.getTopologyUrlsById(),
    AppStore.getNodeDetails()
  );
}

export function clickBackground() {
  AppDispatcher.dispatch({
    type: ActionTypes.CLICK_BACKGROUND
  });
  updateRoute();
}

export function clickCloseDetails(nodeId) {
  AppDispatcher.dispatch({
    type: ActionTypes.CLICK_CLOSE_DETAILS,
    nodeId
  });
  updateRoute();
}

export function clickCloseTerminal(pipeId, closePipe) {
  AppDispatcher.dispatch({
    type: ActionTypes.CLICK_CLOSE_TERMINAL,
    pipeId
  });
  if (closePipe) {
    deletePipe(pipeId);
  }
  updateRoute();
}

export function clickDownloadGraph() {
  saveGraph();
}

export function clickForceRelayout() {
  AppDispatcher.dispatch({
    type: ActionTypes.CLICK_FORCE_RELAYOUT
  });
}

export function clickNode(nodeId, label, origin) {
  AppDispatcher.dispatch({
    type: ActionTypes.CLICK_NODE,
    origin,
    label,
    nodeId
  });
  updateRoute();
  getNodeDetails(
    AppStore.getTopologyUrlsById(),
    AppStore.getNodeDetails()
  );
}

export function clickPauseUpdate() {
  AppDispatcher.dispatch({
    type: ActionTypes.CLICK_PAUSE_UPDATE
  });
}

export function clickRelative(nodeId, topologyId, label, origin) {
  AppDispatcher.dispatch({
    type: ActionTypes.CLICK_RELATIVE,
    label,
    origin,
    nodeId,
    topologyId
  });
  updateRoute();
  getNodeDetails(
    AppStore.getTopologyUrlsById(),
    AppStore.getNodeDetails()
  );
}

export function clickResumeUpdate() {
  AppDispatcher.dispatch({
    type: ActionTypes.CLICK_RESUME_UPDATE
  });
  resumeUpdate();
}

export function clickShowTopologyForNode(topologyId, nodeId) {
  AppDispatcher.dispatch({
    type: ActionTypes.CLICK_SHOW_TOPOLOGY_FOR_NODE,
    topologyId,
    nodeId
  });
  updateRoute();
  resetUpdateBuffer();
  getNodesDelta(
    AppStore.getCurrentTopologyUrl(),
    AppStore.getActiveTopologyOptions()
  );
}

export function clickTopology(topologyId) {
  AppDispatcher.dispatch({
    type: ActionTypes.CLICK_TOPOLOGY,
    topologyId
  });
  updateRoute();
  resetUpdateBuffer();
  getNodesDelta(
    AppStore.getCurrentTopologyUrl(),
    AppStore.getActiveTopologyOptions()
  );
}

export function openWebsocket() {
  AppDispatcher.dispatch({
    type: ActionTypes.OPEN_WEBSOCKET
  });
}

export function clearControlError(nodeId) {
  AppDispatcher.dispatch({
    type: ActionTypes.CLEAR_CONTROL_ERROR,
    nodeId
  });
}

export function closeWebsocket() {
  AppDispatcher.dispatch({
    type: ActionTypes.CLOSE_WEBSOCKET
  });
}

export function doControl(nodeId, control) {
  AppDispatcher.dispatch({
    type: ActionTypes.DO_CONTROL,
    nodeId
  });
  doControlRequest(nodeId, control);
}

export function enterEdge(edgeId) {
  AppDispatcher.dispatch({
    type: ActionTypes.ENTER_EDGE,
    edgeId
  });
}

export function enterNode(nodeId) {
  AppDispatcher.dispatch({
    type: ActionTypes.ENTER_NODE,
    nodeId
  });
}

export function hitEsc() {
  const controlPipe = AppStore.getControlPipe();
  if (AppStore.getShowingHelp()) {
    hideHelp();
  } else if (controlPipe && controlPipe.get('status') === 'PIPE_DELETED') {
    AppDispatcher.dispatch({
      type: ActionTypes.CLICK_CLOSE_TERMINAL,
      pipeId: controlPipe.get('id')
    });
    updateRoute();
    // Don't deselect node on ESC if there is a controlPipe (keep terminal open)
  } else if (AppStore.getMetricQueries()) {
    AppDispatcher.dispatch({
      type: ActionTypes.CLICK_CLOSE_METRICS,
    });
    updateRoute();
  } else if (AppStore.getTopCardNodeId() && !controlPipe) {
    AppDispatcher.dispatch({ type: ActionTypes.DESELECT_NODE });
    updateRoute();
  }
}

export function leaveEdge(edgeId) {
  AppDispatcher.dispatch({
    type: ActionTypes.LEAVE_EDGE,
    edgeId
  });
}

export function leaveNode(nodeId) {
  AppDispatcher.dispatch({
    type: ActionTypes.LEAVE_NODE,
    nodeId
  });
}

export function receiveControlError(nodeId, err) {
  AppDispatcher.dispatch({
    type: ActionTypes.DO_CONTROL_ERROR,
    nodeId,
    error: err
  });
}

export function receiveControlSuccess(nodeId) {
  AppDispatcher.dispatch({
    type: ActionTypes.DO_CONTROL_SUCCESS,
    nodeId
  });
}

export function receiveNodeDetails(details) {
  AppDispatcher.dispatch({
    type: ActionTypes.RECEIVE_NODE_DETAILS,
    details
  });
}

export function receiveNodesDelta(delta) {
  if (AppStore.isUpdatePaused()) {
    bufferDeltaUpdate(delta);
  } else {
    AppDispatcher.dispatch({
      type: ActionTypes.RECEIVE_NODES_DELTA,
      delta
    });
  }
}


export function receiveTopologies(topologies) {
  AppDispatcher.dispatch({
    type: ActionTypes.RECEIVE_TOPOLOGIES,
    topologies
  });
  getNodesDelta(
    AppStore.getCurrentTopologyUrl(),
    AppStore.getActiveTopologyOptions()
  );
  getNodeDetails(
    AppStore.getTopologyUrlsById(),
    AppStore.getNodeDetails()
  );
}

export function receiveApiDetails(apiDetails) {
  AppDispatcher.dispatch({
    type: ActionTypes.RECEIVE_API_DETAILS,
    hostname: apiDetails.hostname,
    version: apiDetails.version
  });
}

export function receiveControlPipeFromParams(pipeId, rawTty) {
  // TODO add nodeId
  AppDispatcher.dispatch({
    type: ActionTypes.RECEIVE_CONTROL_PIPE,
    pipeId,
    rawTty
  });
}

export function receiveQueryData(queryId, data) {
  AppDispatcher.dispatch({
    type: ActionTypes.RECEIVE_QUERY_DATA,
    queryId,
    data
  });
}

export function addMetric(nodeId, nodeTopologyId, metricId) {
  const queryId = [nodeId, metricId].join(',');
  AppDispatcher.dispatch({
    type: ActionTypes.SHOW_METRICS_WINDOW,
    nodeId, metricId, queryId
  });
  updateRoute();
  getQueryData(nodeTopologyId, AppStore.getTopologyUrlsById(),
               nodeId, metricId, queryId);
}

export function selectMetric(nodeId, metricId) {
  AppDispatcher.dispatch({
    type: ActionTypes.SELECT_METRIC,
    nodeId: nodeId,
    metricId: metricId
  });
  updateRoute();
}

export function clickCloseMetrics() {
  AppDispatcher.dispatch({
    type: ActionTypes.CLICK_CLOSE_METRICS,
  });
  updateRoute();
}

export function receiveControlPipe(pipeId, nodeId, rawTty) {
  if (nodeId !== AppStore.getTopCardNodeId()) {
    log('Node was deselected before we could set up control!');
    deletePipe(pipeId);
    return;
  }

  const controlPipe = AppStore.getControlPipe();
  if (controlPipe && controlPipe.get('id') !== pipeId) {
    deletePipe(controlPipe.get('id'));
  }

  AppDispatcher.dispatch({
    type: ActionTypes.RECEIVE_CONTROL_PIPE,
    nodeId,
    pipeId,
    rawTty
  });

  updateRoute();
}

export function receiveControlPipeStatus(pipeId, status) {
  AppDispatcher.dispatch({
    type: ActionTypes.RECEIVE_CONTROL_PIPE_STATUS,
    pipeId,
    status
  });
}

export function receiveError(errorUrl) {
  AppDispatcher.dispatch({
    errorUrl,
    type: ActionTypes.RECEIVE_ERROR
  });
}

export function receiveNotFound(nodeId) {
  AppDispatcher.dispatch({
    nodeId,
    type: ActionTypes.RECEIVE_NOT_FOUND
  });
}

export function route(state) {
  AppDispatcher.dispatch({
    state,
    type: ActionTypes.ROUTE_TOPOLOGY
  });
  getTopologies(
    AppStore.getActiveTopologyOptions()
  );
  getNodesDelta(
    AppStore.getCurrentTopologyUrl(),
    AppStore.getActiveTopologyOptions()
  );
  getNodeDetails(
    AppStore.getTopologyUrlsById(),
    AppStore.getNodeDetails()
  );
}

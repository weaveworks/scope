import debug from 'debug';

import AppDispatcher from '../dispatcher/app-dispatcher';
import ActionTypes from '../constants/action-types';
import { saveGraph } from '../utils/file-utils';
import { updateRoute } from '../utils/router-utils';
import { bufferDeltaUpdate, resumeUpdate,
  resetUpdateBuffer } from '../utils/update-buffer-utils';
import { doControlRequest, getNodesDelta, getNodeDetails,
  getTopologies, deletePipe } from '../utils/web-api-utils';
import AppStore from '../stores/app-store';

const log = debug('scope:app-actions');

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
  if (controlPipe && controlPipe.status === 'PIPE_DELETED') {
    AppDispatcher.dispatch({
      type: ActionTypes.CLICK_CLOSE_TERMINAL,
      pipeId: controlPipe.id
    });
    updateRoute();
    // Don't deselect node on ESC if there is a controlPipe (keep terminal open)
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

export function receiveControlPipe(pipeId, nodeId, rawTty) {
  if (nodeId !== AppStore.getTopCardNodeId()) {
    log('Node was deselected before we could set up control!');
    deletePipe(pipeId);
    return;
  }

  const controlPipe = AppStore.getControlPipe();
  if (controlPipe && controlPipe.id !== pipeId) {
    deletePipe(controlPipe.id);
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

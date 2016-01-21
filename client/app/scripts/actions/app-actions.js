import debug from 'debug';

import AppDispatcher from '../dispatcher/app-dispatcher';
import ActionTypes from '../constants/action-types';
import { updateRoute } from '../utils/router-utils';
import { doControlRequest, getNodesDelta, getNodeDetails,
  getTopologies, deletePipe } from '../utils/web-api-utils';
import AppStore from '../stores/app-store';

const log = debug('scope:app-actions');

export function changeTopologyOption(option, value, topologyId) {
  AppDispatcher.dispatch({
    type: ActionTypes.CHANGE_TOPOLOGY_OPTION,
    topologyId: topologyId,
    option: option,
    value: value
  });
  updateRoute();
  // update all request workers with new options
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
    pipeId: pipeId
  });
  if (closePipe) {
    deletePipe(pipeId);
  }
  updateRoute();
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

export function clickShowTopologyForNode(topologyId, nodeId) {
  AppDispatcher.dispatch({
    type: ActionTypes.CLICK_SHOW_TOPOLOGY_FOR_NODE,
    topologyId,
    nodeId
  });
  updateRoute();
  getNodesDelta(
    AppStore.getCurrentTopologyUrl(),
    AppStore.getActiveTopologyOptions()
  );
}

export function clickTopology(topologyId) {
  AppDispatcher.dispatch({
    type: ActionTypes.CLICK_TOPOLOGY,
    topologyId: topologyId
  });
  updateRoute();
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
    nodeId: nodeId
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
    nodeId: nodeId
  });
  doControlRequest(nodeId, control);
}

export function enterEdge(edgeId) {
  AppDispatcher.dispatch({
    type: ActionTypes.ENTER_EDGE,
    edgeId: edgeId
  });
}

export function enterNode(nodeId) {
  AppDispatcher.dispatch({
    type: ActionTypes.ENTER_NODE,
    nodeId: nodeId
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
    // Dont deselect node on ESC if there is a controlPipe (keep terminal open)
  } else if (AppStore.getTopCardNodeId() && !controlPipe) {
    AppDispatcher.dispatch({type: ActionTypes.DESELECT_NODE});
    updateRoute();
  }
}

export function leaveEdge(edgeId) {
  AppDispatcher.dispatch({
    type: ActionTypes.LEAVE_EDGE,
    edgeId: edgeId
  });
}

export function leaveNode(nodeId) {
  AppDispatcher.dispatch({
    type: ActionTypes.LEAVE_NODE,
    nodeId: nodeId
  });
}

export function receiveControlError(nodeId, err) {
  AppDispatcher.dispatch({
    type: ActionTypes.DO_CONTROL_ERROR,
    nodeId: nodeId,
    error: err
  });
}

export function receiveControlSuccess(nodeId) {
  AppDispatcher.dispatch({
    type: ActionTypes.DO_CONTROL_SUCCESS,
    nodeId: nodeId
  });
}

export function receiveNodeDetails(details) {
  AppDispatcher.dispatch({
    type: ActionTypes.RECEIVE_NODE_DETAILS,
    details: details
  });
}

export function receiveNodesDelta(delta) {
  AppDispatcher.dispatch({
    type: ActionTypes.RECEIVE_NODES_DELTA,
    delta: delta
  });
}

export function receiveTopologies(topologies) {
  AppDispatcher.dispatch({
    type: ActionTypes.RECEIVE_TOPOLOGIES,
    topologies: topologies
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
    pipeId: pipeId,
    rawTty: rawTty
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
    nodeId: nodeId,
    pipeId: pipeId,
    rawTty: rawTty
  });

  updateRoute();
}

export function receiveControlPipeStatus(pipeId, status) {
  AppDispatcher.dispatch({
    type: ActionTypes.RECEIVE_CONTROL_PIPE_STATUS,
    pipeId: pipeId,
    status: status
  });
}

export function receiveError(errorUrl) {
  AppDispatcher.dispatch({
    errorUrl: errorUrl,
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
    state: state,
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

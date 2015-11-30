import AppDispatcher from '../dispatcher/app-dispatcher';
import ActionTypes from '../constants/action-types';

import { updateRoute } from '../utils/router-utils';
import { doControl as doControlRequest, getNodesDelta, getNodeDetails, getTopologies } from '../utils/web-api-utils';
import AppStore from '../stores/app-store';

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
    AppStore.getCurrentTopologyUrl(),
    AppStore.getSelectedNodeId()
  );
}

export function clickCloseDetails() {
  AppDispatcher.dispatch({
    type: ActionTypes.CLICK_CLOSE_DETAILS
  });
  updateRoute();
}

export function clickNode(nodeId) {
  AppDispatcher.dispatch({
    type: ActionTypes.CLICK_NODE,
    nodeId: nodeId
  });
  updateRoute();
  getNodeDetails(
    AppStore.getCurrentTopologyUrl(),
    AppStore.getSelectedNodeId()
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

export function clearControlError() {
  AppDispatcher.dispatch({
    type: ActionTypes.CLEAR_CONTROL_ERROR
  });
}

export function closeWebsocket() {
  AppDispatcher.dispatch({
    type: ActionTypes.CLOSE_WEBSOCKET
  });
}

export function doControl(probeId, nodeId, control) {
  AppDispatcher.dispatch({
    type: ActionTypes.DO_CONTROL
  });
  doControlRequest(
    probeId,
    nodeId,
    control
  );
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
  AppDispatcher.dispatch({
    type: ActionTypes.HIT_ESC_KEY
  });
  updateRoute();
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

export function receiveControlError(err) {
  AppDispatcher.dispatch({
    type: ActionTypes.DO_CONTROL_ERROR,
    error: err
  });
}

export function receiveControlSuccess() {
  AppDispatcher.dispatch({
    type: ActionTypes.DO_CONTROL_SUCCESS
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
    AppStore.getCurrentTopologyUrl(),
    AppStore.getSelectedNodeId()
  );
}

export function receiveApiDetails(apiDetails) {
  AppDispatcher.dispatch({
    type: ActionTypes.RECEIVE_API_DETAILS,
    version: apiDetails.version
  });
}

export function receiveError(errorUrl) {
  AppDispatcher.dispatch({
    errorUrl: errorUrl,
    type: ActionTypes.RECEIVE_ERROR
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
    AppStore.getCurrentTopologyUrl(),
    AppStore.getSelectedNodeId()
  );
}

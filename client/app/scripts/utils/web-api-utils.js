import debug from 'debug';
import reqwest from 'reqwest';
import { defaults } from 'lodash';
import { Map as makeMap, List } from 'immutable';
import stableStringify from 'json-stable-stringify';

import {
  receiveError,
  receiveNodeDetails,
  receiveNotFound, receiveNodesForTopology, receiveNodes,
} from '../actions/app-actions';

import { getCurrentTopologyUrl } from './topology-utils';
import { layersTopologyIdsSelector } from '../selectors/resource-view/layout';
import { activeTopologyOptionsSelector } from '../selectors/topology';

const log = debug('scope:web-api-utils');

const updateFrequency = '5s';
const csrfToken = (() => {
  // Check for token at window level or parent level (for iframe);
  /* eslint-disable no-underscore-dangle */
  const token = typeof window !== 'undefined'
    ? window.__WEAVEWORKS_CSRF_TOKEN || window.parent.__WEAVEWORKS_CSRF_TOKEN
    : null;
  /* eslint-enable no-underscore-dangle */
  if (!token || token === '$__CSRF_TOKEN_PLACEHOLDER__') {
    // Authfe did not replace the token in the static html.
    return null;
  }

  return token;
})();

export function buildUrlQuery(params = makeMap(), state = null) {
  // Attach the time travel timestamp to every request to the backend.
  if (state) {
    params = params.set('timestamp', state.get('pausedAt'));
  }

  // Ignore the entries with values `null` or `undefined`.
  return params.map((value, param) => {
    if (value === undefined || value === null) return null;
    if (List.isList(value)) {
      value = value.join(',');
    }
    return `${param}=${value}`;
  }).filter(s => s).join('&');
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

// TODO: This helper should probably be passed by the 'user' of Scope,
// i.e. in this case Weave Cloud, rather than being hardcoded here.
export function getApiPath(pathname = window.location.pathname) {
  if (process.env.SCOPE_API_PREFIX) {
    // Extract the instance name (pathname in WC context is of format '/:instanceId/explore').
    const instanceId = pathname.split('/')[1];
    return basePath(`${process.env.SCOPE_API_PREFIX}/app/${instanceId}`);
  }

  return basePath(pathname);
}

export function getReportUrl(timestamp) {
  return `${getApiPath()}/api/report?${buildUrlQuery(makeMap({ timestamp }))}`;
}

export function topologiesUrl(state) {
  const activeTopologyOptions = activeTopologyOptionsSelector(state);
  const optionsQuery = buildUrlQuery(activeTopologyOptions, state);
  return `${getApiPath()}/api/topology?${optionsQuery}`;
}

export function getWebsocketUrl(host = window.location.host, pathname = window.location.pathname) {
  const wsProto = window.location.protocol === 'https:' ? 'wss' : 'ws';
  return `${wsProto}://${host}${getApiPath(pathname)}`;
}

export function buildWebsocketUrl(topologyUrl, topologyOptions = makeMap(), state) {
  topologyOptions = topologyOptions.set('t', updateFrequency);
  const optionsQuery = buildUrlQuery(topologyOptions, state);
  return `${getWebsocketUrl()}${topologyUrl}/ws?${optionsQuery}`;
}

/**
  * XHR wrapper. Applies a CSRF token (if it exists) and content-type to all requests.
  * Any opts that get passed in will override the defaults.
  */
export function doRequest(opts) {
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
function getNodesForTopologies(state, dispatch, topologyIds, topologyOptions = makeMap()) {
  // fetch sequentially
  state.get('topologyUrlsById')
    .filter((_, topologyId) => topologyIds.contains(topologyId))
    .reduce(
      (sequence, topologyUrl, topologyId) => sequence
        .then(() => {
          const optionsQuery = buildUrlQuery(topologyOptions.get(topologyId), state);
          return doRequest({ url: `${getApiPath()}${topologyUrl}?${optionsQuery}` });
        })
        .then(json => dispatch(receiveNodesForTopology(json.nodes, topologyId))),
      Promise.resolve()
    );
}

export function getNodesOnce(getState, dispatch) {
  const state = getState();
  const topologyUrl = getCurrentTopologyUrl(state);
  const topologyOptions = activeTopologyOptionsSelector(state);
  const optionsQuery = buildUrlQuery(topologyOptions, state);
  const url = `${getApiPath()}${topologyUrl}?${optionsQuery}`;
  doRequest({
    error: (req) => {
      log(`Error in nodes request: ${req.responseText}`);
      dispatch(receiveError(url));
    },
    success: (res) => {
      dispatch(receiveNodes(res.nodes));
    },
    url
  });
}

/**
 * Gets nodes for all topologies (for search).
 */
export function getAllNodes(state, dispatch) {
  const topologyOptions = state.get('topologyOptions');
  const topologyIds = state.get('topologyUrlsById').keySeq();
  getNodesForTopologies(state, dispatch, topologyIds, topologyOptions);
}

/**
 * One-time update of all the nodes of topologies that appear in the current resource view.
 * TODO: Replace the one-time snapshot with periodic polling.
 */
export function getResourceViewNodesSnapshot(state, dispatch) {
  const topologyIds = layersTopologyIdsSelector(state);
  getNodesForTopologies(state, dispatch, topologyIds);
}

export function getNodeDetails(getState, dispatch) {
  const state = getState();
  const nodeMap = state.get('nodeDetails');
  const topologyUrlsById = state.get('topologyUrlsById');
  const currentTopologyId = state.get('currentTopologyId');
  const requestTimestamp = state.get('pausedAt');

  // get details for all opened nodes
  const obj = nodeMap.last();
  if (obj && topologyUrlsById.has(obj.topologyId)) {
    const topologyUrl = topologyUrlsById.get(obj.topologyId);
    let urlComponents = [getApiPath(), topologyUrl, '/', encodeURIComponent(obj.id)];

    // Only forward filters for nodes in the current topology.
    const topologyOptions = currentTopologyId === obj.topologyId
      ? activeTopologyOptionsSelector(state) : makeMap();

    const query = buildUrlQuery(topologyOptions, state);
    if (query) {
      urlComponents = urlComponents.concat(['?', query]);
    }
    const url = urlComponents.join('');

    doRequest({
      error: (err) => {
        log(`Error in node details request: ${err.responseText}`);
        // dont treat missing node as error
        if (err.status === 404) {
          dispatch(receiveNotFound(obj.id, requestTimestamp));
        } else {
          dispatch(receiveError(topologyUrl));
        }
      },
      success: (res) => {
        // make sure node is still selected
        if (nodeMap.has(res.node.id)) {
          dispatch(receiveNodeDetails(res.node, requestTimestamp));
        }
      },
      url
    });
  } else if (obj) {
    log('No details or url found for ', obj);
  }
}

export function doResizeTty(pipeId, control, cols, rows) {
  const url = `${getApiPath()}/api/control/${encodeURIComponent(control.probeId)}/`
    + `${encodeURIComponent(control.nodeId)}/${control.id}`;

  return doRequest({
    data: stableStringify({ height: rows.toString(), pipeID: pipeId, width: cols.toString() }),
    method: 'POST',
    url,
  })
    .fail((err) => {
      log(`Error resizing pipe: ${err}`);
    });
}


export function deletePipe(pipeId, dispatch) {
  const url = `${getApiPath()}/api/pipe/${encodeURIComponent(pipeId)}`;
  doRequest({
    error: (err) => {
      log(`Error closing pipe:${err}`);
      dispatch(receiveError(url));
    },
    method: 'DELETE',
    success: () => {
      log('Closed the pipe!');
    },
    url
  });
}

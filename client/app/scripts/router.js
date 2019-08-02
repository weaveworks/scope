import page from 'page';
import stableStringify from 'json-stable-stringify';
import { each } from 'lodash';

import { route } from './actions/request-actions';
import { storageGet, storageSet } from './utils/storage-utils';
import {
  decodeURL, encodeURL, isStoreViewStateEnabled, STORAGE_STATE_KEY
} from './utils/router-utils';

// Temporarily detect old topology options to avoid breaking things between releases
// Related to https://github.com/weaveworks/scope/pull/2404
function detectOldOptions(topologyOptions) {
  let bad = false;
  each(topologyOptions, (topology) => {
    each(topology, (option) => {
      if (typeof option === 'string') {
        bad = true;
      }
    });
  });
  return bad;
}

export function getRouter(initialState) {
  return (dispatch, getState) => {
    // strip any trailing '/'s.
    page.base(window.location.pathname.replace(/\/$/, ''));

    page('/', () => {
      // recover from storage state on empty URL
      const storageState = storageGet(STORAGE_STATE_KEY);
      if (storageState && isStoreViewStateEnabled(getState())) {
        const parsedState = JSON.parse(decodeURL(storageState));
        const dirtyOptions = detectOldOptions(parsedState.topologyOptions);
        if (dirtyOptions) {
          dispatch(route(initialState));
        } else {
          const mergedState = Object.assign(initialState, parsedState);
          // push storage state to URL
          window.location.hash = `!/state/${stableStringify(mergedState)}`;
          dispatch(route(mergedState));
        }
      } else {
        dispatch(route(initialState));
      }
    });

    page('/state/:state', (ctx) => {
      const state = JSON.parse(decodeURL(ctx.params.state));
      const dirtyOptions = detectOldOptions(state.topologyOptions);
      const nextState = dirtyOptions ? initialState : state;

      // back up state in storage and redirect
      if (isStoreViewStateEnabled(getState())) {
        storageSet(STORAGE_STATE_KEY, encodeURL(stableStringify(state)));
      }

      dispatch(route(nextState));
    });

    return page;
  };
}

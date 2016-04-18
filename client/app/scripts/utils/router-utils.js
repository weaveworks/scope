import page from 'page';

import { route } from '../actions/app-actions';
import AppStore from '../stores/app-store';

function shouldReplaceState(prevState, nextState) {
  // Opening a new terminal while an existing one is open.
  const terminalToTerminal = (prevState.controlPipe && nextState.controlPipe);
  // Closing a terminal.
  const closingTheTerminal = (prevState.controlPipe && !nextState.controlPipe);

  return terminalToTerminal || closingTheTerminal;
}

export function updateRoute() {
  const state = AppStore.getAppState();
  const stateUrl = encodeURIComponent(encodeURIComponent(JSON.stringify(state)));
  const dispatch = false;
  const urlStateString = window.location.hash
    .replace('#!/state/', '')
    .replace('#!/', '') || '{}';
  const prevState = JSON.parse(decodeURIComponent(decodeURIComponent(urlStateString)));

  if (shouldReplaceState(prevState, state)) {
    // Replace the top of the history rather than pushing on a new item.
    page.replace(`/state/${stateUrl}`, state, dispatch);
  } else {
    page.show(`/state/${stateUrl}`, state, dispatch);
  }
}

page('/', () => {
  updateRoute();
});

page('/state/:state', (ctx) => {
  const state = JSON.parse(ctx.params.state);
  route(state);
});

export function getRouter() {
  // strip any trailing '/'s.
  page.base(window.location.pathname.replace(/\/$/, ''));
  return page;
}

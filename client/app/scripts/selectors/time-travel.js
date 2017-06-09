import { createSelector } from 'reselect';


export const isPausedSelector = createSelector(
  [
    state => state.get('updatePausedAt')
  ],
  updatePausedAt => updatePausedAt !== null
);

export const isWebsocketQueryingCurrentSelector = createSelector(
  [
    state => state.get('websocketQueryMillisecondsInPast')
  ],
  websocketQueryMillisecondsInPast => websocketQueryMillisecondsInPast === 0
);

import { createSelector } from 'reselect';


export const isPausedSelector = createSelector(
  [
    state => state.get('updatePausedAt')
  ],
  updatePausedAt => updatePausedAt !== null
);

export const isNowSelector = createSelector(
  [
    state => state.get('timeTravelMillisecondsInPast')
  ],
  timeTravelMillisecondsInPast => timeTravelMillisecondsInPast === 0
);

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
  // true for values 0, undefined, null, etc...
  timeTravelMillisecondsInPast => !(timeTravelMillisecondsInPast > 0)
);

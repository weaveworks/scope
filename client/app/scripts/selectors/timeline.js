import { createSelector } from 'reselect';


export const isPausedSelector = createSelector(
  [
    state => state.get('updatePausedAt')
  ],
  updatePausedAt => updatePausedAt !== null
);

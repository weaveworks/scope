import { createSelector } from 'reselect';


export const isPausedSelector = createSelector(
  [
    state => state.get('pausedAt')
  ],
  pausedAt => !!pausedAt
);

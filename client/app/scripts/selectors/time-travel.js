import { createSelector } from 'reselect';


export const isPausedSelector = createSelector(
  [
    state => state.get('pausedAt')
  ],
  pausedAt => !!pausedAt
);

export const timeTravelSupportedSelector = state => state.getIn(['capabilities', 'historic_reports']);

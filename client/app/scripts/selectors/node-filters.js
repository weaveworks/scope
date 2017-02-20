import { createSelector } from 'reselect';


export const shownNodesSelector = createSelector(
  [
    state => state.get('nodes'),
  ],
  nodes => nodes.filter(node => !node.get('filtered'))
);

import { createSelector } from 'reselect';

import {
  CANVAS_MARGINS,
  DETAILS_PANEL_WIDTH,
  DETAILS_PANEL_MARGINS
} from '../constants/styles';


export const viewportWidthSelector = createSelector(
  [
    state => state.getIn(['viewport', 'width']),
  ],
  width => width - CANVAS_MARGINS.left - CANVAS_MARGINS.right
);

export const viewportHeightSelector = createSelector(
  [
    state => state.getIn(['viewport', 'height']),
  ],
  height => height - CANVAS_MARGINS.top - CANVAS_MARGINS.bottom
);

const viewportFocusWidthSelector = createSelector(
  [
    viewportWidthSelector,
  ],
  width => width - DETAILS_PANEL_WIDTH - DETAILS_PANEL_MARGINS.right
);

export const viewportFocusHorizontalCenterSelector = createSelector(
  [
    viewportFocusWidthSelector,
  ],
  width => (width / 2) + CANVAS_MARGINS.left
);

export const viewportFocusVerticalCenterSelector = createSelector(
  [
    viewportHeightSelector,
  ],
  height => (height / 2) + CANVAS_MARGINS.top
);

// The narrower dimension of the viewport, used for the circular layout.
export const viewportCircularExpanseSelector = createSelector(
  [
    viewportFocusWidthSelector,
    viewportHeightSelector,
  ],
  (width, height) => Math.min(width, height)
);

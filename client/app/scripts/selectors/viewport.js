import { createSelector } from 'reselect';

import {
  CANVAS_MARGINS,
  DETAILS_PANEL_WIDTH,
  DETAILS_PANEL_MARGINS
} from '../constants/styles';


export const canvasMarginsSelector = createSelector(
  [
    state => state.get('topologyViewMode'),
  ],
  viewMode => CANVAS_MARGINS[viewMode] || { top: 0, left: 0, right: 0, bottom: 0 }
);

export const viewportWidthSelector = createSelector(
  [
    state => state.getIn(['viewport', 'width']),
    canvasMarginsSelector,
  ],
  (width, margins) => width - margins.left - margins.right
);

export const viewportHeightSelector = createSelector(
  [
    state => state.getIn(['viewport', 'height']),
    canvasMarginsSelector,
  ],
  (height, margins) => height - margins.top - margins.bottom
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
    canvasMarginsSelector,
  ],
  (width, margins) => (width / 2) + margins.left
);

export const viewportFocusVerticalCenterSelector = createSelector(
  [
    viewportHeightSelector,
    canvasMarginsSelector,
  ],
  (height, margins) => (height / 2) + margins.top
);

// The narrower dimension of the viewport, used for the circular layout.
export const viewportCircularExpanseSelector = createSelector(
  [
    viewportFocusWidthSelector,
    viewportHeightSelector,
  ],
  (width, height) => Math.min(width, height)
);

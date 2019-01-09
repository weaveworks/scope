import { createSelector } from 'reselect';

import {
  CANVAS_MARGINS,
  DETAILS_PANEL_WIDTH,
  DETAILS_PANEL_MARGINS
} from '../constants/styles';


// Listens in on viewport height state for window height updates.
// Window height is useful as it includes the whole app viewport,
// including e.g. top nav bar in Weave Cloud which is outside of
// Scope app viewport.
export const windowHeightSelector = createSelector(
  [
    state => state.getIn(['viewport', 'height']),
  ],
  () => window.innerHeight,
);

export const canvasMarginsSelector = createSelector(
  [
    state => state.get('topologyViewMode'),
  ],
  viewMode => CANVAS_MARGINS[viewMode] || {
    bottom: 0, left: 0, right: 0, top: 0
  }
);

export const canvasWidthSelector = createSelector(
  [
    state => state.getIn(['viewport', 'width']),
    canvasMarginsSelector,
  ],
  (width, margins) => width - margins.left - margins.right
);

export const canvasHeightSelector = createSelector(
  [
    state => state.getIn(['viewport', 'height']),
    canvasMarginsSelector,
  ],
  (height, margins) => height - margins.top - margins.bottom
);

const canvasWithDetailsWidthSelector = createSelector(
  [
    canvasWidthSelector,
  ],
  width => width - DETAILS_PANEL_WIDTH - DETAILS_PANEL_MARGINS.right
);

export const canvasDetailsHorizontalCenterSelector = createSelector(
  [
    canvasWithDetailsWidthSelector,
    canvasMarginsSelector,
  ],
  (width, margins) => (width / 2) + margins.left
);

export const canvasDetailsVerticalCenterSelector = createSelector(
  [
    canvasHeightSelector,
    canvasMarginsSelector,
  ],
  (height, margins) => (height / 2) + margins.top
);

// The narrower dimension of the viewport, used for the circular layout.
export const canvasCircularExpanseSelector = createSelector(
  [
    canvasWithDetailsWidthSelector,
    canvasHeightSelector,
  ],
  (width, height) => Math.min(width, height)
);


const ZOOM_SENSITIVITY = 0.0025;
const DOM_DELTA_LINE = 1;

// See https://github.com/d3/d3-zoom/blob/807f02c7a5fe496fbd08cc3417b62905a8ce95fa/src/zoom.js
function wheelDelta(ev) {
  // Only Firefox seems to use the line unit (which we assume to
  // be 25px), otherwise the delta is already measured in pixels.
  const unitInPixels = (ev.deltaMode === DOM_DELTA_LINE ? 25 : 1);
  return -ev.deltaY * unitInPixels * ZOOM_SENSITIVITY;
}

export function zoomFactor(ev) {
  return Math.exp(wheelDelta(ev));
}

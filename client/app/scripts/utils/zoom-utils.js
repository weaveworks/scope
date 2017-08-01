
// See https://github.com/d3/d3-zoom/blob/807f02c7a5fe496fbd08cc3417b62905a8ce95fa/src/zoom.js
export function defaultWheelDelta(ev) {
  return ev.deltaY * (ev.deltaMode ? 120 : 1) * 0.002;
}

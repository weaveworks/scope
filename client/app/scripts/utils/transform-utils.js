
const applyTranslateX = ({ scaleX = 1, translateX = 0 }, x) => (x * scaleX) + translateX;
const applyTranslateY = ({ scaleY = 1, translateY = 0 }, y) => (y * scaleY) + translateY;
const applyScaleX = ({ scaleX = 1 }, width) => width * scaleX;
const applyScaleY = ({ scaleY = 1 }, height) => height * scaleY;

export const applyTransform = (transform, { width, height, x, y }) => ({
  x: applyTranslateX(transform, x),
  y: applyTranslateY(transform, y),
  width: applyScaleX(transform, width),
  height: applyScaleY(transform, height),
});

export const transformToString = ({ translateX = 0, translateY = 0, scaleX = 1, scaleY = 1 }) => (
  `translate(${translateX},${translateY}) scale(${scaleX},${scaleY})`
);

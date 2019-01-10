
const applyTranslateX = ({ scaleX = 1, translateX = 0 }, x) => (x * scaleX) + translateX;
const applyTranslateY = ({ scaleY = 1, translateY = 0 }, y) => (y * scaleY) + translateY;
const applyScaleX = ({ scaleX = 1 }, width) => width * scaleX;
const applyScaleY = ({ scaleY = 1 }, height) => height * scaleY;

export const applyTransform = (transform, {
  width = 0, height = 0, x, y
}) => ({
  height: applyScaleY(transform, height),
  width: applyScaleX(transform, width),
  x: applyTranslateX(transform, x),
  y: applyTranslateY(transform, y),
});


const inverseTranslateX = ({ scaleX = 1, translateX = 0 }, x) => (x - translateX) / scaleX;
const inverseTranslateY = ({ scaleY = 1, translateY = 0 }, y) => (y - translateY) / scaleY;
const inverseScaleX = ({ scaleX = 1 }, width) => width / scaleX;
const inverseScaleY = ({ scaleY = 1 }, height) => height / scaleY;

export const inverseTransform = (transform, {
  width = 0, height = 0, x, y
}) => ({
  height: inverseScaleY(transform, height),
  width: inverseScaleX(transform, width),
  x: inverseTranslateX(transform, x),
  y: inverseTranslateY(transform, y),
});


export const transformToString = ({
  translateX = 0, translateY = 0, scaleX = 1, scaleY = 1
}) => (
  `translate(${translateX},${translateY}) scale(${scaleX},${scaleY})`
);


export const applyTransformX = ({ scaleX = 1, translateX = 0 }, x) => (x * scaleX) + translateX;
export const applyTransformY = ({ scaleY = 1, translateY = 0 }, y) => (y * scaleY) + translateY;

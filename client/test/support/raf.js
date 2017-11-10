// https://github.com/facebook/jest/issues/4545#issuecomment-332762365
global.requestAnimationFrame = (callback) => {
  setTimeout(callback, 0);
};

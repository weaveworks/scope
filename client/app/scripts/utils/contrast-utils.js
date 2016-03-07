export const contrastModeUrl = 'contrast.html';

const contrastMode = window.location.pathname.indexOf(contrastModeUrl) > -1;

export function isContrastMode() {
  return contrastMode;
}

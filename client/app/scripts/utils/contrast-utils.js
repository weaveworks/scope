const contrastMode = window.location.pathname.indexOf('contrast') > -1;

export function isContrastMode() {
  return contrastMode;
}

export const contrastModeUrl = 'contrast.html';

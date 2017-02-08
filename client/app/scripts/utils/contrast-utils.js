import { parseHashState } from './router-utils';

export function isContrastMode() {
  return parseHashState().contrastMode;
}

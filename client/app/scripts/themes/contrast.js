import { color } from 'weaveworks-ui-components/lib/theme/selectors';

import defaultTheme from './default';

const contrastTheme = {
  ...defaultTheme,

  /* contrast overrides */
  backgroundColor: color('white'),
  backgroundDarkerColor: color('purple200'),
  backgroundDarkerSecondaryColor: color('purple200'),
  bodyBackgroundColor: color('white'),
  borderLightColor: color('gray600'),
  btnOpacityDefault: 1,
  btnOpacityDisabled: 0.4,
  btnOpacitySelected: 1,
  edgeColor: color('black'),
  edgeHighlightOpacity: 0.3,
  labelBackgroundColor: color('white'),
  linkOpacityDefault: 1,
  nodeElementsInBackgroundOpacity: 0.4,
  textColor: color('black'),
  textDarkerColor: color('black'),
  textSecondaryColor: color('black'),
  textTertiaryColor: color('black'),
};

export default contrastTheme;

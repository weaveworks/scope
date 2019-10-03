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
  edgeOpacityBlurred: 0,
  labelBackgroundColor: color('white'),
  linkOpacityDefault: 1,
  nodeBorderStrokeWidth: 0.2,
  nodeElementsInBackgroundOpacity: 0.4,
  nodeHighlightShadowOpacity: 0.4,
  nodeHighlightStrokeOpacity: 0.5,
  nodeHighlightStrokeWidth: 0.25,
  nodePseudoOpacity: 1,
  searchBorderColor: color('purple200'),
  searchBorderWidth: '2px',
  textColor: color('black'),
  textDarkerColor: color('black'),
  textSecondaryColor: color('black'),
  textTertiaryColor: color('black'),
};

export default contrastTheme;

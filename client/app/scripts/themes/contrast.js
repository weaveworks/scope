import theme from 'weaveworks-ui-components/lib/theme';
import defaultTheme from './default';

const contrastTheme = {
  ...defaultTheme,

  /* contrast overrides */
  backgroundColor: theme.colors.white,
  backgroundDarkerColor: theme.colors.purple200,
  backgroundDarkerSecondaryColor: theme.colors.purple200,
  bodyBackgroundColor: theme.colors.white,
  borderLightColor: theme.colors.gray600,
  btnOpacityDefault: 1,
  btnOpacityDisabled: 0.4,
  btnOpacitySelected: 1,
  edgeColor: theme.colors.black,
  edgeHighlightOpacity: 0.3,
  labelBackgroundColor: theme.colors.white,
  linkOpacityDefault: 1,
  nodeElementsInBackgroundOpacity: 0.4,
  textColor: theme.colors.black,
  textDarkerColor: theme.colors.black,
  textSecondaryColor: theme.colors.black,
  textTertiaryColor: theme.colors.black,
};

export default contrastTheme;

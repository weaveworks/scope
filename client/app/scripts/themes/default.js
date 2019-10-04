import theme from 'weaveworks-ui-components/lib/theme';
import { transparentize } from 'polished';

const defaultTheme = {
  backgroundColor: theme.colors.gray50,
  backgroundDarkColor: theme.colors.purple900,
  backgroundDarkerColor: theme.colors.purple100,
  backgroundDarkerSecondaryColor: theme.colors.gray50,
  baseEase: 'ease-in-out',
  bodyBackgroundColor: theme.colors.purple25,
  borderLightColor: theme.colors.purple100,
  btnOpacityDefault: 0.9,
  btnOpacityDisabled: 0.25,
  btnOpacityHover: 1,
  btnOpacitySelected: 0.9,
  detailsWindowPaddingLeft: '30px',
  detailsWindowWidth: '420px',
  edgeColor: theme.colors.purple500,
  edgeHighlightOpacity: 0.1,
  labelBackgroundColor: transparentize(0.3, theme.colors.purple25),
  linkOpacityDefault: 0.8,
  nodeElementsInBackgroundOpacity: 0.7,
  terminalHeaderHeight: '44px',
  textColor: theme.colors.purple800,
  textDarkerColor: theme.colors.purple900,
  textSecondaryColor: theme.colors.purple600,
  textTertiaryColor: theme.colors.purple400,
  timelineHeight: '55px',
};

export default defaultTheme;

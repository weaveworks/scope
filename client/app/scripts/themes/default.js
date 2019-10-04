import { color } from 'weaveworks-ui-components/lib/theme/selectors';
import { transparentize } from 'polished';

const defaultTheme = {
  backgroundColor: color('gray50'),
  backgroundDarkColor: color('purple900'),
  backgroundDarkerColor: color('purple100'),
  backgroundDarkerSecondaryColor: color('gray50'),
  baseEase: 'ease-in-out',
  bodyBackgroundColor: color('purple25'),
  borderLightColor: color('purple100'),
  btnOpacityDefault: 0.9,
  btnOpacityDisabled: 0.25,
  btnOpacityHover: 1,
  btnOpacitySelected: 0.9,
  detailsWindowPaddingLeft: '30px',
  detailsWindowWidth: '420px',
  edgeColor: color('purple500'),
  edgeHighlightOpacity: 0.1,
  labelBackgroundColor: props => transparentize(0.3, props.theme.colors.purple25),
  linkOpacityDefault: 0.8,
  nodeElementsInBackgroundOpacity: 0.7,
  terminalHeaderHeight: '44px',
  textColor: color('purple800'),
  textDarkerColor: color('purple900'),
  textSecondaryColor: color('purple600'),
  textTertiaryColor: color('purple400'),
  timelineHeight: '55px',
};

export default defaultTheme;

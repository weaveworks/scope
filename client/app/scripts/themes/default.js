import { color } from 'weaveworks-ui-components/lib/theme/selectors';
import { transparentize } from 'polished';

const defaultTheme = {
  background: '#fcc',

  backgroundColor: color('gray50'), // $background-color: $color-gray-50;
  backgroundDarkColor: color('purple900'), // $background-dark-color: $color-purple-900;
  backgroundDarkerColor: color('purple100'), // $background-darker-color: $color-purple-100;
  backgroundDarkerSecondaryColor: color('gray50'), // $background-darker-secondary-color: $color-gray-50;
  backgroundLighterColor: color('white'), // $background-lighter-color: $color-white;
  baseEase: 'ease-in-out', // $base-ease: ease-in-out;
  bodyBackgroundColor: color('purple25'), // $body-background-color: $color-purple-25;
  borderLightColor: color('purple100'), // $border-light-color: $color-purple-100;
  btnOpacityDefault: 0.9, // $btn-opacity-default: 0.9;
  btnOpacityDisabled: 0.25, // $btn-opacity-disabled: 0.25;
  btnOpacityHover: 1, // $btn-opacity-hover: 1;
  btnOpacitySelected: 0.9, // $btn-opacity-selected: 0.9;
  detailsWindowPaddingLeft: '30px', // $details-window-padding-left: 30px;
  detailsWindowWidth: '420px', // $details-window-width: 420px;
  edgeColor: color('purple500'), // $edge-color: $color-purple-500;
  edgeHighlightOpacity: 0.1, // $edge-highlight-opacity: 0.1;
  edgeOpacity: 0.5, // $edge-opacity: 0.5;
  edgeOpacityBlurred: 0.2, // $edge-opacity-blurred: 0.2;
  labelBackgroundColor: props => transparentize(0.3, props.theme.colors.purple25),
  linkOpacityDefault: 0.8, // $link-opacity-default: 0.8;
  nodeBorderStrokeWidth: 0.12, // $node-border-stroke-width: 0.12;
  nodeElementsInBackgroundOpacity: 0.7, // $node-elements-in-background-opacity: 0.7;
  nodeHighlightShadowOpacity: 0.5, // $node-highlight-shadow-opacity: 0.5;
  nodeHighlightStrokeOpacity: 0.4, // $node-highlight-stroke-opacity: 0.4;
  nodeHighlightStrokeWidth: 0.04, // $node-highlight-stroke-width: 0.04;
  nodePseudoOpacity: 0.8, // $node-pseudo-opacity: 0.8;
  nodeShadowStrokeWidth: 0.18, // $node-shadow-stroke-width: 0.18;
  nodeTextScale: 2, // $node-text-scale: 2;
  searchBorderColor: 'transparent', // $search-border-color: transparent;
  searchBorderWidth: '1px', // $search-border-width: 1px;
  terminalHeaderHeight: '44px', // $terminal-header-height: 44px;
  textColor: color('purple800'), // $text-color: $color-purple-800;
  textDarkerColor: color('purple900'), // $text-darker-color: $color-purple-900;
  textSecondaryColor: color('purple600'), // $text-secondary-color: $color-purple-600;
  textTertiaryColor: color('purple400'), // $text-tertiary-color: $color-purple-400;
  timelineHeight: '55px', // $timeline-height: 55px;
};

export default defaultTheme;

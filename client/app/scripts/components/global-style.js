import { createGlobalStyle } from 'styled-components';
import { transparentize } from 'polished';
import { borderRadius, color, fontSize } from 'weaveworks-ui-components/lib/theme/selectors';

import '@fortawesome/fontawesome-free/css/all.css';
import '@fortawesome/fontawesome-free/css/v4-shims.css';
import 'rc-slider/dist/rc-slider.css';

import ProximaNova from '../../fonts/proximanova-regular.woff';
import RobotoMono from '../../fonts/robotomono-regular.ttf';

const scopeTheme = key => props => props.theme.scope[key];

const hideable = props => `
  transition: opacity .5s ${props.theme.scope.baseEase};
`;

const palable = props => `
  transition: all .2s ${props.theme.scope.baseEase};
`;

const blinkable = props => `
  animation: blinking 1.5s infinite ${props.theme.scope.baseEase};
`;

const colorable = props => `
  transition: background-color .3s ${props.theme.scope.baseEase};
`;

/* add this class to truncate text with ellipsis, container needs width */
const truncate = () => `
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`;

const shadow2 = props => `
  box-shadow: 0 3px 10px ${transparentize(0.84, props.theme.colors.black)}, 0 3px 10px ${transparentize(0.77, props.theme.colors.black)};
`;

const btnOpacity = props => `
  ${palable(props)};
  opacity: ${props.theme.scope.btnOpacityDefault};
  &-selected {
    opacity: ${props.theme.scope.btnOpacitySelected};
  }
  &[disabled] {
    cursor: default;
    opacity: ${props.theme.scope.btnOpacityDisabled};
  }
  &:not([disabled]):hover {
    opacity: ${props.theme.scope.btnOpacityHover};
  }
`;

/* From https://stackoverflow.com/a/18294634 */
const fullyPannable = () => `
  width: 100%;
  height: 100%;
  /* stylelint-disable value-no-vendor-prefix */
  /* Grabbable */
  cursor: move; /* fallback if grab cursor is unsupported */
  cursor: grab;
  cursor: -moz-grab;
  cursor: -webkit-grab;

  &.panning {
    /* Grabbing */
    cursor: grabbing;
    cursor: -moz-grabbing;
    cursor: -webkit-grabbing;
  }
  /* stylelint-enable value-no-vendor-prefix */
`;

const overlayWrapper = props => `
  align-items: center;
  background-color: ${transparentize(0.1, props.theme.colors.purple25)};
  border-radius: ${props.theme.borderRadius.soft};
  color: ${props.theme.scope.textTertiaryColor};
  display: flex;
  font-size: ${props.theme.fontSizes.tiny};
  justify-content: center;
  padding: 5px;
  position: fixed;
  bottom: 11px;

  button {
    ${btnOpacity(props)};
    background-color: transparent;
    border: 1px solid transparent;
    border-radius: ${props.theme.borderRadius.soft};
    color: ${props.theme.scope.textSecondaryColor};
    cursor: pointer;
    line-height: 20px;
    padding: 1px 3px;
    outline: 0;

    i {
      font-size: ${props.theme.fontSizes.small};
      position: relative;
      top: 2px;
    }

    &:hover, &.selected {
      border: 1px solid ${props.theme.scope.textTertiaryColor};
    }

    &.active {
      & > * { ${blinkable(props)}; }
      border: 1px solid ${props.theme.scope.textTertiaryColor};
    }
  }
`;

const GlobalStyle = createGlobalStyle`
  /* stylelint-disable sh-waqar/declaration-use-variable */
  @font-face {
    font-family: 'proxima-nova';
    src: url(${ProximaNova});
  }

  @font-face {
    font-family: 'Roboto Mono';
    src: url(${RobotoMono});
  }
  /* stylelint-enable sh-waqar/declaration-use-variable */

  /*
  * Contain all the styles in the root div instead of having them truly
  * global, so that they don't interfere with the app they're injected into.
  */
  .scope-app, .terminal-app {

    /* Extendable classes */
    .hideable { ${hideable}; }
    .palable { ${palable}; }
    .blinkable { ${blinkable}; }
    .colorable { ${colorable}; }
    .truncate { ${truncate}; }
    .shadow-2 { ${shadow2}; }
    .btn-opacity { ${btnOpacity}; }
    .fully-pannable { ${fullyPannable}; }
    .overlay-wrapper { ${overlayWrapper}; }

    /* General styles */
    -webkit-font-smoothing: antialiased;
    background: ${scopeTheme('bodyBackgroundColor')};
    bottom: 0;
    color: ${scopeTheme('textColor')};
    font-family: ${props => props.theme.fontFamilies.regular};
    font-size: ${fontSize('small')};
    height: auto;
    left: 0;
    line-height: 150%;
    margin: 0;
    overflow: auto;
    position: fixed;
    right: 0;
    top: 0;
    width: 100%;
  
    a {
      text-decoration: none;
    }

    * {
      box-sizing: border-box;
      -webkit-tap-highlight-color: transparent;
    }
    *:before,
    *:after {
      box-sizing: border-box;
    }

    p {
      line-height: 20px;
      padding-top: 6px;
      margin-bottom: 14px;
      letter-spacing: 0;
      font-weight: 400;
      color: ${scopeTheme('textColor')};
    }

    h2 {
      font-size: ${fontSize('extraLarge')};
      line-height: 40px;
      padding-top: 8px;
      margin-bottom: 12px;
      font-weight: 400;
    }

    .browsehappy {
      margin: 0.2em 0;
      background: ${color('gray200')};
      color: ${color('black')};
      padding: 0.2em 0;
    }

    .hang-around {
      transition-delay: .5s;
    }

    .overlay {
      ${hideable};

      background-color: ${color('white')};
      position: absolute;
      width: 100%;
      height: 100%;
      top: 0;
      left: 0;
      opacity: 0;
      pointer-events: none;
      z-index: ${props => props.theme.layers.modal};

      &.faded {
        /* NOTE: Not sure if we should block the pointer events here.. */
        pointer-events: all;
        cursor: wait;
        opacity: 0.5;
      }
    }

    .hide {
      opacity: 0;
    }

    &.time-travel-open {
      .details-wrapper {
        margin-top: ${scopeTheme('timelineHeight')} + 50px;
      }
    }

    .header {
      background-color: transparentize(${color('purple25')}, 0.2);
      z-index: ${props => props.theme.layers.front};
      padding: 15px 10px 0;
      position: relative;
      width: 100%;

      .selectors {
        display: flex;
        position: relative;

        > * {
          flex: 1 1;
        }

        .logo {
          margin: -16px 0 0 8px;
          height: 64px;
          max-width: 250px;
          min-width: 0;
        }
      }
    }


    .rc-slider {
      .rc-slider-step { cursor: pointer; }
      .rc-slider-track { background-color: ${scopeTheme('textTertiaryColor')}; }
      .rc-slider-rail { background-color: ${scopeTheme('borderLightColor')}; }
      .rc-slider-handle { border-color: ${scopeTheme('textTertiaryColor')}; }
    }

    .footer {
      ${overlayWrapper};
      right: 43px;

      &-status {
        margin-right: 1em;
      }

      &-label, .pause-text {
        margin: 0 0.25em;
      }

      &-versionupdate {
        margin-right: 0.5em;
      }

      &-tools {
        display: flex;
      }

      &-icon {
        margin-left: 0.5em;
      }

      .tooltip-wrapper {
        position: relative;

        .tooltip {
          display: none;
          background-color: ${color('black')};
          position: absolute;
          color: ${color('white')};
          text-align: center;
          line-height: 22px;
          border-radius: ${borderRadius('soft')};
          font-size: ${fontSize('tiny')};
          margin-bottom: 25px;
          margin-left: -4px;
          opacity: 0.75;
          padding: 10px 20px;
          bottom: 0;
          left: 10px;
          transform: translateX(-50%);
          white-space: nowrap;
          z-index: ${props => props.theme.layers.tooltip};

          /* Adjusted from http://www.cssarrowplease.com/ */
          &:after {
            border: 6px solid transparent;
            content: ' ';
            top: 100%;
            left: 50%;
            height: 0;
            width: 0;
            position: absolute;
            border-color: transparent;
            border-top-color: ${color('black')};
            margin-left: -6px;
          }
        }

        &:hover .tooltip {
          display: block;
        }
      }
    }

    .zoomable-canvas svg {
      ${fullyPannable};
    }

    .topologies-selector {
      margin: 0 4px;
      display: flex;

      .topologies-item {
        margin: 0px 8px;

        &-label {
          font-size: ${fontSize('normal')};
        }

      }

      .topologies-sub {
        &-item {
          &-label {
            font-size: ${fontSize('small')};
          }
        }
      }

      .topologies-item-main,
      .topologies-sub-item {
        pointer-events: all;
        color: ${scopeTheme('textSecondaryColor')};
        ${btnOpacity};
        cursor: pointer;
        padding: 4px 8px;
        border-radius: ${borderRadius('soft')};
        opacity: 0.9;
        margin-bottom: 3px;
        border: 1px solid transparent;
        white-space: nowrap;

        &-active, &:hover {
          color: ${scopeTheme('textColor')};
          background-color: ${scopeTheme('backgroundDarkerColor')};
        }

        &-active {
          opacity: 0.85;
        }

        &-matched {
          border-color: ${color('blue400')};
        }

      }

      .topologies-sub-item {
        padding: 2px 8px;
      }

    }

    .nodes-chart-overlay {
      pointer-events: none;
      opacity: ${scopeTheme('nodeElementsInBackgroundOpacity')};

      &:not(.active) {
        display: none;
      }
    }

    .nodes-chart, .nodes-resources {

      &-error, &-loading {
        ${hideable};
        pointer-events: none;
        position: absolute;
        left: 50%;
        top: 50%;
        margin-left: -16.5%;
        margin-top: -275px;
        color: ${scopeTheme('textSecondaryColor')};
        width: 33%;
        height: 550px;

        display: flex;
        flex-direction: column;
        justify-content: center;

        .heading {
          font-size: ${fontSize('normal')};
        }

        &-icon {
          text-align: center;
          opacity: 0.25;
          font-size: ${props => props.theme.overlayIconSize};
        }

        li { padding-top: 5px; }
      }

      /* Make watermarks blink only if actually shown (otherwise the FF performance decreses weirdly). */
      &-loading:not(.hide) &-error-icon-container {
        ${blinkable};
      }

      &-loading {
        text-align: center;
      }

      svg {
        ${hideable};
        position: absolute;
        top: 0px;
      }

      .logo {
        display: none;
      }

      svg.exported {
        .logo {
          display: inline;
        }
      }

      text {
        font-family: ${props => props.theme.fontFamilies.regular};
        fill: ${scopeTheme('textSecondaryColor')};
      }

      .nodes-chart-elements .matched-results {
        background-color: ${scopeTheme('labelBackgroundColor')};
      }

      .edge {
        .link-none {
          fill: none;
          display: none;
        }
        .link-storage {
          fill: none;
          stroke: ${scopeTheme('edgeColor')};
          stroke-dasharray: 1, 30;
          stroke-linecap: round;
        }
        .link {
          fill: none;
          stroke: ${scopeTheme('edgeColor')};
        }
        .shadow {
          fill: none;
          stroke: ${color('blue400')};
          stroke-opacity: 0;
        }
        &.highlighted {
          .shadow {
            stroke-opacity: ${scopeTheme('edgeHighlightOpacity')};
          }
        }
      }

      .edge-marker {
        color: ${scopeTheme('edgeColor')};
        fill: ${scopeTheme('edgeColor')};
      }
    }

    .matched-results {
      text-align: center;

      &-match {
        font-size: ${fontSize('small')};

        &-wrapper {
          display: inline-block;
          margin: 1px;
          padding: 2px 4px;
          background-color: transparentize(${color('blue400')}, 0.9);
        }

        &-label {
          color: ${scopeTheme('textSecondaryColor')};
          margin-right: 0.5em;
        }
      }

      &-more {
        font-size: ${fontSize('tiny')};
        color: ${color('blue700')};
        margin-top: -2px;
      }
    }

    .details {
      &-wrapper {
        position: fixed;
        display: flex;
        z-index: ${props => props.theme.layers.toolbar};
        right: ${scopeTheme('detailsWindowPaddingLeft')};
        top: 100px;
        bottom: 48px;
        transition: transform 0.33333s cubic-bezier(0,0,0.21,1), margin-top .15s ${scopeTheme('baseEase')};
      }
    }

    .node-details {
      height: 100%;
      width: ${scopeTheme('detailsWindowWidth')};
      display: flex;
      flex-flow: column;
      margin-bottom: 12px;
      padding-bottom: 2px;
      border-radius: ${borderRadius('soft')};
      background-color: ${color('white')};
      ${shadow2};
      /* keep node-details above the terminal. */
      z-index: ${props => props.theme.layers.front};
      overflow: hidden;

      &:last-child {
        margin-bottom: 0;
      }

      &-tools-wrapper {
        position: relative;
      }


      &-tools {
        position: absolute;
        top: 6px;
        right: 8px;

        .close-details {
          position: relative;
          z-index: ${props => props.theme.layers.front};
        }

        > i {
          ${btnOpacity};
          padding: 4px 5px;
          margin-left: 2px;
          font-size: ${fontSize('normal')};
          color: ${color('white')};
          cursor: pointer;
          border: 1px solid transparent;
          border-radius: ${borderRadius('soft')};

          span {
            font-family: ${props => props.theme.fontFamilies.regular};
            font-size: ${fontSize('small')};
            margin-left: 4px;

            span {
              text-transform: capitalize;
              font-size: ${fontSize('normal')};
              margin-left: 0;
            }
          }

          &:hover {
            border-color: ${props => transparentize(0.4, props.theme.colors.white)};
          }
        }
      }

      .match {
        background-color: ${props => transparentize(0.7, props.theme.colors.blue400)};
        border: 1px solid ${color('blue400')};
      }

      &-header {
        ${colorable};

        &-wrapper {
          padding: 36px 36px 8px 36px;
        }

        &-label {
          color: ${color('white')};
          margin: 0;
          width: 348px;
          padding-top: 0;
        }

        .details-tools {
          position: absolute;
          top: 16px;
          right: 24px;
        }

        &-notavailable {
          background-color: ${scopeTheme('backgroundDarkColor')};
        }

      }

      &-relatives {
        margin-top: 4px;
        font-size: ${fontSize('normal')};
        color: ${color('white')};

        &-link {
          ${truncate};
          ${btnOpacity};
          display: inline-block;
          margin-right: 0.5em;
          cursor: pointer;
          text-decoration: underline;
          opacity: ${scopeTheme('linkOpacityDefault')};
          max-width: 12em;
        }

        &-more {
          ${btnOpacity};
          padding: 0 2px;
          cursor: pointer;
          font-size: ${fontSize('tiny')};
          font-weight: bold;
          display: inline-block;
          position: relative;
          top: -5px;
        }
      }

      &-controls {
        white-space: nowrap;
        padding: 8px 0;
        font-size: ${fontSize('small')};

        &-wrapper {
          padding: 0 36px 0 32px;
        }

        .node-control-button {
          color: ${color('white')};
        }

        &-spinner {
          ${hideable};
          color: ${color('white')};
          margin-left: 8px;
        }

        &-error {
          ${truncate};
          float: right;
          width: 55%;
          padding-top: 6px;
          text-align: left;
          color: ${color('white')};

          &-icon {
            ${blinkable};
            margin-right: 0.5em;
          }
        }
      }

      &-content {
        flex: 1;
        padding: 0 36px 0 36px;
        overflow-y: auto;

        &-loading {
          margin-top: 48px;
          text-align: center;
          font-size: ${fontSize('huge')};
          color: ${scopeTheme('textTertiaryColor')};
        }

        &-section {
          margin: 16px 0;

          &-header {
            font-size: ${fontSize('normal')};
            color: ${scopeTheme('textTertiaryColor')};
            margin-bottom: 10px;
          }
        }
      }

      &-health {

        &-wrapper {
          display: flex;
          justify-content: space-around;
          align-content: center;
          text-align: center;
          flex-wrap: wrap;
        }

        &-overflow {
          ${btnOpacity};
          flex-basis: 33%;
          display: flex;
          flex-direction: row;
          flex-wrap: wrap;
          align-items: center;
          opacity: 0.85;
          cursor: pointer;
          position: relative;
          padding-bottom: 16px;

          &-item {
            padding: 4px 8px;
            line-height: 1.2;
            flex-basis: 48%;

            &-value {
              color: ${scopeTheme('textSecondaryColor')};
              font-size: ${fontSize('normal')};
            }

            &-label {
              color: ${scopeTheme('textSecondaryColor')};
              font-size: ${fontSize('tiny')};
            }
          }
        }

        &-item {
          padding: 8px 16px;
          width: 33%;
          display: flex;
          flex-direction: column;
          flex-grow: 1;

          &-label {
            color: ${scopeTheme('textSecondaryColor')};
            font-size: ${fontSize('small')};
          }

          &-sparkline {
            margin-top: auto;
          }

          &-placeholder {
            font-size: ${fontSize('large')};
            opacity: 0.2;
            margin-bottom: 0.2em;
          }
        }

        &-link-item {
          ${btnOpacity};
          cursor: pointer;
          opacity: ${scopeTheme('linkOpacityDefault')};
          width: 33%;
          display: flex;
          color: inherit;

          .node-details-health-item {
            width: auto;
          }
        }
      }

      &-info {

        &-field {
          display: flex;
          align-items: baseline;

          &-label {
            text-align: right;
            width: 30%;
            color: ${scopeTheme('textSecondaryColor')};
            padding: 0 0.5em 0 0;
            white-space: nowrap;
            font-size: ${fontSize('small')};

            &::after {
              content: ':';
            }
          }

          &-value {
            font-size: ${fontSize('small')};
            flex: 1;
            /* Now required (from chrome 48) to get overflow + flexbox behaving: */
            min-width: 0;
            color: ${scopeTheme('textColor')};
          }
        }
      }

      &-property-list {
        &-controls {
          margin-left: -4px;
        }

        &-field {
          display: flex;
          align-items: baseline;

          &-label {
            text-align: right;
            width: 50%;
            color: ${scopeTheme('textSecondaryColor')};
            padding: 0 0.5em 0 0;
            white-space: nowrap;
            font-size: ${fontSize('small')};

            &::after {
              content: ':';
            }
          }

          &-value {
            font-size: ${fontSize('small')};
            flex: 1;
            /* Now required (from chrome 48) to get overflow + flexbox behaving: */
            min-width: 0;
            color: ${scopeTheme('textColor')};
          }
        }
      }

      &-generic-table {
        width: 100%;

        tr {
          display: flex;
          th, td {
            padding: 0 5px;
          }
        }
      }

      &-table {
        width: 100%;
        border-spacing: 0;
        /* need fixed for truncating, but that does not extend wide columns dynamically */
        table-layout: fixed;

        &-wrapper {
          margin: 24px -4px;
        }

        &-header {
          color: ${scopeTheme('textTertiaryColor')};
          font-size: ${fontSize('small')};
          text-align: right;
          padding: 0;

          .node-details-table-header-sortable {
            padding: 3px 4px;
            cursor: pointer;
          }

          &-sorted {
            color: ${scopeTheme('textSecondaryColor')};
          }

          &-sorter {
            margin: 0 0.35em;
          }

          &:first-child {
            margin-right: 0;
            text-align: left;
          }
        }

        tbody {
          position: relative;

          .min-height-constraint {
            position: absolute;
            width: 0 !important;
            opacity: 0;
            top: 0;
          }
        }

        &-node {
          font-size: ${fontSize('small')};
          line-height: 1.5;

          &:hover, &.selected {
            background-color: ${color('white')};
          }

          > * {
            padding: 0 4px;
          }

          &-link {
            ${btnOpacity};
            text-decoration: underline;
            cursor: pointer;
            opacity: ${scopeTheme('linkOpacityDefault')};
            color: ${scopeTheme('textColor')};
          }

          &-value, &-metric {
            flex: 1;
            margin-left: 0.5em;
            text-align: right;
          }

          &-metric-link {
            ${btnOpacity};
            text-decoration: underline;
            cursor: pointer;
            opacity: ${scopeTheme('linkOpacityDefault')};
            color: ${scopeTheme('textColor')};
          }

          &-value-scalar {
            /* width: 2em; */
            text-align: right;
            margin-right: 0.5em;
          }

          &-value-minor,
          &-value-unit {
            font-size: ${fontSize('small')};
            color: ${scopeTheme('textSecondaryColor')};
          }
        }
      }
    }



    .node-resources {
      &-metric-box {
        ${palable};
        cursor: pointer;
        fill: ${props => transparentize(0.6, props.theme.colors.gray600)};

        &-info {
          background-color: ${props => transparentize(0.4, props.theme.colors.white)};
          border-radius: ${borderRadius('soft')};
          cursor: inherit;
          padding: 5px;

          .wrapper {
            display: block;

            &.label { font-size: ${fontSize('small')}; }
            &.consumption { font-size: ${fontSize('tiny')}; }
          }
        }
      }

      &-layer-topology {
        background-color: ${props => transparentize(0.05, props.theme.colors.gray50)};
        border-radius: ${borderRadius('soft')};
        border: 1px solid ${color('gray200')};
        color: ${scopeTheme('textTertiaryColor')};
        font-size: ${fontSize('normal')};
        font-weight: bold;
        padding-right: 20px;
        text-align: right;
        text-transform: capitalize;
      }
    }

    /* This part sets the styles only for the 'real' node details table, not applying
    them to the nodes grid, because there we control hovering from the JS.
    NOTE: Maybe it would be nice to separate the class names between the two places
    where node tables are used - i.e. it doesn't make sense that node-details-table
    can also refer to the tables in the nodes grid. */
    .details-wrapper .node-details-table {
      &-node {
        &:hover, &.selected {
          background-color: ${color('white')};
        }
      }
    }

    .node-control-button {
      ${btnOpacity};
      padding: 6px;
      margin-left: 2px;
      font-size: ${fontSize('small')};
      color: ${scopeTheme('textSecondaryColor')};
      cursor: pointer;
      border: 1px solid transparent;
      border-radius: ${borderRadius('soft')};
      &:hover {
        border-color: ${props => transparentize(0.4, props.theme.colors.white)};
      }
      &-pending, &-pending:hover {
        opacity: 0.2;
        border-color: transparent;
        cursor: not-allowed;
      }
    }

    .terminal {

      &-app {
        display: flex;
        flex-flow: column;
      }

      &-embedded {
        position: relative;
        /* shadow of animation-wrapper is 10px, let it fit in here without being
        overflow hiddened. */
        flex: 1;
        overflow-x: hidden;
      }

      &-animation-wrapper {
        position: absolute;
        /* some room for the drop shadow. */
        top: 10px;
        left: 10px;
        bottom: 10px;
        right: 0;
        transition: transform 0.5s cubic-bezier(0.230, 1.000, 0.320, 1.000);
        ${shadow2};
      }

      &-wrapper {
        width: 100%;
        height: 100%;
        border: 0px solid ${color('black')};
        color: ${color('gray50')};
      }

      &-header {
        ${truncate};
        color: ${color('white')};
        height: ${scopeTheme('terminalHeaderHeight')};
        padding: 8px 24px;
        background-color: ${scopeTheme('textColor')};
        position: relative;
        font-size: ${fontSize('small')};
        line-height: 28px;
        border-top-left-radius: ${borderRadius('soft')};

        &-title {
          cursor: default;
        }

        &-tools {
          position: absolute;
          right: 8px;
          top: 6px;

          &-item, &-item-icon {
            ${palable};
            padding: 4px 5px;
            color: ${color('white')};
            cursor: pointer;
            opacity: 0.7;
            border: 1px solid transparent;
            border-radius: ${borderRadius('soft')};

            font-size: ${fontSize('tiny')};
            font-weight: bold;

            &:hover {
              opacity: 1;
              border-color: ${props => transparentize(0.4, props.theme.colors.white)};
            }
          }

          &-item-icon {
            font-size: ${fontSize('normal')};
          }
        }
      }

      &-embedded { .terminal-inner { top: ${scopeTheme('terminalHeaderHeight')}; } }
      &-inner {
        cursor: text;
        font-family: ${props => props.theme.fontFamilies.monospace};
        position: absolute;
        bottom: 0;
        left: 0;
        right: 0;
        top: 0;
        background-color: ${color('black')};
        padding: 8px;
        box-sizing: content-box;
        border-bottom-left-radius: ${borderRadius('soft')};

        .terminal {
          background-color: transparent !important;
        }
      }

      &-status-bar {
        font-family: ${props => props.theme.fontFamilies.regular};
        position: absolute;
        bottom: 16px;
        right: 16px;
        width: 50%;
        padding: 16px 16px;
        opacity: 0.8;
        border-radius: ${borderRadius('soft')};

        h3 {
          margin: 4px 0;
        }

        &-message {
          margin: 4px 0;
          color: ${color('white')};
        }

        .link {
          font-weight: bold;
          cursor: pointer;
          float: right;
        }
      }

      &-cursor {
        color: ${color('black')};
        background: ${color('gray50')};
      }
    }

    .terminal-inactive .terminal-cursor {
      visibility: hidden;
    }

    .metric {
      &-unit {
        padding-left: 0.25em;
      }
    }

    .show-more {
      ${btnOpacity};
      border-top: 1px dotted ${scopeTheme('borderLightColor')};
      padding: 0px 0;
      margin-top: 4px;
      text-align: right;
      cursor: pointer;
      color: ${scopeTheme('textSecondaryColor')};
      font-size: ${fontSize('small')};

      &-icon {
        color: ${scopeTheme('textTertiaryColor')};
        font-size: ${fontSize('normal')};
        position: relative;
        top: 1px;
      }
    }

    .plugins {
      margin-right: 0.5em;

      &-label {
        margin-right: 0.25em;
      }

      &-plugin {
        cursor: default;
      }

      &-plugin + &-plugin:before {
        content: ', ';
      }

      &-plugin-icon {
        top: 1px;
        position: relative;
        font-size: ${fontSize('large')};
        margin-right: 2px;
      }

      .error {
        animation: blinking 2.0s 60 ${scopeTheme('baseEase')}; /* blink for 2 minutes */
        color: ${scopeTheme('textSecondaryColor')};
      }

      &-empty {
        opacity: ${scopeTheme('textSecondaryColor')};
      }
    }

    .status {
      padding: 2px 12px;
      background-color: ${scopeTheme('bodyBackgroundColor')};
      border-radius: ${borderRadius('soft')};
      color: ${scopeTheme('textSecondaryColor')};
      display: inline-block;
      opacity: 0.9;

      &-icon {
        font-size: ${fontSize('normal')};
        position: relative;
        top: 0.125rem;
        margin-right: 0.25rem;
      }

      &.status-loading {
        animation: blinking 2.0s 150 ${scopeTheme('baseEase')}; /* keep blinking for 5 minutes */
        text-transform: none;
        color: ${scopeTheme('textColor')};
      }
    }

    .topology-option, .metric-selector, .network-selector, .view-mode-selector, .time-control {
      font-size: ${fontSize('normal')};
      color: ${scopeTheme('textSecondaryColor')};
      margin-bottom: 6px;

      &:last-child {
        margin-bottom: 0;
      }

      i {
        font-size: ${fontSize('tiny')};
        margin-left: 4px;
        color: ${color('orange500')};
      }

      &-wrapper {
        pointer-events: all;
        border-radius: ${borderRadius('soft')};
        border: 1px solid ${scopeTheme('backgroundDarkerColor')};
        display: inline-block;
        white-space: nowrap;
      }

      &-action {
        ${btnOpacity};
        padding: 3px 12px;
        cursor: pointer;
        display: inline-block;
        background-color: ${scopeTheme('backgroundColor')};

        &-selected, &:not([disabled]):hover {
          color: ${scopeTheme('textDarkerColor')};
          background-color: ${scopeTheme('backgroundDarkerColor')};
        }

        &:first-child {
          border-left: none;
          border-top-left-radius: ${borderRadius('soft')};
          border-bottom-left-radius: ${borderRadius('soft')};
        }

        &:last-child {
          border-top-right-radius: ${borderRadius('soft')};
          border-bottom-right-radius: ${borderRadius('soft')};
        }
      }
    }

    .metric-selector {
      font-size: ${fontSize('small')};
      margin-top: 6px;
    }

    .view-mode-selector, .time-control {
      margin-left: 20px;

      &-wrapper {
        pointer-events: all;
        border-color: ${scopeTheme('backgroundDarkerSecondaryColor')};
        overflow: hidden;
      }

      &:first-child,
      &:last-child {
        .view-mode-selector-action {
          border-radius: ${borderRadius('none')};
        }
      }

      &-action {
        background-color: transparent;

        &-selected, &:not([disabled]):hover {
          background-color: ${scopeTheme('backgroundDarkerColor')};
        }

        &:not(:last-child) {
          border-right: 1px solid ${scopeTheme('backgroundDarkerSecondaryColor')};
        }
      }
    }

    .time-control {
      margin-right: 20px;

      &-controls {
        align-items: center;
        justify-content: flex-end;
        display: flex;
      }

      &-spinner {
        display: inline-block;
        margin-right: 15px;
        margin-top: 3px;

        i {
          color: ${scopeTheme('textSecondaryColor')};
          font-size: ${fontSize('normal')};
        }
      }

      &-info {
        ${blinkable};
        display: block;
        margin-top: 5px;
        text-align: right;
        pointer-events: all;
        font-size: ${fontSize('small')};
      }
    }

    .topology-option {
      &-wrapper {
        display: inline-flex;
        flex-wrap: wrap;
        overflow: hidden;
        max-height: 27px;
        transition: max-height 0.5s 0s ${scopeTheme('baseEase')};

        .topology-option-action {
          flex: 1 1 auto;
          text-align: center;
        }
      }
      
        &:last-child :hover {
        height: auto;
        max-height: calc((13px * 1.5 + 3px + 3px) * 8); /* expand to display 8 rows */
        overflow: auto;
        transition: max-height 0.5s 0s ${scopeTheme('baseEase')};
      } 

      font-size: ${fontSize('small')};

      &-action {
        &-selected {
          cursor: default;
        }
      }
    }

    .view-mode-selector-wrapper, .time-control-wrapper {
      .label { margin-left: 4px; }
      i {
        margin-left: 0;
        color: ${scopeTheme('textSecondaryColor')};
      }
    }

    .network-selector-action {
      border-top: 3px solid transparent;
      border-bottom: 3px solid ${scopeTheme('backgroundDarkColor')};
    }

    .warning {
      display: inline-block;
      cursor: pointer;
      border: 1px dashed transparent;
      text-transform: none;
      border-radius: ${borderRadius('soft')};
      margin-left: 4px;

      &-wrapper {
        display: flex;
      }

      &-text {
        display: inline-block;
        color: ${scopeTheme('textSecondaryColor')};
        padding-left: 0.5em;
      }

      &-icon {
        ${btnOpacity};
      }

      &-expanded {
        margin-left: 0;
        padding: 2px 4px;
        border-color: ${scopeTheme('textTertiaryColor')};
      }

      &-expanded &-icon {
        position: relative;
        top: 4px;
        left: 2px;
      }

    }

    .sidebar {
      position: fixed;
      bottom: 11px;
      left: 11px;
      padding: 5px;
      font-size: ${fontSize('small')};
      border-radius: ${borderRadius('soft')};
      border: 1px solid transparent;
      pointer-events: none;
      max-width: 50%;
    }

    .sidebar-gridmode {
      background-color: ${color('purple50')};
      border-color: ${scopeTheme('backgroundDarkerColor')};
      opacity: 0.9;
    }

    .search {
      &-wrapper {
        margin: 0 8px;
        min-width: 160px;
        text-align: right;
      }
    }

    @keyframes blinking {
      0%, 50%, 100% {
        opacity: 1.0;
      } 25% {
        opacity: 0.5;
      }
    }

    /*   
    Help panel!
    */

    .help-panel {
      z-index: ${props => props.theme.layers.modal};
      background-color: ${color('white')};
      ${shadow2};
      display: flex;
      position: relative;

      &-wrapper {
        position: absolute;
        width: 100%;
        height: 100%;

        display: flex;
        justify-content: center;
        align-items: flex-start;
      }

      &-header {
        background-color: ${color('blue400')};
        padding: 12px 24px;
        color: ${color('white')};

        h2 {
          margin: 0;
          font-size: ${fontSize('large')};
        }
      }

      &-tools {
        position: absolute;
        top: 6px;
        right: 8px;

        i {
          ${btnOpacity};
          padding: 4px 5px;
          margin-left: 2px;
          font-size: ${fontSize('normal')};
          color: ${color('purple400')};
          cursor: pointer;
          border: 1px solid transparent;
          border-radius: ${borderRadius('soft')};

          &:hover {
            border-color: ${props => transparentize(0.4, props.theme.colors.purple400)};
          }
        }

      }

      &-main {
        display: flex;
        padding: 12px 36px 36px 36px;
        flex-direction: row;
        align-items: stretch;

        h2 {
          line-height: 150%;
          font-size: ${fontSize('large')};
          color: ${color('purple400')};
          padding: 4px 0;
          border-bottom: 1px solid ${props => transparentize(0.9, props.theme.colors.purple400)};
        }

        h3 {
          font-size: ${fontSize('normal')};
          color: ${color('purple400')};
          padding: 4px 0;
        }

        p {
          margin: 0;
        }
      }

      &-shortcuts {
        margin-right: 36px;

        &-shortcut {
          kbd {
            display: inline-block;
            padding: 3px 5px;
            font-size: ${fontSize('tiny')};
            line-height: 10px;
            color: ${color('gray600')};
            vertical-align: middle;
            background-color: ${color('white')};
            border: solid 1px ${color('gray200')};
            border-bottom-color: ${color('gray600')};
            border-radius: ${borderRadius('soft')};
            box-shadow: inset 0 -1px 0 ${color('gray600')};
          }
          div.key {
            width: 60px;
            display: inline-block;
          }
          div.label {
            display: inline-block;
          }
        }
      }

      &-search {
        margin-right: 36px;

        &-row {
          display: flex;
          flex-direction: row;

          &-term {
            flex: 1;
            color: ${scopeTheme('textSecondaryColor')};
            i {
              margin-right: 5px;
            }
          }

          &-term-label {
            flex: 1;
            b {
              color: ${scopeTheme('textSecondaryColor')};
            }
          }
        }
      }

      &-fields {
        display: flex;
        flex-direction: column;

        &-current-topology {
          color: ${color('purple400')};
        }

        &-fields {
          display: flex;
          align-items: stretch;

          &-column {
            display: flex;
            flex-direction: column;
            flex: 1;
            margin-right: 12px;

            &-content {
              overflow: auto;
              /* 160px for top and bottom margins and the rest of the help window
              is about 160px too.
              Notes: Firefox gets a bit messy if you try and bubble
              heights + overflow up (min-height issue + still doesn't work v.well),
              so this is a bit of a hack. */
              max-height: calc(100vh - 160px - 160px - 160px);
            }
          }
        }
      }
    }

    /*   
    Zoom control
    */

    .zoom-control {
      ${overlayWrapper};
      flex-direction: column;
      padding: 5px 7px 0;
      bottom: 50px;
      right: 40px;

      a:hover { border-color: transparent; }

      .rc-slider {
        margin: 5px 0;
        height: 60px;
      }
    }

    /*   
    Debug panel!
    */

    .debug-panel {
      ${shadow2};
      background-color: ${color('white')};
      top: 80px;
      position: absolute;
      padding: 10px;
      left: 10px;
      z-index: ${props => props.theme.layers.modal};

      opacity: 0.3;

      &:hover {
        opacity: 1;
      }

      table {
        display: inline-block;
        border-collapse: collapse;
        margin: 4px 2px;

        td {
          width: 10px;
          height: 10px;
        }
      }
    }

    /*   
    Nodes grid.
    */

    .nodes-grid {
      tr {
        border-radius: ${borderRadius('soft')};
      }

      &-label-minor {
        opacity: 0.7;
      }

      &-id-column {
        margin: -3px -4px;
        padding: 2px 4px;
        display: flex;
        div {
          flex: 1;
        }
      }

      .node-details-table-wrapper-wrapper {

        flex: 1;
        display: flex;
        flex-direction: row;
        width: 100%;

        .node-details-table-wrapper {
          margin: 0;
          flex: 1;
        }

        .nodes-grid-graph {
          position: relative;
          margin-top: 24px;
        }

        .node-details-table-node > * {
          padding: 3px 4px;
        }

        /* Keeping the row height fixed is important for locking the rows on hover. */
        .node-details-table-node, thead tr {
          height: 28px;
        }

        tr:nth-child(even) {
          background: ${color('gray100')};
        }

        tbody tr {
          border: 1px solid transparent;
          border-radius: ${borderRadius('soft')};
          cursor: pointer;
        }

        /* We fully control hovering of the grid rows from JS,
        because we want consistent behaviour between the
        visual and row locking logic that happens on hover. */
        tbody tr.selected, tbody tr.focused {
          background-color: ${props => transparentize(0.85, props.theme.colors.blue400)};
          border: 1px solid ${color('blue400')};
        }
      }

      .scroll-body {

        table {
          border-bottom: 1px solid ${color('gray200')};
        }

        thead, tbody tr {
          display: table;
          width: 100%;
          table-layout: fixed;
        }

        tbody:after {
          content: '';
          display: block;
          /* height of the controls so you can scroll the last row up above them
          and have a good look. */
          height: 140px;
        }

        thead {
          box-shadow: 0 4px 2px -2px ${props => transparentize(0.84, props.theme.colors.black)};
          border-bottom: 1px solid ${color('gray600')};
        }

        tbody {
          display: block;
          overflow-y: scroll;
        }
      }
    }

    .troubleshooting-menu {
      display: flex;
      position: relative;

      &-wrapper {
        height: 100%;
        width: 100%;
        align-items: center;
        display: flex;
        justify-content: center;
        position: absolute;
      }

      &-content {
        position: relative;
        background-color: ${color('white')};
        padding: 20px;
        ${shadow2};
        z-index: ${props => props.theme.layers.modal};
      }

      &-item {
        height: 40px;
        .soft { opacity: 0.6; }
      }

      button {
        border: 0;
        background-color: transparent;
        cursor: pointer;
        padding: 0;
        outline: 0;
      }

      button, a {
        color: ${color('purple900')};

        &:hover {
          color: ${scopeTheme('textColor')};
        }
      }

      i {
        width: 20px;
        text-align: center;
        margin-right: 10px;
      }

      .fa.fa-times {
        width: 25px;
      }
    }

    @media (max-width: 1330px) {
      .view-mode-selector .label { display: none; }
    }
  }


  /**
  * Copyright (c) 2014 The xterm.js authors. All rights reserved.
  * Copyright (c) 2012-2013, Christopher Jeffrey (MIT License)
  * https://github.com/chjj/term.js
  * @license MIT
  *
  * Permission is hereby granted, free of charge, to any person obtaining a copy
  * of this software and associated documentation files (the "Software"), to deal
  * in the Software without restriction, including without limitation the rights
  * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
  * copies of the Software, and to permit persons to whom the Software is
  * furnished to do so, subject to the following conditions:
  *
  * The above copyright notice and this permission notice shall be included in
  * all copies or substantial portions of the Software.
  *
  * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
  * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
  * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
  * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
  * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
  * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
  * THE SOFTWARE.
  *
  * Originally forked from (with the author's permission):
  *   Fabrice Bellard's javascript vt100 for jslinux:
  *   http://bellard.org/jslinux/
  *   Copyright (c) 2011 Fabrice Bellard
  *   The original design remains. The terminal itself
  *   has been extended to include xterm CSI codes, among
  *   other features.
  */

  /**
  *  Default styles for xterm.js
  */

  .xterm {
      font-family: ${props => props.theme.fontFamilies.monospace};
      font-feature-settings: "liga" 0;
      position: relative;
      user-select: none;
      /* stylelint-disable property-no-vendor-prefix */
      -ms-user-select: none;
      -webkit-user-select: none;
      /* stylelint-enable property-no-vendor-prefix */
  }

  .xterm.focus,
  .xterm:focus {
      outline: none;
  }

  .xterm .xterm-helpers {
      position: absolute;
      top: 0;
      /**
      * The z-index of the helpers must be higher than the canvases in order for
      * IMEs to appear on top.
      */
      /* stylelint-disable sh-waqar/declaration-use-variable */
      z-index: 10;
      /* stylelint-enable sh-waqar/declaration-use-variable */
  }

  .xterm .xterm-helper-textarea {
      /*
      * HACK: to fix IE's blinking cursor
      * Move textarea out of the screen to the far left, so that the cursor is not visible.
      */
      position: absolute;
      opacity: 0;
      left: -9999em;
      top: 0;
      width: 0;
      height: 0;
      /* stylelint-disable sh-waqar/declaration-use-variable */
      z-index: -10;
      /* stylelint-enable sh-waqar/declaration-use-variable */
      /** Prevent wrapping so the IME appears against the textarea at the correct position */
      white-space: nowrap;
      overflow: hidden;
      resize: none;
  }

  .xterm .composition-view {
      /* TODO: Composition position got messed up somewhere */
      background: ${color('black')};
      color: ${color('white')};
      display: none;
      position: absolute;
      white-space: nowrap;
      z-index: ${props => props.theme.layers.front};
  }

  .xterm .composition-view.active {
      display: block;
  }

  .xterm .xterm-viewport {
      /* On OS X this is required in order for the scroll bar to appear fully opaque */
      background-color: ${color('black')};
      overflow-y: scroll;
      cursor: default;
      position: absolute;
      right: 0;
      left: 0;
      top: 0;
      bottom: 0;
  }

  .xterm .xterm-screen {
      position: relative;
  }

  .xterm .xterm-screen canvas {
      position: absolute;
      left: 0;
      top: 0;
  }

  .xterm .xterm-scroll-area {
      visibility: hidden;
  }

  .xterm-char-measure-element {
      display: inline-block;
      visibility: hidden;
      position: absolute;
      top: 0;
      left: -9999em;
      line-height: normal;
  }

  .xterm.enable-mouse-events {
      /* When mouse events are enabled (eg. tmux), revert to the standard pointer cursor */
      cursor: default;
  }

  .xterm:not(.enable-mouse-events) {
      cursor: text;
  }

  .xterm .xterm-accessibility,
  .xterm .xterm-message {
      position: absolute;
      left: 0;
      top: 0;
      bottom: 0;
      right: 0;
      /* stylelint-disable sh-waqar/declaration-use-variable */
      z-index: 100;
      /* stylelint-enable sh-waqar/declaration-use-variable */
      color: transparent;
  }

  .xterm .live-region {
      position: absolute;
      left: -9999px;
      width: 1px;
      height: 1px;
      overflow: hidden;
  }

  .xterm-cursor-pointer {
      cursor: pointer;
  }
`;

export default GlobalStyle;

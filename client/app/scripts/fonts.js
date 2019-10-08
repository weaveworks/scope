import { createGlobalStyle } from 'styled-components';

import ProximaNova from '../fonts/proximanova-regular.woff';
import RobotoMono from '../fonts/robotomono-regular.ttf';

const Fonts = createGlobalStyle`
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
`;

export default Fonts;

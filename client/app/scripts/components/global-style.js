import { createGlobalStyle } from 'styled-components';

const GlobalStyle = createGlobalStyle`
  div {
    background: ${props => props.theme.scope.background};
  }
`;

export default GlobalStyle;

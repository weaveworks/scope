import '@babel/polyfill';
import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux';

import '../styles/main.scss';
import '../images/favicon.ico';
import configureStore from './stores/configureStore';

const store = configureStore();

function renderApp() {
  const TerminalApp = require('./components/terminal-app').default;
  ReactDOM.render(
    (
      <Provider store={store}>
        <TerminalApp />
      </Provider>
    ), document.getElementById('app')
  );
}

renderApp();
if (module.hot) {
  module.hot.accept('./components/terminal-app', renderApp);
}

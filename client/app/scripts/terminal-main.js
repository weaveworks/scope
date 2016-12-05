import 'babel-polyfill';
import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux';
import configureStore from './stores/configureStore';

require('../styles/main.less');
require('../images/favicon.ico');

const store = configureStore();

function renderApp() {
  const TerminalApp = require('./components/terminal-app').default;
  ReactDOM.render((
    <Provider store={store}>
      <TerminalApp />
    </Provider>
  ), document.getElementById('app'));
}

renderApp();
if (module.hot) {
  module.hot.accept('./components/terminal-app', renderApp);
}

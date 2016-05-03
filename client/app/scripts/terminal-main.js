require('../styles/main.less');
require('../images/favicon.ico');

import 'babel-polyfill';
import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux';

import configureStore from './stores/configureStore';
import TerminalApp from './components/terminal-app.js';

const store = configureStore();

ReactDOM.render(
  <Provider store={store}>
    <TerminalApp />
  </Provider>,
  document.getElementById('app')
);

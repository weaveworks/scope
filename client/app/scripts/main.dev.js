require('font-awesome-webpack');
require('../styles/main.less');
require('../images/favicon.ico');

import 'babel-polyfill';
import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux';

import configureStore from './stores/configureStore';
import App from './components/app';

import DevTools from './components/dev-tools';
import Immutable from 'immutable';
import installDevTools from 'immutable-devtools';
installDevTools(Immutable);

const store = configureStore();

ReactDOM.render(
  <Provider store={store}>
    <div>
      <App />
      <DevTools />
    </div>
  </Provider>,
  document.getElementById('app')
);

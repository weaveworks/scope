require('font-awesome-webpack');
require('../styles/main.less');
require('../images/favicon.ico');

import 'babel-polyfill';
import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux';
import { AppContainer } from 'react-hot-loader';
import configureStore from './stores/configureStore.dev';

import DevTools from './components/dev-tools';
import Immutable from 'immutable';
import installDevTools from 'immutable-devtools';
installDevTools(Immutable);

const store = configureStore();

function renderApp() {
  const App = require('./components/app').default;
  ReactDOM.render((
    <Provider store={store}>
      <AppContainer>
         <App />
      </AppContainer>
      <DevTools />
    </Provider>
  ), document.getElementById('app'));
}

renderApp();
if (module.hot) {
  module.hot.accept('./components/app', renderApp);
}

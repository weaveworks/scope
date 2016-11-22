require('font-awesome-webpack');
require('../styles/contrast.less');
require('../images/favicon.ico');

import 'babel-polyfill';
import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux';
import { AppContainer } from 'react-hot-loader';
import configureStore from './stores/configureStore';

const store = configureStore();

function renderApp() {
  const App = require('./components/app').default;
  ReactDOM.render((
    <Provider store={store}>
      <AppContainer>
         <App />
      </AppContainer>
    </Provider>
  ), document.getElementById('app'));
}

renderApp();
if (module.hot) {
  module.hot.accept('./components/app', renderApp);
}

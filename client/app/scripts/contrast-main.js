import 'babel-polyfill';
import 'font-awesome-webpack';
import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux';

import '../styles/contrast.scss';
import '../images/favicon.ico';
import configureStore from './stores/configureStore';

const store = configureStore();

function renderApp() {
  const App = require('./components/app').default;
  ReactDOM.render((
    <Provider store={store}>
      <App />
    </Provider>
  ), document.getElementById('app'));
}

renderApp();
if (module.hot) {
  module.hot.accept('./components/app', renderApp);
}

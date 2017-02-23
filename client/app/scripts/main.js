import 'babel-polyfill';
import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux';

import '../styles/main.scss';
import '../images/favicon.ico';
import configureStore from './stores/configureStore';
import runDemo from './utils/demo-utils';

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

if (process.env.DEMO) {
  runDemo(store.getState, store.dispatch);
}

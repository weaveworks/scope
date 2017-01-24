import 'babel-polyfill';
import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux';
import Immutable from 'immutable';
import installDevTools from 'immutable-devtools';

import '../styles/main.scss';
import '../images/favicon.ico';
import configureStore from './stores/configureStore.dev';
import DevTools from './components/dev-tools';

installDevTools(Immutable);
const store = configureStore();

function renderApp() {
  const App = require('./components/app').default;
  ReactDOM.render((
    <Provider store={store}>
      <App />
      <DevTools />
    </Provider>
  ), document.getElementById('app'));
}

renderApp();
if (module.hot) {
  module.hot.accept('./components/app', renderApp);
}

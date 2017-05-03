import { createStore, applyMiddleware, compose } from 'redux';
import thunkMiddleware from 'redux-thunk';

import DevTools from '../components/dev-tools';
import { initialState, rootReducer } from '../reducers/root';

export default function configureStore() {
  const store = createStore(
    rootReducer,
    initialState,
    compose(
      // applyMiddleware(thunkMiddleware, createLogger()),
      applyMiddleware(thunkMiddleware),
      DevTools.instrument()
    )
  );

  if (module.hot) {
    // Enable Webpack hot module replacement for reducers
    module.hot.accept('../reducers/root', () => {
      const nextRootReducer = require('../reducers/root').default;
      store.replaceReducer(nextRootReducer);
    });
  }

  return store;
}

import { createStore, applyMiddleware } from 'redux';
import thunkMiddleware from 'redux-thunk';

import { initialState, rootReducer } from '../reducers/root';

export default function configureStore() {
  return createStore(
    rootReducer,
    initialState,
    applyMiddleware(thunkMiddleware)
  );
}

import { translateUrlParams } from '../router-utils';

describe('router-utils', () => {
  it('translates a query to a view state', () => {
    // const dispatch = createSpy();
    // const state = Object.assign({}, initialState, {
    //   nodes: [{node_1: 'some_id'}]
    // });
    // window.location.search = '?node=node_1';
    // const page = getRouter(dispatch, state);
    // expect(page).toBeTruthy();
    translateUrlParams('?key=value&otherkey=othervalue', []);
  });
});

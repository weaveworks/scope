import expect from 'expect';

import { isCompatibleShape } from '../storage-utils';

describe('storage-utils', () => {
  let state;

  beforeEach(() => {
    state = {
      controlPipe: null,
      nodeDetails: [],
      topologyViewMode: 'topo',
      pinnedMetricType: 'CPU',
      pinnedSearches: [],
      searchQuery: '',
      selectedNodeId: null,
      gridSortedBy: null,
      gridSortedDesc: null,
      topologyId: 'containers',
      topologyOptions: {
        processes: {
          unconnected: 'hide'
        },
        containers: {
          system: [
            'all'
          ],
          stopped: [
            'running'
          ],
          pseudo: [
            'hide'
          ]
        }
      },
      contrastMode: false
    };
  });
  it('is ok when state has not changed', () => {
    // Same state should be ok
    expect(isCompatibleShape(state, state)).toBe(true);
  });
  it('catches state shape changes', () => {
    // State shape changed, should not be compatible;
    const changed = {
      ...state,
      topologyOptions: {
        ...state.topologyOptions,
        processes: {
          // Changes from a string to an array; simulates the actual real-world case
          unconnected: ['hide']
        }
      }
    };

    expect(isCompatibleShape(state, changed)).toBe(false);
  });
  it('ignores trivial shape differences', () => {
    const trivial = {
      ...state,
      nodeDetails: [{ a: 1, b: 2 }],
      controlPipe: { id: 123 }
    };

    expect(isCompatibleShape(state, trivial)).toBe(true);
  });
});

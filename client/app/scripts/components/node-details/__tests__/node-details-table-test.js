import React from 'react';
import TestUtils from 'react/lib/ReactTestUtils';
import { Provider } from 'react-redux';
import configureStore from '../../../stores/configureStore';

// need ES5 require to keep automocking off
const NodeDetailsTable = require('../node-details-table.js').default;

describe('NodeDetailsTable', () => {
  let nodes, columns, component;

  beforeEach(() => {
    columns = [
      { id: 'kubernetes_ip', label: 'IP' },
      { id: 'kubernetes_namespace', label: 'Namespace' },
    ];
    nodes = [
      {
        id: 'node-1',
        metadata: [
          { id: 'kubernetes_ip', label: 'IP', value: '10.244.253.24' },
          { id: 'kubernetes_namespace', label: 'Namespace', value: '10.244.253.24' },
        ]
      }, {
        id: 'node-2',
        metadata: [
          { id: 'kubernetes_ip', label: 'IP', value: '10.244.253.4' },
          { id: 'kubernetes_namespace', label: 'Namespace', value: '10.244.253.4' },
        ]
      }, {
        id: 'node-3',
        metadata: [
          { id: 'kubernetes_ip', label: 'IP', value: '10.44.253.255' },
          { id: 'kubernetes_namespace', label: 'Namespace', value: '10.44.253.255' },
        ]
      }, {
        id: 'node-4',
        metadata: [
          { id: 'kubernetes_ip', label: 'IP', value: '10.244.253.100' },
          { id: 'kubernetes_namespace', label: 'Namespace', value: '10.244.253.100' },
        ]
      },
    ];
  });

  function matchColumnValues(columnLabel, expectedValues) {
    const columnIndex = columns.findIndex(column => column.id === columnLabel);
    const values = TestUtils.scryRenderedDOMComponentsWithClass(component, 'node-details-table-node-value').map(d => d.title);
    const filteredValues = values.filter((element, index) => index % columns.length === columnIndex);
    expect(filteredValues).toEqual(expectedValues);
  }

  describe('Sorts by column', () => {
    describe('kubernetes_ip', () => {
      it('sorts ascendingly', () => {
        component = TestUtils.renderIntoDocument(
          <Provider store={configureStore()}>
            <NodeDetailsTable
              columns={columns}
              sortBy="kubernetes_ip"
              sortedDesc={false}
              nodeIdKey="id"
              nodes={nodes}
            />
          </Provider>
        );

        matchColumnValues('kubernetes_ip', ['10.44.253.255', '10.244.253.4', '10.244.253.24', '10.244.253.100']);
      });

      it('sorts descendingly', () => {
        component = TestUtils.renderIntoDocument(
          <Provider store={configureStore()}>
            <NodeDetailsTable
              columns={columns}
              sortBy="kubernetes_ip"
              sortedDesc={true}
              nodeIdKey="id"
              nodes={nodes}
            />
          </Provider>
        );

        matchColumnValues('kubernetes_ip', ['10.244.253.100', '10.244.253.24', '10.244.253.4', '10.44.253.255']);
      });
    });

    describe('kubernetes_namespace', () => {
      it('sorts ascendingly', () => {
        component = TestUtils.renderIntoDocument(
          <Provider store={configureStore()}>
            <NodeDetailsTable
              columns={columns}
              sortBy="kubernetes_namespace"
              sortedDesc={false}
              nodeIdKey="id"
              nodes={nodes}
            />
          </Provider>
        );

        matchColumnValues('kubernetes_namespace', ['10.244.253.100', '10.244.253.24', '10.244.253.4', '10.44.253.255']);
      });
    });
  });
});

import React from 'react';
import TestUtils from 'react-dom/test-utils';
import { Provider } from 'react-redux';
import configureStore from '../../../stores/configureStore';

// need ES5 require to keep automocking off
const NodeDetailsTable = require('../node-details-table.js').default;

describe('NodeDetailsTable', () => {
  let nodes;
  let columns;
  let component;

  beforeEach(() => {
    columns = [
      { dataType: 'ip', id: 'kubernetes_ip', label: 'IP' },
      { id: 'kubernetes_namespace', label: 'Namespace' },
      { dataType: 'duration', id: 'uptime', label: 'Uptime' },
    ];
    nodes = [
      {
        id: 'node-1',
        metadata: [
          { id: 'kubernetes_ip', label: 'IP', value: '10.244.253.24' },
          { id: 'kubernetes_namespace', label: 'Namespace', value: '1111' },
          {
            dataType: 'duration', id: 'uptime', label: 'Uptime', value: '1'
          },
        ]
      }, {
        id: 'node-2',
        metadata: [
          { id: 'kubernetes_ip', label: 'IP', value: '10.244.253.4' },
          { id: 'kubernetes_namespace', label: 'Namespace', value: '12' },
          {
            dataType: 'duration', id: 'uptime', label: 'Uptime', value: '4'
          },
        ]
      }, {
        id: 'node-3',
        metadata: [
          { id: 'kubernetes_ip', label: 'IP', value: '10.44.253.255' },
          { id: 'kubernetes_namespace', label: 'Namespace', value: '5' },
          {
            dataType: 'duration', id: 'uptime', label: 'Uptime', value: '30'
          },
        ]
      }, {
        id: 'node-4',
        metadata: [
          { id: 'kubernetes_ip', label: 'IP', value: '10.244.253.100' },
          { id: 'kubernetes_namespace', label: 'Namespace', value: '00000' },
          {
            dataType: 'duration', id: 'uptime', label: 'Uptime', value: '22222'
          },
        ]
      },
    ];
  });

  function matchColumnValues(columnLabel, expectedValues) {
    // Get the index of the column whose values we want to match.
    const columnIndex = columns.findIndex(column => column.id === columnLabel);
    // Get all the values rendered in the table.
    const values = TestUtils
      .scryRenderedDOMComponentsWithClass(component, 'node-details-table-node-value')
      .map(d => d.title);
    // Since we are interested only in the values that appear in the column `columnIndex`,
    // we drop the rest. As `values` are ordered by appearance in the DOM structure
    // (that is, first by row and then by column), the indexes we are interested in are of the
    // form columnIndex + n * columns.length, where n >= 0. Therefore we take only the values
    // at the index which divided by columns.length gives a reminder columnIndex.
    const filteredValues = values.filter((element, index) =>
      index % columns.length === columnIndex);
    // Array comparison
    expect(filteredValues).toEqual(expectedValues);
  }

  function clickColumn(title) {
    const node = TestUtils.scryRenderedDOMComponentsWithTag(component, 'td')
      .find(d => d.title === title);
    TestUtils.Simulate.click(node.children[0]);
  }

  describe('kubernetes_ip', () => {
    it('sorts by column', () => {
      component = TestUtils.renderIntoDocument((
        <Provider store={configureStore()}>
          <NodeDetailsTable
            columns={columns}
            sortedBy="kubernetes_ip"
            nodeIdKey="id"
            nodes={nodes}
            />
        </Provider>
      ));

      matchColumnValues('kubernetes_ip', [
        '10.44.253.255',
        '10.244.253.4',
        '10.244.253.24',
        '10.244.253.100'
      ]);
      clickColumn('IP');
      matchColumnValues('kubernetes_ip', [
        '10.244.253.100',
        '10.244.253.24',
        '10.244.253.4',
        '10.44.253.255'
      ]);
      clickColumn('IP');
      matchColumnValues('kubernetes_ip', [
        '10.44.253.255',
        '10.244.253.4',
        '10.244.253.24',
        '10.244.253.100'
      ]);
    });
  });

  describe('kubernetes_namespace', () => {
    it('sorts by column', () => {
      component = TestUtils.renderIntoDocument((
        <Provider store={configureStore()}>
          <NodeDetailsTable
            columns={columns}
            sortedBy="kubernetes_namespace"
            nodeIdKey="id"
            nodes={nodes}
            />
        </Provider>
      ));

      matchColumnValues('kubernetes_namespace', ['00000', '1111', '12', '5']);
      clickColumn('Namespace');
      matchColumnValues('kubernetes_namespace', ['5', '12', '1111', '00000']);
      clickColumn('Namespace');
      matchColumnValues('kubernetes_namespace', ['00000', '1111', '12', '5']);
    });
  });

  describe('uptime duration', () => {
    it('sorts by column', () => {
      component = TestUtils.renderIntoDocument((
        <Provider store={configureStore()}>
          <NodeDetailsTable
            columns={columns}
            sortedBy="uptime"
            nodeIdKey="id"
            nodes={nodes}
          />
        </Provider>
      ));

      matchColumnValues('uptime', ['1 second', '4 seconds', '30 seconds', '6 hours']);
      clickColumn('Uptime');
      matchColumnValues('uptime', ['6 hours', '30 seconds', '4 seconds', '1 second']);
    });
  });
});

import { fromJS, Map as makeMap } from 'immutable';

const SearchSelectors = require('../search');

describe('Search selectors', () => {
  const nodeSets = {
    someNodes: fromJS({
      n1: {
        id: 'n1',
        label: 'node label 1',
        metadata: [{
          id: 'fieldId1',
          label: 'Label 1',
          value: 'value 1'
        }],
        metrics: [{
          id: 'metric1',
          label: 'Metric 1',
          value: 1
        }]
      },
      n2: {
        id: 'n2',
        label: 'node label 2',
        metadata: [{
          id: 'fieldId2',
          label: 'Label 2',
          value: 'value 2'
        }],
        tables: [{
          id: 'metric1',
          rows: [{
            entries: {
              label: 'Label 1',
              value: 'Label Value 1'
            },
            id: 'label1'
          }, {
            entries: {
              label: 'Label 2',
              value: 'Label Value 2'
            },
            id: 'label2'
          }],
          type: 'property-list'
        }, {
          columns: [{
            id: 'a',
            label: 'A'
          }, {
            id: 'c',
            label: 'C'
          }],
          id: 'metric2',
          rows: [{
            entries: {
              a: 'xxxa',
              b: 'yyya',
              c: 'zzz1'
            },
            id: 'row1'
          }, {
            entries: {
              a: 'yyyb',
              b: 'xxxb',
              c: 'zzz2'
            },
            id: 'row2'
          }, {
            entries: {
              a: 'Value 1',
              b: 'Value 2',
              c: 'Value 3'
            },
            id: 'row3'
          }],
          type: 'multicolumn-table'
        }],
      },
    })
  };

  describe('searchNodeMatchesSelector', () => {
    const selector = SearchSelectors.searchNodeMatchesSelector;

    it('should return no matches on an empty topology', () => {
      const result = selector(fromJS({
        nodes: makeMap(),
        searchQuery: '',
      }));
      expect(result.filter(m => !m.isEmpty()).size).toEqual(0);
    });

    it('should return no matches when no query is present', () => {
      const result = selector(fromJS({
        nodes: nodeSets.someNodes,
        searchQuery: '',
      }));
      expect(result.filter(m => !m.isEmpty()).size).toEqual(0);
    });

    it('should return no matches when query matches nothing', () => {
      const result = selector(fromJS({
        nodes: nodeSets.someNodes,
        searchQuery: 'cantmatch',
      }));
      expect(result.filter(m => !m.isEmpty()).size).toEqual(0);
    });

    it('should return a matches when a query matches something', () => {
      const result = selector(fromJS({
        nodes: nodeSets.someNodes,
        searchQuery: 'value 2',
      }));
      expect(result.filter(m => !m.isEmpty()).size).toEqual(1);
    });
  });
});

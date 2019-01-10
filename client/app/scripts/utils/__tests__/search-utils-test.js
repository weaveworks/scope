import { fromJS } from 'immutable';

const SearchUtils = require('../search-utils').testable;

describe('SearchUtils', () => {
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

  describe('applyPinnedSearches', () => {
    const fun = SearchUtils.applyPinnedSearches;

    it('should not filter anything when no pinned searches present', () => {
      let nextState = fromJS({
        nodes: nodeSets.someNodes,
        pinnedSearches: []
      });
      nextState = fun(nextState);
      expect(nextState.get('nodes').filter(node => node.get('filtered')).size).toEqual(0);
    });

    it('should filter nodes if nothing matches a pinned search', () => {
      let nextState = fromJS({
        nodes: nodeSets.someNodes,
        pinnedSearches: ['cantmatch']
      });
      nextState = fun(nextState);
      expect(nextState.get('nodes').filterNot(node => node.get('filtered')).size).toEqual(0);
    });

    it('should filter nodes if nothing matches a combination of pinned searches', () => {
      let nextState = fromJS({
        nodes: nodeSets.someNodes,
        pinnedSearches: ['node label 1', 'node label 2']
      });
      nextState = fun(nextState);
      expect(nextState.get('nodes').filterNot(node => node.get('filtered')).size).toEqual(0);
    });

    it('should filter nodes that do not match a pinned searches', () => {
      let nextState = fromJS({
        nodes: nodeSets.someNodes,
        pinnedSearches: ['Label Value 1']
      });
      nextState = fun(nextState);
      expect(nextState.get('nodes').filter(node => node.get('filtered')).size).toEqual(1);
    });
  });

  describe('findNodeMatch', () => {
    const fun = SearchUtils.findNodeMatch;

    it('does not add a non-matching field', () => {
      let matches = fromJS({});
      matches = fun(
        matches, ['node1', 'field1'],
        'some value', 'some query', null, 'some label'
      );
      expect(matches.size).toBe(0);
    });

    it('adds a matching field', () => {
      let matches = fromJS({});
      matches = fun(
        matches, ['node1', 'field1'],
        'samevalue', 'samevalue', null, 'some label'
      );
      expect(matches.size).toBe(1);
      expect(matches.getIn(['node1', 'field1'])).toBeDefined();
      const {
        text, label, start, length
      } = matches.getIn(['node1', 'field1']);
      expect(text).toBe('samevalue');
      expect(label).toBe('some label');
      expect(start).toBe(0);
      expect(length).toBe(9);
    });

    it('does not add a field when the prefix does not match the label', () => {
      let matches = fromJS({});
      matches = fun(
        matches, ['node1', 'field1'],
        'samevalue', 'samevalue', 'some prefix', 'some label'
      );
      expect(matches.size).toBe(0);
    });

    it('adds a field when the prefix matches the label', () => {
      let matches = fromJS({});
      matches = fun(
        matches, ['node1', 'field1'],
        'samevalue', 'samevalue', 'prefix', 'prefixed label'
      );
      expect(matches.size).toBe(1);
    });
  });

  describe('findNodeMatchMetric', () => {
    const fun = SearchUtils.findNodeMatchMetric;

    it('does not add a non-matching field', () => {
      let matches = fromJS({});
      matches = fun(
        matches, ['node1', 'field1'],
        1, 'metric1', 'metric2', 'lt', 2
      );
      expect(matches.size).toBe(0);
    });

    it('adds a matching field', () => {
      let matches = fromJS({});
      matches = fun(
        matches, ['node1', 'field1'],
        1, 'metric1', 'metric1', 'lt', 2
      );
      expect(matches.size).toBe(1);
      expect(matches.getIn(['node1', 'field1'])).toBeDefined();
      const { metric } = matches.getIn(['node1', 'field1']);
      expect(metric).toBeTruthy();

      matches = fun(
        matches, ['node2', 'field1'],
        1, 'metric1', 'metric1', 'gt', 0
      );
      expect(matches.size).toBe(2);

      matches = fun(
        matches, ['node3', 'field1'],
        1, 'metric1', 'metric1', 'eq', 1
      );
      expect(matches.size).toBe(3);

      matches = fun(
        matches, ['node3', 'field1'],
        1, 'metric1', 'metric1', 'other', 1
      );
      expect(matches.size).toBe(3);
    });
  });

  describe('makeRegExp', () => {
    const fun = SearchUtils.makeRegExp;

    it('should make a regexp from any string', () => {
      expect(fun().source).toEqual((new RegExp()).source);
      expect(fun('que').source).toEqual((new RegExp('que')).source);
      // invalid string
      expect(fun('que[').source).toEqual((new RegExp('que\\[')).source);
    });
  });

  describe('matchPrefix', () => {
    const fun = SearchUtils.matchPrefix;

    it('returns true if the prefix matches the label', () => {
      expect(fun('label', 'prefix')).toBeFalsy();
      expect(fun('memory', 'mem')).toBeTruthy();
      expect(fun('mem', 'memory')).toBeFalsy();
      expect(fun('com.domain.label', 'label')).toBeTruthy();
      expect(fun('com.domain.Label', 'domainlabel')).toBeTruthy();
      expect(fun('com-Domain-label', 'domainlabel')).toBeTruthy();
      expect(fun('memory', 'mem.ry')).toBeTruthy();
    });
  });

  describe('parseQuery', () => {
    const fun = SearchUtils.parseQuery;

    it('should parse a metric value from a string', () => {
      expect(fun('')).toEqual(null);
      expect(fun('text')).toEqual({query: 'text'});
      expect(fun('prefix:text')).toEqual({prefix: 'prefix', query: 'text'});
      expect(fun(':text')).toEqual(null);
      expect(fun('text:')).toEqual(null);
      expect(fun('cpu > 1')).toEqual({comp: 'gt', metric: 'cpu', value: 1});
      expect(fun('cpu >')).toEqual(null);
    });
  });

  describe('parseValue', () => {
    const fun = SearchUtils.parseValue;

    it('should parse a metric value from a string', () => {
      expect(fun('1')).toEqual(1);
      expect(fun('1.34%')).toEqual(1.34);
      expect(fun('10kB')).toEqual(1024 * 10);
      expect(fun('1K')).toEqual(1024);
      expect(fun('2KB')).toEqual(2048);
      expect(fun('1MB')).toEqual(Math.pow(1024, 2));
      expect(fun('1m')).toEqual(Math.pow(1024, 2));
      expect(fun('1GB')).toEqual(Math.pow(1024, 3));
      expect(fun('1TB')).toEqual(Math.pow(1024, 4));
    });
  });

  describe('searchTopology', () => {
    const fun = SearchUtils.searchTopology;

    it('should return no matches on an empty topology', () => {
      const nodes = fromJS({});
      const matches = fun(nodes, {query: 'value'});
      expect(matches.size).toEqual(0);
    });

    it('should match on a node label', () => {
      const nodes = nodeSets.someNodes;
      let matches = fun(nodes, {query: 'node label 1'});
      expect(matches.size).toEqual(1);
      matches = fun(nodes, {query: 'node label'});
      expect(matches.size).toEqual(2);
    });

    it('should match on a metadata field', () => {
      const nodes = nodeSets.someNodes;
      const matches = fun(nodes, {query: 'value'});
      expect(matches.size).toEqual(2);
      expect(matches.getIn(['n1', 'metadata', 'fieldId1']).text).toEqual('value 1');
    });

    it('should match on a metric field', () => {
      const nodes = nodeSets.someNodes;
      const matches = fun(nodes, {comp: 'eq', metric: 'metric1', value: 1});
      expect(matches.size).toEqual(1);
      expect(matches.getIn(['n1', 'metrics', 'metric1']).metric).toBeTruthy();
    });

    it('should match on a property list value', () => {
      const nodes = nodeSets.someNodes;
      const matches = fun(nodes, {query: 'Value 1'});
      expect(matches.size).toEqual(2);
      expect(matches.getIn(['n2', 'property-lists']).size).toEqual(1);
      expect(matches.getIn(['n2', 'property-lists', 'label1']).text).toBe('Label Value 1');
    });

    it('should match on a generic table values', () => {
      const nodes = nodeSets.someNodes;
      const matches1 = fun(nodes, {query: 'xx'}).getIn(['n2', 'tables']);
      const matches2 = fun(nodes, {query: 'yy'}).getIn(['n2', 'tables']);
      const matches3 = fun(nodes, {query: 'zz'}).getIn(['n2', 'tables']);
      const matches4 = fun(nodes, {query: 'a'}).getIn(['n2', 'tables']);
      expect(matches1.size).toEqual(1);
      expect(matches2.size).toEqual(1);
      expect(matches3.size).toEqual(2);
      expect(matches4.size).toEqual(3);
      expect(matches1.get('row1_a').text).toBe('xxxa');
      expect(matches2.get('row2_a').text).toBe('yyyb');
      expect(matches3.get('row1_c').text).toBe('zzz1');
      expect(matches3.get('row2_c').text).toBe('zzz2');
      expect(matches4.get('row1_a').text).toBe('xxxa');
      expect(matches4.get('row3_a').text).toBe('Value 1');
      expect(matches4.get('row3_c').text).toBe('Value 3');
    });
  });
});

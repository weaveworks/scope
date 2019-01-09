import React from 'react';
import { connect } from 'react-redux';

import { searchableFieldsSelector } from '../selectors/search';
import { canvasMarginsSelector } from '../selectors/canvas';
import { hideHelp } from '../actions/app-actions';


const GENERAL_SHORTCUTS = [
  {key: 'esc', label: 'Close active panel'},
  {key: '/', label: 'Activate search field'},
  {key: '?', label: 'Toggle shortcut menu'},
  {key: 'g', label: 'Switch to Graph view'},
  {key: 't', label: 'Switch to Table view'},
  {key: 'r', label: 'Switch to Resources view'},
];


const CANVAS_METRIC_SHORTCUTS = [
  {key: '<', label: 'Select and pin previous metric'},
  {key: '>', label: 'Select and pin next metric'},
  {key: 'q', label: 'Unpin current metric'},
];


function renderShortcuts(cuts) {
  return (
    <div>
      {cuts.map(({key, label}) => (
        <div key={key} className="help-panel-shortcuts-shortcut">
          <div className="key"><kbd>{key}</kbd></div>
          <div className="label">{label}</div>
        </div>
      ))}
    </div>
  );
}


function renderShortcutPanel() {
  return (
    <div className="help-panel-shortcuts">
      <h2>Shortcuts</h2>
      <h3>General</h3>
      {renderShortcuts(GENERAL_SHORTCUTS)}
      <h3>Canvas Metrics</h3>
      {renderShortcuts(CANVAS_METRIC_SHORTCUTS)}
    </div>
  );
}


const BASIC_SEARCHES = [
  {label: 'All fields for foo', term: 'foo'},
  {
    label: <span>Any field matching <b>pid</b> for the value 12345</span>,
    term: 'pid: 12345'
  },
];


const REGEX_SEARCHES = [
  {
    label: 'All fields for foo or bar',
    term: 'foo|bar'
  },
  {
    label: <span><b>command</b> field for foobar or foobaz</span>,
    term: 'command: foo(bar|baz)'
  },
];


const METRIC_SEARCHES = [
  {label: <span><b>CPU</b> greater than 4%</span>, term: 'cpu > 4%'},
  {
    label: <span><b>Memory</b> less than 10 megabytes</span>,
    term: 'memory < 10mb'
  },
];


function renderSearches(searches) {
  return (
    <div>
      {searches.map(({term, label}) => (
        <div key={term} className="help-panel-search-row">
          <div className="help-panel-search-row-term">
            <i className="fa fa-search search-label-icon" />
            {term}
          </div>
          <div className="help-panel-search-row-term-label">{label}</div>
        </div>
      ))}
    </div>
  );
}


function renderSearchPanel() {
  return (
    <div className="help-panel-search">
      <h2>Search</h2>
      <h3>Basics</h3>
      {renderSearches(BASIC_SEARCHES)}

      <h3>Regular expressions</h3>
      {renderSearches(REGEX_SEARCHES)}

      <h3>Metrics</h3>
      {renderSearches(METRIC_SEARCHES)}

    </div>
  );
}


function renderFieldsPanel(currentTopologyName, searchableFields) {
  const none = (
    <span style={{fontStyle: 'italic'}}>None</span>
  );
  const currentTopology = (
    <span className="help-panel-fields-current-topology">
      {currentTopologyName}
    </span>
  );

  return (
    <div className="help-panel-fields">
      <h2>Fields and Metrics</h2>
      <p>
        Searchable fields and metrics in the <br />
        currently selected {currentTopology} topology:
      </p>
      <div className="help-panel-fields-fields">
        <div className="help-panel-fields-fields-column">
          <h3>Fields</h3>
          <div className="help-panel-fields-fields-column-content">
            {searchableFields.get('fields').map(f => (
              <div key={f}>{f}</div>
            ))}
            {searchableFields.get('fields').size === 0 && none}
          </div>
        </div>
        <div className="help-panel-fields-fields-column">
          <h3>Metrics</h3>
          <div className="help-panel-fields-fields-column-content">
            {searchableFields.get('metrics').map(m => (
              <div key={m}>{m}</div>
            ))}
            {searchableFields.get('metrics').size === 0 && none}
          </div>
        </div>
      </div>
    </div>
  );
}


function HelpPanel({
  currentTopologyName, searchableFields, onClickClose, canvasMargins
}) {
  return (
    <div className="help-panel-wrapper">
      <div className="help-panel" style={{marginTop: canvasMargins.top}}>
        <div className="help-panel-header">
          <h2>Help</h2>
        </div>
        <div className="help-panel-main">
          {renderShortcutPanel()}
          {renderSearchPanel()}
          {renderFieldsPanel(currentTopologyName, searchableFields)}
        </div>
        <div className="help-panel-tools">
          <i
            title="Close details"
            className="fa fa-times"
            onClick={onClickClose}
          />
        </div>
      </div>
    </div>
  );
}


function mapStateToProps(state) {
  return {
    canvasMargins: canvasMarginsSelector(state),
    currentTopologyName: state.getIn(['currentTopology', 'fullName']),
    searchableFields: searchableFieldsSelector(state)
  };
}


export default connect(mapStateToProps, {
  onClickClose: hideHelp
})(HelpPanel);

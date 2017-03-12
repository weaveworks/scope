import React from 'react';
import { connect } from 'react-redux';

import { searchableFieldsSelector } from '../selectors/search';
import { canvasMarginsSelector } from '../selectors/viewport';
import { hideHelp } from '../actions/app-actions';


const GENERAL_SHORTCUTS = [
  {key: 'esc', label: 'Close active panel'},
  {key: '/', label: 'Activate search field'},
  {key: '?', label: 'Toggle shortcut menu'},
  {key: 't', label: 'Toggle Table mode'},
  {key: 'g', label: 'Toggle Graph mode'},
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
  {term: 'foo', label: 'All fields for foo'},
  {
    term: 'pid: 12345',
    label: <span>Any field matching <b>pid</b> for the value 12345</span>
  },
];


const REGEX_SEARCHES = [
  {
    term: 'foo|bar',
    label: 'All fields for foo or bar'
  },
  {
    term: 'command: foo(bar|baz)',
    label: <span><b>command</b> field for foobar or foobaz</span>
  },
];


const METRIC_SEARCHES = [
  {term: 'cpu > 4%', label: <span><b>CPU</b> greater than 4%</span>},
  {
    term: 'memory < 10mb',
    label: <span><b>Memory</b> less than 10 megabytes</span>
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
  const none = <span style={{fontStyle: 'italic'}}>None</span>;
  return (
    <div className="help-panel-fields">
      <h2>Fields and Metrics</h2>
      <p>
        Searchable fields and metrics in the <br />
        currently selected <span className="help-panel-fields-current-topology">
          {currentTopologyName}</span> topology:
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


function HelpPanel({currentTopologyName, searchableFields, onClickClose}) {
  return (
    <div className="help-panel-wrapper">
      <div className="help-panel" style={{marginTop: this.props.canvasMargins.top}}>
        <div className="help-panel-header">
          <h2>Help</h2>
        </div>
        <div className="help-panel-main">
          {renderShortcutPanel()}
          {renderSearchPanel()}
          {renderFieldsPanel(currentTopologyName, searchableFields)}
        </div>
        <div className="help-panel-tools">
          <span
            title="Close details"
            className="fa fa-close"
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
    searchableFields: searchableFieldsSelector(state),
    currentTopologyName: state.getIn(['currentTopology', 'fullName'])
  };
}


export default connect(mapStateToProps, {
  onClickClose: hideHelp
})(HelpPanel);

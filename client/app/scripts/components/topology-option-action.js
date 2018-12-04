import React from 'react';

export default class TopologyOptionAction extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.onClick = this.onClick.bind(this);
  }

  onClick(ev) {
    ev.preventDefault();
    const { optionId, topologyId, item } = this.props;
    this.props.onClick(optionId, item.get('value'), topologyId);
  }

  render() {
    const { activeValue, item } = this.props;
    const className = activeValue.includes(item.get('value'))
      ? 'topology-option-action topology-option-action-selected'
      : 'topology-option-action';
    return (
      <div className={className} onClick={this.onClick}>
        {item.get('label')}
      </div>
    );
  }
}

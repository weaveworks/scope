import React from 'react';
import classnames from 'classnames';


class Warning extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.handleClick = this.handleClick.bind(this);
    this.state = {
      expanded: false
    };
  }

  handleClick() {
    const expanded = !this.state.expanded;
    this.setState({ expanded });
  }

  render() {
    const { text } = this.props;
    const { expanded } = this.state;

    const className = classnames('warning', {
      'warning-expanded': expanded
    });

    return (
      <div className={className} onClick={this.handleClick}>
        <div className="warning-wrapper">
          <i className="warning-icon fa fa-exclamation-triangle" title={text} />
          {expanded && <span className="warning-text">{text}</span>}
        </div>
      </div>
    );
  }
}

export default Warning;

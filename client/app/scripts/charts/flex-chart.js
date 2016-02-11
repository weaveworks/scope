import debug from 'debug';
import React from 'react';
import ReactDOM from 'react-dom';
import SimpleChart from './simple-chart';

const log = debug('scope:chart');

export default class FlexChart extends React.Component {

  constructor(props, context) {
    super(props, context);
    log('hi');
    this.updateBCR = this.updateBCR.bind(this);
    this.state = {};
  }

  updateBCR() {
    const bcr = ReactDOM.findDOMNode(this).getBoundingClientRect();
    this.setState({bcr: bcr});
  }

  componentDidMount() {
    this.updateBCR();
    window.addEventListener('resize', this.updateBCR);
  }

  componentWillUnmount() {
    window.removeEventListener('resize', this.updateBCR);
  }

  /*
  componentWillReceiveProps(nextProps) {
    const widthChanged = nextProps.cardsCount !== this.props.cardsCount;
    if (widthChanged) {
      this.updateBCR();
    }
  }
 */

  render() {
    const bcr = this.state.bcr;

    if (!bcr) {
      return <div className="fill-parent" />;
    }

    const props = Object.assign({}, this.props, {width: bcr.width, height: bcr.height});

    return (
      <div className="fill-parent">
        <SimpleChart {...props} />
      </div>
    );
  }
}

/** @jsx React.DOM */

var React = require('react');

var Paper = React.createClass({

    render: function() {
        return (
            <div className="modal-content">
                <div className="modal-body">
                    {this.props.children}
                </div>
            </div>
        );
    }

});

module.exports = Paper;
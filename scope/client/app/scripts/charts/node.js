var _ = require('lodash');
var React = require('react');
var d3 = require('d3');

var colors = d3.scale.category20();

var Node = React.createClass({
    render: function() {
        var transform = "translate(" + this.props.dx + "," + this.props.dy + ")";
        var scale = this.props.scale;
        var textOffsetX = 0;
        var textOffsetY = scale(0.5) + 16; 
        var textAngle = _.isUndefined(this.props.angle) ? 0 : -1 * (this.props.angle * 180 / Math.PI - 90);
        var className = this.props.highlighted ? "node highlighted" : "node";

        return (
            <g className={className} transform={transform} onClick={this.props.onClick} id={this.props.id}>
                <circle r={scale(0.5)} className="border" stroke={colors(this.props.label)}></circle>
                <circle r={scale(0.45)} className="shadow"></circle>
                <circle r={Math.max(2, scale(0.125))} className="node"></circle>
                <text textAnchor="middle" x={textOffsetX} y={textOffsetY}>{this.props.label}</text>
            </g>
        );
    }
});

module.exports = Node;

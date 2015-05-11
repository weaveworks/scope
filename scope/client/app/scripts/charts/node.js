var _ = require('lodash');
var React = require('react');

var NodeColorMixin = require('../mixins/node-color-mixin');

var Node = React.createClass({
    mixins: [
        NodeColorMixin
    ],
    render: function() {
        var transform = "translate(" + this.props.dx + "," + this.props.dy + ")";
        var scale = this.props.scale;
        var textOffsetX = 0;
        var textOffsetY = scale(0.5) + 18;
        var textAngle = _.isUndefined(this.props.angle) ? 0 : -1 * (this.props.angle * 180 / Math.PI - 90);
        var color = this.getNodeColor(this.props.label);
        var className = this.props.highlighted ? "node highlighted" : "node";

        return (
            <g className={className} transform={transform} onClick={this.props.onClick} id={this.props.id}>
                <circle r={scale(0.5)} className="border" stroke={color}></circle>
                <circle r={scale(0.45)} className="shadow"></circle>
                <circle r={Math.max(2, scale(0.125))} className="node"></circle>
                <text className="node-label" textAnchor="middle" x={textOffsetX} y={textOffsetY}>{this.props.label}</text>
                <text className="node-sublabel" textAnchor="middle" x={textOffsetX} y={textOffsetY + 17}>{this.props.subLabel}</text>
            </g>
        );
    }
});

module.exports = Node;

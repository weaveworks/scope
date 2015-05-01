var _ = require('lodash');
var React = require('react');

var Node = React.createClass({
    render: function() {
        var transform = "translate(" + this.props.dx + "," + this.props.dy + ")";
        var scale = this.props.scale;
        var textOffsetX = scale(0.5) + 4;
        var textAngle = _.isUndefined(this.props.angle) ? 0 : -1 * (this.props.angle * 180 / Math.PI - 90);
        var textAnchor = this.props.textAnchor;

        if (textAnchor === "end") {
            textAngle += 180;
            textOffsetX *= -1;
        }

        var rotate = _.isUndefined(this.props.angle) ? "" : "rotate(" + textAngle + ")";
        var className = this.props.highlighted ? "node highlighted" : "node";

        return (
            <g className={className} transform={transform} onClick={this.props.onClick} id={this.props.id}>
                <circle r={scale(0.5)} className="border"></circle>
                <circle r={scale(0.45)} className="shadow"></circle>
                <circle r={Math.max(2, scale(0.125))} className="node"></circle>
                <text textAnchor={textAnchor} transform={rotate}
                    x={textOffsetX} y="0.3em">{this.props.label}</text>
            </g>
        );
    }
});

module.exports = Node;

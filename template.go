package main

const (
	GCVIS_TMPL = `
<html>
<head>
<title>gcvis - {{ .Title }}</title>
<script src="//cdnjs.cloudflare.com/ajax/libs/jquery/2.0.3/jquery.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.2/jquery.flot.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.2/jquery.flot.time.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.2/jquery.flot.selection.min.js"></script>

<script type="text/javascript">

(function() {
	var data = [
		{ label: "gc.heapinuse", data: {{ .HeapUse }} },
		{ label: "scvg.inuse", data: {{ .ScvgInuse }} },
		{ label: "scvg.idle", data: {{ .ScvgIdle }} },
		{ label: "scvg.sys", data: {{ .ScvgSys }} },
		{ label: "scvg.released", data: {{ .ScvgReleased }} },
		{ label: "scvg.consumed", data: {{ .ScvgConsumed }} }
	];

	var options = {
		legend: {
			position: "nw",
			noColumns: 2,
			backgroundOpacity: 0.2
		},
		xaxis: {
			mode: "time",
			timezone: "browser",
			timeformat: "%H:%M:%S "
		},
		selection: {
			mode: "x"
		},
	};

	$(document).ready(function() {
		var plot = $.plot("#placeholder", data, options);

		var overview = $.plot("#overview", data, {
			legend: { show: false},
			series: {
				lines: {
					show: true,
					lineWidth: 1
				},
				shadowSize: 0
			},
			xaxis: {
				ticks: [],
				mode: "time"
			},
			yaxis: {
				ticks: [],
				min: 0,
				autoscaleMargin: 0.1
			},
			selection: {
				mode: "x"
			}
		});

		// now connect the two
		$("#placeholder").bind("plotselected", function (event, ranges) {

			// do the zooming
			$.each(plot.getXAxes(), function(_, axis) {
				var opts = axis.options;
				opts.min = ranges.xaxis.from;
				opts.max = ranges.xaxis.to;
			});
			plot.setupGrid();
			plot.draw();
			plot.clearSelection();

			// don't fire event on the overview to prevent eternal loop

			overview.setSelection(ranges, true);
		});

		$("#overview").bind("plotselected", function (event, ranges) {
			plot.setSelection(ranges);
		});

		// refresh data every second
		pullAndRedraw();

		function pullAndRedraw() {
			$.get(window.location.href + 'graph.json', function(graphData) {
				var data = [
					{ label: "gc.heapinuse", data: graphData.HeapUse },
					{ label: "scvg.inuse", data: graphData.ScvgInuse },
					{ label: "scvg.idle", data: graphData.ScvgIdle },
					{ label: "scvg.sys", data: graphData.ScvgSys },
					{ label: "scvg.released", data: graphData.ScvgReleased },
					{ label: "scvg.consumed", data: graphData.ScvgConsumed }
				];

				plot.setData(data);
				plot.setupGrid();
				plot.draw();

				overview.setData(data);
				overview.setupGrid();
				overview.draw();

				setTimeout(pullAndRedraw, 1000);
			})
		}
	});
})();
</script>
<style>
#content {
	margin: 0 auto;
	padding: 10px;
}

#export {
	float: right;
}

.demo-container {
	box-sizing: border-box;
	width: 1200px;
	height: 450px;
	padding: 20px 15px 15px 15px;
	margin: 15px auto 30px auto;
	border: 1px solid #ddd;
	background: #fff;
	background: linear-gradient(#f6f6f6 0, #fff 50px);
	background: -o-linear-gradient(#f6f6f6 0, #fff 50px);
	background: -ms-linear-gradient(#f6f6f6 0, #fff 50px);
	background: -moz-linear-gradient(#f6f6f6 0, #fff 50px);
	background: -webkit-linear-gradient(#f6f6f6 0, #fff 50px);
	box-shadow: 0 3px 10px rgba(0,0,0,0.15);
	-o-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
	-ms-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
	-moz-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
	-webkit-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
}

.demo-placeholder {
	width: 100%;
	height: 100%;
	font-size: 14px;
	line-height: 1.2em;
}
</style>
</head>
<body>
<pre>{{ .Title }}</pre>
<div id="export">
	<a href="/graph.json">json</a>
</div>
<div id="content">

	<div class="demo-container">
		<div id="placeholder" class="demo-placeholder"></div>
	</div>

	<div class="demo-container" style="height:150px;">
		<div id="overview" class="demo-placeholder"></div>
	</div>

	<p>The smaller plot is linked to the main plot, so it acts as an overview. Try dragging a selection on either plot, and watch the behavior of the other.</p>

</div>

<pre><b>Legend</b>

gc.heapinuse: heap in use after gc
scvg.inuse: virtual memory considered in use by the scavenger
scvg.idle: virtual memory considered unused by the scavenger
scvg.sys: virtual memory requested from the operating system (should aproximate VSS)
scvg.released: virtual memory returned to the operating system by the scavenger
scvg.consumed: virtual memory in use (should roughly match process RSS)
</pre>
</body>
</html>
	`
)

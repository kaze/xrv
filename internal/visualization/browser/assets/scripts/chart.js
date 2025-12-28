window.initializeChart = function(containerId, config) {
    const container = document.getElementById(containerId);
    if (!container) {
        console.error('Chart container not found:', containerId);
        return null;
    }

    const chart = echarts.init(container);

    const option = {
        title: {
            text: config.title.text,
            subtext: config.title.subtext,
            left: 'center'
        },
        tooltip: {
            trigger: config.tooltip.trigger,
            show: config.tooltip.show
        },
        legend: {
            data: config.legend.data,
            show: config.legend.show,
            top: config.legend.top
        },
        toolbox: {
            show: config.toolbox.show,
            feature: {
                saveAsImage: {
                    show: config.toolbox.feature.saveAsImage.show,
                    type: config.toolbox.feature.saveAsImage.type,
                    title: config.toolbox.feature.saveAsImage.title
                },
                dataZoom: {
                    show: config.toolbox.feature.dataZoom.show,
                    title: config.toolbox.feature.dataZoom.title
                }
            }
        },
        dataZoom: config.dataZoom.map(function(dz) {
            return {
                type: dz.type,
                start: dz.start,
                end: dz.end
            };
        }),
        xAxis: {
            type: config.xAxis.type,
            name: config.xAxis.name,
            data: config.xAxis.data
        },
        yAxis: {
            type: config.yAxis.type,
            name: config.yAxis.name
        },
        series: config.series.map(function(s) {
            return {
                name: s.name,
                type: s.type,
                data: s.data,
                smooth: s.smooth
            };
        })
    };

    chart.setOption(option);

    window.addEventListener('resize', function() {
        chart.resize();
    });

    return chart;
};

window.destroyChart = function(containerId) {
    const container = document.getElementById(containerId);
    if (container) {
        const chart = echarts.getInstanceByDom(container);
        if (chart) {
            chart.dispose();
        }
    }
};

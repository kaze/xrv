window.exportChart = function(format) {
    const form = document.getElementById('vizForm');
    if (!form) {
        console.error('Form not found');
        return;
    }

    const formData = new FormData(form);
    const base = formData.get('base') || 'USD';
    const currencies = formData.get('currencies');
    const from = formData.get('from');
    const to = formData.get('to');

    if (!currencies || !from || !to) {
        alert('Please fill in all required fields before exporting');
        return;
    }

    if (format === 'csv') {
        const url = '/export/csv?base=' + encodeURIComponent(base) + '&currencies=' + encodeURIComponent(currencies) + '&from=' + encodeURIComponent(from) + '&to=' + encodeURIComponent(to);
        window.location.href = url;
    } else if (format === 'json') {
        const url = '/export/json?base=' + encodeURIComponent(base) + '&currencies=' + encodeURIComponent(currencies) + '&from=' + encodeURIComponent(from) + '&to=' + encodeURIComponent(to);
        window.location.href = url;
    } else if (format === 'image') {
        exportChartImage();
    }
};

function exportChartImage() {
    const chartCanvas = document.getElementById('chartCanvas');
    if (!chartCanvas) {
        alert('No chart to export. Please generate a chart first.');
        return;
    }

    const chartInstance = echarts.getInstanceByDom(chartCanvas);
    if (!chartInstance) {
        alert('Chart not initialized');
        return;
    }

    const imageDataURL = chartInstance.getDataURL({
        type: 'png',
        pixelRatio: 2,
        backgroundColor: '#fff'
    });

    const link = document.createElement('a');
    link.href = imageDataURL;
    
    const today = new Date().toISOString().split('T')[0];
    link.download = 'xrv-chart-' + today + '.png';
    
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
}

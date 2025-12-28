document.addEventListener('htmx:afterSwap', function(event) {
    if (event.detail.target.id === 'chartContainer') {
        if (window.xrvChartData) {
            const chartCanvas = document.getElementById('chartCanvas');
            if (chartCanvas) {
                window.destroyChart('chartCanvas');
                window.initializeChart('chartCanvas', window.xrvChartData);
            }
        }
    }
});

document.addEventListener('htmx:beforeSwap', function(event) {
    if (event.detail.target.id === 'chartContainer') {
        window.destroyChart('chartCanvas');
    }
});

document.addEventListener('htmx:responseError', function(event) {
    const target = event.detail.target;
    if (target.id === 'chartContainer') {
        target.innerHTML = '<div style="text-align: center; padding: 40px; color: #e53e3e;"><p>Error loading visualization. Please try again.</p></div>';
    }
});

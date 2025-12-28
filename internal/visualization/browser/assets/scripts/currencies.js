window.loadCurrencies = function() {
    fetch('/currencies')
        .then(response => response.json())
        .then(currencies => {
            populateCurrencySelects(currencies);
        })
        .catch(error => {
            console.error('Failed to load currencies:', error);
            showCurrencyError();
        });
};

function populateCurrencySelects(currencies) {
    const baseSelect = document.getElementById('base');
    const currenciesSelect = document.getElementById('currencies');

    if (!baseSelect || !currenciesSelect) {
        return;
    }

    const sortedCurrencies = Object.keys(currencies).sort();

    baseSelect.innerHTML = '';
    currenciesSelect.innerHTML = '';

    sortedCurrencies.forEach(code => {
        const name = currencies[code];
        
        const baseOption = document.createElement('option');
        baseOption.value = code;
        baseOption.textContent = code + ' - ' + name;
        baseSelect.appendChild(baseOption);

        const targetOption = document.createElement('option');
        targetOption.value = code;
        targetOption.textContent = code + ' - ' + name;
        currenciesSelect.appendChild(targetOption);
    });

    setDefaultCurrencies(baseSelect, currenciesSelect);
}

function setDefaultCurrencies(baseSelect, currenciesSelect) {
    baseSelect.value = 'USD';

    const defaultTargets = ['EUR', 'GBP', 'JPY'];
    Array.from(currenciesSelect.options).forEach(option => {
        if (defaultTargets.includes(option.value)) {
            option.selected = true;
        }
    });
}

function showCurrencyError() {
    const baseSelect = document.getElementById('base');
    const currenciesSelect = document.getElementById('currencies');

    if (baseSelect) {
        baseSelect.innerHTML = '<option value="">Failed to load currencies</option>';
    }
    if (currenciesSelect) {
        currenciesSelect.innerHTML = '<option value="">Failed to load currencies</option>';
    }
}

document.addEventListener('DOMContentLoaded', function() {
    loadCurrencies();
    setupCurrencyFormHandler();
});

function setupCurrencyFormHandler() {
    const form = document.getElementById('vizForm');
    if (!form) return;

    form.addEventListener('htmx:configRequest', function(event) {
        const currenciesSelect = document.getElementById('currencies');
        if (!currenciesSelect) return;

        const selectedValues = Array.from(currenciesSelect.selectedOptions)
            .map(option => option.value)
            .filter(value => value !== '');

        if (selectedValues.length > 0) {
            event.detail.parameters.currencies = selectedValues.join(',');
        }
    });
}

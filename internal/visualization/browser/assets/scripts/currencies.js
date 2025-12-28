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

    if (!baseSelect) {
        return;
    }

    const sortedCurrencies = Object.keys(currencies).sort();

    baseSelect.innerHTML = '';

    sortedCurrencies.forEach(code => {
        const name = currencies[code];

        const baseOption = document.createElement('option');
        baseOption.value = code;
        baseOption.textContent = code + ' - ' + name;
        baseSelect.appendChild(baseOption);
    });

    populateCustomMultiSelect(currencies, sortedCurrencies);
    setDefaultCurrencies(baseSelect);
}

function setDefaultCurrencies(baseSelect) {
    baseSelect.value = 'HUF';

    const defaultTargets = ['EUR', 'GBP', 'USD'];
    window.selectedCurrencies = defaultTargets;
    updateMultiSelectDisplay();
}

function populateCustomMultiSelect(currencies, sortedCurrencies) {
    const dropdown = document.getElementById('currenciesDropdown');
    if (!dropdown) return;

    dropdown.innerHTML = '';
    window.currenciesData = {};

    sortedCurrencies.forEach(code => {
        const name = currencies[code];
        window.currenciesData[code] = name;

        const item = document.createElement('div');
        item.className = 'multi-select-item';

        const checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.id = 'currency-' + code;
        checkbox.value = code;
        checkbox.addEventListener('change', handleCurrencySelection);

        const label = document.createElement('label');
        label.htmlFor = 'currency-' + code;
        label.textContent = code + ' - ' + name;

        item.appendChild(checkbox);
        item.appendChild(label);
        dropdown.appendChild(item);
    });

    setupMultiSelectToggle();
}

function setupMultiSelectToggle() {
    const display = document.getElementById('currenciesDisplay');
    const dropdown = document.getElementById('currenciesDropdown');

    if (!display || !dropdown) return;

    display.addEventListener('click', function(e) {
        e.stopPropagation();
        dropdown.classList.toggle('show');
    });

    document.addEventListener('click', function(e) {
        if (!e.target.closest('.custom-multi-select')) {
            dropdown.classList.remove('show');
        }
    });
}

function handleCurrencySelection() {
    const checkboxes = document.querySelectorAll('#currenciesDropdown input[type="checkbox"]');
    window.selectedCurrencies = Array.from(checkboxes)
        .filter(cb => cb.checked)
        .map(cb => cb.value);

    updateMultiSelectDisplay();
}

function updateMultiSelectDisplay() {
    const display = document.getElementById('currenciesDisplay');
    const hiddenInput = document.getElementById('currencies');

    if (!display || !hiddenInput) return;

    const checkboxes = document.querySelectorAll('#currenciesDropdown input[type="checkbox"]');
    checkboxes.forEach(cb => {
        cb.checked = window.selectedCurrencies.includes(cb.value);
    });

    const placeholder = display.querySelector('.multi-select-placeholder');
    if (window.selectedCurrencies.length === 0) {
        placeholder.textContent = 'Select currencies...';
    } else {
        placeholder.textContent = window.selectedCurrencies.join(', ');
    }

    hiddenInput.value = window.selectedCurrencies.join(',');
}

function showCurrencyError() {
    const baseSelect = document.getElementById('base');
    const display = document.getElementById('currenciesDisplay');

    if (baseSelect) {
        baseSelect.innerHTML = '<option value="">Failed to load currencies</option>';
    }
    if (display) {
        const placeholder = display.querySelector('.multi-select-placeholder');
        if (placeholder) {
            placeholder.textContent = 'Failed to load currencies';
        }
    }
}

document.addEventListener('DOMContentLoaded', function() {
    loadCurrencies();
});

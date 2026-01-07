// deals.js - Handles the Detected Deals admin page

const API_BASE = '/api/v1/admin';

document.addEventListener('DOMContentLoaded', () => {
    loadDeals();
    loadAlertStats();

    document.getElementById('refreshBtn').addEventListener('click', () => {
        loadDeals();
        loadAlertStats();
    });

    document.getElementById('filterForm').addEventListener('submit', (e) => {
        e.preventDefault();
        loadDeals();
    });
});

async function loadDeals() {
    const params = new URLSearchParams();

    const origin = document.getElementById('originFilter').value.trim().toUpperCase();
    const destination = document.getElementById('destinationFilter').value.trim().toUpperCase();
    const classification = document.getElementById('classificationFilter').value;
    const status = document.getElementById('statusFilter').value;

    if (origin) params.append('origin', origin);
    if (destination) params.append('destination', destination);
    if (classification) params.append('classification', classification);
    if (status) params.append('status', status);
    params.append('limit', '100');

    try {
        const response = await fetch(`${API_BASE}/deals?${params.toString()}`);
        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Failed to fetch deals');
        }

        renderDealsTable(data.deals || []);
        updateDealStats(data.deals || []);
        document.getElementById('dealsMeta').textContent = `${data.count || 0} deals found`;
    } catch (error) {
        console.error('Error loading deals:', error);
        document.getElementById('dealsTableBody').innerHTML = `
            <tr><td colspan="9" class="text-center py-4 text-danger">
                Error loading deals: ${error.message}
            </td></tr>`;
        document.getElementById('dealsMeta').textContent = 'Error loading';
    }
}

async function loadAlertStats() {
    try {
        const response = await fetch(`${API_BASE}/deal-alerts?limit=100`);
        const data = await response.json();
        const count = data.count || 0;
        document.getElementById('alertsSentCount').textContent = count;
    } catch (error) {
        console.error('Error loading alert stats:', error);
        document.getElementById('alertsSentCount').textContent = '-';
    }
}

function updateDealStats(deals) {
    const activeDeals = deals.filter(d => d.status === 'active').length;
    const amazingDeals = deals.filter(d => d.deal_classification === 'amazing' || d.deal_classification === 'error_fare').length;

    let totalDiscount = 0;
    let discountCount = 0;
    deals.forEach(d => {
        if (d.discount_percent != null) {
            totalDiscount += d.discount_percent;
            discountCount++;
        }
    });
    const avgDiscount = discountCount > 0 ? (totalDiscount / discountCount).toFixed(0) : '-';

    document.getElementById('activeDealsCount').textContent = activeDeals;
    document.getElementById('amazingDealsCount').textContent = amazingDeals;
    document.getElementById('avgDiscount').textContent = avgDiscount !== '-' ? `${avgDiscount}%` : '-';
}

function renderDealsTable(deals) {
    const tbody = document.getElementById('dealsTableBody');

    if (!deals || deals.length === 0) {
        tbody.innerHTML = `
            <tr><td colspan="9" class="text-center py-4 text-muted">
                No deals found. Run a continuous sweep to detect deals.
            </td></tr>`;
        return;
    }

    tbody.innerHTML = deals.map(deal => `
        <tr>
            <td>
                <strong>${deal.origin}</strong> → <strong>${deal.destination}</strong>
            </td>
            <td>${formatDate(deal.departure_date)}</td>
            <td class="fw-bold text-success">
                $${deal.price.toFixed(2)}
                <small class="text-muted">${deal.currency}</small>
            </td>
            <td>
                ${deal.discount_percent != null ? `<span class="text-danger">-${deal.discount_percent.toFixed(0)}%</span>` : '-'}
            </td>
            <td>${renderClassificationBadge(deal.deal_classification)}</td>
            <td>${renderScore(deal.deal_score)}</td>
            <td>${deal.cost_per_mile != null ? `$${deal.cost_per_mile.toFixed(3)}` : '-'}</td>
            <td><small>${formatDateTime(deal.first_seen_at)}</small></td>
            <td class="text-end">
                ${deal.search_url ? `
                    <a href="${deal.search_url}" target="_blank" class="btn btn-sm btn-outline-primary" title="View on Google Flights">
                        <i class="bi bi-box-arrow-up-right"></i>
                    </a>
                ` : ''}
            </td>
        </tr>
    `).join('');
}

function renderClassificationBadge(classification) {
    if (!classification) return '<span class="badge bg-secondary">Unknown</span>';

    const badgeClasses = {
        'good': 'badge-good',
        'great': 'badge-great',
        'amazing': 'badge-amazing',
        'error_fare': 'badge-error-fare'
    };

    const displayNames = {
        'good': 'Good',
        'great': 'Great',
        'amazing': 'Amazing',
        'error_fare': 'Error Fare'
    };

    const badgeClass = badgeClasses[classification] || 'bg-secondary';
    const displayName = displayNames[classification] || classification;

    return `<span class="badge ${badgeClass}">${displayName}</span>`;
}

function renderScore(score) {
    if (score == null) return '-';

    let colorClass = 'text-muted';
    if (score >= 80) colorClass = 'text-success';
    else if (score >= 60) colorClass = 'text-info';
    else if (score >= 40) colorClass = 'text-warning';

    return `<span class="deal-score ${colorClass}">${score}</span>`;
}

function formatDate(dateStr) {
    if (!dateStr) return '-';
    const d = new Date(dateStr);
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

function formatDateTime(dateStr) {
    if (!dateStr) return '-';
    const d = new Date(dateStr);
    return d.toLocaleString('en-US', {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}

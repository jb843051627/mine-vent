async function api(path) {
    const resp = await fetch(path);
    if (!resp.ok) return null;
    return resp.json();
}

async function loadOverview() {
    const data = await api('/api/dashboard');
    if (!data) return;
    document.getElementById('sensor-count').textContent = data.total_sensors || 0;
    document.getElementById('fan-count').textContent = data.total_fans || 0;
    document.getElementById('alert-count').textContent = data.active_alerts || 0;
    document.getElementById('maint-count').textContent = data.scheduled_maintenance || 0;
}

async function loadDashboard() {
    const data = await api('/api/dashboard');
    if (!data) return;
    const grid = document.getElementById('dashboard-data');
    if (grid) {
        grid.innerHTML = `
            <div class="stat-card"><h3>传感器</h3><p>${data.total_sensors}</p></div>
            <div class="stat-card"><h3>风机</h3><p>${data.total_fans}</p></div>
            <div class="stat-card alert"><h3>活跃告警</h3><p>${data.active_alerts}</p></div>
            <div class="stat-card"><h3>待维护</h3><p>${data.scheduled_maintenance}</p></div>
        `;
    }
    const fanStatus = document.getElementById('fan-status');
    if (fanStatus && data.fan_status_counts) {
        fanStatus.innerHTML = Object.entries(data.fan_status_counts)
            .map(([k, v]) => `<span class="badge">${k}: ${v}</span>`).join('');
    }
    const alertSummary = document.getElementById('alert-summary');
    if (alertSummary && data.alert_summary) {
        alertSummary.innerHTML = Object.entries(data.alert_summary.by_level || {})
            .map(([k, v]) => `<span class="badge alert-${k}">${k}: ${v}</span>`).join('');
    }
}

async function loadSensors() {
    const sensors = await api('/api/sensors');
    if (!sensors) return;
    const tbody = document.getElementById('sensor-list');
    if (!tbody) return;
    tbody.innerHTML = sensors.map(s => `
        <tr>
            <td>${s.id}</td>
            <td>${s.name}</td>
            <td>${s.type}</td>
            <td>${s.direction}</td>
            <td>${s.area_id}</td>
            <td>${s.is_active ? '活跃' : '停用'}</td>
        </tr>
    `).join('');
}

async function loadAlerts() {
    const alerts = await api('/api/alerts/active');
    if (!alerts) return;
    const tbody = document.getElementById('alert-list');
    if (!tbody) return;
    tbody.innerHTML = alerts.map(a => `
        <tr class="alert-${a.level}">
            <td>${a.id}</td>
            <td>${a.level}</td>
            <td>${a.title}</td>
            <td>${a.value}</td>
            <td>${a.threshold}</td>
            <td>${new Date(a.triggered_at).toLocaleString()}</td>
        </tr>
    `).join('');
}

if (document.getElementById('overview-stats')) {
    loadOverview();
}

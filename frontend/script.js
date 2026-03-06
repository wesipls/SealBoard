// Modernized frontend: groups pods/containers per host in separate boxes

const statsContainer = document.getElementById('statsContainer');
const searchInput = document.getElementById('searchInput');

async function fetchStats() {
  const resp = await fetch('/stats');
  const raw = await resp.json();
  // Flatten & annotate per host for search and rendering
  let containers = [];
  for (const host in raw) {
    if (Array.isArray(raw[host])) {
      for (const cont of raw[host]) {
        cont.host = host;
        containers.push(cont);
      }
    } else if (raw[host] && raw[host].error) {
      containers.push({ host, error: raw[host].error });
    }
  }
  renderStats(containers);
}

function renderStats(data) {
  if (!data.length) {
    statsContainer.innerHTML = '<p>No containers found.</p>';
    return;
  }
  // Group containers by host
  const grouped = {};
  for (const container of data) {
    const host = container.host || 'Unknown Host';
    if (!grouped[host]) grouped[host] = [];
    grouped[host].push(container);
  }
  let html = '';
  for (const host in grouped) {
    html += `<div class="host-box"><h2>${host}</h2>`;
    const hostContainers = grouped[host];
    if (hostContainers.length === 1 && hostContainers[0].error) {
      html += `<div class='error-box'><span style='color:red'>${hostContainers[0].error}</span></div>`;
    } else {
      for (const container of hostContainers) {
        html += `<div class="container-box">
          
          <div><b>Name:</b> ${container.Names ? container.Names.join(', ') : ''}</div>
          <div><b>Status:</b> ${container.State || container.Status || ''}</div>
          
        </div>`;
      }
    }
    html += '</div>';
  }
  statsContainer.innerHTML = html;
}

function handleSearch() {
  const text = searchInput.value.trim().toLowerCase();
  fetchStatsFiltered(text);
}

async function fetchStatsFiltered(filterText) {
  const resp = await fetch('/stats');
  const raw = await resp.json();
  let containers = [];
  for (const host in raw) {
    if (Array.isArray(raw[host])) {
      for (const cont of raw[host]) {
        cont.host = host;
        containers.push(cont);
      }
    } else if (raw[host] && raw[host].error) {
      containers.push({ host, error: raw[host].error });
    }
  }
  if (filterText) {
    containers = containers.filter(cont =>
      (cont.Names && cont.Names.join(',').toLowerCase().includes(filterText)) ||
      (cont.Id && cont.Id.toLowerCase().includes(filterText)) ||
      (cont.host && cont.host.toLowerCase().includes(filterText))
    );
  }
  renderStats(containers);
}

searchInput.addEventListener('input', handleSearch);


fetchStats();
setInterval(fetchStats, 4000);


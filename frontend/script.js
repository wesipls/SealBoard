// Fetches stats data from /stats and renders it
const statsContainer = document.getElementById('statsContainer');
const searchInput = document.getElementById('searchInput');
let containerData = [];

async function fetchStats() {
  try {
    const response = await fetch('/stats');
    if (!response.ok) throw new Error('Network response was not ok');
    // The backend returns a map label->data, so we will collect and flatten
    const data = await response.json();
    // Flatten to an array
    containerData = [];
    for (const host in data) {
      if (data[host] && data[host].error) {
        containerData.push({host, error: data[host].error});
      } else if (Array.isArray(data[host])) {
        data[host].forEach(container => {
          containerData.push({host, ...container});
        });
      }
    }
    // Always render using current search (if any)
    const q = searchInput.value.toLowerCase();
    const filtered = containerData.filter(c => {
      return (
        (c.Names && c.Names.join(', ').toLowerCase().includes(q)) ||
        (c.Id && c.Id.toLowerCase().includes(q)) ||
        (c.State && c.State.toLowerCase().includes(q)) ||
        (c.Status && c.Status.toLowerCase().includes(q)) ||
        (c.host && c.host.toLowerCase().includes(q))
      );
    });
    renderStats(q ? filtered : containerData);
  } catch (err) {
    statsContainer.innerHTML = `<p class="error">Failed to fetch stats: ${err}</p>`;
  }
}

// Fetches detailed data (stats/config/logs) for a specific container
async function fetchContainerDetail(hostLabel, containerID, type) {
  try {
    const response = await fetch(`/api/host/${hostLabel}/container/${containerID}/${type}`);
    if (!response.ok) throw new Error('Network response was not ok');
    const data = await response.json();
    // For demo: display detail below table; future could open modal
    statsContainer.innerHTML += `<pre>${JSON.stringify(data, null, 2)}</pre>`;
  } catch (err) {
    statsContainer.innerHTML += `<p class="error">Failed to fetch ${type}: ${err}</p>`;
  }
}

function renderStats(data) {
  if (!data.length) {
    statsContainer.innerHTML = '<p>No containers found.</p>';
    return;
  }
  let html = `<table class="table">
    <thead><tr><th>Host</th><th>ID</th><th>Name</th><th>Status</th><th>Details</th></tr></thead>
    <tbody>`;
  for (const container of data) {
    if (container.error) {
      html += `<tr class="error">
        <td colspan="5">${container.host}: <span style='color:red'>${container.error}</span></td>
      </tr>`;
      continue;
    }
    html += `<tr>
      <td>${container.host || ''}</td>
      <td>${container.Id || ''}</td>
      <td>${container.Names ? container.Names.join(', ') : ''}</td>
      <td>${container.State || container.Status || ''}</td>
      <td><button onclick="fetchContainerDetail('${container.host}','${container.Id}','stats')">Stats</button>
          <button onclick="fetchContainerDetail('${container.host}','${container.Id}','config')">Config</button>
          <button onclick="fetchContainerDetail('${container.host}','${container.Id}','logs')">Logs</button></td>
    </tr>`;
  }
  html += '</tbody></table>';
  statsContainer.innerHTML = html;
}

searchInput.addEventListener('input', function() {
  const q = searchInput.value.toLowerCase();
  const filtered = containerData.filter(c => {
    return (
      (c.Names && c.Names.join(', ').toLowerCase().includes(q)) ||
      (c.Id && c.Id.toLowerCase().includes(q)) ||
      (c.State && c.State.toLowerCase().includes(q)) ||
      (c.Status && c.Status.toLowerCase().includes(q)) ||
      (c.host && c.host.toLowerCase().includes(q))
    );
  });
  renderStats(filtered);
});

// Initial fetch
fetchStats();
setInterval(fetchStats, 15000);


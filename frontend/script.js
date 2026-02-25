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

function renderStats(data) {
  if (!data.length) {
    statsContainer.innerHTML = '<p>No containers found.</p>';
    return;
  }
  let html = `<table class="table">
    <thead><tr><th>Host</th><th>ID</th><th>Name</th><th>Status</th></tr></thead>
    <tbody>`;
  for (const container of data) {
    if (container.error) {
      html += `<tr class="error">
        <td colspan="4">${container.host}: <span style='color:red'>${container.error}</span></td>
      </tr>`;
      continue;
    }
    // Display as error if status is 'error' (regardless of error msg)
    if (container.status === 'error' || container.State === 'error' || container.Status === 'error') {
      html += `<tr class="error">
        <td>${container.host || ''}</td>
        <td>${container.Id || ''}</td>
        <td>${container.Names ? container.Names.join(', ') : ''}</td>
        <td><span style='color:red'>error</span></td>
      </tr>`;
      continue;
    }
    html += `<tr>
      <td>${container.host || ''}</td>
      <td>${container.Id || ''}</td>
      <td>${container.Names ? container.Names.join(', ') : ''}</td>
      <td>${container.State || container.Status || ''}</td>
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
// Optionally refresh every X seconds
setInterval(fetchStats, 15000);


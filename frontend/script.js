// Modernized frontend: groups pods/containers per host in separate boxes

function renderStatsByPod(hostsData) {
  if (!hostsData.length) {
    statsContainer.innerHTML = '<p>No data found.</p>';
    return;
  }
  let html = '';
  for (const hostObj of hostsData) {
    html += `<div class="host-box"><h2>${hostObj.host}</h2>`;
    if (hostObj.pods && hostObj.pods.length > 0) {
      for (const pod of hostObj.pods) {
        html += `<div style='margin-bottom:1em;'><div style='font-weight:bold;'>Pod: ${pod.Name || pod.Id}</div>`;
        let podContainers = (hostObj.containers || []).filter(c => c.Pod === pod.Id);
        for (const container of podContainers) {
          html += `<div class="container-box">
            <div><b>Name:</b> ${container.Names ? container.Names.join(', ') : ''}</div>
            <div><b>Status:</b> ${container.State || container.Status || ''}</div>
          </div>`;
        }
        if (!podContainers.length) {
          html += `<div style='padding:0.5em;'>No containers in this pod.</div>`;
        }
        html += '</div>';
      }
    }
    // Standalone containers not in any pod
    // Build set of all pod IDs
    let podIds = new Set((hostObj.pods||[]).map(pod => pod.Id));
    // Standalone containers are those without a Pod property matching any known pod ID
    const standalone = (hostObj.containers||[]).filter(c => !c.Pod || !podIds.has(c.Pod));
    if (standalone.length) {
      html += `<div style='margin-bottom:1em;'><div style='font-weight:bold;'>No Pod</div>`;
      for (const container of standalone) {
        html += `<div class="container-box"><div><b>Name:</b> ${container.Names ? container.Names.join(', ') : ''}</div><div><b>Status:</b> ${container.State || container.Status || ''}</div></div>`;
      }
      html += '</div>';
    }
    html += '</div>';
  }
  statsContainer.innerHTML = html;
}

const statsContainer = document.getElementById('statsContainer');
const searchInput = document.getElementById('searchInput');

async function fetchStats() {
  const resp = await fetch('/stats');
  const raw = await resp.json();
  let hosts = Object.keys(raw);
  let allHostData = [];
  for (const host of hosts) {
    let hostObj = { host };
    // Fetch pods and containers for each host
    let pods = [], containers = [];
    try {
      let podsResp = await fetch(`/api/host/${host}/pods`);
      if (podsResp.ok) pods = await podsResp.json();
    } catch {}
    try {
      let containersResp = await fetch(`/api/host/${host}/containers`);
      if (containersResp.ok) containers = await containersResp.json();
    } catch {}
    hostObj.pods = pods;
        hostObj.containers = containers;
        console.log('debug:', { host, pods, containers });
    allHostData.push(hostObj);
  }
  renderStatsByPod(allHostData);
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


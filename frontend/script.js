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

// Fetch specific Podman container endpoint data for a container
async function fetchContainerEndpoint(host, containerId, endpointType) {
  // endpointType: 'inspect', 'logs', 'stats', or 'top'
  let url = `/api/host/${host}/containers/${containerId}/${endpointType}`;
  try {
    let resp = await fetch(url);
    if (resp.ok) return await resp.json();
    else throw new Error(`Request failed for ${url}`);
  } catch (err) {
    console.error('API error:', err);
    return null;
  }
}


fetchStats();
setInterval(fetchStats, 4000);


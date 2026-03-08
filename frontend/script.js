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
                        <button onclick="showStats('${hostObj.host}', '${container.Id}')">Stats</button>
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

// Show container stats when the Stats button is clicked
async function showStats(host, containerId) {
  const containerStats = await fetchContainerEndpoint(host, containerId, 'stats');
  let statsStr = '';
  if (!containerStats) {
    statsStr = 'No stats data.';
  } else {
    // Common Podman stats object surface: look for memory and cpu breakdowns
    const cpu = containerStats.CPU_stats || containerStats.CPU || containerStats.cpu_stats || containerStats.cpu;
    const mem = containerStats.Memory_stats || containerStats.Memory || containerStats.mem_stats || containerStats.mem;
    if (cpu || mem) {
      if (cpu) {
        statsStr += 'CPU: ';
        if (cpu.cpu_percent !== undefined) statsStr += cpu.cpu_percent + '%\n';
        else statsStr += JSON.stringify(cpu) + '\n';
      }
      if (mem) {
        statsStr += 'Memory: ';
        if (mem.usage !== undefined && mem.limit !== undefined) statsStr += `${(mem.usage/1048576).toFixed(2)} MiB / ${(mem.limit/1048576).toFixed(2)} MiB (use/limit)\n`;
        else statsStr += JSON.stringify(mem) + '\n';
      }
    } else if (containerStats.cpu_percent !== undefined || containerStats.mem_usage !== undefined) {
      // Flat Docker-like
      if (containerStats.cpu_percent !== undefined) statsStr += `CPU: ${containerStats.cpu_percent}%\n`;
      if (containerStats.mem_usage !== undefined) statsStr += `Mem: ${(containerStats.mem_usage/1048576).toFixed(2)} MiB\n`;
    } else {
      statsStr = JSON.stringify(containerStats, null, 2);
    }
  }
  // Try to find (or create) a stats display div directly under the container-box corresponding to containerId
  const allBoxes = document.querySelectorAll('.container-box');
  for (const box of allBoxes) {
    if (box.innerHTML.includes(containerId)) {
      let statsDiv = box.querySelector('.container-stats-output');
      if (!statsDiv) {
        statsDiv = document.createElement('div');
        statsDiv.className = 'container-stats-output';
        statsDiv.style = 'white-space:pre;font-size:smaller;padding:0.5em;background:#181F26;color:#BFF;border-radius:4px;margin-top:4px;';
        box.appendChild(statsDiv);
      }
      statsDiv.textContent = statsStr;
      break;
    }
  }
}

fetchStats();
setInterval(fetchStats, 4000);


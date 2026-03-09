// Fetches all hosts, then for each host fetches pods, displays each pod in a side-by-side div
async function loadPodsAllHosts() {
  const res = await fetch('/stats');
  const stats = await res.json();
  const hosts = Object.keys(stats);
  const podsDiv = document.getElementById('pods');
  podsDiv.innerHTML = 'Loading...';
  let results = [];
  for (let host of hosts) {
    try {
      const [podResp, containersResp] = await Promise.all([
        fetch(`/api/host/${host}/pods`),
        fetch(`/api/host/${host}/containers`)
      ]);
      const pods = await podResp.json();
      const containers = await containersResp.json();
      results.push({host, pods, containers});
    } catch {
      results.push({host, pods: [], containers: []});
    }
  }
  podsDiv.innerHTML = '';
  results.forEach(({host, pods, containers}) => {
    // Gather all IDs of containers that are in any pod
    const podContainerIds = new Set();
    pods.forEach(pod => {
      (pod.Containers || pod.containers || []).forEach(c => {
        podContainerIds.add(c.Id || c.ID || c.id);
      });
    });
    // Render each pod
    pods.forEach(pod => {
      const podName = pod.Name || pod.name || pod.Id || pod.ID || 'Unknown';
      const podBlock = document.createElement('div');
      podBlock.className = 'pod-block';
      podBlock.style.border = '1px solid #ccc';
      podBlock.style.padding = '1em';
      podBlock.style.marginRight = '1em';
      const header = document.createElement('h2');
      header.textContent = `Pod: ${podName}`;
      podBlock.appendChild(header);
      const containerList = document.createElement('ul');
      const containersInPod = (pod.Containers || pod.containers || []);
      containersInPod.forEach(c => {
        const cname = c.Name || c.Names || c.name || c.Id || c.ID || JSON.stringify(c);
        const cId = c.Id || c.ID || c.id;
        const li = document.createElement('li');
        li.textContent = cname;
        const ramSpan = document.createElement('div');
        ramSpan.textContent = 'RAM: Loading...';
        ramSpan.className = 'ram-usage';
        const cpuSpan = document.createElement('div');
        cpuSpan.textContent = 'CPU: Loading...';
        cpuSpan.className = 'cpu-usage';
        li.appendChild(ramSpan);
        li.appendChild(cpuSpan);
        if (cId) fetchAndShowStats(host, cId, ramSpan, cpuSpan);
        containerList.appendChild(li);
      });
      podBlock.appendChild(containerList);
      podsDiv.appendChild(podBlock);
    });
    // Now render containers not assigned to any pod
    const podless = containers.filter(c => !podContainerIds.has(c.Id || c.ID || c.id));
    if (podless.length) {
      const podlessBlock = document.createElement('div');
      podlessBlock.className = 'pod-block';
      podlessBlock.style.border = '1px solid #833';
      podlessBlock.style.padding = '1em';
      podlessBlock.style.marginRight = '1em';
      const header = document.createElement('h2');
      header.textContent = 'Containers not in any pod';
      podlessBlock.appendChild(header);
      const ul = document.createElement('ul');
      podless.forEach(c => {
        const cname = c.Name || c.Names || c.name || c.Id || c.ID || JSON.stringify(c);
        const cId = c.Id || c.ID || c.id;
        const li = document.createElement('li');
        li.textContent = cname;
        const ramSpan = document.createElement('div');
        ramSpan.textContent = 'RAM: Loading...';
        ramSpan.className = 'ram-usage';
        const cpuSpan = document.createElement('div');
        cpuSpan.textContent = 'CPU: Loading...';
        cpuSpan.className = 'cpu-usage';
        li.appendChild(ramSpan);
        li.appendChild(cpuSpan);
        if (cId) fetchAndShowStats(host, cId, ramSpan, cpuSpan);
        ul.appendChild(li);
      });
      podlessBlock.appendChild(ul);
      podsDiv.appendChild(podlessBlock);
    }
    if (pods.length === 0 && !podless.length) {
      const msg = document.createElement('div');
      msg.textContent = `No pods or containers found for host ${host}`;
      podsDiv.appendChild(msg);
    }
  });
}

async function fetchAndShowStats(host, containerId, ramSpan, cpuSpan) {
  try {
    const resp = await fetch(`/api/host/${host}/container/${containerId}/stats`);
    const stat = await resp.json();
    // Defensive parsing for RAM
    let ram = undefined;
    if (stat.memory_stats && typeof stat.memory_stats.usage === 'number') {
      ram = stat.memory_stats.usage;
    } else if (stat.memory && typeof stat.memory.usage === 'number') {
      ram = stat.memory.usage;
    } else if (typeof stat.usage === 'number') {
      ram = stat.usage;
    }
    if (typeof ram === 'number') {
      let msg = 'RAM: ' + (ram / 1024 / 1024).toFixed(1) + ' MB';
      let limit = undefined;
      if (stat.memory_stats && typeof stat.memory_stats.limit === 'number') {
        limit = stat.memory_stats.limit;
      } else if (stat.memory && typeof stat.memory.limit === 'number') {
        limit = stat.memory.limit;
      }
      if (typeof limit === 'number' && limit > 0) {
        msg += ' / ' + (limit / 1024 / 1024).toFixed(1) + ' MB';
      }
      ramSpan.textContent = msg;
    } else {
      ramSpan.textContent = 'RAM: N/A';
    }
    // Defensive parsing for CPU
    let cpuUsage = undefined;
    // Podman v4/stat API: .cpu_stats.cpu_usage.total_usage (in nanoseconds?), system_cpu_usage, .cpu_stats.online_cpus, etc
    if (stat.cpu_stats && stat.cpu_stats.cpu_usage && typeof stat.cpu_stats.cpu_usage.total_usage === 'number') {
      const usage = stat.cpu_stats.cpu_usage.total_usage;
      let cpus = stat.cpu_stats.online_cpus || (stat.cpu_stats.cpu_count || stat.cpu_count) || 1;
      // Show as nanoseconds, try to render core count
      cpuUsage = `CPU: ${(usage/1e9).toFixed(2)}s total`;
      if (cpus && typeof cpus === 'number') {
        cpuUsage += `, ${cpus} CPU${cpus > 1 ? 's' : ''}`;
      }
    }
    // Fallback
    if (!cpuUsage && typeof stat.cpu_percent === 'number') {
      cpuUsage = `CPU: ${stat.cpu_percent.toFixed(1)}%`;
    }
    cpuSpan.textContent = cpuUsage || 'CPU: N/A';
  } catch(e) {
    ramSpan.textContent = 'RAM: Error';
    cpuSpan.textContent = 'CPU: Error';
  }
}

loadPodsAllHosts();


// API Configuration
const API_BASE_URL = 'http://localhost:8888/api/v1';

// Task Status Mapping
const TaskStatus = {
    1: 'Queued',
    2: 'Running',
    3: 'Completed',
    4: 'Failed',
    5: 'Canceled'
};

// DOM Elements
const createTaskForm = document.getElementById('createTaskForm');
const tasksList = document.getElementById('tasksList');
const refreshButton = document.getElementById('refreshTasks');
const logsModal = document.getElementById('logsModal');
const closeModalBtn = document.querySelector('.close');
const logsContent = document.getElementById('logsContent');
const refreshLogsBtn = document.getElementById('refreshLogs');
const fromLineInput = document.getElementById('fromLine');
const toLineInput = document.getElementById('toLine');
const fetchRangeBtn = document.getElementById('fetchRange');

let currentTaskId = null;
let isLoadingLogs = false;
let currentLogState = {
    taskId: null,
    totalLines: 0,
    loadedLines: {
        from: 0,
        to: 0
    },
    hasMoreAbove: true,
    batchSize: 100,  // Initial batch size
    prefetchedLogs: null  // Store prefetched logs
};

// Event Listeners
createTaskForm.addEventListener('submit', handleCreateTask);
refreshButton.addEventListener('click', fetchTasks);
closeModalBtn.addEventListener('click', () => {
    logsModal.style.display = 'none';
    currentTaskId = null;
    resetLogState();
});
refreshLogsBtn.addEventListener('click', () => {
    if (currentTaskId) {
        resetLogState();
        fetchLogs(currentTaskId);
    }
});
fetchRangeBtn.addEventListener('click', () => {
    if (currentTaskId) {
        const from = parseInt(fromLineInput.value) || 0;
        const to = parseInt(toLineInput.value) || 0;
        resetLogState();
        fetchLogs(currentTaskId, from, to);
    }
});

// Close modal when clicking outside
window.addEventListener('click', (e) => {
    if (e.target === logsModal) {
        logsModal.style.display = 'none';
        currentTaskId = null;
        resetLogState();
    }
});

// Add scroll event listener for infinite scrolling
logsContent.addEventListener('scroll', handleLogsScroll);

function resetLogState() {
    currentLogState = {
        taskId: null,
        totalLines: 0,
        loadedLines: {
            from: 0,
            to: 0
        },
        hasMoreAbove: true,
        batchSize: 100,
        prefetchedLogs: null
    };
}

// Fetch tasks on page load
document.addEventListener('DOMContentLoaded', fetchTasks);

async function handleCreateTask(e) {
    e.preventDefault();
    
    const taskName = document.getElementById('taskName').value;
    const taskCommand = document.getElementById('taskCommand').value;
    
    try {
        const response = await fetch(`${API_BASE_URL}/tasks`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                name: taskName,
                command: taskCommand
            })
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const result = await response.json();
        if (result.success) {
            createTaskForm.reset();
            fetchTasks(); // Refresh the tasks list
        } else {
            alert(`Failed to create task: ${result.error}`);
        }
    } catch (error) {
        console.error('Error creating task:', error);
        alert('Failed to create task. Please try again.');
    }
}

async function fetchTasks() {
    try {
        const response = await fetch(`${API_BASE_URL}/tasks`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const result = await response.json();
        if (result.success) {
            renderTasks(result.data);
        } else {
            throw new Error(result.error);
        }
    } catch (error) {
        console.error('Error fetching tasks:', error);
        tasksList.innerHTML = '<p class="error">Failed to load tasks. Please try again.</p>';
    }
}

function renderTasks(tasks) {
    if (!tasks || tasks.length === 0) {
        tasksList.innerHTML = '<p>No tasks found.</p>';
        return;
    }

    tasksList.innerHTML = tasks.map(task => `
        <div class="task-item">
            <h3>${task.name}</h3>
            <p><strong>Command:</strong> ${task.command}</p>
            <p><strong>Status:</strong> <span class="task-status status-${TaskStatus[task.status].toLowerCase()}">${TaskStatus[task.status]}</span></p>
            ${task.exit_code !== undefined ? `<p><strong>Exit Code:</strong> ${task.exit_code}</p>` : ''}
            ${task.reason ? `<p><strong>Reason:</strong> ${task.reason}</p>` : ''}
            <button onclick="showLogs(${task.id})" class="view-logs-btn">View Logs</button>
        </div>
    `).join('');
}

async function showLogs(taskId) {
    currentTaskId = taskId;
    resetLogState();
    currentLogState.taskId = taskId;
    logsModal.style.display = 'block';
    logsContent.innerHTML = 'Loading logs...';
    await fetchLogs(taskId);
}

async function handleLogsScroll() {
    const currentScrollPosition = logsContent.scrollTop;
    const halfwayPoint = (currentLogState.loadedLines.to - currentLogState.loadedLines.from) / 2;
    const scrollThreshold = logsContent.clientHeight * 0.4; // Increased threshold for smoother experience

    // If we're halfway through the current content and haven't prefetched yet
    if (currentScrollPosition < scrollThreshold && !isLoadingLogs && currentLogState.hasMoreAbove && !currentLogState.prefetchedLogs) {
        const newTo = currentLogState.loadedLines.from - 1;
        const newFrom = Math.max(1, newTo - currentLogState.batchSize);
        
        if (newFrom < currentLogState.loadedLines.from) {
            // Start prefetching
            await prefetchLogs(currentTaskId, newFrom, newTo);
            
            // Increase batch size for next fetch, but cap at 500
            currentLogState.batchSize = Math.min(500, currentLogState.batchSize * 2);
        }
    }

    // If we're near the top and have prefetched logs, insert them
    if (currentScrollPosition < 100 && currentLogState.prefetchedLogs) {
        insertPrefetchedLogs();
    }
}

async function prefetchLogs(taskId, from, to) {
    isLoadingLogs = true;
    try {
        const queryParams = new URLSearchParams();
        if (from > 0) queryParams.append('from', from);
        if (to > 0) queryParams.append('to', to);
        
        const response = await fetch(`${API_BASE_URL}/tasks/${taskId}/logs?${queryParams}`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const result = await response.json();
        if (result.success && result.data) {
            const newLogs = result.data.logs || [];
            if (newLogs.length > 0) {
                // Store the prefetched logs and their range
                currentLogState.prefetchedLogs = {
                    logs: newLogs,
                    from: from,
                    to: to
                };
                currentLogState.hasMoreAbove = from > 1;
            } else {
                currentLogState.hasMoreAbove = false;
            }
        }
    } catch (error) {
        console.error('Error prefetching logs:', error);
    } finally {
        isLoadingLogs = false;
    }
}

function insertPrefetchedLogs() {
    if (!currentLogState.prefetchedLogs) return;

    // Preserve scroll position
    const scrollHeight = logsContent.scrollHeight;
    const scrollTop = logsContent.scrollTop;
    
    // Update header
    const header = document.querySelector('.logs-header');
    if (header) {
        header.textContent = `Showing lines ${currentLogState.prefetchedLogs.from}-${currentLogState.loadedLines.to} of ${currentLogState.totalLines}`;
    }
    
    // Prepend new logs
    const logsContainer = document.createElement('div');
    logsContainer.innerHTML = currentLogState.prefetchedLogs.logs.map(log => `<p>${log}</p>`).join('');
    logsContent.insertBefore(logsContainer, header.nextSibling);
    
    // Restore scroll position
    logsContent.scrollTop = logsContent.scrollHeight - scrollHeight + scrollTop;
    
    // Update state
    currentLogState.loadedLines.from = currentLogState.prefetchedLogs.from;
    
    // Clear prefetched logs
    currentLogState.prefetchedLogs = null;
}

async function fetchLogs(taskId, from = 0, to = 0) {
    currentLogState.prefetchedLogs = null; // Clear any prefetched logs
    try {
        const queryParams = new URLSearchParams();
        if (from > 0) queryParams.append('from', from);
        if (to > 0) queryParams.append('to', to);
        
        const response = await fetch(`${API_BASE_URL}/tasks/${taskId}/logs?${queryParams}`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const result = await response.json();
        if (result.success && result.data) {
            const logs = result.data.logs || [];
            const totalLines = result.data.total_lines || 0;
            
            // Update state
            currentLogState.totalLines = totalLines;
            if (from === 0 && to === 0) {
                // Default case - showing last 100 lines
                currentLogState.loadedLines.from = Math.max(1, totalLines - 100);
                currentLogState.loadedLines.to = totalLines;
            } else {
                currentLogState.loadedLines.from = from || currentLogState.loadedLines.from;
                currentLogState.loadedLines.to = to || currentLogState.loadedLines.to;
            }
            currentLogState.hasMoreAbove = currentLogState.loadedLines.from > 1;
            
            renderLogs(logs, totalLines);
        } else {
            throw new Error(result.error || 'No logs data found');
        }
    } catch (error) {
        console.error('Error fetching logs:', error);
        logsContent.innerHTML = `<p class="error">Failed to load logs: ${error.message}</p>`;
    }
}

function renderLogs(logs, totalLines) {
    if (!logs || logs.length === 0) {
        logsContent.innerHTML = '<p>No logs available.</p>';
        return;
    }

    const header = `<div class="logs-header">Showing lines ${currentLogState.loadedLines.from}-${currentLogState.loadedLines.to} of ${totalLines}</div>`;
    const logsHtml = logs.map(log => `<p>${log}</p>`).join('');
    logsContent.innerHTML = header + logsHtml;
    
    // Scroll to bottom of logs on initial load
    if (currentLogState.loadedLines.to === currentLogState.totalLines) {
        logsContent.scrollTop = logsContent.scrollHeight;
    }
}

// Auto-refresh tasks every 5 seconds
setInterval(fetchTasks, 20000); 
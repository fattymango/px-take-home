// API Configuration
const API_BASE_URL = 'http://localhost:8888/api/v1';

// Task Status Mapping
const TaskStatus = {
    1: 'Queued',
    2: 'Running',
    3: 'Completed',
    4: 'Failed',
    5: 'Cancelled'
};

// Task Commands
const TaskCommands = {
    'generate_100_random_numbers': 1,
    'print_100000_prime_numbers': 2,
    'fail_randomly': 3
};

// Event Types
const MsgTypeTaskStatus = 1;
const MsgTypeLog = 2;

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

// SSE connection
let eventSource = null;

// Buffer for collecting log messages before rendering
let logBuffer = [];
let isRenderPending = false;
const LOG_BUFFER_SIZE = 50; // Number of logs to collect before forcing a render
const RENDER_DELAY = 16; // Roughly 60fps

// Event Listeners
createTaskForm.addEventListener('submit', handleCreateTask);
refreshButton.addEventListener('click', () => fetchTasks());
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
        const from = parseInt(fromLineInput.value) || 1;
        const to = parseInt(toLineInput.value) || 100;
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

async function handleCreateTask(event) {
    event.preventDefault();
    
    const name = document.getElementById('taskName').value;
    const commandStr = document.getElementById('taskCommand').value;
    
    if (!name || !commandStr) {
        alert('Please fill in all fields');
        return;
    }
    
    const command = TaskCommands[commandStr];
    if (!command) {
        alert('Invalid command selected');
        return;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/tasks`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                name: name,
                command: command
            }),
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const result = await response.json();
        if (result.success) {
            document.getElementById('createTaskForm').reset();
            await fetchTasks();
        } else {
            alert(`Failed to create task: ${result.error}`);
        }
    } catch (error) {
        console.error('Error:', error);
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

function formatTimestamp(timestamp) {
    if (!timestamp) return 'N/A';
    return new Date(timestamp).toLocaleString();
}

function getStatusTimestamp(task) {
    switch (task.status) {
        case 3: // Completed
            return formatTimestamp(task.completed_at);
        case 4: // Failed
            return formatTimestamp(task.failed_at);
        case 5: // Cancelled
            return formatTimestamp(task.canceled_at);
        case 2: // Running
            return formatTimestamp(task.started_at);
        default:
            return formatTimestamp(task.created_at);
    }
}

function getCommandName(commandId) {
    // Reverse lookup in TaskCommands object
    return Object.entries(TaskCommands).find(([name, id]) => id === commandId)?.[0] || 'Unknown';
}

function renderTask(task) {
    const taskElement = document.createElement('div');
    taskElement.className = `task-item status-${TaskStatus[task.status].toLowerCase()}`;
    
    const statusTime = getStatusTimestamp(task);
    const statusText = `${TaskStatus[task.status]}${task.reason ? ` (${task.reason})` : ''}`;
    const commandName = getCommandName(task.command);
    
    taskElement.innerHTML = `
        <div class="task-header">
            <h3>${task.name}</h3>
            <span class="task-status">${statusText}</span>
        </div>
        <div class="task-details">
            <p><strong>ID:</strong> ${task.id}</p>
            <p><strong>Command:</strong> ${commandName}</p>
            <p><strong>Created:</strong> ${formatTimestamp(task.created_at)}</p>
            <p><strong>Status Time:</strong> ${statusTime}</p>
        </div>
        <div class="task-actions">
            ${task.status === 2 ? `<button onclick="cancelTask('${task.id}')">Cancel</button>` : ''}
            <button onclick="showLogs('${task.id}')">View Logs</button>
        </div>
    `;
    
    return taskElement;
}

function renderTasks(data) {
    const { tasks } = data;
    if (!tasks || tasks.length === 0) {
        tasksList.innerHTML = '<p>No tasks found.</p>';
        return;
    }

    tasksList.innerHTML = tasks.map(task => {
        const status = TaskStatus[task.status];
        const statusLower = status.toLowerCase();
        const commandName = getCommandName(task.command);
        
        let actionButtons = '';
        if (statusLower === 'queued' || statusLower === 'running') {
            actionButtons = `<button onclick="cancelTask('${task.id}')" class="cancel-btn">Cancel</button>`;
        }

        return `
        <div class="task-item" data-status="${statusLower}" data-task-id="${task.id}">
            <div class="task-header">
                <h3>${task.name}</h3>
            </div>
            <div class="task-info">
                <p><strong>Command:</strong> ${commandName}</p>
                <p><strong>Status:</strong> <span class="task-status status-${statusLower}">${status}</span></p>
                <p class="exit-code"${task.exit_code === undefined ? ' style="display:none"' : ''}><strong>Exit Code:</strong> ${task.exit_code !== undefined ? task.exit_code : ''}</p>
                <p class="reason"${!task.reason ? ' style="display:none"' : ''}><strong>Reason:</strong> ${task.reason || ''}</p>
                <div class="task-actions">
                    <button onclick="showLogs('${task.id}')" class="view-logs-btn">View Logs</button>
                    ${actionButtons}
                </div>
            </div>
        </div>
    `}).join('');
}

async function showLogs(taskId) {
    currentTaskId = taskId;
    resetLogState();
    currentLogState.taskId = taskId;
    logsModal.style.display = 'block';
    logsContent.innerHTML = '<p class="logs-loading">Loading logs...</p>';
    
    try {
        const response = await fetch(`${API_BASE_URL}/tasks/${taskId}/logs`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            },
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const result = await response.json();
        if (result.success && result.data) {
            const { logs, total_lines } = result.data;
            currentLogState.totalLines = total_lines;
            currentLogState.loadedLines.from = 1;
            currentLogState.loadedLines.to = logs.length;
            currentLogState.hasMoreAbove = logs.length < total_lines;
            
            renderLogs(logs, total_lines);
        } else {
            throw new Error(result.error || 'Failed to fetch logs');
        }
    } catch (error) {
        console.error('Error fetching logs:', error);
        logsContent.innerHTML = `<p class="error">Failed to load logs: ${error.message}</p>`;
    }
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

async function fetchLogs(taskId, from = 1, to = 100) {
    if (isLoadingLogs) return;
    isLoadingLogs = true;
    
    try {
        const queryParams = new URLSearchParams();
        if (from > 0) queryParams.append('from', from);
        if (to > 0) queryParams.append('to', to);
        
        const response = await fetch(`${API_BASE_URL}/tasks/${taskId}/logs?${queryParams}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const result = await response.json();
        if (result.success && result.data) {
            const { logs, total_lines } = result.data;
            
            // Update state
            currentLogState.totalLines = total_lines;
            currentLogState.loadedLines.from = from;
            currentLogState.loadedLines.to = to;
            currentLogState.hasMoreAbove = from > 1;
            
            renderLogs(logs, total_lines);
        } else {
            throw new Error(result.error || 'No logs data found');
        }
    } catch (error) {
        console.error('Error fetching logs:', error);
        logsContent.innerHTML = `<p class="error">Failed to load logs: ${error.message}</p>`;
    } finally {
        isLoadingLogs = false;
    }
}

function renderLogs(logs, totalLines) {
    if (!logs || logs.length === 0) {
        logsContent.innerHTML = '<p class="no-logs">No logs available.</p>';
        return;
    }

    const header = `<div class="logs-header">Showing lines ${currentLogState.loadedLines.from}-${currentLogState.loadedLines.to} of ${totalLines}</div>`;
    const logsHtml = logs.map((log, index) => {
        const lineNumber = currentLogState.loadedLines.from + index;
        return `<div class="log-line">
            <span class="line-number">${lineNumber}</span>
            <span class="log-message">${log}</span>
        </div>`;
    }).join('');
    
    logsContent.innerHTML = header + logsHtml;
}

function connectToSSE() {
    if (eventSource) {
        eventSource.close();
    }

    console.log('Connecting to SSE...');
    eventSource = new EventSource(`${API_BASE_URL}/events`);

    // Handle connection event
    eventSource.addEventListener('connect', (e) => {
        console.log('SSE Connected:', e.data);
    });

    // Handle all messages through onmessage
    eventSource.onmessage = (e) => {
        try {
            // Remove "data: " prefix and parse the JSON
            const rawData = e.data.replace(/^data: /, '');
            const data = JSON.parse(rawData);
            
            // Log the parsed message for debugging
            console.debug('Parsed SSE message:', data);
            
            // Handle ping messages
            if (data.ping !== undefined) {
                console.debug('Received ping:', data.ping);
                return;
            }
            
            // Handle regular messages
            const eventType = parseInt(data.event);
            if (isNaN(eventType)) {
                console.warn('Invalid event type:', data.event);
                return;
            }

            switch (eventType) {
                case MsgTypeLog:
                    handleLogMessage(data);
                    break;
                case MsgTypeTaskStatus:
                    handleTaskStatusMessage(data);
                    break;
                default:
                    console.warn('Unknown event type:', eventType);
            }
        } catch (error) {
            console.error('Error processing SSE message:', error, e.data);
        }
    };

    eventSource.onopen = () => {
        console.log('SSE connection opened');
    };

    eventSource.onerror = (error) => {
        console.error('SSE connection error:', error);
        eventSource.close();
        // Try to reconnect after a delay
        setTimeout(connectToSSE, 5000);
    };
}

function handleLogMessage(data) {
    const logData = data.value;
    if (!logData || !logData.task_id || !logData.line) return;

    // If this log is for the currently viewed task
    if (currentTaskId === logData.task_id) {
        // Update total lines if needed
        if (logData.line_number > currentLogState.totalLines) {
            currentLogState.totalLines = logData.line_number;
        }

        // Add to buffer
        logBuffer.push({
            lineNumber: logData.line_number,
            message: logData.line.trim()
        });

        // If buffer is full or no render is pending, schedule a render
        if (logBuffer.length >= LOG_BUFFER_SIZE && !isRenderPending) {
            isRenderPending = true;
            requestAnimationFrame(flushLogBuffer);
        } else if (!isRenderPending) {
            isRenderPending = true;
            setTimeout(flushLogBuffer, RENDER_DELAY);
        }
    }
}

function flushLogBuffer() {
    if (logBuffer.length === 0) {
        isRenderPending = false;
        return;
    }

    const logsContainer = document.getElementById('logsContent');
    if (!logsContainer) {
        logBuffer = [];
        isRenderPending = false;
        return;
    }

    // Get or create the logs header
    let header = logsContainer.querySelector('.logs-header');
    if (!header) {
        header = document.createElement('div');
        header.className = 'logs-header';
        logsContainer.insertBefore(header, logsContainer.firstChild);
    }

    // Update the header
    header.textContent = `Showing lines 1-${currentLogState.totalLines} of ${currentLogState.totalLines}`;

    // Sort buffer by line number to ensure correct order
    logBuffer.sort((a, b) => a.lineNumber - b.lineNumber);

    // Check if we should auto-scroll
    const shouldAutoScroll = isScrolledToBottom(logsContainer);

    // Create or update log lines
    for (const log of logBuffer) {
        let logLine = logsContainer.querySelector(`.log-line[data-line="${log.lineNumber}"]`);
        
        if (!logLine) {
            logLine = document.createElement('div');
            logLine.className = 'log-line';
            logLine.setAttribute('data-line', log.lineNumber);
            
            const lineNumber = document.createElement('span');
            lineNumber.className = 'line-number';
            lineNumber.textContent = log.lineNumber;
            
            const message = document.createElement('span');
            message.className = 'log-message';
            message.textContent = log.message;
            
            logLine.appendChild(lineNumber);
            logLine.appendChild(message);
            logsContainer.appendChild(logLine);
        } else {
            // Update existing line if content is different
            const messageSpan = logLine.querySelector('.log-message');
            if (messageSpan && messageSpan.textContent !== log.message) {
                messageSpan.textContent = log.message;
            }
        }
    }

    // Auto-scroll if we were at the bottom
    if (shouldAutoScroll) {
        logsContainer.scrollTop = logsContainer.scrollHeight;
    }

    // Clear the buffer and reset the pending flag
    logBuffer = [];
    isRenderPending = false;

    // Trim old logs if there are too many
    trimOldLogs();
}

function isScrolledToBottom(element) {
    const threshold = 50; // pixels from bottom to consider "scrolled to bottom"
    return element.scrollHeight - element.scrollTop - element.clientHeight < threshold;
}

function trimOldLogs() {
    const logsContainer = document.getElementById('logsContent');
    const maxLines = 1000; // Maximum number of lines to keep in DOM
    const logLines = logsContainer.querySelectorAll('.log-line');
    
    if (logLines.length > maxLines) {
        const linesToRemove = logLines.length - maxLines;
        for (let i = 0; i < linesToRemove; i++) {
            logLines[i].remove();
        }
    }
}

function handleTaskStatusMessage(data) {
    const taskId = parseInt(data.task_id);
    const taskValue = data.value;
    
    // Find the task element and update its status
    const taskElement = document.querySelector(`.task-item[data-task-id="${taskId}"]`);
    if (taskElement) {
        const statusElement = taskElement.querySelector('.task-status');
        const newStatus = TaskStatus[taskValue.status];
        statusElement.textContent = newStatus;
        statusElement.className = `task-status status-${newStatus.toLowerCase()}`;

        // Get the task info container
        const taskInfo = taskElement.querySelector('.task-info');

        // Update exit code
        const exitCodeElement = taskElement.querySelector('.exit-code');
        if (exitCodeElement) {
            if (taskValue.exit_code !== undefined && taskValue.exit_code !== null) {
                exitCodeElement.innerHTML = `<strong>Exit Code:</strong> ${taskValue.exit_code}`;
                exitCodeElement.style.display = '';
            } else {
                exitCodeElement.style.display = 'none';
            }
        }

        // Update reason
        const reasonElement = taskElement.querySelector('.reason');
        if (reasonElement) {
            if (taskValue.reason) {
                reasonElement.innerHTML = `<strong>Reason:</strong> ${taskValue.reason}`;
                reasonElement.style.display = '';
            } else {
                reasonElement.style.display = 'none';
            }
        }

        // Update action buttons based on new status
        const actionButtons = taskElement.querySelector('.task-actions');
        const statusLower = newStatus.toLowerCase();
        if (statusLower === 'queued' || statusLower === 'running') {
            if (!actionButtons.querySelector('.cancel-btn')) {
                const cancelBtn = document.createElement('button');
                cancelBtn.className = 'cancel-btn';
                cancelBtn.textContent = 'Cancel';
                cancelBtn.onclick = () => cancelTask(taskId);
                actionButtons.appendChild(cancelBtn);
            }
        } else {
            const cancelBtn = actionButtons.querySelector('.cancel-btn');
            if (cancelBtn) {
                cancelBtn.remove();
            }
        }
    } else {
        // If the task element doesn't exist, refresh the entire task list
        fetchTasks();
    }
}

// Connect to SSE when the page loads
connectToSSE();

// Update cleanup when closing modal
function closeLogsModal() {
    logsModal.style.display = 'none';
    currentTaskId = null;
    resetLogState();
    logsContent.onscroll = null;
}

closeModalBtn.addEventListener('click', closeLogsModal);

// Cleanup SSE connection when the page is unloaded
window.addEventListener('beforeunload', () => {
    if (eventSource) {
        eventSource.close();
    }
});

async function cancelTask(taskId) {
    try {
        const response = await fetch(`${API_BASE_URL}/tasks/${taskId}/cancel`, {
            method: 'DELETE',
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const result = await response.json();
        if (!result.success) {
            throw new Error(result.error || 'Failed to cancel task');
        }

        // Task cancellation initiated successfully
        // The actual status update will come through SSE
    } catch (error) {
        console.error('Error cancelling task:', error);
        alert('Failed to cancel task. Please try again.');
    }
} 
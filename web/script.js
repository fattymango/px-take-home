// API Configuration
const API_BASE_URL = 'http://localhost:8888/api/v1';

// Pagination State
let paginationState = {
    currentPage: 1,
    pageSize: 10,
    totalTasks: 0
};

// Task Status Mapping
const TaskStatus = {
    1: 'Queued',
    2: 'Running',
    3: 'Completed',
    4: 'Failed',
    5: 'Canceled'
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
const prevPageBtn = document.getElementById('prevPage');
const nextPageBtn = document.getElementById('nextPage');
const pageSizeSelect = document.getElementById('pageSize');
const pageInfo = document.getElementById('pageInfo');

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
        const from = parseInt(fromLineInput.value) || 0;
        const to = parseInt(toLineInput.value) || 0;
        resetLogState();
        fetchLogs(currentTaskId, from, to);
    }
});
prevPageBtn.addEventListener('click', () => {
    if (paginationState.currentPage > 1) {
        paginationState.currentPage--;
        fetchTasks();
    }
});
nextPageBtn.addEventListener('click', () => {
    paginationState.currentPage++;
    fetchTasks();
});
pageSizeSelect.addEventListener('change', (e) => {
    paginationState.pageSize = parseInt(e.target.value);
    paginationState.currentPage = 1;  // Reset to first page when changing page size
    fetchTasks();
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
        // For page 1: offset should be 0
        // For page 2: offset should be pageSize
        // For page 3: offset should be pageSize * 2
        // etc.
        const offset = Math.max(0, (paginationState.currentPage - 1));
        const limit = paginationState.pageSize;
        const response = await fetch(`${API_BASE_URL}/tasks?offset=${offset}&limit=${limit}`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const result = await response.json();
        if (result.success) {
            renderTasks(result.data);
            updatePaginationControls(result.data);
        } else {
            throw new Error(result.error);
        }
    } catch (error) {
        console.error('Error fetching tasks:', error);
        tasksList.innerHTML = '<p class="error">Failed to load tasks. Please try again.</p>';
    }
}

// Helper function to format timestamps
function formatTimestamp(timestamp) {
    if (!timestamp) return 'N/A';
    const date = new Date(timestamp * 1000); // Convert from Unix timestamp to milliseconds
    return date.toLocaleString();
}

async function downloadLogs(taskId) {
    try {
        const response = await fetch(`${API_BASE_URL}/tasks/${taskId}/logs/download`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        // Get the filename from the Content-Disposition header if available
        const contentDisposition = response.headers.get('Content-Disposition');
        let filename = `task-${taskId}-logs.txt`;
        if (contentDisposition) {
            const matches = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/.exec(contentDisposition);
            if (matches != null && matches[1]) {
                filename = matches[1].replace(/['"]/g, '');
            }
        }

        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
    } catch (error) {
        console.error('Error downloading logs:', error);
        alert('Failed to download logs. Please try again.');
    }
}

function renderTasks(data) {
    const { tasks, total } = data;
    if (!tasks || tasks.length === 0) {
        tasksList.innerHTML = '<p>No tasks found.</p>';
        return;
    }

    tasksList.innerHTML = tasks.map(task => {
        const status = TaskStatus[task.status];
        const statusLower = status.toLowerCase();
        const isRunning = task.status === 1 || task.status === 2; // Queued or Running
        
        return `
            <div class="task-item">
                <div class="task-header">
                    <h3>${task.name}</h3>
                    <span class="status ${statusLower}">${status}</span>
                </div>
                <div class="task-details">
                    <p><strong>Command:</strong> ${task.command}</p>
                    <p><strong>Start Time:</strong> ${formatTimestamp(task.start_time)}</p>
                    <p><strong>End Time:</strong> ${formatTimestamp(task.end_time)}</p>
                    ${task.exit_code !== undefined && !isRunning ? 
                        `<p><strong>Exit Code:</strong> ${task.exit_code}</p>` : ''}
                    ${task.reason ? `<p><strong>Reason:</strong> ${task.reason}</p>` : ''}
                </div>
                <div class="task-actions">
                    <button onclick="showLogs(${task.id})">View Logs</button>
                    ${!isRunning ? 
                        `<button class="download-btn" onclick="downloadLogs(${task.id})">Download Logs</button>` : ''}
                    ${isRunning ? 
                        `<button class="cancel-btn" onclick="cancelTask(${task.id})">Cancel</button>` : ''}
                </div>
            </div>
        `;
    }).join('');
}

function updatePaginationControls(data) {
    const { tasks, total } = data;
    paginationState.totalTasks = total;
    
    // Disable previous button if we're on the first page
    prevPageBtn.disabled = paginationState.currentPage === 1;
    
    // Calculate total pages
    const totalPages = Math.ceil(total / paginationState.pageSize);
    
    // Disable next button if we're on the last page
    nextPageBtn.disabled = paginationState.currentPage >= totalPages;
    
    // Update page info with more details
    pageInfo.textContent = `Page ${paginationState.currentPage} of ${totalPages} (${total} total tasks)`;
}

async function showLogs(taskId) {
    currentTaskId = taskId;
    resetLogState();
    currentLogState.taskId = taskId;
    logsModal.style.display = 'block';
    logsContent.innerHTML = 'Loading logs...';
    
    // Fetch initial logs
    await fetchLogs(taskId);
    
    // Clear any existing scroll event listener
    logsContent.onscroll = null;
    
    // Add scroll event listener for infinite scrolling
    logsContent.onscroll = () => {
        if (logsContent.scrollTop === 0 && currentLogState.hasMoreAbove && !isLoadingLogs) {
            const newFrom = Math.max(1, currentLogState.loadedLines.from - currentLogState.batchSize);
            const newTo = currentLogState.loadedLines.from - 1;
            if (newFrom < newTo) {
                fetchLogs(taskId, newFrom, newTo);
            }
        }
    };
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
            // console.debug('Parsed SSE message:', data);
            
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
    const taskId = parseInt(data.task_id);
    if (taskId === currentTaskId) {
        // Add to buffer with line number
        const logValue = data.value;
        const logEntry = {
            line: logValue.line,
            lineNumber: logValue.line_number
        };

        logBuffer.push(logEntry);
        
        // Update state
        currentLogState.totalLines++;
        currentLogState.loadedLines.to = currentLogState.totalLines;

        // Schedule render if not already pending
        if (!isRenderPending) {
            isRenderPending = true;
            requestAnimationFrame(flushLogBuffer);
        }

        // Force render if buffer is getting too large
        if (logBuffer.length >= LOG_BUFFER_SIZE) {
            flushLogBuffer();
        }
    }
}

function flushLogBuffer() {
    if (logBuffer.length === 0) {
        isRenderPending = false;
        return;
    }

    const fragment = document.createDocumentFragment();
    const shouldScroll = isScrolledToBottom();

    // Create elements for all buffered logs
    logBuffer.forEach(log => {
        const logLine = document.createElement('div');
        logLine.className = 'log-line';
        logLine.innerHTML = `<span class="log-message">${log.line}</span>`;
        fragment.appendChild(logLine);
    });

    // Update DOM in a single operation
    logsContent.appendChild(fragment);

    // Update header
    const header = document.querySelector('.logs-header');
    if (header) {
        header.textContent = `Showing lines ${currentLogState.loadedLines.from}-${currentLogState.loadedLines.to} of ${currentLogState.totalLines}`;
    }

    // Maintain scroll position if user was at bottom
    if (shouldScroll) {
        logsContent.scrollTop = logsContent.scrollHeight;
    }

    // Clear buffer
    logBuffer = [];
    isRenderPending = false;
}

function isScrolledToBottom() {
    const threshold = 50; // pixels from bottom to consider "scrolled to bottom"
    return logsContent.scrollHeight - logsContent.scrollTop - logsContent.clientHeight < threshold;
}

// Virtual scrolling implementation
const VIRTUAL_SCROLL_BUFFER = 1000; // Maximum number of visible log lines
function trimOldLogs() {
    const logElements = logsContent.getElementsByTagName('div');
    if (logElements.length > VIRTUAL_SCROLL_BUFFER) {
        const numToRemove = logElements.length - VIRTUAL_SCROLL_BUFFER;
        for (let i = 0; i < numToRemove; i++) {
            if (logElements[1]) { // Skip header element
                logElements[1].remove();
            }
        }
    }
}

// Add periodic cleanup to prevent memory issues
setInterval(trimOldLogs, 5000);

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
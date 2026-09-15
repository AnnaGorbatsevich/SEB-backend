import React, { useState, useEffect } from 'react';
import './App.css';

const BASE_URL = process.env.BACKEND_BASE_URL;

const getSessionEvents = async (sessionId) => {
  const response = await fetch(`${BASE_URL}/session-events?session_id=${sessionId}`);
  if (!response.ok) throw new Error(`Error: ${response.status}`);
  return response.json();
};

const getCursorEvents = async (sessionId, email) => {
  const response = await fetch(`${BASE_URL}/cursors?session_id=${sessionId}&email=${email}`);
  if (!response.ok) throw new Error(`Error: ${response.status}`);
  return response.json();
};

const getKeypressEvents = async (sessionId, email) => {
  const response = await fetch(`${BASE_URL}/keypresses?session_id=${sessionId}&email=${email}`);
  if (!response.ok) throw new Error(`Error: ${response.status}`);
  return response.json();
};

const getLogEvents = async (sessionId, email) => {
  const response = await fetch(`${BASE_URL}/logs?session_id=${sessionId}&email=${email}`);
  if (!response.ok) throw new Error(`Error: ${response.status}`);
  return response.json();
};

const getDiagnosticSummary = async (sessionId) => {
  const response = await fetch(`${BASE_URL}/diagnostic-summary?session_id=${sessionId}`);
  if (!response.ok) throw new Error(`Error: ${response.status}`);
  return response.json();
};

const getDiagnosticsEvents = async (sessionId, email) => {
  const response = await fetch(`${BASE_URL}/diagnostics?session_id=${sessionId}&email=${email}&limit=1000`);
  if (!response.ok) throw new Error(`Error: ${response.status}`);
  return response.json();
};

// Component for session entry
const SessionEntry = ({ onSubmit }) => {
  const [sessionId, setSessionId] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = (e) => {
    e.preventDefault();
    if (!sessionId.trim()) {
      setError('Session ID is required');
      return;
    }
    setError('');
    onSubmit(sessionId);
  };

  return (
    <div className="container">
      <div className="card">
        <div className="header">
          <h1>Teacher Monitoring Dashboard</h1>
        </div>
        <div className="card-body" style={{ padding: '2rem' }}>
          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label htmlFor="sessionId">Session ID</label>
              <input
                type="text"
                id="sessionId"
                className="form-control"
                value={sessionId}
                onChange={(e) => setSessionId(e.target.value)}
                placeholder="Enter session ID"
              />
              {error && <div style={{ color: 'red', fontSize: '0.875rem', marginTop: '0.25rem' }}>{error}</div>}
            </div>
            <button type="submit" className="btn">Enter Session</button>
          </form>
        </div>
      </div>
    </div>
  );
};

// Component for student list
const StudentList = ({ sessionId, onStudentClick }) => {
  const [students, setStudents] = useState([]);
  const [diagSummary, setDiagSummary] = useState({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const prevStudentsRef = React.useRef([]);

  const fetchStudents = async () => {
    try {
      const [eventsData, summaryData] = await Promise.all([
        getSessionEvents(sessionId),
        getDiagnosticSummary(sessionId),
      ]);
      if (JSON.stringify(prevStudentsRef.current) !== JSON.stringify(eventsData)) {
        setStudents(eventsData);
        prevStudentsRef.current = eventsData;
      }
      const summaryMap = {};
      summaryData.forEach(s => { summaryMap[s.email] = s; });
      setDiagSummary(summaryMap);
      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // Initial fetch
    fetchStudents();
    
    // Set up polling to refresh every second
    const interval = setInterval(fetchStudents, 1000);
    
    // Clean up interval on unmount
    return () => clearInterval(interval);
  }, [sessionId]);

  const getEventTypeName = (eventType) => {
    switch (eventType) {
      case 'cursor_position': return 'Cursor Movement';
      case 'key_press': return 'Keypress';
      case 'log': return 'System Log';
      default: return eventType;
    }
  };

  const getEventBadgeClass = (eventType) => {
    switch (eventType) {
      case 'cursor_position': return 'event-type-badge cursor-position';
      case 'key_press': return 'event-type-badge key-press';
      case 'log': return 'event-type-badge log';
      default: return 'event-type-badge';
    }
  };

  if (loading) {
    return (
      <div className="container">
        <div className="card">
          <div className="header">
            <h1>Student Monitoring</h1>
          </div>
          <div className="card-body" style={{ padding: '2rem', textAlign: 'center' }}>
            <p>Loading students...</p>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container">
        <div className="card">
          <div className="header">
            <h1>Student Monitoring</h1>
          </div>
          <div className="card-body" style={{ padding: '2rem' }}>
            <div style={{ color: 'red', marginBottom: '1rem' }}>{error}</div>
            <button className="btn" onClick={fetchStudents}>Retry</button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="container">
      <div className="card">
        <div className="header">
          <h1>Student Monitoring</h1>
        </div>
        <div className="card-body">
          <ul className="student-list" style={{ padding: '0 1rem' }}>
            {students.length === 0 ? (
              <li className="student-item" style={{ textAlign: 'center', color: '#718096' }}>
                No students found
              </li>
            ) : (
              students.map((student) => (
                <li 
                  key={student.email} 
                  className="student-item"
                  onClick={() => onStudentClick(student.email)}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <div className="student-email">{student.email}</div>
                    {diagSummary[student.email] && (
                      <div className="diag-dots">
                        {diagSummary[student.email].crit > 0 && (
                          <span className="diag-dot" style={{ color: statusColor('CRIT'), backgroundColor: statusBg('CRIT') }}>
                            {diagSummary[student.email].crit}
                          </span>
                        )}
                        {diagSummary[student.email].warn > 0 && (
                          <span className="diag-dot" style={{ color: statusColor('WARN'), backgroundColor: statusBg('WARN') }}>
                            {diagSummary[student.email].warn}
                          </span>
                        )}
                        {diagSummary[student.email].ok > 0 && (
                          <span className="diag-dot" style={{ color: statusColor('OK'), backgroundColor: statusBg('OK') }}>
                            {diagSummary[student.email].ok}
                          </span>
                        )}
                      </div>
                    )}
                  </div>
                  <div className="event-type">
                    <span className={getEventBadgeClass(student.event_type)}>
                      {getEventTypeName(student.event_type)}
                    </span>
                  </div>
                  <div className="event-time">
                    {new Date(student.event_time).toLocaleDateString('en-US', {
                      year: 'numeric',
                      month: 'short',
                      day: 'numeric',
                      hour: '2-digit',
                      minute: '2-digit'
                    })}
                  </div>
                </li>
              ))
            )}
          </ul>
        </div>
      </div>
    </div>
  );
};

const formatTimestamp = (ts) => {
  return new Date(ts).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });
};

const statusColor = (status) => {
  switch ((status || '').toUpperCase()) {
    case 'OK':   return '#16a34a';
    case 'WARN': return '#ca8a04';
    case 'CRIT': return '#dc2626';
    default:     return '#64748b';
  }
};

const statusBg = (status) => {
  switch ((status || '').toUpperCase()) {
    case 'OK':   return '#dcfce7';
    case 'WARN': return '#fef9c3';
    case 'CRIT': return '#fee2e2';
    default:     return '#f1f5f9';
  }
};

const worstStatus = (items) => {
  const statuses = items.map(i => (i.status || '').toUpperCase());
  if (statuses.includes('CRIT')) return 'CRIT';
  if (statuses.includes('WARN')) return 'WARN';
  if (statuses.includes('OK')) return 'OK';
  return items[0]?.status || '';
};

const buildIntervals = (items) => {
  const sorted = [...items].reverse(); // items come DESC, need ASC
  if (sorted.length === 0) return [];

  const lastMs = new Date(sorted[sorted.length - 1].received_at).getTime();
  const lastEnd = new Date(lastMs + ACTIVE_MS).toISOString();

  // Pass 1: one raw segment per item, with gap detection
  const raw = [];
  for (let i = 0; i < sorted.length; i++) {
    const cur = sorted[i];
    const next = sorted[i + 1];
    const curMs = new Date(cur.received_at).getTime();

    if (!next) {
      raw.push({ status: cur.status, details: cur.details, from: cur.received_at, to: lastEnd, isGap: false });
    } else {
      const gap = new Date(next.received_at).getTime() - curMs;
      if (gap > GAP_MS) {
        const activeEnd = new Date(curMs + ACTIVE_MS).toISOString();
        raw.push({ status: cur.status, details: cur.details, from: cur.received_at, to: activeEnd, isGap: false });
        raw.push({ status: null, details: null, from: activeEnd, to: next.received_at, isGap: true });
      } else {
        raw.push({ status: cur.status, details: cur.details, from: cur.received_at, to: next.received_at, isGap: false });
      }
    }
  }

  // Pass 2: merge consecutive non-gap segments with same (status, details)
  const merged = [];
  for (const seg of raw) {
    const prev = merged[merged.length - 1];
    if (prev && !seg.isGap && !prev.isGap &&
        seg.status === prev.status &&
        JSON.stringify(seg.details) === JSON.stringify(prev.details)) {
      prev.to = seg.to;
    } else {
      merged.push({ ...seg });
    }
  }
  return merged;
};

const GAP_MS = 15000; // >15s between diagnostics = gap (3× expected 5s interval)
const ACTIVE_MS = 5000; // one expected diagnostic interval

const DiagnosticTimeline = ({ items }) => {
  if (!items || items.length === 0) return null;

  const sorted = [...items].reverse(); // ASC

  const firstMs = new Date(sorted[0].received_at).getTime();
  const lastMs = new Date(sorted[sorted.length - 1].received_at).getTime();
  // Add one active interval after the last item so it's visible
  const endMs = lastMs + ACTIVE_MS;
  const totalMs = endMs - firstMs;

  if (totalMs <= 0) return null;

  const segments = [];
  for (let i = 0; i < sorted.length; i++) {
    const segStart = new Date(sorted[i].received_at).getTime();
    const segEnd = i < sorted.length - 1
      ? new Date(sorted[i + 1].received_at).getTime()
      : endMs;
    const duration = segEnd - segStart;

    if (i < sorted.length - 1 && duration > GAP_MS) {
      // Show active period, then gray gap
      const activeWidth = (ACTIVE_MS / totalMs) * 100;
      const gapWidth = ((duration - ACTIVE_MS) / totalMs) * 100;
      segments.push({ color: statusColor(sorted[i].status), width: activeWidth, label: `${sorted[i].status}: ${formatTimestamp(sorted[i].received_at)}` });
      segments.push({ color: '#cbd5e1', width: gapWidth, label: `No data: ${Math.round(duration / 1000)}s` });
    } else {
      segments.push({ color: statusColor(sorted[i].status), width: (duration / totalMs) * 100, label: `${sorted[i].status}: ${formatTimestamp(sorted[i].received_at)}` });
    }
  }

  return (
    <div className="diagnostic-timeline">
      <div className="timeline-bar">
        {segments.map((seg, i) => (
          <div
            key={i}
            className="timeline-segment"
            style={{ width: `${seg.width}%`, backgroundColor: seg.color }}
            title={seg.label}
          />
        ))}
      </div>
      <div className="timeline-labels">
        <span>{formatTimestamp(sorted[0].received_at)}</span>
        <span>{formatTimestamp(sorted[sorted.length - 1].received_at)}</span>
      </div>
    </div>
  );
};

// Component for student detail
const StudentDetail = ({ email, sessionId, onBack }) => {
  const [events, setEvents] = useState({
    cursor: [],
    keypress: [],
    log: [],
    diagnostic: [],
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [openSections, setOpenSections] = useState({
    cursor: true,
    keypress: true,
    log: true,
  });
  const [activeTab, setActiveTab] = useState('events');
  const [openCodes, setOpenCodes] = useState({});

  const fetchAllEvents = async () => {
    try {
      const [cursorResponse, keypressResponse, logResponse, diagnosticResponse] = await Promise.all([
        getCursorEvents(sessionId, email),
        getKeypressEvents(sessionId, email),
        getLogEvents(sessionId, email),
        getDiagnosticsEvents(sessionId, email),
      ]);

      setEvents({
        cursor: cursorResponse.data,
        keypress: keypressResponse.data,
        log: logResponse.data,
        diagnostic: diagnosticResponse.data,
      });

      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // Initial fetch
    fetchAllEvents();

    // Set up polling to refresh every second
    const interval = setInterval(fetchAllEvents, 1000);

    // Clean up interval on unmount
    return () => clearInterval(interval);
  }, [sessionId, email]);

  const toggleSection = (section) => {
    setOpenSections(prev => ({
      ...prev,
      [section]: !prev[section]
    }));
  };

  if (loading) {
    return (
      <div className="container">
        <div className="card">
          <div className="header">
            <button className="back-button" onClick={onBack}>
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M19 12H5M12 19l-7-7 7-7" />
              </svg>
              Back
            </button>
            <h1>Details for {email}</h1>
          </div>
          <div className="card-body" style={{ padding: '2rem', textAlign: 'center' }}>
            <p>Loading events...</p>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container">
        <div className="card">
          <div className="header">
            <button className="back-button" onClick={onBack}>
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M19 12H5M12 19l-7-7 7-7" />
              </svg>
              Back
            </button>
            <h1>Details for {email}</h1>
          </div>
          <div className="card-body" style={{ padding: '2rem' }}>
            <div style={{ color: 'red', marginBottom: '1rem' }}>{error}</div>
            <button className="btn" onClick={fetchAllEvents}>Retry</button>
          </div>
        </div>
      </div>
    );
  }

  const groupedDiagnostics = events.diagnostic.reduce((acc, item) => {
    if (!acc[item.code]) acc[item.code] = [];
    acc[item.code].push(item);
    return acc;
  }, {});

  const toggleCode = (code) => {
    setOpenCodes(prev => ({ ...prev, [code]: !prev[code] }));
  };

  return (
    <div className="container">
      <div className="card">
        <div className="header">
          <button className="back-button" onClick={onBack}>
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M19 12H5M12 19l-7-7 7-7" />
            </svg>
            Back
          </button>
          <h1>Details for {email}</h1>
        </div>

        <div className="tab-nav">
          <button
            className={`tab-btn ${activeTab === 'events' ? 'active' : ''}`}
            onClick={() => setActiveTab('events')}
          >
            Events
          </button>
          <button
            className={`tab-btn ${activeTab === 'diagnostics' ? 'active' : ''}`}
            onClick={() => setActiveTab('diagnostics')}
          >
            Diagnostics
          </button>
        </div>

        <div className="card-body">
          {activeTab === 'events' && (
            <div className="events-container" style={{ padding: '0 1rem' }}>
              {/* Cursor Events */}
              <div className="collapsible-section">
                <div
                  className={`collapsible-header ${openSections.cursor ? 'open' : ''}`}
                  onClick={() => toggleSection('cursor')}
                >
                  <div className="event-type-header">
                    <span>Cursor Movements events</span>
                  </div>
                </div>
                <div className={`collapsible-content ${openSections.cursor ? 'open' : ''}`}>
                  {events.cursor.length > 0 ? (
                    events.cursor.map((event) => (
                      <div key={event.id} className="event-item">
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                          <div>
                            <span className="event-key">Position:</span>
                            <span className="event-value" style={{ marginLeft: '0.5rem' }}>
                              ({event.x}, {event.y})
                            </span>
                          </div>
                          <div className="event-timestamp">{formatTimestamp(event.ts)}</div>
                        </div>
                      </div>
                    ))
                  ) : (
                    <div className="no-events-message">No cursor movements recorded</div>
                  )}
                </div>
              </div>

              {/* Keypress Events */}
              <div className="collapsible-section">
                <div
                  className={`collapsible-header ${openSections.keypress ? 'open' : ''}`}
                  onClick={() => toggleSection('keypress')}
                >
                  <div className="event-type-header">
                    <span>Keypresses events</span>
                  </div>
                </div>
                <div className={`collapsible-content ${openSections.keypress ? 'open' : ''}`}>
                  {events.keypress.length > 0 ? (
                    events.keypress.map((event) => (
                      <div key={event.id} className="event-item">
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                          <div>
                            <span className="event-key">Key:</span>
                            <span className="event-value" style={{ marginLeft: '0.5rem' }}>
                              {event.key_name}
                              {event.modifiers.length > 0 && ` (${event.modifiers.join(', ')})`}
                            </span>
                          </div>
                          <div className="event-timestamp">{formatTimestamp(event.ts)}</div>
                        </div>
                      </div>
                    ))
                  ) : (
                    <div className="no-events-message">No keypresses recorded</div>
                  )}
                </div>
              </div>

              {/* Log Events */}
              <div className="collapsible-section">
                <div
                  className={`collapsible-header ${openSections.log ? 'open' : ''}`}
                  onClick={() => toggleSection('log')}
                >
                  <div className="event-type-header">
                    <span>System Logs events</span>
                  </div>
                </div>
                <div className={`collapsible-content ${openSections.log ? 'open' : ''}`}>
                  {events.log.length > 0 ? (
                    events.log.map((event) => (
                      <div key={event.id} className="event-item">
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                          <div>
                            <span className="event-key">{event.level}:</span>
                            <span className="event-value" style={{ marginLeft: '0.5rem' }}>{event.message}</span>
                          </div>
                          <div className="event-timestamp">{formatTimestamp(event.ts)}</div>
                        </div>
                      </div>
                    ))
                  ) : (
                    <div className="no-events-message">No system logs recorded</div>
                  )}
                </div>
              </div>

              {events.cursor.length === 0 && events.keypress.length === 0 && events.log.length === 0 && (
                <div style={{ textAlign: 'center', color: '#64748b', padding: '2rem' }}>
                  No events recorded for this student yet.
                </div>
              )}
            </div>
          )}

          {activeTab === 'diagnostics' && (() => {
            if (Object.keys(groupedDiagnostics).length === 0) {
              return <div className="events-container" style={{ padding: '0 1rem' }}><div className="no-events-message">No diagnostics recorded</div></div>;
            }
            const summary = { CRIT: 0, WARN: 0, OK: 0 };
            Object.values(groupedDiagnostics).forEach(items => {
              const w = worstStatus(items).toUpperCase();
              if (w in summary) summary[w]++;
            });
            return (
              <div className="events-container" style={{ padding: '0 1rem' }}>
                <div className="diag-summary">
                  {summary.CRIT > 0 && (
                    <span className="diag-summary-item" style={{ color: statusColor('CRIT'), backgroundColor: statusBg('CRIT') }}>
                      CRIT: {summary.CRIT}
                    </span>
                  )}
                  {summary.WARN > 0 && (
                    <span className="diag-summary-item" style={{ color: statusColor('WARN'), backgroundColor: statusBg('WARN') }}>
                      WARN: {summary.WARN}
                    </span>
                  )}
                  {summary.OK > 0 && (
                    <span className="diag-summary-item" style={{ color: statusColor('OK'), backgroundColor: statusBg('OK') }}>
                      OK: {summary.OK}
                    </span>
                  )}
                </div>
                {Object.entries(groupedDiagnostics).map(([code, items]) => {
                  const worst = worstStatus(items);
                  const intervals = buildIntervals(items);
                  const isOpen = !!openCodes[code];
                  return (
                    <div key={code} className="collapsible-section">
                      <div
                        className={`collapsible-header ${isOpen ? 'open' : ''}`}
                        onClick={() => toggleCode(code)}
                      >
                        <span>{code}</span>
                        <span
                          className="diag-status-badge"
                          style={{
                            color: statusColor(worst),
                            backgroundColor: statusBg(worst),
                          }}
                        >
                          {worst}
                        </span>
                      </div>
                      <DiagnosticTimeline items={items} />
                      {isOpen && (
                        <div className="diagnostic-items">
                          {intervals.map((interval, i) => (
                            <div key={i} className={`event-item diagnostic-item${interval.isGap ? ' diagnostic-gap-row' : ''}`}>
                              {interval.isGap ? (
                                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                  <span style={{ color: '#94a3b8', fontSize: '0.8rem', fontStyle: 'italic' }}>no data</span>
                                  <span className="event-timestamp" style={{ whiteSpace: 'nowrap' }}>
                                    {formatTimestamp(interval.from)} — {formatTimestamp(interval.to)}
                                  </span>
                                </div>
                              ) : (
                                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: '1rem' }}>
                                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', minWidth: 0 }}>
                                    <span
                                      className="diag-status-badge"
                                      style={{
                                        color: statusColor(interval.status),
                                        backgroundColor: statusBg(interval.status),
                                        flexShrink: 0,
                                      }}
                                    >
                                      {interval.status}
                                    </span>
                                    <span className="interval-details">{JSON.stringify(interval.details)}</span>
                                  </div>
                                  <span className="event-timestamp" style={{ whiteSpace: 'nowrap', flexShrink: 0 }}>
                                    {formatTimestamp(interval.from)} — {formatTimestamp(interval.to)}
                                  </span>
                                </div>
                              )}
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            );
          })()}
        </div>
      </div>
    </div>
  );
};

// Main App component
const App = () => {
  const [sessionId, setSessionId] = useState('');
  const [currentEmail, setCurrentEmail] = useState(null);

  const handleSessionSubmit = (id) => {
    setSessionId(id);
  };

  const handleStudentClick = (email) => {
    setCurrentEmail(email);
  };

  const handleBack = () => {
    setCurrentEmail(null);
  };

  if (!sessionId) {
    return <SessionEntry onSubmit={handleSessionSubmit} />;
  }

  if (currentEmail) {
    return <StudentDetail email={currentEmail} sessionId={sessionId} onBack={handleBack} />;
  }

  return <StudentList sessionId={sessionId} onStudentClick={handleStudentClick} />;
};

export default App;
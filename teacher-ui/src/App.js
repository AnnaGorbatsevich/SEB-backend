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

const StudentList = ({ sessionId, onStudentClick }) => {
  const [students, setStudents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const prevStudentsRef = React.useRef([]);

  const fetchStudents = async () => {
    try {
      const data = await getSessionEvents(sessionId);
      if (JSON.stringify(prevStudentsRef.current) !== JSON.stringify(data)) {
        setStudents(data);
        prevStudentsRef.current = data;
      }
      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStudents();
    
    const interval = setInterval(fetchStudents, 1000);
    
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
                  <div className="student-email">{student.email}</div>
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

const StudentDetail = ({ email, sessionId, onBack }) => {
  const [events, setEvents] = useState({
    cursor: [],
    keypress: [],
    log: []
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [openSections, setOpenSections] = useState({
    cursor: true,
    keypress: true,
    log: true
  });

  const fetchAllEvents = async () => {
    try {
      const [cursorResponse, keypressResponse, logResponse] = await Promise.all([
        getCursorEvents(sessionId, email),
        getKeypressEvents(sessionId, email),
        getLogEvents(sessionId, email)
      ]);
      
      setEvents({
        cursor: cursorResponse.data,
        keypress: keypressResponse.data,
        log: logResponse.data
      });
      
      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAllEvents();

    const interval = setInterval(fetchAllEvents, 1000);

    return () => clearInterval(interval);
  }, [sessionId, email]);

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
        <div className="card-body">
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
                        <div className="event-timestamp">
                          {formatTimestamp(event.ts)}
                        </div>
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
                        <div className="event-timestamp">
                          {formatTimestamp(event.ts)}
                        </div>
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
                        <div className="event-timestamp">
                          {formatTimestamp(event.ts)}
                        </div>
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="no-events-message">No system logs recorded</div>
                )}
              </div>
            </div>
            
            {/* No events message - only show if ALL event types are empty */}
            {events.cursor.length === 0 && events.keypress.length === 0 && events.log.length === 0 && (
              <div style={{ textAlign: 'center', color: '#64748b', padding: '2rem' }}>
                No events recorded for this student yet.
              </div>
            )}
            
          </div>
        </div>
      </div>
    </div>
  );
};

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
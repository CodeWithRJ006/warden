import { useEffect, useState } from 'react';
import { Shield, Activity, Database, CheckCircle, XCircle, AlertTriangle } from 'lucide-react';
import './App.css';

interface Trace {
  request_id: string;
  actor: string;
  requested_tool: string;
  policy_version: string;
  data_classification: string;
  policy: {
    decision: string;
    rule: string;
  };
  model: {
    invoked: boolean;
    model_name?: string;
    status?: string;
  };
  tool: {
    executed: boolean;
    status?: string;
  };
  final_decision: string;
  timestamp: string;
}

function DecisionBadge({ decision }: { decision: string }) {
  if (decision === 'ALLOW') return <div className="badge allow"><CheckCircle size={12} className="mr-1" /> ALLOW</div>;
  if (decision === 'DENY' || decision === 'ERROR') return <div className="badge deny"><XCircle size={12} className="mr-1" /> {decision}</div>;
  if (decision === 'REDACT_AND_ALLOW') return <div className="badge redact"><AlertTriangle size={12} className="mr-1" /> REDACT</div>;
  return <div className="badge clean">{decision}</div>;
}

export default function App() {
  const [traces, setTraces] = useState<Trace[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchLogs = async () => {
      try {
        const res = await fetch('http://localhost:8081/v1/audit/logs');
        if (res.ok) {
          const data = await res.json();
          setTraces(data || []);
        }
      } catch (err) {
        console.error("Failed to fetch logs", err);
      } finally {
        setLoading(false);
      }
    };

    fetchLogs();
    const interval = setInterval(fetchLogs, 2000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="dashboard">
      <header className="header">
        <div className="title">
          <Shield color="var(--accent)" />
          Warden Control Plane
        </div>
          <div className="tabs">
            <button className="tab active">Live Decision Trace</button>
          </div>
      </header>

      <main>
        {loading ? (
          <div style={{color: 'var(--text-muted)'}}>Loading traces...</div>
        ) : traces.length === 0 ? (
          <div style={{color: 'var(--text-muted)'}}>No audit events found. Send a request to the gateway to see traces.</div>
        ) : (
          <div className="trace-list">
            {traces.map((trace) => (
              <div key={trace.request_id} className="trace-card">
                <div className="trace-header">
                  <div className="trace-id">
                    <span style={{color: 'var(--text-muted)', marginRight: '8px'}}>{new Date(trace.timestamp).toLocaleTimeString()}</span>
                    {trace.request_id}
                  </div>
                  <DecisionBadge decision={trace.final_decision} />
                </div>
                
                <div className="trace-body">
                  <div className="trace-section">
                    <span className="section-label">Identity & Context</span>
                    <div className="code-block">Actor: {trace.actor}</div>
                  </div>

                  <div className="trace-section">
                    <span className="section-label">PII Processing</span>
                    <div className={`badge ${trace.data_classification}`}>
                      {trace.data_classification.toUpperCase()}
                    </div>
                    <div className="code-block" style={{marginTop: '0.5rem', wordBreak: 'break-all'}}>
                      {trace.requested_tool}
                    </div>
                  </div>

                  <div className="trace-section">
                    <span className="section-label">Model Routing</span>
                    {trace.model.invoked ? (
                       <div className="code-block">
                         Routed to: {trace.model.model_name || 'unknown'}<br/>
                         Status: {trace.model.status || 'OK'}
                       </div>
                    ) : (
                       <div className="code-block" style={{color: 'var(--text-muted)'}}>
                         Bypassed / Failed
                       </div>
                    )}
                  </div>

                  <div className="trace-section">
                    <span className="section-label">Policy Engine</span>
                    <div className="code-block">
                      Rule: {trace.policy.rule}<br/>
                      Result: {trace.policy.decision}
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}

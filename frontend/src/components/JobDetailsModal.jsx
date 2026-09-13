import React, { useEffect, useState } from 'react';
import { fetchJobAttempts } from '../api/client';

export default function JobDetailsModal({ job, onClose }) {
  const [attempts, setAttempts] = useState([]);
  const [loadingAttempts, setLoadingAttempts] = useState(true);

  useEffect(() => {
    if (!job) return;

    let isMounted = true;
    fetchJobAttempts(job.id)
      .then((data) => {
        if (isMounted) {
          setAttempts(data || []);
          setLoadingAttempts(false);
        }
      })
      .catch(() => {
        if (isMounted) setLoadingAttempts(false);
      });

    return () => {
      isMounted = false;
    };
  }, [job]);

  if (!job) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-sm">
      <div className="bg-slate-900 border border-slate-800 rounded-xl max-w-2xl w-full p-6 max-h-[90vh] overflow-y-auto shadow-2xl">
        <div className="flex items-center justify-between border-b border-slate-800 pb-4 mb-4">
          <div>
            <h3 className="text-base font-semibold text-white">Job Details</h3>
            <p className="text-xs text-slate-400 font-mono mt-0.5">{job.id}</p>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Overview Grid */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-5">
          <div className="p-3 bg-slate-950/60 rounded-lg border border-slate-800">
            <span className="text-[10px] uppercase text-slate-400 font-medium">Type</span>
            <p className="text-xs font-semibold text-white capitalize mt-0.5">{job.type}</p>
          </div>
          <div className="p-3 bg-slate-950/60 rounded-lg border border-slate-800">
            <span className="text-[10px] uppercase text-slate-400 font-medium">Status</span>
            <p className="text-xs font-semibold text-white capitalize mt-0.5">{job.status}</p>
          </div>
          <div className="p-3 bg-slate-950/60 rounded-lg border border-slate-800">
            <span className="text-[10px] uppercase text-slate-400 font-medium">Attempts</span>
            <p className="text-xs font-semibold text-white mt-0.5">{job.attempts} / {job.max_attempts}</p>
          </div>
          <div className="p-3 bg-slate-950/60 rounded-lg border border-slate-800">
            <span className="text-[10px] uppercase text-slate-400 font-medium">Created</span>
            <p className="text-xs font-semibold text-white mt-0.5">{new Date(job.created_at).toLocaleTimeString()}</p>
          </div>
        </div>

        {/* Timestamps */}
        <div className="text-xs text-slate-400 space-y-1 mb-5 p-3 bg-slate-950/30 rounded-lg border border-slate-800/60">
          <div><span className="text-slate-500">Started At:</span> {job.started_at ? new Date(job.started_at).toLocaleString() : 'Not started'}</div>
          <div><span className="text-slate-500">Completed At:</span> {job.completed_at ? new Date(job.completed_at).toLocaleString() : 'In-progress / Pending'}</div>
          {job.last_error && (
            <div className="text-rose-400 mt-2">
              <span className="text-rose-500 font-medium">Last Error:</span> {job.last_error}
            </div>
          )}
        </div>

        {/* Payload */}
        <div className="mb-5">
          <h4 className="text-xs font-medium text-slate-300 uppercase tracking-wider mb-2">Payload</h4>
          <pre className="p-3 bg-slate-950 rounded-lg border border-slate-800 font-mono text-xs text-slate-200 overflow-x-auto">
            {JSON.stringify(job.payload, null, 2)}
          </pre>
        </div>

        {/* Attempt History */}
        <div>
          <h4 className="text-xs font-medium text-slate-300 uppercase tracking-wider mb-2 flex items-center justify-between">
            <span>Attempt History</span>
            <span className="text-slate-500">{attempts.length} attempts recorded</span>
          </h4>

          {loadingAttempts ? (
            <div className="text-xs text-slate-500 p-4 text-center">Loading attempt records...</div>
          ) : attempts.length === 0 ? (
            <div className="text-xs text-slate-500 p-4 text-center border border-dashed border-slate-800 rounded-lg">
              No attempt history logged yet.
            </div>
          ) : (
            <div className="space-y-2">
              {attempts.map((att) => (
                <div key={att.id} className="p-3 rounded-lg bg-slate-950/60 border border-slate-800 text-xs flex items-start justify-between">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-semibold text-white">Attempt #{att.attempt}</span>
                      <span className="text-[10px] px-2 py-0.5 rounded-full bg-slate-800 border border-slate-700 capitalize text-slate-300">
                        {att.status}
                      </span>
                    </div>
                    {att.error && <p className="text-rose-400 mt-1 font-mono">{att.error}</p>}
                  </div>
                  <div className="text-right text-[11px] text-slate-500">
                    <div>{new Date(att.started_at).toLocaleTimeString()}</div>
                    {att.completed_at && <div>Finished: {new Date(att.completed_at).toLocaleTimeString()}</div>}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

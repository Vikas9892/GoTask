import React from 'react';

const STATUS_BADGES = {
  pending: 'bg-amber-950/80 text-amber-300 border-amber-800',
  processing: 'bg-blue-950/80 text-blue-300 border-blue-800 animate-pulse',
  completed: 'bg-emerald-950/80 text-emerald-300 border-emerald-800',
  failed: 'bg-rose-950/80 text-rose-300 border-rose-800',
};

export default function JobList({ jobs, total, loading, error, onRefresh, onSelectJob, onDeleteJob }) {
  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-sm">
      <div className="p-6 border-b border-slate-800 flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-white">Job Queue History</h2>
          <p className="text-xs text-slate-400 mt-0.5">
            Total {total} jobs stored in PostgreSQL
          </p>
        </div>

        <button
          onClick={onRefresh}
          disabled={loading}
          className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-xs font-medium text-slate-300 border border-slate-700 transition flex items-center gap-1.5"
        >
          <svg className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          <span>Refresh</span>
        </button>
      </div>

      {error ? (
        <div className="p-8 text-center">
          <p className="text-rose-400 text-sm">{error}</p>
          <button
            onClick={onRefresh}
            className="mt-3 text-xs text-indigo-400 underline hover:text-indigo-300"
          >
            Try Again
          </button>
        </div>
      ) : loading && jobs.length === 0 ? (
        <div className="p-12 text-center text-slate-400 text-sm">
          <div className="inline-block animate-spin h-6 w-6 border-2 border-indigo-500 border-t-transparent rounded-full mb-2"></div>
          <p>Loading jobs...</p>
        </div>
      ) : jobs.length === 0 ? (
        <div className="p-12 text-center text-slate-500">
          <svg className="w-10 h-10 mx-auto text-slate-600 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
          </svg>
          <p className="text-sm">No jobs in queue yet</p>
          <p className="text-xs text-slate-600 mt-1">Submit your first job using the form above</p>
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-slate-950/60 text-slate-400 font-medium border-b border-slate-800">
              <tr>
                <th className="py-3 px-4">Job ID</th>
                <th className="py-3 px-4">Type</th>
                <th className="py-3 px-4">Status</th>
                <th className="py-3 px-4">Attempts</th>
                <th className="py-3 px-4">Created</th>
                <th className="py-3 px-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60 font-mono">
              {jobs.map((job) => (
                <tr key={job.id} className="hover:bg-slate-800/30 transition">
                  <td className="py-3.5 px-4 text-slate-300">
                    <button
                      onClick={() => onSelectJob(job)}
                      className="text-indigo-400 hover:underline hover:text-indigo-300 font-medium"
                    >
                      {job.id.substring(0, 8)}...
                    </button>
                  </td>
                  <td className="py-3.5 px-4 capitalize text-slate-200">{job.type}</td>
                  <td className="py-3.5 px-4">
                    <span
                      className={`inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium border capitalize ${
                        STATUS_BADGES[job.status] || 'bg-slate-800 text-slate-400'
                      }`}
                    >
                      {job.status}
                    </span>
                  </td>
                  <td className="py-3.5 px-4 text-slate-400">
                    {job.attempts} / {job.max_attempts}
                  </td>
                  <td className="py-3.5 px-4 text-slate-400">
                    {new Date(job.created_at).toLocaleTimeString()}
                  </td>
                  <td className="py-3.5 px-4 text-right space-x-2">
                    <button
                      onClick={() => onSelectJob(job)}
                      className="px-2 py-1 rounded bg-slate-800 hover:bg-slate-700 text-slate-300 text-[11px] transition font-sans"
                    >
                      Details
                    </button>
                    <button
                      onClick={() => onDeleteJob(job.id)}
                      className="px-2 py-1 rounded bg-rose-950/40 hover:bg-rose-900/60 text-rose-400 text-[11px] border border-rose-900/40 transition font-sans"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

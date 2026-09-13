import React from 'react';

export default function Navbar({ healthStatus, checkingHealth, onRefreshHealth }) {
  return (
    <header className="border-b border-slate-800 bg-slate-900/80 backdrop-blur sticky top-0 z-40">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <div className="h-9 w-9 rounded-lg bg-gradient-to-tr from-indigo-500 to-violet-500 flex items-center justify-center shadow-lg shadow-indigo-500/30">
            <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
          <div>
            <div className="flex items-center space-x-2">
              <span className="font-bold text-lg tracking-tight text-white">GoTask</span>
              <span className="text-xs px-2 py-0.5 rounded-full bg-indigo-950 text-indigo-400 border border-indigo-800 font-mono">v1.0</span>
            </div>
            <p className="text-xs text-slate-400 hidden sm:block">Concurrent Background Job Processor</p>
          </div>
        </div>

        <div className="flex items-center space-x-4">
          <button
            onClick={onRefreshHealth}
            disabled={checkingHealth}
            className="flex items-center space-x-2 px-3 py-1.5 rounded-full text-xs font-medium bg-slate-800 hover:bg-slate-700/80 border border-slate-700 transition"
            title="Click to re-check API & Database health"
          >
            <span
              className={`h-2 w-2 rounded-full ${
                healthStatus === 'ready'
                  ? 'bg-emerald-400 animate-pulse'
                  : healthStatus === 'checking'
                  ? 'bg-amber-400 animate-ping'
                  : 'bg-rose-500'
              }`}
            />
            <span className="text-slate-300">
              {checkingHealth ? 'Checking...' : healthStatus === 'ready' ? 'PostgreSQL Connected' : 'Backend Disconnected'}
            </span>
          </button>
        </div>
      </div>
    </header>
  );
}

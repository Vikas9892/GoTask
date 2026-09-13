import React, { useState } from 'react';
import { submitJob } from '../api/client';

const DEFAULT_TEMPLATES = {
  email: '{\n  "to": "dev@example.com",\n  "subject": "System Alert: Process Complete"\n}',
  webhook: '{\n  "url": "https://httpbin.org/post"\n}',
  report: '{\n  "format": "pdf",\n  "month": "September 2026"\n}',
};

export default function JobForm({ onJobCreated }) {
  const [type, setType] = useState('email');
  const [payloadText, setPayloadText] = useState(DEFAULT_TEMPLATES.email);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [feedback, setFeedback] = useState({ type: '', message: '' });

  const handleTypeChange = (newType) => {
    setType(newType);
    setPayloadText(DEFAULT_TEMPLATES[newType] || '{\n\n}');
    setFeedback({ type: '', message: '' });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setFeedback({ type: '', message: '' });

    // Validate JSON
    let parsedPayload;
    try {
      parsedPayload = JSON.parse(payloadText);
    } catch (err) {
      setFeedback({ type: 'error', message: 'Invalid JSON format: ' + err.message });
      return;
    }

    setIsSubmitting(true);
    try {
      const created = await submitJob(type, parsedPayload);
      setFeedback({
        type: 'success',
        message: `Job ${created.id.substring(0, 8)}... created successfully!`,
      });
      onJobCreated?.();
    } catch (err) {
      setFeedback({ type: 'error', message: err.message });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl p-6 shadow-sm">
      <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
        <svg className="w-5 h-5 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 9v3m0 0v3m0-3h3m-3 0H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        Submit New Background Job
      </h2>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-xs font-medium text-slate-300 uppercase tracking-wider mb-1.5">
            Job Type
          </label>
          <div className="grid grid-cols-3 gap-2">
            {['email', 'webhook', 'report'].map((t) => (
              <button
                key={t}
                type="button"
                onClick={() => handleTypeChange(t)}
                className={`py-2 px-3 text-xs font-medium rounded-lg border transition capitalize ${
                  type === t
                    ? 'bg-indigo-600 text-white border-indigo-500 shadow-md shadow-indigo-600/20'
                    : 'bg-slate-800 text-slate-400 border-slate-700 hover:bg-slate-700'
                }`}
              >
                {t}
              </button>
            ))}
          </div>
        </div>

        <div>
          <div className="flex items-center justify-between mb-1.5">
            <label className="block text-xs font-medium text-slate-300 uppercase tracking-wider">
              JSON Payload
            </label>
            <button
              type="button"
              onClick={() => setPayloadText(DEFAULT_TEMPLATES[type])}
              className="text-xs text-indigo-400 hover:text-indigo-300 transition"
            >
              Reset Template
            </button>
          </div>
          <textarea
            rows="5"
            value={payloadText}
            onChange={(e) => setPayloadText(e.target.value)}
            className="w-full bg-slate-950 text-slate-200 font-mono text-xs rounded-lg border border-slate-800 p-3 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition"
            placeholder="Enter JSON payload"
          />
        </div>

        {feedback.message && (
          <div
            className={`p-3 rounded-lg text-xs font-medium flex items-center gap-2 ${
              feedback.type === 'success'
                ? 'bg-emerald-950/60 border border-emerald-800 text-emerald-300'
                : 'bg-rose-950/60 border border-rose-800 text-rose-300'
            }`}
          >
            <span>{feedback.message}</span>
          </div>
        )}

        <button
          type="submit"
          disabled={isSubmitting}
          className="w-full py-2.5 px-4 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-semibold transition disabled:opacity-50 disabled:cursor-not-allowed shadow-md shadow-indigo-600/30 flex items-center justify-center gap-2"
        >
          {isSubmitting ? (
            <>
              <svg className="animate-spin h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
              </svg>
              <span>Enqueuing Job...</span>
            </>
          ) : (
            <span>Submit Job</span>
          )}
        </button>
      </form>
    </div>
  );
}

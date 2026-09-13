import React, { useCallback, useEffect, useState } from 'react';
import Navbar from './components/Navbar';
import JobForm from './components/JobForm';
import JobList from './components/JobList';
import JobDetailsModal from './components/JobDetailsModal';
import { fetchHealth, fetchJobs, deleteJob } from './api/client';

export default function App() {
  const [healthStatus, setHealthStatus] = useState('checking');
  const [checkingHealth, setCheckingHealth] = useState(false);
  const [jobs, setJobs] = useState([]);
  const [totalJobs, setTotalJobs] = useState(0);
  const [loadingJobs, setLoadingJobs] = useState(true);
  const [jobsError, setJobsError] = useState('');
  const [selectedJob, setSelectedJob] = useState(null);

  const checkHealth = useCallback(async () => {
    setCheckingHealth(true);
    try {
      const data = await fetchHealth();
      setHealthStatus(data.status === 'ready' ? 'ready' : 'degraded');
    } catch {
      setHealthStatus('disconnected');
    } finally {
      setCheckingHealth(false);
    }
  }, []);

  const loadJobs = useCallback(async () => {
    try {
      const data = await fetchJobs(50, 0);
      setJobs(data.jobs || []);
      setTotalJobs(data.total || 0);
      setJobsError('');
    } catch (err) {
      setJobsError(err.message);
    } finally {
      setLoadingJobs(false);
    }
  }, []);

  useEffect(() => {
    checkHealth();
    loadJobs();
  }, [checkHealth, loadJobs]);

  // Prompt 27: Status polling for active (pending / processing) jobs
  useEffect(() => {
    const hasActiveJobs = jobs.some(
      (j) => j.status === 'pending' || j.status === 'processing'
    );

    if (!hasActiveJobs) return;

    const intervalId = setInterval(() => {
      loadJobs();
    }, 2000);

    return () => clearInterval(intervalId);
  }, [jobs, loadJobs]);

  const handleDeleteJob = async (id) => {
    if (!confirm('Are you sure you want to delete this job?')) return;
    try {
      await deleteJob(id);
      loadJobs();
      if (selectedJob?.id === id) setSelectedJob(null);
    } catch (err) {
      alert(err.message);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col">
      <Navbar
        healthStatus={healthStatus}
        checkingHealth={checkingHealth}
        onRefreshHealth={checkHealth}
      />

      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 items-start">
          <div className="lg:col-span-1 sticky top-24">
            <JobForm onJobCreated={loadJobs} />
          </div>

          <div className="lg:col-span-2">
            <JobList
              jobs={jobs}
              total={totalJobs}
              loading={loadingJobs}
              error={jobsError}
              onRefresh={loadJobs}
              onSelectJob={setSelectedJob}
              onDeleteJob={handleDeleteJob}
            />
          </div>
        </div>
      </main>

      <JobDetailsModal
        job={selectedJob}
        onClose={() => setSelectedJob(null)}
      />
    </div>
  );
}

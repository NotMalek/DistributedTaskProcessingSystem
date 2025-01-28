'use client';

import React, { useState, useEffect } from 'react';
import { MetricsCards } from '@/components/dashboard/metrics-cards';
import { QueueChart } from '@/components/dashboard/queue-chart';
import { WorkerStatus } from '@/components/dashboard/worker-status';
import { TaskSubmissionForm } from '@/components/dashboard/task-submission';
import { SystemManagement } from '@/components/dashboard/system-management';

interface SystemMetrics {
  activeWorkers: number;
  totalTasks: number;
  processedTasks: number;
  failedTasks: number;
  queueLengths: Record<string, number>;
  workerMetrics: Record<string, any>;
}

export default function Home() {
  const [metrics, setMetrics] = useState<SystemMetrics>({
    activeWorkers: 0,
    totalTasks: 0,
    processedTasks: 0,
    failedTasks: 0,
    queueLengths: {},
    workerMetrics: {}
  });
  const [error, setError] = useState<string | null>(null);

  const fetchMetrics = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/metrics');
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const data = await response.json();
      setMetrics(data);
      setError(null);
    } catch (error) {
      console.error('Error fetching metrics:', error);
      setError('Failed to fetch metrics data');
    }
  };

  useEffect(() => {
    // Initial fetch
    fetchMetrics();

    // Set up polling every second
    const intervalId = setInterval(fetchMetrics, 1000);

    // Cleanup interval on component unmount
    return () => clearInterval(intervalId);
  }, []);

  return (
      <main className="p-6 max-w-7xl mx-auto space-y-6">
        <h1 className="text-3xl font-bold mb-8">Task Processing System Dashboard</h1>

        {error && (
            <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
              {error}
            </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 space-y-6">
            <MetricsCards
                activeWorkers={metrics.activeWorkers}
                totalTasks={metrics.totalTasks}
                processedTasks={metrics.processedTasks}
                failedTasks={metrics.failedTasks}
            />

            <QueueChart queueLengths={metrics.queueLengths} />

            <WorkerStatus workers={metrics.workerMetrics} />
          </div>

          <div className="space-y-6">
            <TaskSubmissionForm onSuccess={fetchMetrics} />
            <SystemManagement onSystemReset={fetchMetrics} />
          </div>
        </div>
      </main>
  );
}
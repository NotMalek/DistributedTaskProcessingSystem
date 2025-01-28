'use client';

import React, { useState, useEffect } from 'react';
import { MetricsCards } from '@/components/dashboard/metrics-cards';
import { QueueChart } from '@/components/dashboard/queue-chart';
import { WorkerStatus } from '@/components/dashboard/worker-status';

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

  useEffect(() => {
    const fetchMetrics = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/metrics');
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        console.log('Received metrics:', data); // Debug log
        setMetrics(data);
        setError(null);
      } catch (error) {
        console.error('Error fetching metrics:', error);
        setError('Failed to fetch metrics data');
      }
    };

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

        <MetricsCards
            activeWorkers={metrics.activeWorkers}
            totalTasks={metrics.totalTasks}
            processedTasks={metrics.processedTasks}
            failedTasks={metrics.failedTasks}
        />

        <QueueChart queueLengths={metrics.queueLengths} />

        <WorkerStatus workers={metrics.workerMetrics} />
      </main>
  );
}
import React, { useState, useEffect } from 'react';
import { Settings, MapPin, Clock, Activity } from 'lucide-react';
import { Machine } from '../types';
import { StatusIndicator } from './StatusIndicator';
import { MetricsCard } from './MetricsCard';
import { machinesApi, handleApiError } from '../services/api';

interface MachineStatusProps {
  selectedMachine?: string;
  onMachineSelect?: (machineId: string) => void;
  className?: string;
}

export const MachineStatus: React.FC<MachineStatusProps> = ({
  selectedMachine,
  onMachineSelect,
  className = '',
}) => {
  const [machines, setMachines] = useState<Machine[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchMachines();
  }, []);

  const fetchMachines = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await machinesApi.getMachines();
      setMachines(response.machines);
    } catch (err) {
      setError(handleApiError(err));
    } finally {
      setLoading(false);
    }
  };

  const handleMachineClick = (machineId: string) => {
    if (onMachineSelect) {
      onMachineSelect(machineId);
    }
  };

  const getRealTimeStats = (machine: Machine) => {
    return machine.config?.real_time_stats || null;
  };

  const formatLastEventTime = (timestamp: string) => {
    if (!timestamp) return 'No recent data';

    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / (1000 * 60));

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;

    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `${diffHours}h ago`;

    const diffDays = Math.floor(diffHours / 24);
    return `${diffDays}d ago`;
  };

  if (loading) {
    return (
      <div className={`card ${className}`}>
        <div className="card-header">
          <h3 className="text-lg font-semibold text-gray-900">Machine Status</h3>
        </div>
        <div className="flex items-center justify-center h-32">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
        </div>
      </div>
    );
  }

  return (
    <div className={`card ${className}`}>
      <div className="card-header flex items-center justify-between">
        <h3 className="text-lg font-semibold text-gray-900">Machine Status</h3>
        <button
          onClick={fetchMachines}
          className="text-primary-600 hover:text-primary-700 p-1"
          title="Refresh machine status"
        >
          <Activity size={16} />
        </button>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-danger-50 border border-danger-200 rounded-md">
          <span className="text-sm text-danger-700">{error}</span>
        </div>
      )}

      <div className="space-y-4">
        {machines.map((machine) => {
          const stats = getRealTimeStats(machine);
          const isSelected = selectedMachine === machine.machine_id;

          return (
            <div
              key={machine.machine_id}
              className={`border rounded-lg p-4 cursor-pointer transition-all ${
                isSelected
                  ? 'border-primary-500 bg-primary-50 shadow-md'
                  : 'border-gray-200 hover:border-gray-300 hover:shadow-sm'
              }`}
              onClick={() => handleMachineClick(machine.machine_id)}
            >
              <div className="flex items-start justify-between mb-3">
                <div>
                  <div className="flex items-center gap-2 mb-1">
                    <h4 className="font-medium text-gray-900">{machine.machine_id}</h4>
                    <StatusIndicator
                      status={machine.status === 'active' ? 'ok' : 'warning'}
                      size="sm"
                    />
                  </div>
                  <div className="flex items-center gap-4 text-sm text-gray-500">
                    <div className="flex items-center gap-1">
                      <Settings size={12} />
                      <span className="capitalize">{machine.machine_type}</span>
                    </div>
                    <div className="flex items-center gap-1">
                      <MapPin size={12} />
                      <span>{machine.location}</span>
                    </div>
                  </div>
                </div>
                {stats && (
                  <div className="text-right">
                    <div className="flex items-center gap-1 text-sm text-gray-500">
                      <Clock size={12} />
                      <span>{formatLastEventTime(stats.last_event_time)}</span>
                    </div>
                  </div>
                )}
              </div>

              {stats && (
                <div className="grid grid-cols-2 gap-3 mt-3">
                  <div className="bg-white rounded-md p-3 border border-gray-100">
                    <div className="text-xs text-gray-500 mb-1">Temperature</div>
                    <div className="text-lg font-semibold text-gray-900">
                      {stats.avg_temperature?.toFixed(1) || 'N/A'}°C
                    </div>
                  </div>
                  <div className="bg-white rounded-md p-3 border border-gray-100">
                    <div className="text-xs text-gray-500 mb-1">Speed</div>
                    <div className="text-lg font-semibold text-gray-900">
                      {stats.avg_conveyor_speed?.toFixed(2) || 'N/A'} m/s
                    </div>
                  </div>
                  <div className="bg-white rounded-md p-3 border border-gray-100">
                    <div className="text-xs text-gray-500 mb-1">Events</div>
                    <div className="text-lg font-semibold text-gray-900">
                      {stats.event_count || 0}
                    </div>
                  </div>
                  <div className="bg-white rounded-md p-3 border border-gray-100">
                    <div className="text-xs text-gray-500 mb-1">Fault Rate</div>
                    <div className={`text-lg font-semibold ${
                      (stats.fault_rate || 0) > 0.1 ? 'text-danger-600' : 'text-success-600'
                    }`}>
                      {((stats.fault_rate || 0) * 100).toFixed(1)}%
                    </div>
                  </div>
                </div>
              )}

              {!stats && (
                <div className="mt-3 text-center text-sm text-gray-400 py-2">
                  No real-time data available
                </div>
              )}
            </div>
          );
        })}

        {machines.length === 0 && !loading && (
          <div className="text-center text-gray-500 py-8">
            <Settings size={48} className="mx-auto mb-2 text-gray-300" />
            <p className="text-sm">No machines found</p>
          </div>
        )}
      </div>
    </div>
  );
};
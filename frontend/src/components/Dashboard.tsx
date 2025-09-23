import React, { useState, useEffect, useMemo } from 'react';
import { Activity, Wifi, WifiOff } from 'lucide-react';
import { SensorEvent, Alert, EventStats, ChartDataPoint } from '../types';
import { useWebSocket } from '../hooks/useWebSocket';
import { eventsApi, systemApi, handleApiError } from '../services/api';
import { RealtimeChart } from './RealtimeChart';
import { MetricsCard } from './MetricsCard';
import { AlertsPanel } from './AlertsPanel';
import { MachineStatus } from './MachineStatus';
import { StatusIndicator } from './StatusIndicator';

const WS_URL = process.env.REACT_APP_WS_URL || 'ws://localhost:8080/ws';

export const Dashboard: React.FC = () => {
  const [events, setEvents] = useState<SensorEvent[]>([]);
  const [newAlerts, setNewAlerts] = useState<Alert[]>([]);
  const [stats, setStats] = useState<EventStats | null>(null);
  const [selectedMachine, setSelectedMachine] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // WebSocket connection
  const { isConnected, connectionStatus, lastMessage } = useWebSocket({
    url: WS_URL,
    onMessage: (message) => {
      switch (message.type) {
        case 'sensor_event':
          handleNewEvent(message.data);
          break;
        case 'alert':
          handleNewAlert(message.data);
          break;
        case 'stats':
          if (message.data.system_stats) {
            setStats(message.data.system_stats);
          }
          break;
        default:
          console.log('Unhandled WebSocket message type:', message.type);
      }
    },
    onConnect: () => {
      console.log('Connected to WebSocket');
      // Subscribe to relevant topics
      // subscribe(['sensor_events', 'alerts', 'stats']);
    },
    onDisconnect: () => {
      console.log('Disconnected from WebSocket');
    },
    reconnectInterval: 3000,
    maxReconnectAttempts: 5,
  });

  useEffect(() => {
    fetchInitialData();
  }, [selectedMachine]);

  const fetchInitialData = async () => {
    setLoading(true);
    setError(null);

    try {
      // Fetch recent events
      const eventsResponse = await eventsApi.getEvents({
        limit: 100,
        machine_id: selectedMachine || undefined,
      });
      setEvents(eventsResponse.events);

      // Fetch system stats
      const statsResponse = await eventsApi.getEventStats({
        machine_id: selectedMachine || undefined,
        since: '1h',
      });
      setStats(statsResponse.stats);
    } catch (err) {
      setError(handleApiError(err));
    } finally {
      setLoading(false);
    }
  };

  const handleNewEvent = (event: SensorEvent) => {
    // Add new event to the beginning of the list
    setEvents(prev => {
      const newEvents = [event, ...prev];
      // Keep only last 100 events for performance
      return newEvents.slice(0, 100);
    });
  };

  const handleNewAlert = (alert: Alert) => {
    // Add new alert to the beginning of the list
    setNewAlerts(prev => [alert, ...prev.slice(0, 4)]); // Keep only 5 most recent
  };

  const chartData = useMemo((): ChartDataPoint[] => {
    return events
      .filter(event => !selectedMachine || event.machine_id === selectedMachine)
      .map(event => ({
        timestamp: event.timestamp,
        conveyor_speed: event.conveyor_speed,
        temperature: event.temperature,
        robot_arm_angle: event.robot_arm_angle,
        status: event.status,
      }))
      .slice(0, 50) // Last 50 points for performance
      .reverse(); // Show chronological order
  }, [events, selectedMachine]);

  const latestEvent = useMemo(() => {
    const filteredEvents = selectedMachine
      ? events.filter(e => e.machine_id === selectedMachine)
      : events;
    return filteredEvents[0] || null;
  }, [events, selectedMachine]);

  const getConnectionStatusColor = () => {
    switch (connectionStatus) {
      case 'connected':
        return 'text-success-600';
      case 'connecting':
        return 'text-warning-600';
      default:
        return 'text-danger-600';
    }
  };

  const getConnectionStatusText = () => {
    switch (connectionStatus) {
      case 'connected':
        return 'Connected';
      case 'connecting':
        return 'Connecting...';
      case 'disconnected':
        return 'Disconnected';
      default:
        return 'Error';
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-3">
              <Activity className="text-primary-600" size={24} />
              <h1 className="text-xl font-bold text-gray-900">FactoryFlow Dashboard</h1>
            </div>

            <div className="flex items-center gap-4">
              {/* Machine Filter */}
              <select
                value={selectedMachine}
                onChange={(e) => setSelectedMachine(e.target.value)}
                className="text-sm border border-gray-300 rounded-md px-3 py-1.5 focus:ring-primary-500 focus:border-primary-500"
              >
                <option value="">All Machines</option>
                <option value="sensor_hub_001">Sensor Hub 001</option>
                <option value="conveyor_001">Conveyor 001</option>
                <option value="robot_arm_001">Robot Arm 001</option>
              </select>

              {/* Connection Status */}
              <div className="flex items-center gap-2">
                {isConnected ? <Wifi size={16} /> : <WifiOff size={16} />}
                <span className={`text-sm font-medium ${getConnectionStatusColor()}`}>
                  {getConnectionStatusText()}
                </span>
              </div>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {error && (
          <div className="mb-6 p-4 bg-danger-50 border border-danger-200 rounded-lg">
            <div className="flex items-center gap-2">
              <Activity size={16} className="text-danger-600" />
              <span className="text-sm text-danger-700">{error}</span>
            </div>
          </div>
        )}

        {/* Metrics Row */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <MetricsCard
            title="Current Temperature"
            value={latestEvent?.temperature || 0}
            unit="°C"
            status={
              latestEvent?.temperature && latestEvent.temperature > 80
                ? 'critical'
                : latestEvent?.temperature && latestEvent.temperature > 75
                ? 'warning'
                : 'normal'
            }
            icon={<Activity size={20} />}
          />

          <MetricsCard
            title="Conveyor Speed"
            value={latestEvent?.conveyor_speed || 0}
            unit="m/s"
            status={
              latestEvent?.conveyor_speed && latestEvent.conveyor_speed < 0.5
                ? 'warning'
                : 'normal'
            }
            icon={<Activity size={20} />}
          />

          <MetricsCard
            title="Robot Arm Angle"
            value={latestEvent?.robot_arm_angle || 0}
            unit="°"
            icon={<Activity size={20} />}
          />

          <MetricsCard
            title="System Status"
            value={latestEvent?.status ? (
              <StatusIndicator
                status={latestEvent.status}
                size="sm"
                showIcon={false}
              />
            ) : 'Unknown'}
            unit=""
            icon={<Activity size={20} />}
          />
        </div>

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Left Column - Charts */}
          <div className="lg:col-span-2 space-y-8">
            <RealtimeChart
              data={chartData}
              title={`Real-time Sensor Data ${selectedMachine ? `- ${selectedMachine}` : ''}`}
              height={400}
            />

            {/* System Statistics */}
            {stats && (
              <div className="card">
                <div className="card-header">
                  <h3 className="text-lg font-semibold text-gray-900">System Statistics (Last Hour)</h3>
                </div>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  <div className="text-center">
                    <div className="text-2xl font-bold text-gray-900">{stats.total_events}</div>
                    <div className="text-sm text-gray-500">Total Events</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold text-danger-600">{stats.fault_events}</div>
                    <div className="text-sm text-gray-500">Fault Events</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold text-warning-600">{stats.warning_events}</div>
                    <div className="text-sm text-gray-500">Warning Events</div>
                  </div>
                  <div className="text-center">
                    <div className={`text-2xl font-bold ${
                      stats.uptime_percent >= 95 ? 'text-success-600' : 'text-warning-600'
                    }`}>
                      {stats.uptime_percent.toFixed(1)}%
                    </div>
                    <div className="text-sm text-gray-500">Uptime</div>
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Right Column - Alerts & Machine Status */}
          <div className="space-y-8">
            <AlertsPanel newAlerts={newAlerts} />
            <MachineStatus
              selectedMachine={selectedMachine}
              onMachineSelect={setSelectedMachine}
            />
          </div>
        </div>
      </main>
    </div>
  );
};
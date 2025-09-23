import React, { useState, useEffect } from "react";
import { AlertCircle, AlertTriangle, CheckCircle, X } from "lucide-react";
import { Alert } from "../types";
import { alertsApi, handleApiError } from "../services/api";

interface AlertsPanelProps {
  newAlerts?: Alert[];
  className?: string;
}

export const AlertsPanel: React.FC<AlertsPanelProps> = ({
  newAlerts = [],
  className = "",
}) => {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchAlerts();
  }, []);

  useEffect(() => {
    // Add new alerts from WebSocket
    if (newAlerts.length > 0) {
      setAlerts((prev) => [...newAlerts, ...prev]);
    }
  }, [newAlerts]);

  const fetchAlerts = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await alertsApi.getAlerts();
      setAlerts(response.alerts);
    } catch (err) {
      setError(handleApiError(err));
    } finally {
      setLoading(false);
    }
  };

  const acknowledgeAlert = async (alertId: number) => {
    try {
      await alertsApi.acknowledgeAlert(alertId);
      setAlerts((prev) => prev.filter((alert) => alert.id !== alertId));
    } catch (err) {
      setError(handleApiError(err));
    }
  };

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case "high":
        return <AlertCircle size={16} className="text-danger-600" />;
      case "medium":
        return <AlertTriangle size={16} className="text-warning-600" />;
      default:
        return <AlertCircle size={16} className="text-primary-600" />;
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case "high":
        return "border-l-danger-500 bg-danger-50";
      case "medium":
        return "border-l-warning-500 bg-warning-50";
      default:
        return "border-l-primary-500 bg-primary-50";
    }
  };

  const formatTime = (timestamp: string) => {
    return new Date(timestamp).toLocaleString([], {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  if (loading) {
    return (
      <div className={`card ${className}`}>
        <div className="card-header">
          <h3 className="text-lg font-semibold text-gray-900">Active Alerts</h3>
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
        <h3 className="text-lg font-semibold text-gray-900">Active Alerts</h3>
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-500">
            {alerts.length} alert{alerts.length !== 1 ? "s" : ""}
          </span>
          <button
            onClick={fetchAlerts}
            className="text-primary-600 hover:text-primary-700 p-1"
            title="Refresh alerts"
          >
            <svg
              className="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
          </button>
        </div>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-danger-50 border border-danger-200 rounded-md">
          <div className="flex items-center gap-2">
            <AlertCircle size={16} className="text-danger-600 flex-shrink-0" />
            <span className="text-sm text-danger-700">{error}</span>
          </div>
        </div>
      )}

      <div className="space-y-3 max-h-96 overflow-y-auto">
        {alerts.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-8 text-gray-500">
            <CheckCircle size={48} className="text-success-400 mb-2" />
            <p className="text-sm">No active alerts</p>
            <p className="text-xs text-gray-400">
              All systems running normally
            </p>
          </div>
        ) : (
          alerts.map((alert) => (
            <div
              key={alert.id}
              className={`p-3 border-l-4 rounded-r-lg ${getSeverityColor(
                alert.severity
              )}`}
            >
              <div className="flex items-start justify-between">
                <div className="flex items-start gap-2 flex-1">
                  {getSeverityIcon(alert.severity)}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-sm font-medium text-gray-900 capitalize">
                        {alert.alert_type.replace("_", " ")}
                      </span>
                      <span
                        className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                          alert.severity === "high"
                            ? "bg-danger-100 text-danger-800"
                            : alert.severity === "medium"
                            ? "bg-warning-100 text-warning-800"
                            : "bg-primary-100 text-primary-800"
                        }`}
                      >
                        {alert.severity}
                      </span>
                    </div>
                    <p className="text-sm text-gray-700 mb-2">
                      {alert.message}
                    </p>
                    <p className="text-xs text-gray-500">
                      {formatTime(alert.created_at)}
                    </p>
                  </div>
                </div>
                <button
                  onClick={() => acknowledgeAlert(alert.id)}
                  className="text-gray-400 hover:text-gray-600 p-1 flex-shrink-0"
                  title="Acknowledge alert"
                >
                  <X size={16} />
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

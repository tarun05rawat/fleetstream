import React from 'react';
import { AlertCircle, CheckCircle, AlertTriangle } from 'lucide-react';

interface StatusIndicatorProps {
  status: 'ok' | 'warning' | 'fault';
  className?: string;
  size?: 'sm' | 'md' | 'lg';
  showIcon?: boolean;
}

export const StatusIndicator: React.FC<StatusIndicatorProps> = ({
  status,
  className = '',
  size = 'md',
  showIcon = true,
}) => {
  const getStatusConfig = () => {
    switch (status) {
      case 'ok':
        return {
          className: 'status-ok',
          icon: CheckCircle,
          text: 'OK',
        };
      case 'warning':
        return {
          className: 'status-warning',
          icon: AlertTriangle,
          text: 'Warning',
        };
      case 'fault':
        return {
          className: 'status-fault',
          icon: AlertCircle,
          text: 'Fault',
        };
      default:
        return {
          className: 'status-ok',
          icon: CheckCircle,
          text: 'Unknown',
        };
    }
  };

  const getSizeClasses = () => {
    switch (size) {
      case 'sm':
        return 'text-xs px-2 py-0.5';
      case 'lg':
        return 'text-sm px-3 py-1';
      default:
        return 'text-xs px-2.5 py-0.5';
    }
  };

  const getIconSize = () => {
    switch (size) {
      case 'sm':
        return 12;
      case 'lg':
        return 18;
      default:
        return 14;
    }
  };

  const config = getStatusConfig();
  const Icon = config.icon;

  return (
    <span className={`status-indicator ${config.className} ${getSizeClasses()} ${className}`}>
      {showIcon && <Icon size={getIconSize()} className="mr-1" />}
      {config.text}
    </span>
  );
};